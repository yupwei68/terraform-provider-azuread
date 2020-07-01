package graph

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"

	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"
)

func UserGetByObjectId(client *graphrbac.UsersClient, ctx context.Context, objectId string) (*graphrbac.User, error) {
	user, err := userGetByFilter(client, ctx, fmt.Sprintf("objectId eq '%s'", objectId))
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}

	if *user.ObjectID != objectId {
		return nil, fmt.Errorf("objectID for returned AD User does is does not match (%q!=%q)", *user.ObjectID, objectId)
	}

	return user, nil
}

func UserGetByMail(client *graphrbac.UsersClient, ctx context.Context, mail string) (*graphrbac.User, error) {
	return userGetByFilter(client, ctx, fmt.Sprintf("mail eq '%s'", mail))
}

func UserGetByMailNickname(client *graphrbac.UsersClient, ctx context.Context, mailNickname string) (*graphrbac.User, error) {
	return userGetByFilter(client, ctx, fmt.Sprintf("mailNickname eq '%s'", mailNickname))
}

func userGetByFilter(client *graphrbac.UsersClient, ctx context.Context, filter string) (*graphrbac.User, error) {
	resp, err := client.ListComplete(ctx, filter, "")
	if err != nil {
		if ar.ResponseWasNotFound(resp.Response().Response) {
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

	return &user, nil
}
