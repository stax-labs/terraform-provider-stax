package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAccountResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stax_account.test", "name", "one"),
					resource.TestCheckResourceAttr("stax_account.test", "status", "new"),
					resource.TestCheckResourceAttr("stax_account.test", "id", "example-id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "stax_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"status", "name"},
			},
			// // Update and Read testing
			// {
			// 	Config: testAccAccountResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("stax_account.test", "name", "two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAccountResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`	
	provider "stax" {
		installation = "dev"
		api_token_access_key = "test_access_key"
		api_token_secret_key = "test_secret_key"
	}
	
	resource "stax_account" "test" {
	name = %[1]q
}
`, configurableAttribute)
}
