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

func TestAccGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_basic(data.RandomInteger),
				Check:  testCheckGroupBasic(data.RandomInteger, "0", "0"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_complete(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "1", "1"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_owners(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithThreeOwners(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "0", "3"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_members(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithThreeMembers(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "3", "0"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_membersAndOwners(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithOwnersAndMembers(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "2", "1"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_membersDiverse(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithDiverseMembers(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "3", "0"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_ownersDiverse(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithDiverseOwners(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "0", "2"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_membersUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_msgraph", "test")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			// Empty group with 0 members
			{
				Config: testAccGroup_basic(data.RandomInteger),
				Check:  testCheckGroupBasic(data.RandomInteger, "0", "0"),
			},
			data.ImportStep(),
			// Group with 1 member
			{
				Config: testAccGroupWithOneMember(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "1", "0"),
			},
			data.ImportStep(),
			// Group with multiple members
			{
				Config: testAccGroupWithThreeMembers(data.RandomInteger, pw),
				Check:  testCheckGroupBasic(data.RandomInteger, "3", "0"),
			},
			data.ImportStep(),
			// Group with a different member
			{
				Config: testAccGroupWithServicePrincipalMember(data.RandomInteger),
				Check:  testCheckGroupBasic(data.RandomInteger, "1", "0"),
			},
			data.ImportStep(),
			// Empty group with 0 members
			{
				Config: testAccGroup_basic(data.RandomInteger),
				Check:  testCheckGroupBasic(data.RandomInteger, "0", "0"),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroup_ownersUpdate(t *testing.T) {
	rn := "azuread_group_msgraph.test"
	id := tf.AccRandTimeInt()
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			// Empty group with 0 owners
			{
				Config: testAccGroup_basic(id),
				Check:  testCheckGroupBasic(id, "0", "0"),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Group with multiple owners
			{
				Config: testAccGroupWithThreeOwners(id, pw),
				Check:  testCheckGroupBasic(id, "0", "3"),
			},
			// Group with 1 owners
			{
				Config: testAccGroupWithOneOwners(id, pw),
				Check:  testCheckGroupBasic(id, "0", "1"),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Group with a different owners
			{
				Config: testAccGroupWithServicePrincipalOwner(id),
				Check:  testCheckGroupBasic(id, "0", "1"),
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Empty group with 0 owners is not possible
		},
	})
}

func TestAccGroup_preventDuplicateNames(t *testing.T) {
	ri := tf.AccRandTimeInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccGroup_duplicateName(ri),
				ExpectError: regexp.MustCompile("[0-9]+ existing groups? (was|were) found with display_name"),
			},
		},
	})
}

func testCheckGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.GroupsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		_, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return fmt.Errorf("Group does not exist: %q", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Could not Get Group: %+v", err)
		}

		return nil
	}
}

func testCheckGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_group_msgraph" {
			continue
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.GroupsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext
		resp, status, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if status == http.StatusNotFound {
				return nil
			}
			return fmt.Errorf("BAD Get request on Group %q: %+v", rs.Primary.ID, err)
		}

		return fmt.Errorf("Group still exists:\n%#v", resp)
	}

	return nil
}

func testCheckGroupBasic(id int, memberCount, ownerCount string) resource.TestCheckFunc {
	resourceName := "azuread_group_msgraph.test"

	return resource.ComposeTestCheckFunc(
		testCheckGroupExists(resourceName),
		resource.TestCheckResourceAttr(resourceName, "display_name", fmt.Sprintf("acctestGroup-%d", id)),
		resource.TestCheckResourceAttrSet(resourceName, "object_id"),
		resource.TestCheckResourceAttr(resourceName, "members.#", memberCount),
		resource.TestCheckResourceAttr(resourceName, "owners.#", ownerCount),
	)
}

func testAccGroup_basic(id int) string {
	return fmt.Sprintf(`
resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%d"
  members      = []
}
`, id)
}

func testAccGroup_complete(id int, password string) string {
	return fmt.Sprintf(`
%s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%d"
  description  = "Please delete me as this is a.test.AD group!"
  members      = [azuread_user_msgraph.test.object_id]
  owners       = [azuread_user_msgraph.test.object_id]
}
`, testAccUser_basic(id, password), id)
}

func testAccDiverseDirectoryObjects(id int, password string) string {
	return fmt.Sprintf(`
data "azuread_domains" "tenant_domain" {
  only_initial = true
}

resource "azuread_application" "test" {
  name = "acctestApp-%[1]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azuread_group_msgraph" "member" {
  display_name = "acctestGroup-%[1]d-Member"
}

resource "azuread_user_msgraph" "test" {
  user_principal_name = "acctestUser.%[1]d@${data.azuread_domains.tenant_domain.domains.0.domain_name}"
  display_name        = "acctestUser-%[1]d"
  password            = "%[2]s"
}
`, id, password)
}

func testAccGroupWithDiverseMembers(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  members      = [azuread_user_msgraph.test.object_id, azuread_group_msgraph.member.object_id, azuread_service_principal.test.object_id]
}
`, testAccDiverseDirectoryObjects(id, password), id)
}

func testAccGroupWithDiverseOwners(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  owners       = [azuread_user_msgraph.test.object_id, azuread_service_principal.test.object_id]
}
`, testAccDiverseDirectoryObjects(id, password), id)
}

func testAccGroupWithOneMember(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  members      = [azuread_user_msgraph.test.object_id]
}
`, testAccUser_basic(id, password), id)
}

func testAccGroupWithOneOwners(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  owners       = [azuread_user_msgraph.test.object_id]
}
`, testAccUser_basic(id, password), id)
}

func testAccGroupWithThreeMembers(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  members      = [azuread_user_msgraph.testA.object_id, azuread_user_msgraph.testB.object_id, azuread_user_msgraph.testC.object_id]
}
`, testAccUser_threeUsersABC(id, password), id)
}

func testAccGroupWithThreeOwners(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  owners       = [azuread_user_msgraph.testA.object_id, azuread_user_msgraph.testB.object_id, azuread_user_msgraph.testC.object_id]
}
`, testAccUser_threeUsersABC(id, password), id)
}

func testAccGroupWithOwnersAndMembers(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
  owners       = [azuread_user_msgraph.testA.object_id]
  members      = [azuread_user_msgraph.testB.object_id, azuread_user_msgraph.testC.object_id]
}
`, testAccUser_threeUsersABC(id, password), id)
}

func testAccGroupWithServicePrincipalMember(id int) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctestApp-%[1]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[1]d"
  members      = [azuread_service_principal.test.object_id]
}
`, id)
}

func testAccGroupWithServicePrincipalOwner(id int) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctestApp-%[1]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[1]d"
  owners       = [azuread_service_principal.test.object_id]
}
`, id)
}

func testAccGroup_duplicateName(id int) string {
	return fmt.Sprintf(`
%s

resource "azuread_group_msgraph" "duplicate" {
  display_name            = azuread_group_msgraph.test.display_name
  prevent_duplicate_names = true
}
`, testAccGroup_basic(id))
}
