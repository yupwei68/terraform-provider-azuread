package groups

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/helpers/aadgraph"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/validate"
)

const groupMemberResourceName = "azuread_group_member"

func groupMemberResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: groupMemberResourceCreate,
		ReadContext:   groupMemberResourceRead,
		DeleteContext: groupMemberResourceDelete,

		Importer: tf.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := aadgraph.ParseGroupMemberId(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"group_object_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},

			"member_object_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.UUID,
			},
		},
	}
}

func groupMemberResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if useMsGraph := meta.(*clients.Client).EnableMsGraphBeta; useMsGraph {
		return groupMemberResourceCreateMsGraph(ctx, d, meta)
	}
	return groupMemberResourceCreateAadGraph(ctx, d, meta)
}

func groupMemberResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if useMsGraph := meta.(*clients.Client).EnableMsGraphBeta; useMsGraph {
		return groupMemberResourceReadMsGraph(ctx, d, meta)
	}
	return groupMemberResourceReadAadGraph(ctx, d, meta)
}

func groupMemberResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if useMsGraph := meta.(*clients.Client).EnableMsGraphBeta; useMsGraph {
		return groupMemberResourceDeleteMsGraph(ctx, d, meta)
	}
	return groupMemberResourceDeleteAadGraph(ctx, d, meta)
}
