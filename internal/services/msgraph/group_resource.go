package msgraph

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	clients2 "github.com/manicminer/hamilton/clients"
	"github.com/manicminer/hamilton/models"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

func GroupResource() *schema.Resource {
	return &schema.Resource{
		Create: groupResourceCreate,
		Read:   groupResourceRead,
		Update: groupResourceUpdate,
		Delete: groupResourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"object_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"display_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"members": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"owners": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validate.UUID,
				},
			},

			"mail_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"mail_nickname": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"prevent_duplicate_names": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func groupResourceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	displayName := d.Get("display_name").(string)

	if preventDuplicates := d.Get("prevent_duplicate_names").(bool); preventDuplicates {
		if err := groupCheckExistingDisplayName(ctx, client, displayName, nil); err != nil {
			return err
		}
	}

	// default value matches portal behaviour
	mailNickname := uuid.New().String()
	if v, ok := d.GetOk("mail_nickname"); ok {
		mailNickname = v.(string)
	}

	properties := models.Group{
		DisplayName:  utils.String(displayName),
		MailNickname: utils.String(mailNickname),

		// API only supports creation of security groups
		SecurityEnabled: utils.Bool(true),
		MailEnabled:     utils.Bool(false),
	}

	if v, ok := d.GetOk("description"); ok {
		properties.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("members"); ok {
		members := v.(*schema.Set).List()
		for _, o := range members {
			properties.AppendMember(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, o.(string))
		}
	}

	if v, ok := d.GetOk("owners"); ok {
		owners := v.(*schema.Set).List()
		for _, o := range owners {
			properties.AppendOwner(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, o.(string))
		}
	}

	group, _, err := client.Create(ctx, properties)
	if err != nil {
		return fmt.Errorf("creating Group %q: %+v", displayName, err)
	}

	if group.ID == nil {
		return fmt.Errorf("null ID returned for Group %q", displayName)
	}

	d.SetId(*group.ID)

	return groupResourceRead(d, meta)
}

func groupResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	group := models.Group{ID: utils.String(d.Id())}

	displayName := d.Get("display_name").(string)

	if d.HasChange("display_name") {
		if preventDuplicates := d.Get("prevent_duplicate_names").(bool); preventDuplicates {
			if err := groupCheckExistingDisplayName(ctx, client, displayName, group.ID); err != nil {
				return err
			}
		}
		group.DisplayName = utils.String(displayName)
	}

	if d.HasChange("description") {
		group.Description = utils.String(d.Get("description").(string))
	}

	if _, err := client.Update(ctx, group); err != nil {
		return fmt.Errorf("updating Group with ID %q: %+v", d.Id(), err)
	}

	if v, ok := d.GetOkExists("members"); ok && d.HasChange("members") {
		desiredMembers := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())

		members, _, err := client.ListMembers(ctx, *group.ID)
		if err != nil {
			return fmt.Errorf("retrieving members for Group with ID %q: %+v", d.Id(), err)
		}

		existingMembers := *members
		membersForRemoval := utils.Difference(existingMembers, desiredMembers)
		membersToAdd := utils.Difference(desiredMembers, existingMembers)

		if membersForRemoval != nil {
			if _, err = client.RemoveMembers(ctx, d.Id(), &membersForRemoval); err != nil {
				return fmt.Errorf("removing member from Group with ID %q: %+v", d.Id(), err)
			}
		}

		if membersToAdd != nil {
			for _, m := range membersToAdd {
				group.AppendMember(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
			}

			if _, err := client.AddMembers(ctx, &group); err != nil {
				return err
			}
		}
	}

	if v, ok := d.GetOkExists("owners"); ok && d.HasChange("owners") {
		desiredOwners := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())

		owners, _, err := client.ListOwners(ctx, *group.ID)
		if err != nil {
			return fmt.Errorf("retrieving owners for Group with ID %q: %+v", d.Id(), err)
		}

		existingOwners := *owners
		ownersForRemoval := utils.Difference(existingOwners, desiredOwners)
		ownersToAdd := utils.Difference(desiredOwners, existingOwners)

		if ownersForRemoval != nil {
			if _, err = client.RemoveOwners(ctx, d.Id(), &ownersForRemoval); err != nil {
				return fmt.Errorf("removing owner from Group with ID %q: %+v", d.Id(), err)
			}
		}

		if ownersToAdd != nil {
			for _, m := range ownersToAdd {
				group.AppendOwner(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
			}

			if _, err := client.AddOwners(ctx, &group); err != nil {
				return err
			}
		}
	}

	return groupResourceRead(d, meta)
}

func groupResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	group, status, err := client.Get(ctx, d.Id())
	if err != nil {
		if status == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("display_name", group.DisplayName)
	d.Set("object_id", group.ID)
	d.Set("description", group.Description)
	d.Set("mail_enabled", group.MailEnabled)
	d.Set("mail_nickname", group.MailNickname)
	d.Set("security_enabled", group.SecurityEnabled)

	owners, _, err := client.ListOwners(ctx, *group.ID)
	if err != nil {
		return fmt.Errorf("retrieving Owners for Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("owners", owners)

	members, _, err := client.ListMembers(ctx, *group.ID)
	if err != nil {
		return fmt.Errorf("retrieving Owners for Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("members", members)

	if preventDuplicates := d.Get("prevent_duplicate_names").(bool); !preventDuplicates {
		d.Set("prevent_duplicate_names", false)
	}

	return nil
}

func groupResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	if _, err := client.Delete(ctx, d.Id()); err != nil {
		return fmt.Errorf("deleting Group with ID %q: %+v", d.Id(), err)
	}

	return nil
}

func groupCheckExistingDisplayName(ctx context.Context, client *clients2.GroupsClient, displayName string, existingId *string) error {
	filter := fmt.Sprintf("displayName eq '%s'", displayName)
	result, _, err := client.List(ctx, filter)
	if err != nil {
		return fmt.Errorf("unable to list existing groups: %+v", err)
	}

	var existing []models.Group
	for _, r := range *result {
		if existingId != nil && *r.ID == *existingId {
			continue
		}
		if strings.EqualFold(displayName, *r.DisplayName) {
			existing = append(existing, r)
		}
	}
	count := len(existing)
	if count > 0 {
		noun := "group was"
		if count > 1 {
			noun = "groups were"
		}
		return fmt.Errorf("`prevent_duplicate_names` was specified and %d existing %s found with display_name %q", count, noun, displayName)
	}
	return nil
}
