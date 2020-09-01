package msgraph

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/manicminer/hamilton/models"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
	"github.com/terraform-providers/terraform-provider-azuread/internal/utils"
)

func SchemaOptionalClaims() *schema.Schema {
	return &schema.Schema{
		Type:       schema.TypeList,
		Optional:   true,
		ConfigMode: schema.SchemaConfigModeAttr,
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

func expandApplicationApi(input []interface{}) *models.ApplicationApi {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	result := models.ApplicationApi{}
	web := input[0].(map[string]interface{})

	result.AcceptMappedClaims = utils.Bool(web["accept_mapped_claims"].(bool))
	result.KnownClientApplications = tf.ExpandStringSlicePtr(web["known_client_applications"].(*schema.Set).List())
	result.OAuth2PermissionScopes = expandApplicationOAuth2Permissions(web["oauth2_permission_scope"])

	if web["requested_access_token_version"] != nil && web["requested_access_token_version"].(int) > 0 {
		result.RequestedAccessTokenVersion = utils.Int32(int32(web["requested_access_token_version"].(int)))
	}

	return &result
}

func expandApplicationAppRoles(input []interface{}) *[]models.ApplicationAppRole {
	if len(input) == 0 {
		return nil
	}

	result := make([]models.ApplicationAppRole, 0, len(input))
	for _, appRoleRaw := range input {
		appRole := appRoleRaw.(map[string]interface{})

		var allowedMemberTypes []string
		for _, allowedMemberType := range appRole["allowed_member_types"].(*schema.Set).List() {
			allowedMemberTypes = append(allowedMemberTypes, allowedMemberType.(string))
		}

		newAppRole := models.ApplicationAppRole{
			ID:                 utils.String(uuid.New().String()),
			AllowedMemberTypes: &allowedMemberTypes,
			Description:        utils.String(appRole["description"].(string)),
			DisplayName:        utils.String(appRole["display_name"].(string)),
			IsEnabled:          utils.Bool(appRole["is_enabled"].(bool)),
		}

		if v, ok := appRole["value"]; ok {
			newAppRole.Value = utils.String(v.(string))
		}

		result = append(result, newAppRole)
	}

	return &result
}

func expandApplicationOAuth2Permissions(i interface{}) *[]models.ApplicationApiPermissionScope {
	input := i.(*schema.Set).List()
	result := make([]models.ApplicationApiPermissionScope, 0, len(input))

	for _, raw := range input {
		scope := raw.(map[string]interface{})

		result = append(result, models.ApplicationApiPermissionScope{
			ID:                      utils.String(uuid.New().String()),
			AdminConsentDescription: utils.String(scope["admin_consent_description"].(string)),
			AdminConsentDisplayName: utils.String(scope["admin_consent_display_name"].(string)),
			IsEnabled:               utils.Bool(scope["is_enabled"].(bool)),
			Type:                    utils.String(scope["type"].(string)),
			UserConsentDescription:  utils.String(scope["user_consent_description"].(string)),
			UserConsentDisplayName:  utils.String(scope["user_consent_display_name"].(string)),
			Value:                   utils.String(scope["value"].(string)),
		})
	}
	return &result
}

func expandApplicationOptionalClaims(input []interface{}) *models.ApplicationOptionalClaims {
	emptyVal := []models.ApplicationOptionalClaim{}
	result := models.ApplicationOptionalClaims{
		AccessToken: &emptyVal,
		IdToken:     &emptyVal,
		Saml2Token:  &emptyVal,
	}

	if len(input) == 0 || input[0] == nil {
		return &result
	}

	optionalClaims := input[0].(map[string]interface{})

	result.AccessToken = expandApplicationOptionalClaim(optionalClaims["access_token"].([]interface{}))
	result.IdToken = expandApplicationOptionalClaim(optionalClaims["id_token"].([]interface{}))
	result.Saml2Token = expandApplicationOptionalClaim(optionalClaims["saml2_token"].([]interface{}))

	return &result
}

func expandApplicationOptionalClaim(input []interface{}) *[]models.ApplicationOptionalClaim {
	result := make([]models.ApplicationOptionalClaim, 0, len(input))

	for _, optionalClaimRaw := range input {
		optionalClaim := optionalClaimRaw.(map[string]interface{})

		additionalProps := []string{}

		if props := optionalClaim["additional_properties"]; props != nil {
			for _, prop := range props.([]interface{}) {
				additionalProps = append(additionalProps, prop.(string))
			}
		}

		newClaim := models.ApplicationOptionalClaim{
			Name:                 utils.String(optionalClaim["name"].(string)),
			Essential:            utils.Bool(optionalClaim["essential"].(bool)),
			AdditionalProperties: &additionalProps,
		}

		if source := optionalClaim["source"].(string); source != "" {
			newClaim.Source = &source
		}

		result = append(result, newClaim)
	}

	return &result
}

func expandApplicationPublicClient(input []interface{}) *models.ApplicationPublicClient {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	result := models.ApplicationPublicClient{}
	web := input[0].(map[string]interface{})

	result.RedirectUris = tf.ExpandStringSlicePtr(web["redirect_uris"].(*schema.Set).List())

	return &result
}

func expandApplicationRequiredResourceAccess(input []interface{}) *[]models.ApplicationRequiredResourceAccess {
	result := make([]models.ApplicationRequiredResourceAccess, 0, len(input))

	for _, raw := range input {
		requiredResourceAccess := raw.(map[string]interface{})

		result = append(result, models.ApplicationRequiredResourceAccess{
			ResourceAppId: utils.String(requiredResourceAccess["resource_app_id"].(string)),
			ResourceAccess: expandApplicationResourceAccess(
				requiredResourceAccess["resource_access"].([]interface{}),
			),
		})
	}
	return &result
}

func expandApplicationResourceAccess(input []interface{}) *[]models.ApplicationResourceAccess {
	result := make([]models.ApplicationResourceAccess, 0, len(input))

	for _, resourceAccessRaw := range input {
		resourceAccess := resourceAccessRaw.(map[string]interface{})

		result = append(result, models.ApplicationResourceAccess{
			ID:   utils.String(resourceAccess["id"].(string)),
			Type: utils.String(resourceAccess["type"].(string)),
		})
	}

	return &result
}

func expandApplicationWeb(input []interface{}) *models.ApplicationWeb {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	result := models.ApplicationWeb{}
	web := input[0].(map[string]interface{})

	if web["homepage_url"].(string) == "" {
		result.HomePageUrl = nil
	} else {
		result.HomePageUrl = utils.String(web["homepage_url"].(string))
	}
	if web["logout_url"].(string) == "" {
		result.LogoutUrl = nil
	} else {
		result.LogoutUrl = utils.String(web["logout_url"].(string))
	}
	result.RedirectUris = tf.ExpandStringSlicePtr(web["redirect_uris"].(*schema.Set).List())
	result.ImplicitGrantSettings = &models.ApplicationImplicitGrantSettings{
		EnableAccessTokenIssuance: utils.Bool(web["enable_access_token_issuance"].(bool)),
		EnableIdTokenIssuance:     utils.Bool(web["enable_id_token_issuance"].(bool)),
	}

	return &result
}

func flattenApplicationApi(in *models.ApplicationApi) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	api := map[string]interface{}{}

	if v := in.AcceptMappedClaims; v != nil {
		api["accept_mapped_claims"] = *v
	}
	if v := in.KnownClientApplications; v != nil {
		api["known_client_applications"] = *v
	}
	if v := in.OAuth2PermissionScopes; v != nil {
		api["oauth2_permission_scope"] = flattenApplicationOAuth2Permissions(v)
	}
	if v := in.RequestedAccessTokenVersion; v != nil {
		api["requested_access_token_version"] = *v
	}

	return []interface{}{api}
}

func flattenApplicationAppRoles(in *[]models.ApplicationAppRole) []map[string]interface{} {
	if in == nil {
		return []map[string]interface{}{}
	}

	appRoles := make([]map[string]interface{}, 0, len(*in))
	for _, role := range *in {
		appRole := make(map[string]interface{})
		if v := role.ID; v != nil {
			appRole["id"] = v
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

func flattenApplicationOAuth2Permissions(in *[]models.ApplicationApiPermissionScope) []map[string]interface{} {
	if in == nil {
		return []map[string]interface{}{}
	}

	permissions := make([]map[string]interface{}, 0, len(*in))
	for _, p := range *in {
		scope := make(map[string]interface{})
		if v := p.AdminConsentDescription; v != nil {
			scope["admin_consent_description"] = v
		}
		if v := p.AdminConsentDisplayName; v != nil {
			scope["admin_consent_display_name"] = v
		}
		if v := p.ID; v != nil {
			scope["id"] = v
		}
		if v := p.IsEnabled; v != nil {
			scope["is_enabled"] = *v
		}
		if v := p.Type; v != nil {
			scope["type"] = v
		}
		if v := p.UserConsentDescription; v != nil {
			scope["user_consent_description"] = v
		}
		if v := p.UserConsentDisplayName; v != nil {
			scope["user_consent_display_name"] = v
		}
		if v := p.Value; v != nil {
			scope["value"] = v
		}

		permissions = append(permissions, scope)
	}

	return permissions
}

func flattenApplicationOptionalClaims(in *models.ApplicationOptionalClaims) interface{} {
	var result []map[string]interface{}

	if in == nil {
		return result
	}

	optionalClaims := make(map[string]interface{})
	if claims := flattenApplicationOptionalClaim(in.AccessToken); len(claims) > 0 {
		optionalClaims["access_token"] = claims
	}
	if claims := flattenApplicationOptionalClaim(in.IdToken); len(claims) > 0 {
		optionalClaims["id_token"] = claims
	}
	if claims := flattenApplicationOptionalClaim(in.Saml2Token); len(claims) > 0 {
		optionalClaims["saml2_token"] = claims
	}
	if len(optionalClaims) == 0 {
		return result
	}
	result = append(result, optionalClaims)
	return result
}

func flattenApplicationOptionalClaim(in *[]models.ApplicationOptionalClaim) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	optionalClaims := make([]interface{}, 0, len(*in))
	for _, claim := range *in {
		optionalClaim := make(map[string]interface{})
		if claim.Name != nil {
			optionalClaim["name"] = *claim.Name
		}
		if claim.Source != nil {
			optionalClaim["source"] = *claim.Source
		}
		if claim.Essential != nil {
			optionalClaim["essential"] = *claim.Essential
		}
		additionalProperties := make([]string, 0)
		if props := claim.AdditionalProperties; props != nil {
			for _, prop := range *props {
				additionalProperties = append(additionalProperties, prop)
			}
		}
		optionalClaim["additional_properties"] = additionalProperties
		optionalClaims = append(optionalClaims, optionalClaim)
	}

	return optionalClaims
}

func flattenApplicationPublicClient(in *models.ApplicationPublicClient) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	publicClient := map[string]interface{}{}

	if v := in.RedirectUris; v != nil {
		publicClient["redirect_uris"] = *v
	}

	return []interface{}{publicClient}
}

func flattenApplicationRequiredResourceAccess(in *[]models.ApplicationRequiredResourceAccess) []map[string]interface{} {
	if in == nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0, len(*in))
	for _, requiredResourceAccess := range *in {
		resource := make(map[string]interface{})
		if requiredResourceAccess.ResourceAppId != nil {
			resource["resource_app_id"] = *requiredResourceAccess.ResourceAppId
		}

		resource["resource_access"] = flattenApplicationResourceAccess(requiredResourceAccess.ResourceAccess)

		result = append(result, resource)
	}

	return result
}

func flattenApplicationResourceAccess(in *[]models.ApplicationResourceAccess) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	accesses := make([]interface{}, 0, len(*in))
	for _, resourceAccess := range *in {
		access := make(map[string]interface{})
		if resourceAccess.ID != nil {
			access["id"] = *resourceAccess.ID
		}
		if resourceAccess.Type != nil {
			access["type"] = *resourceAccess.Type
		}
		accesses = append(accesses, access)
	}

	return accesses
}

func flattenApplicationWeb(in *models.ApplicationWeb) []interface{} {
	if in == nil {
		return []interface{}{}
	}

	web := map[string]interface{}{}

	if v := in.HomePageUrl; v != nil {
		web["homepage_url"] = *v
	}
	if v := in.LogoutUrl; v != nil {
		web["logout_url"] = *v
	}
	if v := in.RedirectUris; v != nil {
		web["redirect_uris"] = *v
	}

	if v := in.ImplicitGrantSettings; v != nil {
		if m := v.EnableAccessTokenIssuance; m != nil {
			web["enable_access_token_issuance"] = m
		}
		if m := v.EnableIdTokenIssuance; m != nil {
			web["enable_id_token_issuance"] = m
		}
	}

	return []interface{}{web}
}
