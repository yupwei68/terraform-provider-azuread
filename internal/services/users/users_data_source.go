package users

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func usersData() *schema.Resource {
	return &schema.Resource{
		ReadContext: usersDataRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"object_ids": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"object_ids", "user_principal_names", "mail_nicknames"},
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validate.UUID,
				},
			},

			"user_principal_names": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"object_ids", "user_principal_names", "mail_nicknames"},
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validate.NoEmptyStrings,
				},
			},

			"mail_nicknames": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"object_ids", "user_principal_names", "mail_nicknames"},
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validate.NoEmptyStrings,
				},
			},

			"ignore_missing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"immutable_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"mail": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"mail_nickname": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"object_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"onpremises_sam_account_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"onpremises_user_principal_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"usage_location": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"user_principal_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func usersDataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Users.UsersClient

	var users []*graphrbac.User
	expectedCount := 0

	ignoreMissing := d.Get("ignore_missing").(bool)
	if upns, ok := d.Get("user_principal_names").([]interface{}); ok && len(upns) > 0 {
		expectedCount = len(upns)
		for _, v := range upns {
			u, err := client.Get(ctx, v.(string))
			if err != nil {
				if ignoreMissing && utils.ResponseWasNotFound(u.Response) {
					continue
				}

				return tf.ErrorDiag(fmt.Sprintf("Retrieving user with UPN: %q", v), err.Error(), "")
			}
			users = append(users, &u)
		}
	} else {
		if oids, ok := d.Get("object_ids").([]interface{}); ok && len(oids) > 0 {
			expectedCount = len(oids)
			for _, v := range oids {
				u, err := aadgraph.UserGetByObjectId(ctx, client, v.(string))
				if err != nil {
					return tf.ErrorDiag(fmt.Sprintf("Finding user with object ID: %q", v), err.Error(), "object_ids")
				}
				if u == nil {
					if ignoreMissing {
						continue
					} else {
						return tf.ErrorDiag(fmt.Sprintf("User not found with object ID: %q", v), "", "object_ids")
					}
				}
				users = append(users, u)
			}
		} else if mailNicknames, ok := d.Get("mail_nicknames").([]interface{}); ok && len(mailNicknames) > 0 {
			expectedCount = len(mailNicknames)
			for _, v := range mailNicknames {
				u, err := aadgraph.UserGetByMailNickname(ctx, client, v.(string))
				if err != nil {
					return tf.ErrorDiag(fmt.Sprintf("Finding user with email alias: %q", v), err.Error(), "mail_nicknames")
				}
				if u == nil {
					if ignoreMissing {
						continue
					} else {
						return tf.ErrorDiag(fmt.Sprintf("User not found with email alias: %q", v), "", "mail_nicknames")
					}
				}
				users = append(users, u)
			}
		}
	}

	if !ignoreMissing && len(users) != expectedCount {
		return tf.ErrorDiag("Unexpected number of users returned", fmt.Sprintf("Expected: %d, Actual: %d", expectedCount, len(users)), "")
	}

	upns := make([]string, 0, len(users))
	oids := make([]string, 0, len(users))
	mailNicknames := make([]string, 0, len(users))
	userList := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		if u.ObjectID == nil || u.UserPrincipalName == nil {
			return tf.ErrorDiag("Bad API response", "API returned user with nil object ID", "")
		}

		oids = append(oids, *u.ObjectID)
		upns = append(upns, *u.UserPrincipalName)
		if u.MailNickname != nil {
			mailNicknames = append(mailNicknames, *u.MailNickname)
		}

		user := make(map[string]interface{})
		user["account_enabled"] = u.AccountEnabled
		user["display_name"] = u.DisplayName
		user["immutable_id"] = u.ImmutableID
		user["mail"] = u.Mail
		user["mail_nickname"] = u.MailNickname
		user["object_id"] = u.ObjectID
		user["onpremises_sam_account_name"] = u.AdditionalProperties["onPremisesSamAccountName"]
		user["onpremises_user_principal_name"] = u.AdditionalProperties["onPremisesUserPrincipalName"]
		user["usage_location"] = u.UsageLocation
		user["user_principal_name"] = u.UserPrincipalName
		userList = append(userList, user)
	}

	h := sha1.New()
	if _, err := h.Write([]byte(strings.Join(upns, "-"))); err != nil {
		return tf.ErrorDiag("Unable to compute hash for UPNs", err.Error(), "")
	}

	d.SetId("users#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))

	if err := d.Set("object_ids", oids); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "object_ids")
	}

	if err := d.Set("user_principal_names", upns); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "user_principal_name")
	}

	if err := d.Set("mail_nicknames", mailNicknames); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "mail_nicknames")
	}

	if err := d.Set("users", userList); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "users")
	}

	return nil
}
