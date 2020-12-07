package aadgraph

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"

	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func UserGetByObjectId(ctx context.Context, client *graphrbac.UsersClient, objectId string) (*graphrbac.User, error) {
	filter := fmt.Sprintf("objectId eq '%s'", objectId)
	resp, err := client.ListComplete(ctx, filter, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response().Response) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing Azure AD Users for filter %q: %+v", filter, err)
	}

	values := resp.Response().Value
	if values == nil {
		return nil, fmt.Errorf("nil values for AD Users matching %q", filter)
	}
	if len(*values) == 0 {
		return nil, nil
	}
	if len(*values) > 2 {
		return nil, fmt.Errorf("found multiple AD Users matching %q", filter)
	}

	user := (*values)[0]
	if user.DisplayName == nil {
		return nil, fmt.Errorf("nil DisplayName for AD Users matching %q", filter)
	}
	if user.ObjectID == nil {
		return nil, fmt.Errorf("nil ObjectID for AD Users matching %q", filter)
	}
	if *user.ObjectID != objectId {
		return nil, fmt.Errorf("objectID for AD Users matching %q does is does not match(%q!=%q)", filter, *user.ObjectID, objectId)
	}

	return &user, nil
}

func UserGetByMailNickname(ctx context.Context, client *graphrbac.UsersClient, mailNickname string) (*graphrbac.User, error) {
	filter := fmt.Sprintf("mailNickname eq '%s'", mailNickname)
	resp, err := client.ListComplete(ctx, filter, "")
	if err != nil {
		return nil, fmt.Errorf("listing Azure AD Users for filter %q: %+v", filter, err)
	}

	values := resp.Response().Value
	if values == nil {
		return nil, fmt.Errorf("nil values for AD Users matching %q", filter)
	}
	if len(*values) == 0 {
		return nil, nil
	}
	if len(*values) > 2 {
		return nil, fmt.Errorf("found multiple AD Users matching %q", filter)
	}

	user := (*values)[0]
	if user.DisplayName == nil {
		return nil, fmt.Errorf("nil DisplayName for AD Users matching %q", filter)
	}

	return &user, nil
}
