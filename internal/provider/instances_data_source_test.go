// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInstancesDataSource(t *testing.T) {
	// Pre-existing "golden" instance, used exclusively for acceptance testing.
	// User: robot$acotten.unikraft.io.users.kraftcloud
	const tUUID = "3ce45bbf-5921-4590-ba1f-611da83871a0"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "kraftcloud_instances" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kraftcloud_instances.test", "uuids.#", "1"),
					resource.TestCheckResourceAttr("data.kraftcloud_instances.test", "uuids.0", tUUID),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccInstancesDataSourceConfig(`["stopped", "stopping"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kraftcloud_instances.test", "uuids.#", "1"),
					resource.TestCheckResourceAttr("data.kraftcloud_instances.test", "uuids.0", tUUID),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccInstancesDataSourceConfig(`["running", "starting"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kraftcloud_instances.test", "uuids.#", "0"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig(statesAttr string) string {
	return fmt.Sprintf(`
data "kraftcloud_instances" "test" {
  states = %[1]s
}
`, statesAttr)
}
