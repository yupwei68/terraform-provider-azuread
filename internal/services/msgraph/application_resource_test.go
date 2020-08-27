package msgraph_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func TestAccApplication_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")
	//appName := fmt.Sprintf("acctest-APP-%d", data.RandomInteger)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_basic(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					//resource.TestCheckResourceAttr(data.ResourceName, "name", appName),
					//resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://%s", appName)),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_allow_implicit_flow", "false"),
					//resource.TestCheckResourceAttr(data.ResourceName, "type", "webapp/api"),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "1"),
					//resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
					//resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_complete(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					//resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://homepage-%d", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_allow_implicit_flow", "true"),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.0", fmt.Sprintf("http://%d.hashicorptest.com/00000000-0000-0000-0000-00000000", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "group_membership_claims", "All"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.0.access_token.#", "2"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.0.id_token.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "required_resource_access.#", "2"),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "2"),
					//resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
					//resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")
	updatedri := tf.AccRandTimeInt()
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_basic(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					//resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://acctest-APP-%d", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "0"),
					//resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "0"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_complete(updatedri, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					//resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", updatedri)),
					//resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://homepage-%d", updatedri)),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.0", fmt.Sprintf("http://%d.hashicorptest.com/00000000-0000-0000-0000-00000000", updatedri)),
					//resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.3714513888", "http://unittest.hashicorptest.com"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.0.access_token.#", "2"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.0.id_token.#", "1"),
					//resource.TestCheckResourceAttr(data.ResourceName, "required_resource_access.#", "2"),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "2"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_basicEmpty(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					//resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					//resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "0"),
					//resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "0"),
					//resource.TestCheckResourceAttr(data.ResourceName, "optional_claims.#", "0"),
					//resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "0"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_appRolesNoValue(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_appRolesNoValue(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "app_role.#", "1"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_groupMembershipClaimsUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_basic(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_withGroupMembershipClaimsDirectoryRole(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "group_membership_claims", "DirectoryRole"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_withGroupMembershipClaimsSecurityGroup(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "group_membership_claims", "SecurityGroup"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_withGroupMembershipClaimsApplicationGroup(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "group_membership_claims", "ApplicationGroup"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_native(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_native(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "homepage", ""),
					resource.TestCheckResourceAttr(data.ResourceName, "type", "native"),
					resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "0"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_nativeReplyUrls(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_nativeReplyUrls(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "type", "native"),
					resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.3637476042", "urn:ietf:wg:oauth:2.0:oob"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_nativeUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_basic(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://acctest-APP-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "type", "webapp/api"),
					resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "0"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_native(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://acctest-APP-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "type", "native"),
					resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "0"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_complete(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "homepage", fmt.Sprintf("https://homepage-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.#", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "identifier_uris.0", fmt.Sprintf("http://%d.hashicorptest.com/00000000-0000-0000-0000-00000000", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "reply_urls.#", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "required_resource_access.#", "2"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "application_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_oauth2PermissionsUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplication_basic(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "1"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_oauth2Permissions(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "2"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccApplication_oauth2PermissionsEmpty(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					testCheckApplicationExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", fmt.Sprintf("acctest-APP-%[1]d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "oauth2_permissions.#", "0"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccApplication_preventDuplicateNames(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplication_duplicateName(data.RandomInteger),
				ExpectError: regexp.MustCompile("[0-9]+ existing applications? (was|were) found with display_name"),
			},
		},
	})
}

func TestAccApplication_duplicateAppRolesOauth2PermissionsValues(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_application_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplication_duplicateAppRolesOauth2PermissionsValues(data.RandomInteger),
				ExpectError: regexp.MustCompile("validation failed: duplicate app_role / oauth2_permission_scope value found:"),
			},
		},
	})
}

func testCheckApplicationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.ApplicationsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		_, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return fmt.Errorf("Bad: Application %q does not exist", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Get on ApplicationsClient: %+v", err)
		}

		return nil
	}
}

func testCheckApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_application_msgraph" {
			continue
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.ApplicationsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		_, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return nil
			}

			return err
		}

		return fmt.Errorf("Application %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccApplication_basic(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name = "acctest-APP-%[1]d"
  web {
    redirect_uris = ["https://blah", "https://blurh"]
  }
}
`, ri)
}

func testAccApplication_basicEmpty(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name            = "acctest-APP-%[1]d"
  identifier_uris         = []
  group_membership_claims = "None"
  owners = []

  //api {
  //  //oauth2_permission_scope = []
  //}

  optional_claims = []
  required_resource_access = []

  web {
    redirect_uris = []
    homepage_url  = ""
    logout_url    = ""
  }
}
`, ri)
}

func testAccApplication_withGroupMembershipClaimsDirectoryRole(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name                    = "acctest-APP-%[1]d"
  group_membership_claims = "DirectoryRole"
}
`, ri)
}

func testAccApplication_withGroupMembershipClaimsSecurityGroup(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name                    = "acctest-APP-%[1]d"
  group_membership_claims = "SecurityGroup"
}
`, ri)
}

func testAccApplication_withGroupMembershipClaimsApplicationGroup(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name                    = "acctest-APP-%[1]d"
  group_membership_claims = "ApplicationGroup"
}
`, ri)
}

func testAccApplication_complete(ri int, pw string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_application_msgraph" "test" {
  display_name    = "acctest-APP-%[2]d"
  identifier_uris = ["api://%[2]d"]
  group_membership_claims = "All"
  owners = [azuread_user_msgraph.test.object_id]

  api {
    oauth2_permission_scope {
      admin_consent_description  = "Administer the application"
      admin_consent_display_name = "Administer"
      is_enabled                 = true
      type                       = "Admin"
      value                      = "administer"
    }

    oauth2_permission_scope {
      admin_consent_description  = "Allow the application to access acctest-APP-%[2]d on behalf of the signed-in user."
      admin_consent_display_name = "Access acctest-APP-%[2]d"
      is_enabled                 = true
      type                       = "User"
      user_consent_description   = "Allow the application to access acctest-APP-%[2]d on your behalf."
      user_consent_display_name  = "Access acctest-APP-%[2]d"
      value                      = "user_impersonation"
    }
  }

  app_role {
    allowed_member_types = ["User"]
    description          = "Admins can manage roles and perform all task actions"
    display_name         = "Admin"
    is_enabled           = true
    value                = ""
  }

  app_role {
    allowed_member_types = ["User"]
    description          = "ReadOnly roles have limited query access"
    display_name         = "ReadOnly"
    is_enabled           = true
    value                = "User"
  }

  optional_claims {
    access_token {
      name = "myclaim"
    }

    access_token {
      name = "otherclaim"
    }

    id_token {
      name                  = "userclaim"
      source                = "user"
      essential             = true
      additional_properties = ["emit_as_roles"]
    }
  }

  required_resource_access {
    resource_app_id = "00000003-0000-0000-c000-000000000000"

    resource_access {
      id   = "7ab1d382-f21e-4acd-a863-ba3e13f7da61"
      type = "Role"
    }

    resource_access {
      id   = "e1fe6dd8-ba31-4d61-89e7-88639da4683d"
      type = "Scope"
    }

    resource_access {
      id   = "06da0dbc-49e2-44d2-8312-53f166ab848a"
      type = "Scope"
    }
  }

  required_resource_access {
    resource_app_id = "00000002-0000-0000-c000-000000000000"

    resource_access {
      id   = "311a71cc-e848-46a1-bdf8-97ff7156d8e6"
      type = "Scope"
    }
  }

  web {
    homepage_url  = "https://homepage-%[2]d"
    logout_url    = "https://log.me.out"
    redirect_uris = ["https://unittest.hashicorptest.com"]
  }
}
`, testAccUser_basic(ri, pw), ri)
}

func testAccApplication_appRolesNoValue(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name = "acctest-APP-%[1]d"

  app_role {
    allowed_member_types = ["User"]
    description          = "Admins can manage roles and perform all task actions"
    display_name         = "Admin"
    is_enabled           = true
  }
}
`, ri)
}

func testAccApplication_oauth2Permissions(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name = "acctest-APP-%[1]d"

  oauth2_permissions {
    admin_consent_description  = "Allow the application to access acctest-APP-%[1]d on behalf of the signed-in user."
    admin_consent_display_name = "Access acctest-APP-%[1]d"
    is_enabled                 = true
    type                       = "User"
    user_consent_description   = "Allow the application to access acctest-APP-%[1]d on your behalf."
    user_consent_display_name  = "Access acctest-APP-%[1]d"
    value                      = "user_impersonation"
  }

  oauth2_permissions {
    admin_consent_description  = "Administer the application"
    admin_consent_display_name = "Administer"
    is_enabled                 = true
    type                       = "Admin"
    value                      = "administer"
  }
}
`, ri)
}

func testAccApplication_oauth2PermissionsEmpty(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name               = "acctest-APP-%[1]d"
  oauth2_permissions = []
}
`, ri)
}

func testAccApplication_native(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name = "acctest-APP-%[1]d"
  type = "native"
}
`, ri)
}

func testAccApplication_nativeReplyUrls(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name       = "acctest-APP-%[1]d"
  type       = "native"
  reply_urls = ["urn:ietf:wg:oauth:2.0:oob"]
}
`, ri)
}

func testAccApplication_duplicateName(ri int) string {
	return fmt.Sprintf(`
%s

resource "azuread_application_msgraph" "duplicate" {
  display_name            = azuread_application_msgraph.test.display_name
  prevent_duplicate_names = true
}
`, testAccApplication_basic(ri))
}

func testAccApplication_duplicateAppRolesOauth2PermissionsValues(ri int) string {
	return fmt.Sprintf(`
resource "azuread_application_msgraph" "test" {
  display_name = "acctest-APP-%[1]d"

  api {
    oauth2_permission_scope {
      admin_consent_description  = "Administer the application"
      admin_consent_display_name = "Administer"
      is_enabled                 = true
      type                       = "Admin"
      value                      = "administer"
    }
  }

  app_role {
    allowed_member_types = ["User"]
    description          = "Admins can manage roles and perform all task actions"
    display_name         = "Admin"
    is_enabled           = true
    value                = "administer"
  }
}
`, ri)
}
