package aadgraph

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func SchemaAppRolesComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"allowed_member_types": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},

				"description": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"display_name": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"is_enabled": {
					Type:     schema.TypeBool,
					Computed: true,
				},

				"value": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func SchemaOauth2PermissionsComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"admin_consent_description": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"admin_consent_display_name": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"id": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"is_enabled": {
					Type:     schema.TypeBool,
					Computed: true,
				},

				"type": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"user_consent_description": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"user_consent_display_name": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"value": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func SchemaOptionalClaims() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},

				"source": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.StringInSlice(
						[]string{"user"},
						false,
					),
				},
				"essential": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"additional_properties": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.StringInSlice(
							[]string{
								"dns_domain_and_sam_account_name",
								"emit_as_roles",
								"netbios_domain_and_sam_account_name",
								"sam_account_name",
							},
							false,
						),
					},
				},
			},
		},
	}
}

func FlattenAppRoles(in *[]graphrbac.AppRole) []map[string]interface{} {
	if in == nil {
		return []map[string]interface{}{}
	}

	appRoles := make([]map[string]interface{}, 0, len(*in))
	for _, role := range *in {
		appRole := map[string]interface{}{
			"id":                   "",
			"allowed_member_types": []interface{}{},
			"description":          "",
			"display_name":         "",
			"is_enabled":           false,
			"value":                "",
		}

		if v := role.ID; v != nil {
			appRole["id"] = *v
		}

		if v := role.AllowedMemberTypes; v != nil {
			memberTypes := make([]interface{}, 0, len(*v))
			for _, m := range *v {
				memberTypes = append(memberTypes, m)
			}
			appRole["allowed_member_types"] = memberTypes
		}

		if v := role.Description; v != nil {
			appRole["description"] = v
		}

		if v := role.DisplayName; v != nil {
			appRole["display_name"] = v
		}

		if v := role.IsEnabled; v != nil {
			appRole["is_enabled"] = v
		}

		if v := role.Value; v != nil {
			appRole["value"] = v
		}

		appRoles = append(appRoles, appRole)
	}

	return appRoles
}

func FlattenOauth2Permissions(in *[]graphrbac.OAuth2Permission) []map[string]interface{} {
	if in == nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0, len(*in))
	for _, p := range *in {
		permission := map[string]interface{}{
			"admin_consent_description":  "",
			"admin_consent_display_name": "",
			"id":                         "",
			"is_enabled":                 false,
			"type":                       "",
			"user_consent_description":   "",
			"user_consent_display_name":  "",
			"value":                      "",
		}

		if v := p.AdminConsentDescription; v != nil {
			permission["admin_consent_description"] = v
		}

		if v := p.AdminConsentDisplayName; v != nil {
			permission["admin_consent_display_name"] = v
		}

		if v := p.ID; v != nil {
			permission["id"] = v
		}

		if v := p.IsEnabled; v != nil {
			permission["is_enabled"] = *v
		}

		if v := p.Type; v != nil {
			permission["type"] = v
		}

		if v := p.UserConsentDescription; v != nil {
			permission["user_consent_description"] = v
		}

		if v := p.UserConsentDisplayName; v != nil {
			permission["user_consent_display_name"] = v
		}

		if v := p.Value; v != nil {
			permission["value"] = v
		}

		result = append(result, permission)
	}

	return result
}

func ApplicationAllOwners(ctx context.Context, client *graphrbac.ApplicationsClient, appId string) ([]string, error) {
	owners, err := client.ListOwnersComplete(ctx, appId)

	if err != nil {
		return nil, fmt.Errorf("listing existing owners for Application with ID %q: %+v", appId, err)
	}

	existingMembers, err := DirectoryObjectListToIDs(ctx, owners)
	if err != nil {
		return nil, fmt.Errorf("getting object IDs of owners for Application with ID %q: %+v", appId, err)
	}

	return existingMembers, nil
}

func ApplicationAddOwner(ctx context.Context, client *graphrbac.ApplicationsClient, appId string, owner string) error {
	ownerGraphURL := fmt.Sprintf("%s/%s/directoryObjects/%s", strings.TrimRight(client.BaseURI, "/"), client.TenantID, owner)

	properties := graphrbac.AddOwnerParameters{
		URL: &ownerGraphURL,
	}

	if _, err := client.AddOwner(ctx, appId, properties); err != nil {
		return fmt.Errorf("adding owner %q to Application with ID %q: %+v", owner, appId, err)
	}

	return nil
}

func ApplicationAddOwners(ctx context.Context, client *graphrbac.ApplicationsClient, appId string, owner []string) error {
	for _, ownerUuid := range owner {
		err := ApplicationAddOwner(ctx, client, appId, ownerUuid)

		if err != nil {
			return fmt.Errorf("adding owners to Application with ID %q: %+v", appId, err)
		}
	}

	return nil
}

func ApplicationFindByName(ctx context.Context, client *graphrbac.ApplicationsClient, name string) (*graphrbac.Application, error) {
	nameFilter := fmt.Sprintf("displayName eq '%s'", name)
	resp, err := client.List(ctx, nameFilter)

	if err != nil {
		return nil, fmt.Errorf("unable to list Applications with filter %q: %+v", nameFilter, err)
	}

	for _, app := range resp.Values() {
		if *app.DisplayName == name {
			return &app, nil
		}
	}

	return nil, nil
}

func ApplicationCheckNameAvailability(ctx context.Context, client *graphrbac.ApplicationsClient, name string) error {
	existingApp, err := ApplicationFindByName(ctx, client, name)
	if err != nil {
		return err
	}
	if existingApp != nil {
		return fmt.Errorf("existing Application with name %q (AppID: %q) was found and `prevent_duplicate_names` was specified", name, *existingApp.AppID)
	}
	return nil
}

type AppRoleId struct {
	ObjectId string
	RoleId   string
}

func (id AppRoleId) String() string {
	return id.ObjectId + "/role/" + id.RoleId
}

func AppRoleIdFrom(objectId, roleId string) AppRoleId {
	return AppRoleId{
		ObjectId: objectId,
		RoleId:   roleId,
	}
}

func ParseAppRoleId(idString string) (*AppRoleId, error) {
	id, err := ParseObjectSubResourceId(idString, "role")
	if err != nil {
		return nil, fmt.Errorf("unable to parse App Role ID: %v", err)
	}

	return &AppRoleId{
		ObjectId: id.objectId,
		RoleId:   id.subId,
	}, nil
}

func AppRoleFindById(app graphrbac.Application, roleId string) (*graphrbac.AppRole, error) {
	if app.AppRoles == nil {
		return nil, nil
	}

	if roleId == "" {
		return nil, errors.New("specified role ID is blank")
	}

	for _, r := range *app.AppRoles {
		if r.ID == nil {
			continue
		}
		if *r.ID == roleId {
			return &r, nil
		}
	}
	return nil, nil
}

func AppRoleAdd(roles *[]graphrbac.AppRole, role *graphrbac.AppRole) (*[]graphrbac.AppRole, error) {
	if role == nil {
		return nil, errors.New("role to be added is null")
	} else if role.ID == nil {
		return nil, errors.New("ID of new role is null")
	}

	cap := 1
	if roles != nil {
		cap += len(*roles)
	}

	newRoles := make([]graphrbac.AppRole, 1, cap)
	newRoles[0] = *role

	for _, v := range *roles {
		if v.ID != nil && *v.ID == *role.ID {
			return nil, &AlreadyExistsError{"App Role", *role.ID}
		}
		newRoles = append(newRoles, v)
	}

	return &newRoles, nil
}

func AppRoleUpdate(roles *[]graphrbac.AppRole, role *graphrbac.AppRole) (*[]graphrbac.AppRole, error) {
	newRoles := make([]graphrbac.AppRole, len(*roles))

	if role.ID == nil {
		return nil, errors.New("ID of role to be updated is null")
	}

	for i, v := range *roles {
		if v.ID == nil {
			continue
		}
		if *v.ID == *role.ID {
			newRoles[i] = *role
			continue
		}
		newRoles[i] = v
	}

	return &newRoles, nil
}

func AppRoleResultDisableById(existing *[]graphrbac.AppRole, roleId string) (*[]graphrbac.AppRole, error) {
	if existing == nil {
		return nil, errors.New("existing roles are null")
	} else if roleId == "" {
		return nil, errors.New("ID of role to be updated is blank")
	}

	newRoles := make([]graphrbac.AppRole, len(*existing))

	for i, v := range *existing {
		if v.ID == nil {
			continue
		}
		if *v.ID == roleId {
			v.IsEnabled = utils.Bool(false)
		}
		newRoles[i] = v
	}

	return &newRoles, nil
}

func AppRoleResultRemoveById(existing *[]graphrbac.AppRole, roleId string) (*[]graphrbac.AppRole, error) {
	if existing == nil {
		return nil, errors.New("existing roles are null")
	} else if roleId == "" {
		return nil, errors.New("ID of role to be disabled is empty")
	}

	newRoles := make([]graphrbac.AppRole, 0)

	for _, v := range *existing {
		if v.ID == nil {
			continue
		}
		if *v.ID == roleId {
			continue
		}
		newRoles = append(newRoles, v)
	}

	return &newRoles, nil
}

func AppRolesSet(ctx context.Context, client *graphrbac.ApplicationsClient, appId string, newRoles *[]graphrbac.AppRole) error {
	// don't support setting nil roles, the sdk ignores them
	// should instead be zero length slice of AppRole
	if newRoles == nil {
		return fmt.Errorf("cannot set nil App Roles for Application with ID %q", appId)
	}

	// Roles must be disabled before they can be edited or removed.
	// Since we cannot match them by ID, we have to disable all the roles, and replace them in one pass.
	app, err := client.Get(ctx, appId)
	if err != nil {
		if utils.ResponseWasNotFound(app.Response) {
			return fmt.Errorf("application with ID %q was not found", appId)
		}

		return fmt.Errorf("retrieving Application with ID %q: %+v", appId, err)
	}

	// don't update if no changes to be made
	if app.AppRoles != nil && reflect.DeepEqual(*app.AppRoles, *newRoles) {
		return nil
	}

	// first disable any existing permissions
	properties := graphrbac.ApplicationUpdateParameters{
		AppRoles: app.AppRoles,
	}

	if properties.AppRoles != nil {
		for _, role := range *properties.AppRoles {
			*role.IsEnabled = false
		}

		if _, err := client.Patch(ctx, appId, properties); err != nil {
			return fmt.Errorf("disabling App Roles for Application with ID %q: %+v", appId, err)
		}
	}

	// then set the new permissions
	properties = graphrbac.ApplicationUpdateParameters{
		AppRoles: newRoles,
	}

	if _, err := client.Patch(ctx, appId, properties); err != nil {
		return fmt.Errorf("setting App Roles for Application with ID %q: %+v", appId, err)
	}

	return nil
}

type OAuth2PermissionId struct {
	ObjectId     string
	PermissionId string
}

func (id OAuth2PermissionId) String() string {
	return id.ObjectId + "/scope/" + id.PermissionId
}

func OAuth2PermissionIdFrom(objectId, permissionId string) OAuth2PermissionId {
	return OAuth2PermissionId{
		ObjectId:     objectId,
		PermissionId: permissionId,
	}
}

func ParseOAuth2PermissionId(idString string) (*OAuth2PermissionId, error) {
	id, err := ParseObjectSubResourceId(idString, "scope")
	if err != nil {
		return nil, fmt.Errorf("unable to parse OAuth2 Permission ID: %v", err)
	}

	return &OAuth2PermissionId{
		ObjectId:     id.objectId,
		PermissionId: id.subId,
	}, nil
}

func OAuth2PermissionFindById(app graphrbac.Application, permissionId string) (*graphrbac.OAuth2Permission, error) {
	if app.Oauth2Permissions == nil {
		return nil, nil
	}

	if permissionId == "" {
		return nil, errors.New("specified permission ID is blank")
	}

	for _, r := range *app.Oauth2Permissions {
		if r.ID == nil {
			continue
		}
		if r.ID != nil && *r.ID == permissionId {
			return &r, nil
		}
	}
	return nil, nil
}

func OAuth2PermissionAdd(permissions *[]graphrbac.OAuth2Permission, permission *graphrbac.OAuth2Permission) (*[]graphrbac.OAuth2Permission, error) {
	if permission == nil {
		return nil, errors.New("permission to be added is null")
	} else if permission.ID == nil {
		return nil, errors.New("ID of new permission is null")
	}

	cap := 1
	if permissions != nil {
		cap += len(*permissions)
	}

	newPermissions := make([]graphrbac.OAuth2Permission, 1, cap)
	newPermissions[0] = *permission

	if permissions != nil {
		for _, v := range *permissions {
			if v.ID != nil && *v.ID == *permission.ID {
				return nil, &AlreadyExistsError{"App Permission", *permission.ID}
			}
			newPermissions = append(newPermissions, v)
		}
	}

	return &newPermissions, nil
}

func OAuth2PermissionUpdate(permissions *[]graphrbac.OAuth2Permission, permission *graphrbac.OAuth2Permission) (*[]graphrbac.OAuth2Permission, error) {
	if permission == nil {
		return nil, errors.New("permission to be added is null")
	} else if permission.ID == nil {
		return nil, errors.New("ID of new permission is null")
	} else if permissions == nil {
		return nil, errors.New("permissions cannot be null when updating")
	}

	newPermissions := make([]graphrbac.OAuth2Permission, len(*permissions))

	for i, v := range *permissions {
		if v.ID == nil {
			continue
		}
		if *v.ID == *permission.ID {
			newPermissions[i] = *permission
			continue
		}
		newPermissions[i] = v
	}

	return &newPermissions, nil
}

func OAuth2PermissionResultDisableById(existing *[]graphrbac.OAuth2Permission, permissionId string) (*[]graphrbac.OAuth2Permission, error) {
	if existing == nil {
		return nil, errors.New("existing permissions are null")
	} else if permissionId == "" {
		return nil, errors.New("ID of permission to be disabled is empty")
	}

	newPermissions := make([]graphrbac.OAuth2Permission, len(*existing))

	for i, v := range *existing {
		if v.ID == nil {
			continue
		}
		if *v.ID == permissionId {
			v.IsEnabled = utils.Bool(false)
		}
		newPermissions[i] = v
	}

	return &newPermissions, nil
}

func OAuth2PermissionResultRemoveById(existing *[]graphrbac.OAuth2Permission, permissionId string) (*[]graphrbac.OAuth2Permission, error) {
	if existing == nil {
		return nil, errors.New("existing permissions are null")
	} else if permissionId == "" {
		return nil, errors.New("ID of permission to be disabled is empty")
	}

	newPermissions := make([]graphrbac.OAuth2Permission, 0)

	for _, v := range *existing {
		if v.ID == nil {
			continue
		}
		if *v.ID == permissionId {
			continue
		}
		newPermissions = append(newPermissions, v)
	}

	return &newPermissions, nil
}

func OAuth2PermissionsSet(ctx context.Context, client *graphrbac.ApplicationsClient, appId string, newPermissions *[]graphrbac.OAuth2Permission) error {
	// don't support setting nil permissions, the sdk ignores them
	// should instead be zero length slice of OAuth2Permission
	if newPermissions == nil {
		return fmt.Errorf("cannot set nil OAuth2 Permissions for Application with ID %q", appId)
	}

	// Permissions must be disabled before they can be edited or removed.
	// Since we cannot match them by ID, we have to disable all the permissions, and replace them in one pass.
	app, err := client.Get(ctx, appId)
	if err != nil {
		if utils.ResponseWasNotFound(app.Response) {
			return fmt.Errorf("application with ID %q was not found", appId)
		}

		return fmt.Errorf("retrieving Application with ID %q: %+v", appId, err)
	}

	// don't update if no changes to be made
	if app.Oauth2Permissions != nil && reflect.DeepEqual(*app.Oauth2Permissions, *newPermissions) {
		return nil
	}

	// first disable any existing permissions
	properties := graphrbac.ApplicationUpdateParameters{
		Oauth2Permissions: app.Oauth2Permissions,
	}

	if properties.Oauth2Permissions != nil {
		for _, permission := range *properties.Oauth2Permissions {
			*permission.IsEnabled = false
		}

		if _, err := client.Patch(ctx, appId, properties); err != nil {
			return fmt.Errorf("disabling OAuth2 Permissions for Application with ID %q: %+v", appId, err)
		}
	}

	// then set the new permissions
	properties = graphrbac.ApplicationUpdateParameters{
		Oauth2Permissions: newPermissions,
	}

	if _, err := client.Patch(ctx, appId, properties); err != nil {
		return fmt.Errorf("setting OAuth2 Permissions for Application with ID %q: %+v", appId, err)
	}

	return nil
}
