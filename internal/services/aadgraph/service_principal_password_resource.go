package aadgraph

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
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph/graph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func servicePrincipalPasswordResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: servicePrincipalPasswordResourceCreate,
		ReadContext:   servicePrincipalPasswordResourceRead,
		DeleteContext: servicePrincipalPasswordResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := graph.ParsePasswordId(id)
			return err
		}),

		Schema: graph.PasswordResourceSchema("service_principal_id"),

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceServicePrincipalPasswordInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceServicePrincipalPasswordInstanceStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func servicePrincipalPasswordResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.AadClient).AadGraph.ServicePrincipalsClient

	objectId := d.Get("service_principal_id").(string)

	cred, err := graph.PasswordCredentialForResource(d)
	if err != nil {
		attr := ""
		if kerr, ok := err.(graph.CredentialError); ok {
			attr = kerr.Attr()
		}
		return tf.ErrorDiagPathF(err, attr, "Generating password credentials for service principal with object ID %q", objectId)
	}
	id := graph.CredentialIdFrom(objectId, "password", *cred.KeyID)

	tf.LockByName(servicePrincipalResourceName, id.ObjectId)
	defer tf.UnlockByName(servicePrincipalResourceName, id.ObjectId)

	existingCreds, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiagPathF(err, "service_principal_id", "Listing password credentials for service principal with ID %q", objectId)
	}

	newCreds, err := graph.PasswordCredentialResultAdd(existingCreds, cred)
	if err != nil {
		if _, ok := err.(*graph.AlreadyExistsError); ok {
			return tf.ImportAsExistsDiag("azuread_service_principal_password", id.String())
		}
		return tf.ErrorDiagF(err, "Adding service principal password")
	}

	if _, err = client.UpdatePasswordCredentials(ctx, objectId, graphrbac.PasswordCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiagF(err, "Creating password credentials %q for service principal with object ID %q", id.KeyId, id.ObjectId)
	}

	d.SetId(id.String())

	_, err = graph.WaitForPasswordCredentialReplication(ctx, id.KeyId, d.Timeout(schema.TimeoutCreate), func() (graphrbac.PasswordCredentialListResult, error) {
		return client.ListPasswordCredentials(ctx, id.ObjectId)
	})
	if err != nil {
		return tf.ErrorDiagF(err, "Waiting for password credential replication for service principal (ObjectID %q, KeyID %q)", id.ObjectId, id.KeyId)
	}

	return servicePrincipalPasswordResourceRead(ctx, d, meta)
}

func servicePrincipalPasswordResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.AadClient).AadGraph.ServicePrincipalsClient

	id, err := graph.ParsePasswordId(d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "id", "Parsing password credential with ID %q", d.Id())
	}

	// ensure the parent Service Principal exists
	servicePrincipal, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Service Principal has been removed - skip it
		if utils.ResponseWasNotFound(servicePrincipal.Response) {
			log.Printf("[DEBUG] Service Principal with Object ID %q was not found - removing from state!", id.ObjectId)
			d.SetId("")
			return nil
		}
		return tf.ErrorDiagPathF(err, "service_principal_id", "Retrieving service principal with object ID %q", id.ObjectId)
	}

	credentials, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiagPathF(err, "service_principal_id", "Listing password credentials for service principal with object ID %q", id.ObjectId)
	}

	credential := graph.PasswordCredentialResultFindByKeyId(credentials, id.KeyId)
	if credential == nil {
		log.Printf("[DEBUG] Service Principal %q (ID %q) was not found - removing from state!", id.KeyId, id.ObjectId)
		d.SetId("")
		return nil
	}

	if dg := tf.Set(d, "service_principal_id", id.ObjectId); dg != nil {
		return dg
	}

	if dg := tf.Set(d, "key_id", id.KeyId); dg != nil {
		return dg
	}

	description := ""
	if v := credential.CustomKeyIdentifier; v != nil {
		description = string(*v)
	}
	if dg := tf.Set(d, "description", description); dg != nil {
		return dg
	}

	startDate := ""
	if v := credential.StartDate; v != nil {
		startDate = v.Format(time.RFC3339)
	}
	if dg := tf.Set(d, "start_date", startDate); dg != nil {
		return dg
	}

	endDate := ""
	if v := credential.EndDate; v != nil {
		endDate = v.Format(time.RFC3339)
	}
	if dg := tf.Set(d, "end_date", endDate); dg != nil {
		return dg
	}

	return nil
}

func servicePrincipalPasswordResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.AadClient).AadGraph.ServicePrincipalsClient

	id, err := graph.ParsePasswordId(d.Id())
	if err != nil {
		return tf.ErrorDiagPathF(err, "id", "Parsing password credential with ID %q", d.Id())
	}

	tf.LockByName(servicePrincipalResourceName, id.ObjectId)
	defer tf.UnlockByName(servicePrincipalResourceName, id.ObjectId)

	// ensure the parent Service Principal exists
	servicePrincipal, err := client.Get(ctx, id.ObjectId)
	if err != nil {
		// the parent Service Principal was removed - skip it
		if utils.ResponseWasNotFound(servicePrincipal.Response) {
			log.Printf("[DEBUG] Service Principal with Object ID %q was not found - removing from state!", id.ObjectId)
			return nil
		}
		return tf.ErrorDiagPathF(err, "service_principal_id", "Retrieving service principal with object ID %q", id.ObjectId)
	}

	existing, err := client.ListPasswordCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiagPathF(err, "service_principal_id", "Listing password credentials for service principal with object ID %q", id.ObjectId)
	}

	newCreds, err := graph.PasswordCredentialResultRemoveByKeyId(existing, id.KeyId)
	if err != nil {
		return tf.ErrorDiagF(err, "Removing password credential %q from service principal with object ID %q", id.KeyId, id.ObjectId)
	}

	if _, err = client.UpdatePasswordCredentials(ctx, id.ObjectId, graphrbac.PasswordCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiagF(err, "Removing password credential %q from service principal with object ID %q", id.KeyId, id.ObjectId)
	}

	return nil
}

func resourceServicePrincipalPasswordInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"service_principal_id": {
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

func resourceServicePrincipalPasswordInstanceStateUpgradeV0(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	log.Println("[DEBUG] Migrating ID from v0 to v1 format")
	newId, err := graph.ParseOldPasswordId(rawState["id"].(string))
	if err != nil {
		return rawState, fmt.Errorf("generating new ID: %s", err)
	}

	rawState["id"] = newId.String()
	return rawState, nil
}