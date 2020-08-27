package msgraph

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
)

func DomainsData() *schema.Resource {
	return &schema.Resource{
		Read: domainsDataRead,

		Schema: map[string]*schema.Schema{
			"include_unverified": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_default", "only_initial", "only_root"}, // default, initial or root domains have to be verified
			},

			"only_default": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_initial", "only_root"},
			},

			"only_initial": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_default", "only_root"},
			},

			"only_root": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"only_default", "only_initial"},
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

						"is_root": {
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

func domainsDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.DomainsClient
	ctx := meta.(*clients.AadClient).StopContext

	result, _, err := client.List(ctx)
	if err != nil {
		return fmt.Errorf("listing Domains: %+v", err)
	} else if result == nil {
		return errors.New("null Domains response from server")
	} else if len(*result) == 0 {
		return errors.New("no Domains were returned")
	}

	domains := flattenDomains(d, result)
	if len(domains) == 0 {
		return errors.New("no domains matched the specified filters")
	}

	d.SetId(domainsDataId(d, meta))

	if err = d.Set("domains", domains); err != nil {
		return fmt.Errorf("setting `result`: %+v", err)
	}

	return nil
}

func domainsDataId(d *schema.ResourceData, meta interface{}) (id string) {
	a := []string{
		"include_unverified",
		"only_default",
		"only_initial",
		"only_root",
	}
	tenantId := meta.(*clients.AadClient).TenantID
	id = fmt.Sprintf("domains-%s-", tenantId)
	for i, v := range a {
		if d.Get(v).(bool) {
			id = id + strconv.Itoa(i)
		}
	}
	return id
}

func flattenDomains(d *schema.ResourceData, input *[]models.Domain) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	includeUnverified := d.Get("include_unverified").(bool)
	onlyDefault := d.Get("only_default").(bool)
	onlyInitial := d.Get("only_initial").(bool)
	onlyRoot := d.Get("only_root").(bool)

	domains := make([]interface{}, 0, len(*input))

	for _, v := range *input {
		authenticationType := ""
		if v.AuthenticationType != nil {
			authenticationType = *v.AuthenticationType
		}

		isDefault := false
		if v.IsDefault != nil {
			isDefault = *v.IsDefault
		}

		isInitial := false
		if v.IsInitial != nil {
			isInitial = *v.IsInitial
		}

		isRoot := false
		if v.IsRoot != nil {
			isRoot = *v.IsRoot
		}

		isVerified := false
		if v.IsVerified != nil {
			isVerified = *v.IsVerified
		}

		if !includeUnverified {
			if onlyDefault && !isDefault {
				continue
			} else if onlyInitial && !isInitial {
				continue
			} else if onlyRoot && !isRoot {
				continue
			}
		} else if !isVerified {
			continue
		}

		domains = append(domains, map[string]interface{}{
			"domain_name":         *v.ID,
			"authentication_type": authenticationType,
			"is_default":          isDefault,
			"is_initial":          isInitial,
			"is_root":             isRoot,
			"is_verified":         isVerified,
		})
	}

	return domains
}
