package msgraph_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/terraform-providers/terraform-provider-azuread/internal/acceptance"
)

func TestAccGroupsDataSource_byDisplayNames(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_groups_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSource_byDisplayNames(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "display_names.#", "2"),
					resource.TestCheckResourceAttr(data.ResourceName, "object_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource_byObjectIds(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_groups_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSource_byObjectIds(data.RandomInteger),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "display_names.#", "2"),
					resource.TestCheckResourceAttr(data.ResourceName, "object_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccGroupsDataSource_noNames(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azuread_groups_msgraph", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSource_noNames(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "displayNames.#", "0"),
					resource.TestCheckResourceAttr(data.ResourceName, "object_ids.#", "0"),
				),
			},
		},
	})
}

func testAccGroup_multiple(id int) string {
	id1 := acctest.RandomWithPrefix(strconv.Itoa(id))
	id2 := acctest.RandomWithPrefix(strconv.Itoa(id))

	return fmt.Sprintf(`
resource "azuread_group_msgraph" "testA" {
  display_name = "acctestGroup-%s"
}

resource "azuread_group_msgraph" "testB" {
  display_name = "acctestGroup-%s"
}
`, id1, id2)
}

func testAccGroupsDataSource_byDisplayNames(id int) string {
	return fmt.Sprintf(`
%s

data "azuread_groups_msgraph" "test" {
  display_names = [azuread_group_msgraph.testA.display_name, azuread_group_msgraph.testB.display_name]
}
`, testAccGroup_multiple(id))
}

func testAccGroupsDataSource_byObjectIds(id int) string {
	return fmt.Sprintf(`
%s

data "azuread_groups_msgraph" "test" {
  object_ids = [azuread_group_msgraph.testA.object_id, azuread_group_msgraph.testB.object_id]
}
`, testAccGroup_multiple(id))
}

func testAccGroupsDataSource_noNames() string {
	return `
data "azuread_groups_msgraph" "test" {
  display_names = []
}
`
}
