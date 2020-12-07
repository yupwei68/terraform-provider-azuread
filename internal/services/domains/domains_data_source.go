package domains

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func domainsData() *schema.Resource {
	return &schema.Resource{
		ReadContext: domainsDataRead,

		Schema: map[string]*schema.Schema{
			"include_unverified": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_default", "only_initial"}, // default or initial domains have to be verified
			},
			"only_default": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_initial"},
			},
			"only_initial": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_default"},
			},
			"domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"authentication_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_default": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_initial": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_verified": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func domainsDataRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := meta.(*clients.Client).TenantID
	client := meta.(*clients.Client).Domains.DomainsClient

	includeUnverified := d.Get("include_unverified").(bool)
	onlyDefault := d.Get("only_default").(bool)
	onlyInitial := d.Get("only_initial").(bool)

	results, err := client.List(ctx, "")
	if err != nil {
		return tf.ErrorDiag("Listing domains", err.Error(), "")
	}

	d.SetId("domains-" + tenantId) // todo this should be more unique

	domains := flattenDomains(results.Value, includeUnverified, onlyDefault, onlyInitial)
	if len(domains) == 0 {
		return tf.ErrorDiag("No domains were returned for the provided filters", "", "")
	}

	if err = d.Set("domains", domains); err != nil {
		return tf.ErrorDiag("Could not set attribute", err.Error(), "domains")
	}

	return nil
}

func flattenDomains(input *[]graphrbac.Domain, includeUnverified, onlyDefault, onlyInitial bool) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	domains := make([]interface{}, 0)
	for _, v := range *input {
		if v.Name == nil {
			log.Printf("[DEBUG] Domain Name was nil - skipping")
			continue
		}

		domainName := *v.Name

		authenticationType := "undefined"
		if v.AuthenticationType != nil {
			authenticationType = *v.AuthenticationType
		}

		isDefault := false
		if v.IsDefault != nil {
			isDefault = *v.IsDefault
		}

		isInitial := false
		if v.AdditionalProperties["isInitial"] != nil {
			isInitial = v.AdditionalProperties["isInitial"].(bool)
		}

		isVerified := false
		if v.IsVerified != nil {
			isVerified = *v.IsVerified
		}

		// Filters
		if !isDefault && onlyDefault {
			// skip all domains except the initial domain
			log.Printf("[DEBUG] Skipping %q since the filter requires the default domain", domainName)
			continue
		}

		if !isInitial && onlyInitial {
			// skip all domains except the initial domain
			log.Printf("[DEBUG] Skipping %q since the filter requires the initial domain", domainName)
			continue
		}

		if !isVerified && !includeUnverified {
			//skip unverified domains
			log.Printf("[DEBUG] Skipping %q since the filter requires verified domains", domainName)
			continue
		}

		domain := map[string]interface{}{
			"authentication_type": authenticationType,
			"domain_name":         domainName,
			"is_default":          isDefault,
			"is_initial":          isInitial,
			"is_verified":         isVerified,
		}

		domains = append(domains, domain)
	}

	return domains
}
