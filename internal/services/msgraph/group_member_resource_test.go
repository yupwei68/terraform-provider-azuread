package msgraph_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azuread/internal/clients"
	"github.com/terraform-providers/terraform-provider-azuread/internal/tf"
)

func TestAccGroupMember_user(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_member_msgraph", "testA")
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMember_oneUser(data.RandomInteger, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(data.ResourceName, "group_object_id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "member_object_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroupMember_multipleUser(t *testing.T) {
	rna := "azuread_group_member_msgraph.testA"
	rnb := "azuread_group_member_msgraph.testB"
	id := tf.AccRandTimeInt()
	pw := "utils@$$wR2" + acctest.RandStringFromCharSet(7, acctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMember_oneUser(id, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rna, "group_object_id"),
					resource.TestCheckResourceAttrSet(rna, "member_object_id"),
				),
			},
			{
				ResourceName:      rna,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupMember_twoUsers(id, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rna, "group_object_id"),
					resource.TestCheckResourceAttrSet(rna, "member_object_id"),
					resource.TestCheckResourceAttrSet(rnb, "group_object_id"),
					resource.TestCheckResourceAttrSet(rnb, "member_object_id"),
				),
			},
			// we rerun the config so the group resource updates with the number of members
			{
				Config: testAccGroupMember_twoUsers(id, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuread_group_msgraph.test", "members.#", "2"),
				),
			},
			{
				ResourceName:      rna,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupMember_oneUser(id, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rna, "group_object_id"),
					resource.TestCheckResourceAttrSet(rna, "member_object_id"),
				),
			},
			{
				Config: testAccGroupMember_oneUser(id, pw),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuread_group_msgraph.test", "members.#", "1"),
				),
			},
		},
	})
}

func TestAccGroupMember_group(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_member_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMember_group(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(data.ResourceName, "group_object_id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "member_object_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccGroupMember_servicePrincipal(t *testing.T) {
	data := acceptance.BuildTestData(t, "azuread_group_member_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMember_servicePrincipal(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(data.ResourceName, "group_object_id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "member_object_id"),
				),
			},
			data.ImportStep(),
		},
	})
}

func testCheckGroupMemberDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_group_member_msgraph" {
			continue
		}

		client := acceptance.AzureADProvider.Meta().(*clients.AadClient).MsGraph.GroupsClient
		ctx := acceptance.AzureADProvider.Meta().(*clients.AadClient).StopContext

		groupID := rs.Primary.Attributes["group_object_id"]
		memberID := rs.Primary.Attributes["member_object_id"]

		// see if group exists
		if _, err := client.Get(ctx, groupID); err != nil {
			continue
		}

		members, err := client.ListMembers(ctx, groupID)
		if err != nil {
			return fmt.Errorf("retrieving Group members (groupObjectId: %q): %+v", groupID, err)
		}

		var memberObjectID string
		for _, objectID := range *members {
			if objectID == memberID {
				memberObjectID = objectID
			}
		}

		if memberObjectID != "" {
			return fmt.Errorf("Group member still exists:\n%#v", memberObjectID)
		}
	}

	return nil
}

func testAccGroupMember_oneUser(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
}

resource "azuread_group_member_msgraph" "testA" {
  group_object_id  = azuread_group_msgraph.test.object_id
  member_object_id = azuread_user_msgraph.testA.object_id
}

`, testAccUser_threeUsersABC(id, password), id)
}

func testAccGroupMember_twoUsers(id int, password string) string {
	return fmt.Sprintf(`
%[1]s

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[2]d"
}

resource "azuread_group_member_msgraph" "testA" {
  group_object_id  = azuread_group_msgraph.test.object_id
  member_object_id = azuread_user_msgraph.testA.object_id
}

resource "azuread_group_member_msgraph" "testB" {
  group_object_id  = azuread_group_msgraph.test.object_id
  member_object_id = azuread_user_msgraph.testB.object_id
}

`, testAccUser_threeUsersABC(id, password), id)
}

func testAccGroupMember_group(id int) string {
	return fmt.Sprintf(`

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[1]d"
}

resource "azuread_group_msgraph" "member" {
  display_name = "acctestGroup-%[1]d-Member"
}

resource "azuread_group_member_msgraph" "test" {
  group_object_id  = azuread_group_msgraph.test.object_id
  member_object_id = azuread_group_msgraph.member.object_id
}

`, id)
}

func testAccGroupMember_servicePrincipal(id int) string {
	return fmt.Sprintf(`

resource "azuread_application" "test" {
  name = "acctestApp-%[1]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azuread_group_msgraph" "test" {
  display_name = "acctestGroup-%[1]d"
}

resource "azuread_group_member_msgraph" "test" {
  group_object_id  = azuread_group_msgraph.test.object_id
  member_object_id = azuread_service_principal.test.object_id
}

`, id)
}
