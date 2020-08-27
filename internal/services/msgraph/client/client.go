package client

import (
	"github.com/manicminer/hamilton/auth"
	"github.com/manicminer/hamilton/clients"
)

type Client struct {
	ApplicationsClient *clients.ApplicationsClient
	DomainsClient      *clients.DomainsClient
	GroupsClient       *clients.GroupsClient
	UsersClient        *clients.UsersClient
}

func BuildClient(authorizer auth.Authorizer, tenantId string) *Client {
	return &Client{
		ApplicationsClient: clients.NewApplicationsClient(authorizer, tenantId),
		DomainsClient:      clients.NewDomainsClient(authorizer, tenantId),
		GroupsClient:       clients.NewGroupsClient(authorizer, tenantId),
		UsersClient:        clients.NewUsersClient(authorizer, tenantId),
	}
}
