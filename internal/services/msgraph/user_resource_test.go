package msgraph_test

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func TestAccUser_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_user_msgraph", "test")
	pw := "utils@$$wRd" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUser_basic(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "user_principal_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestUser-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "mail_nickname", fmt.Sprintf("acctestUser.%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "account_enabled", "true"),
				),
			},
			data.ImportStep("password"),
		},
	})
}

func TestAccUser_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_user_msgraph", "test")
	pw := "utils@$$wRd" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUser_complete(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "user_principal_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestUser-%d-Updated", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "mail_nickname", fmt.Sprintf("acctestUser-%d-Updated", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "account_enabled", "false"),
					resource.TestCheckResourceAttr(data.ResourceName, "onpremises_immutable_id", strconv.Itoa(data.RandomInteger)),
				),
			},
			data.ImportStep("password"),
		},
	})
}

func TestAccUser_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_user_msgraph", "test")
	pw1 := "utils@$$wRd" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)
	pw2 := "utils@$$wRd2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUser_basic(data.RandomInteger, pw1),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "user_principal_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestUser-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "mail_nickname", fmt.Sprintf("acctestUser.%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "account_enabled", "true"),
				),
			},
			data.ImportStep("password"),
			{
				Config: testAccUser_complete(data.RandomInteger, pw2),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "user_principal_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "object_id"),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestUser-%d-Updated", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "mail_nickname", fmt.Sprintf("acctestUser-%d-Updated", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "account_enabled", "false"),
					resource.TestCheckResourceAttr(data.ResourceName, "onpremises_immutable_id", strconv.Itoa(data.RandomInteger)),
				),
			},
			data.ImportStep("password"),
		},
	})
}

func TestAccUser_threeUsersABC(t *testing.T) {
	ri := tf.AccRandTimeInt()
	pw := "utils@$$wRd" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUser_threeUsersABC(ri, pw),
				Check: resource.ComposeTestCheckFunc(
					testCheckUserExists("azuread_user_msgraph.testA"),
					testCheckUserExists("azuread_user_msgraph.testB"),
					resource.TestCheckResourceAttrSet("azuread_user_msgraph.testA", "user_principal_name"),
					resource.TestCheckResourceAttr("azuread_user_msgraph.testA", "display_name", fmt.Sprintf("acctestUser-%d-A", ri)),
					resource.TestCheckResourceAttr("azuread_user_msgraph.testA", "mail_nickname", fmt.Sprintf("acctestUser.%d.A", ri)),
					resource.TestCheckResourceAttrSet("azuread_user_msgraph.testB", "user_principal_name"),
					resource.TestCheckResourceAttr("azuread_user_msgraph.testB", "display_name", fmt.Sprintf("acctestUser-%d-B", ri)),
					resource.TestCheckResourceAttr("azuread_user_msgraph.testB", "mail_nickname", fmt.Sprintf("acctestUser-%d-B", ri)),
				),
			},
			{
				ResourceName:            "azuread_user_msgraph.testA",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				ResourceName:            "azuread_user_msgraph.testB",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				ResourceName:            "azuread_user_msgraph.testC",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testCheckUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.UsersClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		_, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return fmt.Errorf("User does not exist: %q", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Unable to get User %q: %+v", rs.Primary.ID, err)
		}

		return nil
	}
}

func testCheckUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_user_msgraph" {
			continue
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.UsersClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		resp, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return nil
			}
			return fmt.Errorf("BAD Get request on Group %q: %+v", rs.Primary.ID, err)
		}

		return fmt.Errorf("User still exists:\n%#v", resp)
	}

	return nil
}

func testAccUser_basic(id int, password string) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

resource "azuread_user_msgraph" "test" {
  user_principal_name = "acctestUser.%[1]d@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestUser-%[1]d"
  password            = "%[2]s"
}
`, id, password)
}

func testAccUser_complete(id int, password string) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

resource "azuread_user_msgraph" "test" {
  user_principal_name       = "acctestUser.%[1]d@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name              = "acctestUser-%[1]d-Updated"
  mail_nickname             = "acctestUser-%[1]d-Updated"
  account_enabled           = false
  password                  = "%[2]s"
  force_password_change     = true
  usage_location            = "NO"
  onpremises_immutable_id   = "%[1]d"
}
`, id, password)
}

func testAccUser_threeUsersABC(id int, password string) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

resource "azuread_user_msgraph" "testA" {
  user_principal_name = "acctestUser.%[1]d.A@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestUser-%[1]d-A"
  password            = "%[2]s"
}

resource "azuread_user_msgraph" "testB" {
  user_principal_name = "acctestUser.%[1]d.B@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestUser-%[1]d-B"
  mail_nickname       = "acctestUser-%[1]d-B"
  password            = "%[2]s"
}

resource "azuread_user_msgraph" "testC" {
  user_principal_name = "acctestUser.%[1]d.C@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestUser-%[1]d-C"
  password            = "%[2]s"
}
`, id, password)
}
