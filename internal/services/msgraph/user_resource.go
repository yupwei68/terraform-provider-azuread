package msgraph

import (
	"fmt"
	"github.com/manicminer/hamilton/models"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func UserResource() *schema.Resource {
	return &schema.Resource{
		Create: userResourceCreate,
		Read:   userResourceRead,
		Update: userResourceUpdate,
		Delete: userResourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_principal_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StringIsEmailAddress,
			},

			"display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"mail_nickname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"account_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 256), // currently the max length for AAD passwords is 256
			},

			"force_password_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"mail": {
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

			"onpremises_immutable_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "This must be specified if you are using a federated domain for the user's userPrincipalName (UPN) property when creating a new user account. " +
					"It is used to associate an on-premises Active Directory user account with their Azure AD user object.",
			},

			"object_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"usage_location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "A two letter country code (ISO standard 3166). " +
					"Required for users that will be assigned licenses due to legal requirement to check for availability of services in countries. " +
					"Examples include: `NO`, `JP`, and `GB`. Not nullable.",
			},
		},
	}
}

func userResourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	upn := d.Get("user_principal_name").(string)
	mailNickName := d.Get("mail_nickname").(string)

	// default mail nickname to the first part of the UPN (matches the portal)
	if mailNickName == "" {
		mailNickName = strings.Split(upn, "@")[0]
	}

	properties := models.User{
		AccountEnabled: utils.BoolI(d.Get("account_enabled")),
		DisplayName:    utils.StringI(d.Get("display_name")),
		MailNickname:   &mailNickName,
		PasswordProfile: &models.UserPasswordProfile{
			ForceChangePasswordNextSignIn: utils.BoolI(d.Get("force_password_change")),
			Password:                      utils.StringI(d.Get("password")),
		},
		UserPrincipalName: &upn,
	}

	if v, ok := d.GetOk("usage_location"); ok {
		properties.UsageLocation = utils.StringI(v)
	}

	if v, ok := d.GetOk("onpremises_immutable_id"); ok {
		properties.OnPremisesImmutableId = utils.StringI(v)
	}

	user, err := client.Create(ctx, properties)
	if err != nil {
		return fmt.Errorf("creating User %q: %+v", upn, err)
	}
	if user.ID == nil {
		return fmt.Errorf("null ID returned for User %q", upn)
	}

	d.SetId(*user.ID)

	return userResourceRead(d, meta)
}

func userResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	user := models.User{ID: utils.String(d.Id())}

	if d.HasChange("display_name") {
		user.DisplayName = utils.StringI(d.Get("display_name"))
	}

	if d.HasChange("mail_nickname") {
		user.MailNickname = utils.StringI(d.Get("mail_nickname"))
	}

	if d.HasChange("account_enabled") {
		user.AccountEnabled = utils.BoolI(d.Get("account_enabled"))
	}

	if d.HasChange("password") || d.HasChange("force_password_change") {
		user.PasswordProfile = &models.UserPasswordProfile{
			ForceChangePasswordNextSignIn: utils.BoolI(d.Get("force_password_change")),
			Password:                      utils.StringI(d.Get("password")),
		}
	}

	if d.HasChange("usage_location") {
		user.UsageLocation = utils.StringI(d.Get("usage_location"))
	}

	if d.HasChange("onpremises_immutable_id") {
		user.OnPremisesImmutableId = utils.StringI(d.Get("onpremises_immutable_id"))
	}

	if err := client.Update(ctx, user); err != nil {
		return fmt.Errorf("updating User with ID %q: %+v", d.Id(), err)
	}

	return userResourceRead(d, meta)
}

func userResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	objectId := d.Id()

	user, err := client.Get(ctx, objectId)
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("user_principal_name", user.UserPrincipalName)
	d.Set("display_name", user.DisplayName)
	d.Set("mail", user.Mail)
	d.Set("mail_nickname", user.MailNickname)
	d.Set("account_enabled", user.AccountEnabled)
	d.Set("object_id", user.ID)
	d.Set("usage_location", user.UsageLocation)
	d.Set("onpremises_immutable_id", user.OnPremisesImmutableId)
	d.Set("onpremises_sam_account_name", user.OnPremisesSamAccountName)
	d.Set("onpremises_user_principal_name", user.OnPremisesUserPrincipalName)

	forcePasswordChange := false
	if user.PasswordProfile != nil {
		forcePasswordChange = *user.PasswordProfile.ForceChangePasswordNextSignIn
	}
	d.Set("force_password_change", forcePasswordChange)

	return nil
}

func userResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.UsersClient
	ctx := meta.(*clients.AadClient).StopContext

	err := client.Delete(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("deleting User with ID %q: %+v", d.Id(), err)
	}

	return nil
}
