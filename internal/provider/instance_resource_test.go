// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				ResourceName:      "kraftcloud_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				// FIXME(antoineco): ImportState still receives req.ID="id-attribute-not-set"
				// despite this option, and despite the attribute having UseStateForUnknown
				// set in the schema.
				ImportStateVerifyIdentifierAttribute: "uuid",
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
