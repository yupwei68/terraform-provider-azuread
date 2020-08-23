package msgraph

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Optional: true,
				Default:  false,
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
				Optional: true,
				Default:  true,
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

	name := d.Get("display_name").(string)

	// default value matches portal behaviour
	mailNickname := uuid.New().String()
	if v, ok := d.GetOk("mail_nickname"); ok {
		mailNickname = v.(string)
	}

	properties := models.Group{
		DisplayName:     utils.String(name),
		MailEnabled:     utils.Bool(d.Get("mail_enabled").(bool)),
		MailNickname:    utils.String(mailNickname),
		SecurityEnabled: utils.Bool(d.Get("security_enabled").(bool)),
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

	group, err := client.Create(ctx, properties)
	if err != nil {
		return err
	}

	if group.ID == nil {
		return fmt.Errorf("nil Group ID for %q", name)
	}

	d.SetId(*group.ID)

	return groupResourceRead(d, meta)
}

func groupResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	group, err := client.Get(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("retrieving Group with ID %q: %+v", d.Id(), err)
	}

	group.DisplayName = utils.String(d.Get("display_name").(string))
	group.MailEnabled = utils.Bool(d.Get("mail_enabled").(bool))
	group.SecurityEnabled = utils.Bool(d.Get("security_enabled").(bool))

	group.Description = nil
	if v, ok := d.GetOk("description"); ok {
		group.Description = utils.String(v.(string))
	}

	if err := client.Update(ctx, group); err != nil {
		return fmt.Errorf("updating Group with ID %q: %+v", d.Id(), err)
	}

	if v, ok := d.GetOkExists("members"); ok && d.HasChange("members") {
		desiredMembers := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())

		members, err := client.ListMembers(ctx, *group.ID)
		if err != nil {
			return fmt.Errorf("retrieving members for Group with ID %q: %+v", d.Id(), err)
		}

		existingMembers := *members
		membersForRemoval := utils.Difference(existingMembers, desiredMembers)
		membersToAdd := utils.Difference(desiredMembers, existingMembers)

		if membersForRemoval != nil {
			if err = client.RemoveMembers(ctx, d.Id(), &membersForRemoval); err != nil {
				return fmt.Errorf("removing member from Group with ID %q: %+v", d.Id(), err)
			}
		}

		if membersToAdd != nil {
			for _, m := range membersToAdd {
				group.AppendMember(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
			}

			if err := client.AddMembers(ctx, group); err != nil {
				return err
			}
		}
	}

	if v, ok := d.GetOkExists("owners"); ok && d.HasChange("owners") {
		desiredOwners := *tf.ExpandStringSlicePtr(v.(*schema.Set).List())

		owners, err := client.ListOwners(ctx, *group.ID)
		if err != nil {
			return fmt.Errorf("retrieving owners for Group with ID %q: %+v", d.Id(), err)
		}

		existingOwners := *owners
		ownersForRemoval := utils.Difference(existingOwners, desiredOwners)
		ownersToAdd := utils.Difference(desiredOwners, existingOwners)

		if ownersForRemoval != nil {
			if err = client.RemoveOwners(ctx, d.Id(), &ownersForRemoval); err != nil {
				return fmt.Errorf("removing owner from Group with ID %q: %+v", d.Id(), err)
			}
		}

		if ownersToAdd != nil {
			for _, m := range ownersToAdd {
				group.AppendOwner(client.BaseClient.Endpoint, client.BaseClient.ApiVersion, m)
			}

			if err := client.AddOwners(ctx, group); err != nil {
				return err
			}
		}
	}

	return groupResourceRead(d, meta)
}

func groupResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.AadClient).MsGraph.GroupsClient
	ctx := meta.(*clients.AadClient).StopContext

	group, err := client.Get(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("retrieving Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("display_name", group.DisplayName)
	d.Set("object_id", group.ID)
	d.Set("description", group.Description)
	d.Set("mail_enabled", group.MailEnabled)
	d.Set("mail_nickname", group.MailNickname)
	d.Set("security_enabled", group.SecurityEnabled)

	owners, err := client.ListOwners(ctx, *group.ID)
	if err != nil {
		return fmt.Errorf("retrieving Owners for Group with ID %q: %+v", d.Id(), err)
	}

	d.Set("owners", owners)

	members, err := client.ListMembers(ctx, *group.ID)
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

	if err := client.Delete(ctx, d.Id()); err != nil {
		return fmt.Errorf("deleting Group with ID %q: %+v", d.Id(), err)
	}

	return nil
}
