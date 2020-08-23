package provider

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/manicminer/hamilton/auth"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/msgraph"

	msgraphclient "github.com/terraform-providers/terraform-provider-azuread/internal/services/msgraph/client"
)

type ServiceRegistration interface {
	// Name is the name of this Service
	Name() string

	// WebsiteCategories returns a list of categories which can be used for the sidebar
	WebsiteCategories() []string

	// SupportedDataSources returns the supported Data Sources supported by this Service
	SupportedDataSources() map[string]*schema.Resource

	// SupportedResources returns the supported Resources supported by this Service
	SupportedResources() map[string]*schema.Resource
}

// Provider returns a terraform.ResourceProvider.
func AzureADProvider() terraform.ResourceProvider {
	// avoids this showing up in test output
	var debugLog = func(f string, v ...interface{}) {
		if os.Getenv("TF_LOG") == "" {
			return
		}

		if os.Getenv("TF_ACC") != "" {
			return
		}

		log.Printf(f, v...)
	}

	// only one for now so keeping it simple, eventually we will need a way to differentiate between aadgraph and msgraph?
	// looks like only an env var will work?
	services := []ServiceRegistration{
		aadgraph.Registration{},
		msgraph.Registration{},
	}

	dataSources := make(map[string]*schema.Resource)
	resources := make(map[string]*schema.Resource)
	for _, service := range services {
		debugLog("[DEBUG] Registering Resources for %q..", service.Name())
		for k, v := range service.SupportedResources() {
			if existing := resources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Resource exists for %q", k))
			}

			resources[k] = v
		}

		debugLog("[DEBUG] Registering Data Sources for %q..", service.Name())
		for k, v := range service.SupportedDataSources() {
			if existing := dataSources[k]; existing != nil {
				panic(fmt.Sprintf("An existing Data Source exists for %q", k))
			}

			dataSources[k] = v
		}
	}

	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},

			"metadata_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOSTNAME", ""),
				Description: "The Hostname which should be used to fetch environment metadata from.",
			},

			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
			},

			// Client Certificate specific fields
			"client_certificate_password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", ""),
			},

			"client_certificate_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", ""),
			},

			// Client Secret specific fields
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
			},

			// Managed Service Identity specific fields
			"use_msi": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
			},

			"msi_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
			},
		},

		ResourcesMap:   resources,
		DataSourcesMap: dataSources,
	}

	p.ConfigureFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		clientId := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)
		tenantId := d.Get("tenant_id").(string)

		builder := &authentication.Builder{
			ClientID:           clientId,
			ClientSecret:       clientSecret,
			TenantID:           tenantId,
			SubscriptionID:     tenantId, // TODO: delete in v1.1
			MetadataURL:        d.Get("metadata_host").(string),
			Environment:        d.Get("environment").(string),
			MsiEndpoint:        d.Get("msi_endpoint").(string),
			ClientCertPassword: d.Get("client_certificate_password").(string),
			ClientCertPath:     d.Get("client_certificate_path").(string),

			// Feature Toggles
			SupportsClientCertAuth:         true,
			SupportsClientSecretAuth:       true,
			SupportsManagedServiceIdentity: d.Get("use_msi").(bool),
			SupportsAzureCliToken:          true,
			//TenantOnly:                     true, // TODO: enable in v1.1
		}

		config, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("building AzureAD Client: %s", err)
		}

		clientBuilder := clients.ClientBuilder{
			AuthConfig:       config,
			TerraformVersion: p.TerraformVersion,
		}

		client, err := clientBuilder.Build(p.StopContext())
		if err != nil {
			return nil, err
		}

		// MS Graph
		msGraphAuthorizer := auth.NewClientSecretAuthorizer(p.StopContext(), clientId, clientSecret, tenantId)
		client.MsGraph = msgraphclient.BuildClient(msGraphAuthorizer, tenantId)

		// replaces the context between tests
		p.MetaReset = func() error { //nolint unparam
			client.StopContext = p.StopContext()
			return nil
		}

		return client, nil
	}
}
