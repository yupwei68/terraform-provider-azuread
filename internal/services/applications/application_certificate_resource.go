package applications

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func applicationCertificateResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: applicationCertificateResourceCreate,
		ReadContext:   applicationCertificateResourceRead,
		DeleteContext: applicationCertificateResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := aadgraph.ParseCertificateId(id)
			return err
		}),

		Schema: aadgraph.CertificateResourceSchema("application_object_id"),
	}
}

func applicationCertificateResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	objectId := d.Get("application_object_id").(string)

	cred, err := aadgraph.KeyCredentialForResource(d)
	if err != nil {
		attr := ""
		if kerr, ok := err.(aadgraph.CredentialError); ok {
			attr = kerr.Attr()
		}
		return tf.ErrorDiag(fmt.Sprintf("Generating certificate credentials for application with object ID %q", objectId), err.Error(), attr)
	}

	id := aadgraph.CredentialIdFrom(objectId, "certificate", *cred.KeyID)

	tf.LockByName(resourceApplicationName, id.ObjectId)
	defer tf.UnlockByName(resourceApplicationName, id.ObjectId)

	existingCreds, err := client.ListKeyCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing certificate credentials for application with ID %q", objectId), err.Error(), "application_object_id")
	}

	newCreds, err := aadgraph.KeyCredentialResultAdd(existingCreds, cred)
	if err != nil {
		if _, ok := err.(*aadgraph.AlreadyExistsError); ok {
			return tf.ImportAsExistsDiag("azuread_application_certificate", id.String())
		}
		return tf.ErrorDiag("Adding application certificate", err.Error(), "")
	}

	if _, err = client.UpdateKeyCredentials(ctx, id.ObjectId, graphrbac.KeyCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Creating certificate credentials %q for application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "")
	}

	_, err = aadgraph.WaitForKeyCredentialReplication(ctx, id.KeyId, d.Timeout(schema.TimeoutCreate), func() (graphrbac.KeyCredentialListResult, error) {
		return client.ListKeyCredentials(ctx, id.ObjectId)
	})
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Waiting for certificate credential replication for application (AppID %q, KeyID %q)", id.ObjectId, id.KeyId), err.Error(), "")
	}

	d.SetId(id.String())

	return applicationCertificateResourceRead(ctx, d, meta)
}

func applicationCertificateResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParseCertificateId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing certificate credential with ID %q", d.Id()), err.Error(), "id")
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
		return tf.ErrorDiag(fmt.Sprintf("Retrieving application with ID %q", id.ObjectId), err.Error(), "name")
	}

	credentials, err := client.ListKeyCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing certificate credentials for application with object ID %q", id.ObjectId), err.Error(), "name")
	}

	credential := aadgraph.KeyCredentialResultFindByKeyId(credentials, id.KeyId)
	if credential == nil {
		log.Printf("[DEBUG] Certificate credential %q (ID %q) was not found - removing from state!", id.KeyId, id.ObjectId)
		d.SetId("")
		return nil
	}

	if err := d.Set("application_object_id", id.ObjectId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "application_object_id")
	}

	if err := d.Set("key_id", id.KeyId); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "key_id")
	}

	keyType := ""
	if v := credential.Type; v != nil {
		keyType = *v
	}
	if err := d.Set("type", keyType); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "type")
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

func applicationCertificateResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clients.Client).Applications.ApplicationsClient

	id, err := aadgraph.ParseCertificateId(d.Id())
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Parsing certificate credential with ID %q", d.Id()), err.Error(), "id")
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

	existing, err := client.ListKeyCredentials(ctx, id.ObjectId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Listing certificate credential for application with object ID %q", id.ObjectId), err.Error(), "name")
	}

	newCreds, err := aadgraph.KeyCredentialResultRemoveByKeyId(existing, id.KeyId)
	if err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Removing certificate credential %q from application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "name")
	}

	if _, err = client.UpdateKeyCredentials(ctx, id.ObjectId, graphrbac.KeyCredentialsUpdateParameters{Value: newCreds}); err != nil {
		return tf.ErrorDiag(fmt.Sprintf("Removing certificate credential %q from application with object ID %q", id.KeyId, id.ObjectId), err.Error(), "name")
	}

	return nil
}
