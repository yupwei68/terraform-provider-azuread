package msgraph

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func GroupsData() *schema.Resource {
	return &schema.Resource{
		Read: groupsDataRead,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_ids": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"display_names", "object_ids"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"display_names": {
				Type:         schema.TypeList,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"display_names", "object_ids"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.NoEmptyStrings,
				},
			},
		},
	}
}

func groupsDataRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	var groups []models.Group
	var expectedCount int

	if displayNames, ok := d.Get("display_names").([]interface{}); ok && len(displayNames) > 0 {
		expectedCount = len(displayNames)
		for _, v := range displayNames {
			displayName := v.(string)
			filter := fmt.Sprintf("displayName eq '%s'", displayName)
			result, _, err := client.List(ctx, filter)
			if err != nil {
				return fmt.Errorf("finding Group with display name %q: %+v", displayName, err)
			}

			count := len(*result)
			if count > 1 {
				return fmt.Errorf("more than one group found with display name: %q", displayName)
			} else if count == 0 {
				return fmt.Errorf("no groups found with display name: %q", displayName)
			}

			groups = append(groups, (*result)[0])
		}
	} else if objectIds, ok := d.Get("object_ids").([]interface{}); ok && len(objectIds) > 0 {
		expectedCount = len(objectIds)
		for _, v := range objectIds {
			objectId := v.(string)
			group, status, err := client.Get(ctx, objectId)
			if err != nil {
				if status == http.StatusNotFound {
					return fmt.Errorf("Group with ID %q was not found", objectId)
				}
				return fmt.Errorf("reading Group with ID %q: %+v", objectId, err)
			}

			groups = append(groups, *group)
		}
	}

	if len(groups) != expectedCount {
		return fmt.Errorf("unexpected number of groups returned (%d != %d)", len(groups), expectedCount)
	}

	displayNames := make([]string, 0, len(groups))
	objectIds := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.ID == nil || group.DisplayName == nil {
			return fmt.Errorf("Group with null ID was found: %v", group)
		}

		objectIds = append(objectIds, *group.ID)
		displayNames = append(displayNames, *group.DisplayName)
	}

	h := sha1.New()
	if _, err := h.Write([]byte(strings.Join(displayNames, "-"))); err != nil {
		return fmt.Errorf("unable to compute hash for displayNames: %v", err)
	}

	d.SetId("groups#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	d.Set("object_ids", objectIds)
	d.Set("display_names", displayNames)
	return nil
}
