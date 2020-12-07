package applications

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func applicationPasswordResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: applicationPasswordResourceCreate,
		ReadContext:   applicationPasswordResourceRead,
		DeleteContext: applicationPasswordResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := aadgraph.ParsePasswordId(id)
			return err
		}),

		Schema: aadgraph.PasswordResourceSchema("application_object_id"),

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceApplicationPasswordInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceApplicationPasswordInstanceStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func applicationPasswordResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	objectId := d.Get("application_object_id").(string)

	cred, err := aadgraph.PasswordCredentialForResource(d)
	if err != nil {
		attr := ""
		if kerr, ok := err.(aadgraph.CredentialError); ok {
			attr = kerr.Attr()
		}
		return tf.ErrorDiag(fmt.Sprintf("Generating password credentials for application with object ID %q", objectId), err.Error(), attr)
	}
	id := aadgraph.CredentialIdFrom(objectId, "password", *cred.KeyID)

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	existingCreds, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing password credentials for application with ID %q", objectId), err.Error(), "application_object_id")
	}

	newCreds, err := aadgraph.PasswordCredentialResultAdd(existingCreds, cred)
	if err != nil {
		if _, ok := err.(*aadgraph.AlreadyExistsError); ok {
			return tf.ImportAsExistsDiag("azuread_application_password", id.String())
		}
		return tf.ErrorDiag("Adding application password", err.Error(), "")
	}

	if _, err = client.UpdatePasswordCredentials(ctx, id.ObjectId, graphrbac.PasswordCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Creating password credentials %q for application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "")
	}

	_, err = aadgraph.WaitForPasswordCredentialReplication(ctx, id.KeyId, d.Timeout(schema.TimeoutCreate), func() (graphrbac.PasswordCredentialListResult, error) {
		return client.ListPasswordCredentials(ctx, id.ObjectId)
	})
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Waiting for certificate credential replication for application (AppID %q, KeyID %q)", id.ObjectId, id.KeyId), err.Error(), "")
	}

	d.SetId(id.String())

	return applicationPasswordResourceRead(ctx, d, meta)
}

func applicationPasswordResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParsePasswordId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing password credential with ID %q", d.Id()), err.Error(), "id")
	}

	app, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Application has been removed - skip it
		if utils.ResponseWasNotFound(app.Response) {
			log.Printf("[DEBUG] Application with Object ID %q was not found - removing from state!", id.ObjectId)
			d.SetId("")
			return nil
		}
		return tf.ErrorDiag(fmt.Sprintf("Retrieving application with ID %q", id.ObjectId), err.Error(), "name")
	}

	credentials, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing password credentials for application with object ID %q", id.ObjectId), err.Error(), "name")
	}

	credential := aadgraph.PasswordCredentialResultFindByKeyId(credentials, id.KeyId)
	if credential == nil {
		log.Printf("[DEBUG] Password credential %q (ID %q) was not found - removing from state!", id.KeyId, id.ObjectId)
		d.SetId("")
		return nil
	}

	if err := d.Set("application_object_id", id.ObjectId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "application_object_id")
	}

	if err := d.Set("key_id", id.KeyId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "key_id")
	}

	description := ""
	if v := credential.CustomKeyIdentifier; v != nil {
		description = string(*v)
	}
	if err := d.Set("description", description); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "description")
	}

	startDate := ""
	if v := credential.StartDate; v != nil {
		startDate = v.Format(time.RFC3339)
	}
	if err := d.Set("start_date", startDate); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "start_date")
	}

	endDate := ""
	if v := credential.EndDate; v != nil {
		endDate = v.Format(time.RFC3339)
	}
	if err := d.Set("end_date", endDate); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "end_date")
	}

	return nil
}

func applicationPasswordResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParsePasswordId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing password credential with ID %q", d.Id()), err.Error(), "id")
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
		return tf.ErrorDiag(fmt.Sprintf("Retrieving application with ID %q", id.ObjectId), err.Error(), "name")
	}

	existing, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing password credentials for application with object ID %q", id.ObjectId), err.Error(), "name")
	}

	newCreds, err := aadgraph.PasswordCredentialResultRemoveByKeyId(existing, id.KeyId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Removing password credential %q from application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "name")
	}

	if _, err = client.UpdatePasswordCredentials(ctx, id.ObjectId, graphrbac.PasswordCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Removing password credential %q from application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "name")
	}

	return nil
}

func resourceApplicationPasswordInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"application_object_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},

			"key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 863),
			},

			"start_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},

			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"end_date_relative"},
				ValidateFunc: validation.IsRFC3339Time,
			},

			"end_date_relative": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ExactlyOneOf:     []string{"end_date"},
				ValidateDiagFunc: validate.NoEmptyStrings,
			},
		},
	}
}

func resourceApplicationPasswordInstanceStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	log.Println("[DEBUG] Migrating ID from v0 to v1 format")
	newId, err := aadgraph.ParseOldPasswordId(rawState["id"].(string))
	if err != nil {
		return rawState, fmt.Errorf("generating new ID: %s", err)
	}

	rawState["id"] = newId.String()
	return rawState, nil
}
