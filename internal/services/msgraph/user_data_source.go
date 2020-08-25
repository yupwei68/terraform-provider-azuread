package msgraph

import (
	"fmt"
	"github.com/manicminer/hamilton/models"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func UserData() *schema.Resource {
	return &schema.Resource{
		Read: userDataRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"mail_nickname": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.NoEmptyStrings,
				ExactlyOneOf: []string{"mail_nickname", "object_id", "user_principal_name"},
			},

			"object_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.UUID,
				ExactlyOneOf: []string{"mail_nickname", "object_id", "user_principal_name"},
			},

			"user_principal_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.NoEmptyStrings,
				ExactlyOneOf: []string{"mail_nickname", "object_id", "user_principal_name"},
			},

			"account_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mail": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"onpremises_immutable_id": {
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
		},
	}
}

func userDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	var user models.User

	if objectId, ok := d.Get("object_id").(string); ok && objectId != "" {
		u, err := client.Get(ctx, objectId)
		if err != nil {
			return fmt.Errorf("reading User with ID %q: %+v", objectId, err)
		}
		user = *u
	} else if mailNickname, ok := d.Get("mail_nickname").(string); ok && mailNickname != "" {
		filter := fmt.Sprintf("mailNickname eq '%s'", mailNickname)
		users, err := client.List(ctx, filter)
		if err != nil {
			return fmt.Errorf("identifying User with mail nickname %q: %+v", mailNickname, err)
		}
		if len(*users) == 0 {
			return fmt.Errorf("no user found with mail nickname: %q", mailNickname)
		}
		user = (*users)[0]
	} else if upn, ok := d.Get("user_principal_name").(string); ok && upn != "" {
		filter := fmt.Sprintf("userPrincipalName eq '%s'", upn)
		users, err := client.List(ctx, filter)
		if err != nil {
			return fmt.Errorf("identifying User with user principal name %q: %+v", upn, err)
		}
		if len(*users) == 0 {
			return fmt.Errorf("no user found with user principal name: %q", upn)
		}
		user = (*users)[0]
	}

	if user.ID == nil {
		return fmt.Errorf("User ID is null")
	}
	d.SetId(*user.ID)

	d.Set("account_enabled", user.AccountEnabled)
	d.Set("display_name", user.DisplayName)
	d.Set("immutable_id", user.OnPremisesImmutableId)
	d.Set("mail", user.Mail)
	d.Set("mail_nickname", user.MailNickname)
	d.Set("object_id", user.ID)
	d.Set("onpremises_immutable_id", user.OnPremisesImmutableId)
	d.Set("onpremises_sam_account_name", user.OnPremisesSamAccountName)
	d.Set("onpremises_user_principal_name", user.OnPremisesUserPrincipalName)
	d.Set("usage_location", user.UsageLocation)
	d.Set("user_principal_name", user.UserPrincipalName)

	return nil
}
