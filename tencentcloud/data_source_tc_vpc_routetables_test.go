package tencentcloud

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceTencentCloudVpcSubnets_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: TestAccDataSourceTencentCloudVpcRoutetables,

				Check: resource.ComposeTestCheckFunc(
					//id filter
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_vpc_routetables.id_instances"),
					resource.TestCheckResourceAttr("data.tencentcloud_vpc_routetables.id_instances", "instance_list.#", "1"),
					resource.TestCheckResourceAttr("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.name", "ci-temp-test-rt"),

					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.vpc_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.routetable_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.is_default"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.subnet_ids.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.route_entry_infos.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.create_time"),

					//name filter ,Every routable with a "ci-temp-test-rt" name will be found
					testAccCheckTencentCloudDataSourceID("data.tencentcloud_vpc_routetables.name_instances"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.name_instances", "instance_list.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.name"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.vpc_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.routetable_id"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.is_default"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.subnet_ids.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.route_entry_infos.#"),
					resource.TestCheckResourceAttrSet("data.tencentcloud_vpc_routetables.id_instances", "instance_list.0.create_time"),
				),
			},
		},
	})
}

const TestAccDataSourceTencentCloudVpcRoutetables = `
variable "availability_zone" {
	default = "ap-guangzhou-3"
}

resource "tencentcloud_vpc" "foo" {
    name="guagua-ci-temp-test"
    cidr_block="10.0.0.0/16"
}

resource "tencentcloud_route_table" "routetable" {
   vpc_id = "${tencentcloud_vpc.foo.id}"
   name = "ci-temp-test-rt"
}

data "tencentcloud_vpc_routetables" "id_instances" {
	routetable_id="${tencentcloud_route_table.routetable.id}"
}
data "tencentcloud_vpc_routetables" "name_instances" {
	name="${tencentcloud_route_table.routetable.name}"
}

`
