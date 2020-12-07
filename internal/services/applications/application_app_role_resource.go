package applications

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func applicationAppRoleResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: applicationAppRoleResourceCreateUpdate,
		UpdateContext: applicationAppRoleResourceCreateUpdate,
		ReadContext:   applicationAppRoleResourceRead,
		DeleteContext: applicationAppRoleResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := aadgraph.ParseAppRoleId(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"application_object_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.NoEmptyStrings,
			},

			"display_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.NoEmptyStrings,
			},

			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"role_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},

			"value": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func applicationAppRoleResourceCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	objectId := d.Get("application_object_id").(string)

	// errors should be handled by the validation
	var roleId string
	if v, ok := d.GetOk("role_id"); ok {
		roleId = v.(string)
	} else {
		rid, err := uuid.GenerateUUID()
		if err != nil {
			return tf.ErrorDiag(fmt.Sprintf("Generating App Role for application with object ID %q", objectId), err.Error(), "")
		}
		roleId = rid
	}

	allowedMemberTypesRaw := d.Get("allowed_member_types").(*schema.Set).List()
	allowedMemberTypes := make([]string, 0, len(allowedMemberTypesRaw))
	for _, a := range allowedMemberTypesRaw {
		allowedMemberTypes = append(allowedMemberTypes, a.(string))
	}

	role := graphrbac.AppRole{
		AllowedMemberTypes: &allowedMemberTypes,
		ID:                 utils.String(roleId),
		Description:        utils.String(d.Get("description").(string)),
		DisplayName:        utils.String(d.Get("display_name").(string)),
		IsEnabled:          utils.Bool(d.Get("is_enabled").(bool)),
	}

	if v, ok := d.GetOk("value"); ok {
		role.Value = utils.String(v.(string))
	}

	id := aadgraph.AppRoleIdFrom(objectId, *role.ID)

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	// ensure the Application Object exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		if utils.ResponseWasNotFound(app.Response) {
			return tf.ErrorDiag(fmt.Sprintf("Application with object ID %q was not found", id.ObjectId), "", "application_object_id")
		}
		return tf.ErrorDiag(fmt.Sprintf("retrieving Application with object ID %q", id.ObjectId), err.Error(), "application_object_id")
	}

	var newRoles *[]graphrbac.AppRole

	if d.IsNewResource() {
		newRoles, err = aadgraph.AppRoleAdd(app.AppRoles, &role)
		if err != nil {
			if _, ok := err.(*aadgraph.AlreadyExistsError); ok {
				return tf.ImportAsExistsDiag("azuread_application_app_role", id.String())
			}
			return tf.ErrorDiag("Failed to add App Role", err.Error(), "")
		}
	} else {
		if existing, _ := aadgraph.AppRoleFindById(app, id.RoleId); existing == nil {
			return tf.ErrorDiag(fmt.Sprintf("App Role with ID %q was not found for Application %q", id.RoleId, id.ObjectId), "", "role_id")
		}

		newRoles, err = aadgraph.AppRoleUpdate(app.AppRoles, &role)
		if err != nil {
			return tf.ErrorDiag(fmt.Sprintf("Updating App Role with ID %q", *role.ID), err.Error(), "")
		}
	}

	properties := graphrbac.ApplicationUpdateParameters{
		AppRoles: newRoles,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Updating Application with ID %q", id.ObjectId), err.Error(), "")
	}

	d.SetId(id.String())

	return applicationAppRoleResourceRead(ctx, d, meta)
}

func applicationAppRoleResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParseAppRoleId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing App Role ID %q", d.Id()), err.Error(), "id")
	}

	// ensure the Application Object exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Application has been removed - skip it
		if utils.ResponseWasNotFound(app.Response) {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			d.SetId("")
			return nil
		}
		return tf.ErrorDiag(fmt.Sprintf("Retrieving Application with ID %q", id.ObjectId), err.Error(), "application_object_id")
	}

	role, err := aadgraph.AppRoleFindById(app, id.RoleId)
	if err != nil {
		return tf.ErrorDiag("Identifying App Role", err.Error(), "")
	}

	if role == nil {
		log.Printf("[DEBUG] App Role %q (ID %q) was not found - removing from state!", id.RoleId, id.ObjectId)
		d.SetId("")
		return nil
	}

	if err := d.Set("application_object_id", id.ObjectId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "application_object_id")
	}

	if err := d.Set("role_id", id.RoleId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "role_id")
	}

	if err := d.Set("allowed_member_types", role.AllowedMemberTypes); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "allowed_member_types")
	}

	if err := d.Set("description", role.Description); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "description")
	}

	if err := d.Set("display_name", role.DisplayName); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "display_name")
	}

	if err := d.Set("is_enabled", role.IsEnabled); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "is_enabled")
	}

	if err := d.Set("value", role.Value); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "value")
	}

	return nil
}

func applicationAppRoleResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParseAppRoleId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing App Role ID %q", d.Id()), err.Error(), "id")
	}

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	// ensure the parent Application exists
	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Application has been removed - skip it
		if utils.ResponseWasNotFound(app.Response) {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			return nil
		}
		return tf.ErrorDiag(fmt.Sprintf("Retrieving Application with ID %q", id.ObjectId), err.Error(), "application_object_id")
	}

	log.Printf("[DEBUG] Disabling App Role %q for Application %q prior to removal", id.RoleId, id.ObjectId)
	newRoles, err := aadgraph.AppRoleResultDisableById(app.AppRoles, id.RoleId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Disabling App Role with ID %q for application %q", id.RoleId, id.ObjectId), err.Error(), "")
	}

	properties := graphrbac.ApplicationUpdateParameters{
		AppRoles: newRoles,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Updating Application with ID %q", id.ObjectId), err.Error(), "")
	}

	log.Printf("[DEBUG] Removing App Role %q for Application %q", id.RoleId, id.ObjectId)
	newRoles, err = aadgraph.AppRoleResultRemoveById(app.AppRoles, id.RoleId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Removing App Role with ID %q for application %q", id.RoleId, id.ObjectId), err.Error(), "")
	}

	properties = graphrbac.ApplicationUpdateParameters{
		AppRoles: newRoles,
	}
	if _, err := client.Patch(ctx, id.ObjectId, properties); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Updating Application with ID %q", id.ObjectId), err.Error(), "")
	}

	return nil
}
