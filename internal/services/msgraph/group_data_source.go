package msgraph

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func GroupData() *schema.Resource {
	return &schema.Resource{
		Read: groupDataRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.UUID,
				ExactlyOneOf: []string{"display_name", "object_id"},
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validate.NoEmptyStrings,
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
		},
	}
}

func groupDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	var group models.Group

	if objectId, ok := d.Get("object_id").(string); ok && objectId != "" {
		g, err := client.Get(ctx, objectId)
		if err != nil {
			return fmt.Errorf("reading Group with ID %q: %+v", objectId, err)
		}
		group = *g
	} else if displayName, ok := d.Get("display_name").(string); ok && displayName != "" {
		filter := fmt.Sprintf("displayName eq '%s'", displayName)
		groups, err := client.List(ctx, filter)
		if err != nil {
			return fmt.Errorf("identifying Group with display name %q: %+v", displayName, err)
		}

		if len(*groups) > 1 {
			return fmt.Errorf("more than one group found with display name: %q", displayName)
		} else if len(*groups) == 0 {
			return fmt.Errorf("no groups found with display name: %q", displayName)
		}

		group = (*groups)[0]
	}

	if group.ID == nil {
		return fmt.Errorf("Group objectId is nil")
	}

	d.SetId(*group.ID)

	d.Set("object_id", group.ID)
	d.Set("display_name", group.DisplayName)

	if v := group.Description; v != nil {
		d.Set("description", v)
	}

	members, err := client.ListMembers(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("unable to retrieve group members for %s: %+v", d.Id(), err)
	}
	d.Set("members", members)

	owners, err := client.ListOwners(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("unable to retrieve group owners for %s: %+v", d.Id(), err)
	}
	d.Set("owners", owners)

	return nil
}
