package msgraph

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "MS Graph"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"MS Graph",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azuread_group_msgraph": GroupData(),
		"azuread_groups_msgraph": GroupsData(),
		"azuread_user_msgraph": UserData(),
		"azuread_users_msgraph": UsersData(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azuread_group_msgraph": GroupResource(),
		"azuread_group_member_msgraph": GroupMemberResource(),
		"azuread_user_msgraph": UserResource(),
	}
}
