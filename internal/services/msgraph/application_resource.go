package msgraph

import (
	"context"
	"fmt"
	clients2 "github.com/manicminer/hamilton/clients"
	"github.com/manicminer/hamilton/models"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

const resourceApplicationName = "azuread_application"

func ApplicationResource() *schema.Resource {
	return &schema.Resource{
		Create: applicationResourceCreate,
		Read:   applicationResourceRead,
		Update: applicationResourceUpdate,
		Delete: applicationResourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"object_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"prevent_duplicate_names": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"api": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				//ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accept_mapped_claims": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"known_client_applications": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.UUID,
							},
						},

						"oauth2_permission_scope": {
							Type:       schema.TypeSet,
							Optional:   true,
							//Computed:   true,
							//ConfigMode: schema.SchemaConfigModeAttr,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"admin_consent_description": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validate.NoEmptyStrings,
									},

									"admin_consent_display_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validate.NoEmptyStrings,
									},

									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"is_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},

									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"Admin", "User"}, false),
									},

									"user_consent_description": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"user_consent_display_name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validate.NoEmptyStrings,
									},
								},
							},
						},

						"requested_access_token_version": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							ValidateFunc: validate.ValidInt32,
						},
					},
				},
			},

			"app_role": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"allowed_member_types": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice(
									[]string{"User", "Application"},
									false,
								),
							},
						},

						"description": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},

						"display_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validate.NoEmptyStrings,
						},

						"is_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"group_membership_claims": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"All",
					"ApplicationGroup", // not documented
					"DirectoryRole",    // not documented
					"None",
					"SecurityGroup",
				}, false),
			},

			"identifier_uris": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"optional_claims": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_token": SchemaOptionalClaims(),
						"id_token":     SchemaOptionalClaims(),
						"saml2_token":  SchemaOptionalClaims(),
					},
				},
			},

			"owners": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"public_client": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"redirect_uris": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.URLIsHTTPOrHTTPS,
							},
						},
					},
				},
			},

			"required_resource_access": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_app_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"resource_access": {
							Type:       schema.TypeList,
							Required:   true,
							ConfigMode: schema.SchemaConfigModeAttr,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validate.UUID,
									},

									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice(
											[]string{"Scope", "Role"},
											false,
										),
									},
								},
							},
						},
					},
				},
			},

			"sign_in_audience": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AzureADMyOrg",
					"AzureADMultipleOrgs",
					"AzureADandPersonalMicrosoftAccount",
				}, false),
			},

			"web": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				MaxItems:   1,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_access_token_issuance": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"enable_id_token_issuance": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"homepage_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validate.URLIsHTTPOrHTTPSorEmpty,
						},

						"logout_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validate.URLIsHTTPOrHTTPSorEmpty,
						},

						"redirect_uris": {
							Type:       schema.TypeSet,
							Optional:   true,
							ConfigMode: schema.SchemaConfigModeAttr,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.URLIsHTTPOrHTTPSorEmpty,
							},
						},
					},
				},
			},
		},
	}
}

func applicationResourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	displayName := d.Get("display_name").(string)

	if d.Get("prevent_duplicate_names").(bool) {
		if err := applicationCheckExistingDisplayName(ctx, client, displayName, nil); err != nil {
			return err
		}
	}

	if err := applicationValidateRolesScopes(d.Get("app_role"), d.Get("api.0.oauth2_permission_scope")); err != nil {
		return err
	}

	properties := models.Application{
		DisplayName: utils.String(displayName),
	}

	if v, ok := d.GetOk("api"); ok {
		properties.Api = expandApplicationApi(v.([]interface{}))
	}

	if v, ok := d.GetOk("app_role"); ok {
		properties.AppRoles = expandApplicationAppRoles(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("group_membership_claims"); ok {
		properties.GroupMembershipClaims = utils.String(v.(string))
	}

	if v, ok := d.GetOk("identifier_uris"); ok {
		properties.IdentifierUris = tf.ExpandStringSlicePtr(v.([]interface{}))
	}

	if v, ok := d.GetOk("optional_claims"); ok {
		properties.OptionalClaims = expandApplicationOptionalClaims(v.([]interface{}))
	}

	if v, ok := d.GetOk("public_client"); ok {
		properties.PublicClient = expandApplicationPublicClient(v.([]interface{}))
	}

	if v, ok := d.GetOk("required_resource_access"); ok {
		properties.RequiredResourceAccess = expandApplicationRequiredResourceAccess(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("sign_in_audience"); ok {
		properties.SignInAudience = utils.String(v.(string))
	}

	if v, ok := d.GetOk("web"); ok {
		properties.Web = expandApplicationWeb(v.([]interface{}))
	}

	app, _, err := client.Create(ctx, properties)
	if err != nil {
		return fmt.Errorf("creating Application: %+v", err)
	}

	if app.ID == nil {
		return fmt.Errorf("after creating, Application object ID was null")
	}

	d.SetId(*app.ID)

	if v, ok := d.GetOk("owners"); ok {
		owners := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())
		if err := applicationSetOwners(ctx, client, app, owners); err != nil {
			return err
		}
	}

	return applicationResourceRead(d, meta)
}

func applicationResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	app := models.Application{ID: utils.String(d.Id())}
	displayName := d.Get("display_name").(string)

	if d.HasChange("display_name") {
		if preventDuplicates := d.Get("prevent_duplicate_names").(bool); preventDuplicates {
			if err := applicationCheckExistingDisplayName(ctx, client, displayName, app.ID); err != nil {
				return err
			}
		}
		app.DisplayName = utils.String(displayName)
	}

	if err := applicationValidateRolesScopes(d.Get("app_role"), d.Get("api.0.oauth2_permission_scope")); err != nil {
		return err
	}

	if v, ok := d.GetOkExists("api"); ok && d.HasChange("api") {
		app.Api = expandApplicationApi(v.([]interface{}))
	}

	if v, ok := d.GetOkExists("app_role"); ok && d.HasChange("app_role") {
		app.AppRoles = expandApplicationAppRoles(v.(*schema.Set).List())
	}

	if v, ok := d.GetOkExists("group_membership_claims"); ok && d.HasChange("group_membership_claims") {
		app.GroupMembershipClaims = utils.String(v.(string))
	}

	if v, ok := d.GetOkExists("identifier_uris"); ok && d.HasChange("identifier_uris") {
		app.IdentifierUris = tf.ExpandStringSlicePtr(v.([]interface{}))
	}

	if v, ok := d.GetOkExists("optional_claims"); ok && d.HasChange("optional_claims") {
		app.OptionalClaims = expandApplicationOptionalClaims(v.([]interface{}))
	}

	if v, ok := d.GetOkExists("public_client"); ok && d.HasChange("public_client") {
		app.PublicClient = expandApplicationPublicClient(v.([]interface{}))
	}

	if v, ok := d.GetOkExists("required_resource_access"); ok && d.HasChange("required_resource_access") {
		app.RequiredResourceAccess = expandApplicationRequiredResourceAccess(v.(*schema.Set).List())
	}

	if v, ok := d.GetOkExists("sign_in_audience"); ok && d.HasChange("sign_in_audience") {
		app.SignInAudience = utils.String(v.(string))
	}

	if v, ok := d.GetOkExists("web"); ok && d.HasChange("web") {
		app.Web = expandApplicationWeb(v.([]interface{}))
	}

	if _, err := client.Update(ctx, app); err != nil {
		return fmt.Errorf("updating Application with ID %q: %+v", d.Id(), err)
	}

	if v, ok := d.GetOkExists("owners"); ok && d.HasChange("owners") {
		owners := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())
		if err := applicationSetOwners(ctx, client, &app, owners); err != nil {
			return err
		}
	}
	//if d.HasChange("oauth2_permissions") {
	//	// if the permission already exists then it must be disabled
	//	// with no other changes before it can be edited or deleted
	//	var app graphrbac.Application
	//	var appProperties graphrbac.ApplicationUpdateParameters
	//	resp, err := client.Get(ctx, d.Id())
	//	if err != nil {
	//		if utils.ResponseWasNotFound(resp.Response) {
	//			return fmt.Errorf("Application with ID %q was not found", d.Id())
	//		}
	//
	//		return fmt.Errorf("making Read request on Application with ID %q: %+v", d.Id(), err)
	//	}
	//	app = resp
	//	for _, OAuth2Permission := range *app.Oauth2Permissions {
	//		*OAuth2Permission.IsEnabled = false
	//	}
	//	appProperties.Oauth2Permissions = app.Oauth2Permissions
	//	if _, err := client.Patch(ctx, d.Id(), appProperties); err != nil {
	//		return fmt.Errorf("disabling OAuth2 permissions for Application with ID %q: %+v", d.Id(), err)
	//	}
	//
	//	// now we can set the new state of the permission
	//	properties.Oauth2Permissions = expandApplicationOAuth2Permissions(d.Get("oauth2_permissions"))
	//}
	//
	//if d.HasChange("app_role") {
	//	// if the app role already exists then it must be disabled
	//	// with no other changes before it can be edited or deleted
	//	var app graphrbac.Application
	//	var appRolesProperties graphrbac.ApplicationUpdateParameters
	//	resp, err := client.Get(ctx, d.Id())
	//	if err != nil {
	//		if utils.ResponseWasNotFound(resp.Response) {
	//			return fmt.Errorf("Application with ID %q was not found", d.Id())
	//		}
	//
	//		return fmt.Errorf("making Read request on Application with ID %q: %+v", d.Id(), err)
	//	}
	//	app = resp
	//	for _, appRole := range *app.AppRoles {
	//		*appRole.IsEnabled = false
	//	}
	//	appRolesProperties.AppRoles = app.AppRoles
	//	if _, err := client.Patch(ctx, d.Id(), appRolesProperties); err != nil {
	//		return fmt.Errorf("disabling App Roles for Application with ID %q: %+v", d.Id(), err)
	//	}
	//
	//	// now we can set the new state of the app role
	//	properties.AppRoles = expandApplicationAppRoles(d.Get("app_role"))
	//}
	//
	//if d.HasChange("group_membership_claims") {
	//	properties.GroupMembershipClaims = graphrbac.GroupMembershipClaimTypes(d.Get("group_membership_claims").(string))
	//}
	//
	//if d.HasChange("type") {
	//	switch appType := d.Get("type"); appType {
	//	case "webapp/api":
	//		properties.PublicClient = utils.Bool(false)
	//		properties.IdentifierUris = tf.ExpandStringSlicePtr(d.Get("identifier_uris").([]interface{}))
	//	case "native":
	//		properties.PublicClient = utils.Bool(true)
	//		properties.IdentifierUris = &[]string{}
	//	default:
	//		return fmt.Errorf("patching Application with ID %q: Unknow application type %v. Supported types are [webapp/api, native]", d.Id(), appType)
	//	}
	//}
	//
	//if _, err := client.Patch(ctx, d.Id(), properties); err != nil {
	//	return fmt.Errorf("patching Application with ID %q: %+v", d.Id(), err)
	//}

	return applicationResourceRead(d, meta)
}

func applicationResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	app, status, err := client.Get(ctx, d.Id())
	if err != nil {
		if status == http.StatusNoContent {
			log.Printf("[DEBUG] Application with ID %q was not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Application with ID %q: %+v", d.Id(), err)
	}

	d.Set("object_id", app.ID)
	d.Set("application_id", app.AppId)
	d.Set("display_name", app.DisplayName)

	if err := d.Set("api", flattenApplicationApi(app.Api)); err != nil {
		return fmt.Errorf("setting `api`: %+v", err)
	}

	if err := d.Set("group_membership_claims", app.GroupMembershipClaims); err != nil {
		return fmt.Errorf("setting `group_membership_claims`: %+v", err)
	}

	if err := d.Set("identifier_uris", tf.FlattenStringSlicePtr(app.IdentifierUris)); err != nil {
		return fmt.Errorf("setting `identifier_uris`: %+v", err)
	}

	if err := d.Set("optional_claims", flattenApplicationOptionalClaims(app.OptionalClaims)); err != nil {
		return fmt.Errorf("setting `optional_claims`: %+v", err)
	}

	if err := d.Set("public_client", flattenApplicationPublicClient(app.PublicClient)); err != nil {
		return fmt.Errorf("setting `public_client`: %+v", err)
	}

	requiredResourceAccess := flattenApplicationRequiredResourceAccess(app.RequiredResourceAccess)
	if err := d.Set("required_resource_access", requiredResourceAccess); err != nil {
		return fmt.Errorf("setting `required_resource_access`: %+v", err)
	}

	appRoles := flattenApplicationAppRoles(app.AppRoles)
	if err := d.Set("app_role", appRoles); err != nil {
		return fmt.Errorf("setting `app_role`: %+v", err)
	}

	if err := d.Set("sign_in_audience", app.SignInAudience); err != nil {
		return fmt.Errorf("setting `sign_in_audience`: %+v", err)
	}

	if err := d.Set("web", flattenApplicationWeb(app.Web)); err != nil {
		return fmt.Errorf("setting `web`: %+v", err)
	}

	owners, _, err := client.ListOwners(ctx, *app.ID)
	if err != nil {
		return fmt.Errorf("retrieving Owners for Application with ID %q: %+v", d.Id(), err)
	}

	d.Set("owners", owners)

	if preventDuplicates := d.Get("prevent_duplicate_names").(bool); !preventDuplicates {
		d.Set("prevent_duplicate_names", false)
	}

	return nil
}

func applicationResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.ApplicationsClient
	ctx := meta.(*clients.AadClient).StopContext

	//// in order to delete an application which is available to other tenants, we first have to disable this setting
	//availableToOtherTenants := d.Get("available_to_other_tenants").(bool)
	//if availableToOtherTenants {
	//	log.Printf("[DEBUG] Application is available to other tenants - disabling that feature before deleting.")
	//	properties := graphrbac.ApplicationUpdateParameters{
	//		AvailableToOtherTenants: utils.Bool(false),
	//	}
	//
	//	if _, err := client.Patch(ctx, d.Id(), properties); err != nil {
	//		return fmt.Errorf("patching Application with ID %q: %+v", d.Id(), err)
	//	}
	//}

	status, err := client.Delete(ctx, d.Id())
	if err != nil {
		if status != http.StatusNotFound {
			return fmt.Errorf("deleting Application with ID %q: %+v", d.Id(), err)
		}
	}

	return nil
}

func applicationCheckExistingDisplayName(ctx context.Context, client *clients2.ApplicationsClient, displayName string, existingId *string) error {
	filter := fmt.Sprintf("displayName eq '%s'", displayName)
	result, _, err := client.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("unable to list existing applications: %+v", err)
	}

	var existing []models.Application
	for _, r := range *result {
		if existingId != nil && *r.ID == *existingId {
			continue
		}
		if strings.EqualFold(displayName, *r.DisplayName) {
			existing = append(existing, r)
		}
	}
	count := len(existing)
	if count > 0 {
		noun := "application was"
		if count > 1 {
			noun = "applications were"
		}
		return fmt.Errorf("`prevent_duplicate_names` was specified and %d existing %s found with display_name %q", count, noun, displayName)
	}
	return nil
}

func applicationSetOwners(ctx context.Context, client *clients2.ApplicationsClient, application *models.Application, desiredOwners []string) error {
	owners, _, err := client.ListOwners(ctx, *application.ID)
	if err != nil {
		return fmt.Errorf("retrieving owners for Application with ID %q: %+v", *application.ID, err)
	}

	existingOwners := *owners
	ownersForRemoval := utils.Difference(existingOwners, desiredOwners)
	ownersToAdd := utils.Difference(desiredOwners, existingOwners)

	if ownersForRemoval != nil {
		if _, err = client.RemoveOwners(ctx, *application.ID, &ownersForRemoval); err != nil {
			return fmt.Errorf("removing owner from Application with ID %q: %+v", *application.ID, err)
		}
	}

	if ownersToAdd != nil {
		for _, m := range ownersToAdd {
			application.AppendOwner(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
		}

		if _, err := client.AddOwners(ctx, application); err != nil {
			return err
		}
	}
	return nil
}

func applicationValidateRolesScopes(appRoles, oauth2Permissions interface{}) error {
	var values []string

	if appRoles != nil {
		for _, roleRaw := range appRoles.(*schema.Set).List() {
			role := roleRaw.(map[string]interface{})
			if val := role["value"].(string); val != "" {
				values = append(values, val)
			}
		}
	}

	if oauth2Permissions != nil {
		for _, scopeRaw := range oauth2Permissions.(*schema.Set).List() {
			scope := scopeRaw.(map[string]interface{})
			if val := scope["value"].(string); val != "" {
				values = append(values, val)
			}
		}
	}

	encountered := make([]string, len(values))
	for _, val := range values {
		for _, en := range encountered {
			if en == val {
				return fmt.Errorf("validation failed: duplicate app_role / oauth2_permission_scope value found: %q", val)
			}
		}
		encountered = append(encountered, val)
	}

	return nil
}
