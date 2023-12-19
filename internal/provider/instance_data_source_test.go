// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInstanceDataSource(t *testing.T) {
	// Pre-existing "golden" instance, used exclusively for acceptance testing.
	// User: robot$acotten.unikraft.io.users.kraftcloud
	const (
		tUUID     = "3ce45bbf-5921-4590-ba1f-611da83871a0"
		tState    = "stopped"
		tImg      = "acotten.unikraft.io/tf-acc-nginx@sha256:7b2185c79f8f0ff64b1ee29ed5e741342d84e84d53f6429c7155e489f9eb5b28"
		tMem      = "16"
		tCreated  = "2023-10-13T06:12:27Z"
		tFQDN     = "billowing-sunset-scgoixma.fra0.kraft.cloud"
		tPrivIP   = "172.16.26.1"
		tSvcGrp   = "98901e2d-7d4a-4244-a40d-86b7fd3e246c"
		tNetIface = "e900f68f-3651-4fa5-8743-3ecb1acda440"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccInstanceDataSourceConfig(tUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "uuid", tUUID),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "state", tState),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "image", tImg),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "memory_mb", tMem),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "created_at", tCreated),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "fqdn", tFQDN),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "private_ip", tPrivIP),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "service_group.uuid", tSvcGrp),
					resource.TestCheckResourceAttr("data.kraftcloud_instance.test", "network_interfaces.0.uuid", tNetIface),
				),
			},
		},
	})
}

func testAccInstanceDataSourceConfig(uuidAttr string) string {
	return fmt.Sprintf(`
data "kraftcloud_instance" "test" {
  uuid = %[1]q
}
`, uuidAttr)
}
