package msgraph

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/services/msgraph/helper"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

const groupMemberResourceName = "azuread_group_member_msgraph"

func GroupMemberResource() *schema.Resource {
	return &schema.Resource{
		Create: groupMemberResourceCreate,
		Read:   groupMemberResourceRead,
		Delete: groupMemberResourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group_object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},

			"member_object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},
		},
	}
}

func groupMemberResourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	groupID := d.Get("group_object_id").(string)
	memberID := d.Get("member_object_id").(string)

	tf.LockByName(groupMemberResourceName, groupID)
	defer tf.UnlockByName(groupMemberResourceName, groupID)

	group, status, err := client.Get(ctx, groupID)
	if err != nil {
		if status == http.StatusNotFound {
			return fmt.Errorf("Group with ID %q was not found", groupID)
		}
		return fmt.Errorf("could not retrieve group details for %q: %+v", groupID, err)
	}

	group.AppendMember(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, memberID)

	if _, err := client.AddMembers(ctx, group); err != nil {
		return fmt.Errorf("adding member %q to group %q: %+v", memberID, groupID, err)
	}

	d.SetId(helper.GroupMemberIdFrom(groupID, memberID).String())

	return groupMemberResourceRead(d, meta)
}

func groupMemberResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	id, err := helper.ParseGroupMemberId(d.Id())
	if err != nil {
		return fmt.Errorf("unable to parse ID: %v", err)
	}

	members, _, err := client.ListMembers(ctx, id.GroupId)
	if err != nil {
		return fmt.Errorf("retrieving Group members (group object ID: %q): %+v", id.GroupId, err)
	}

	var memberObjectID string
	for _, objectID := range *members {
		if objectID == id.MemberId {
			memberObjectID = objectID
		}
	}

	if memberObjectID == "" {
		d.SetId("")
		return nil
	}

	d.Set("group_object_id", id.GroupId)
	d.Set("member_object_id", memberObjectID)

	return nil
}

func groupMemberResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	id, err := helper.ParseGroupMemberId(d.Id())
	if err != nil {
		return fmt.Errorf("Unable to parse ID: %v", err)
	}

	tf.LockByName(groupMemberResourceName, id.GroupId)
	defer tf.UnlockByName(groupMemberResourceName, id.GroupId)

	if _, err := client.RemoveMembers(ctx, id.GroupId, &[]string{id.MemberId}); err != nil {
		return fmt.Errorf("removing Member (member object ID: %q) from Group (group object ID: %q): %+v", id.MemberId, id.GroupId, err)
	}

	return nil
}
