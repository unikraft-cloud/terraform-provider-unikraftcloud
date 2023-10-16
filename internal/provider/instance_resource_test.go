// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccInstanceResource(t *testing.T) {
	// "golden" image, used exclusively for acceptance testing.
	const tImg = "unikraft.io/acotten.unikraft.io/tf-acc-nginx/be23de32"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccInstanceResourceConfig(tImg, 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("kraftcloud_instance.test", "uuid"),
					resource.TestCheckResourceAttr("kraftcloud_instance.test", "image", tImg),
					resource.TestCheckResourceAttr("kraftcloud_instance.test", "port", "80"),
					resource.TestCheckResourceAttr("kraftcloud_instance.test", "memory_mb", "128"),    // defaulted
					resource.TestCheckResourceAttr("kraftcloud_instance.test", "internal_port", "80"), // defaulted
				),
			},
			// ImportState testing
			{
				ResourceName:                         "kraftcloud_instance.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateIdFunc:                    instanceUUID("kraftcloud_instance.test"),
				ImportStateVerifyIgnore: []string{
					// not returned for instances in "stopped" state
					"port", "internal_port", "handlers",
					// differs from given value if the image references a tag (vs. a digest)
					"image",
				},
			},
			// Update and Read testing
			{
				Config: testAccInstanceResourceConfig(tImg, 81),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kraftcloud_instance.test", "port", "81"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccInstanceResourceConfig(imageAttr string, portAttr uint16) string {
	return fmt.Sprintf(`
resource "kraftcloud_instance" "test" {
  image = %[1]q
  port  = %[2]d
}
`, imageAttr, portAttr)
}

// instanceUUID returns a resource.ImportStateIdFunc which retrieves the uuid
// attribute of the given "instance" resource from the Terraform state.
// This is used in ImportState tests, because resources of type "instance" must
// be imported using their uuid.
func instanceUUID(resourceName string) resource.ImportStateIdFunc {
	const attrName = "uuid"
	return func(st *terraform.State) (string, error) {
		res, ok := st.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("could not find resource %q in Terraform state", resourceName)
		}
		attr, ok := res.Primary.Attributes[attrName]
		if !ok {
			return "", fmt.Errorf("attribute %q not set on resource %q", attrName, resourceName)
		}
		return attr, nil
	}
}
