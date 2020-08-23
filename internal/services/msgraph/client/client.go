package client

import (
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/clients"
)

type Client struct {
	GroupsClient *clients.GroupsClient
}

func BuildClient(authorizer auth.Authorizer, tenantId string) *Client {
	return &Client{
		GroupsClient: clients.NewGroupsClient(authorizer, tenantId),
	}
}
