package shim

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/msgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func GroupDataSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"object_id": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.UUID,
			ExactlyOneOf: []string{"display_name", "object_id"},
		},

		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"display_name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validate.NoEmptyStrings,
			ExactlyOneOf: []string{"display_name", "object_id"},
		},

		"members": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"owners": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}

func GroupData() *schema.Resource {
	return &schema.Resource{
		Read: groupDataRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: GroupDataSchema(),
	}
}

func groupDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient)

	if client.UseMsGraph {
		return msgraph.GroupDataRead(d, meta)
	} else {
		return aadgraph.GroupsDataRead(d, meta)
	}
}
