package tencentcloud

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccTencentCloudVpc_basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("tencentcloud_vpc.foo"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "name", "ci-temp-test"),
				),
			},
			{
				ResourceName:      "tencentcloud_vpc.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTencentCloudVpc_update(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("tencentcloud_vpc.foo"),

					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "name", "ci-temp-test"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "is_multicast", "true"),

					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "is_default"),
					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "create_time"),
					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "dns_servers.#"),
				),
			},
			{
				Config: testAccVpcConfigUpdate,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "name", "ci-temp-test-updated"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "is_multicast", "false"),

					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "is_default"),
					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "create_time"),
					resource.TestCheckResourceAttrSet("tencentcloud_vpc.foo", "dns_servers.#"),
					resource.TestCheckResourceAttr("tencentcloud_vpc.foo", "dns_servers.0", "119.29.29.29"),
				),
			},
		},
	})
}

const testAccVpcConfig = `
resource "tencentcloud_vpc" "foo" {
    name = "ci-temp-test"
    cidr_block = "10.0.0.0/16"
}
`

const testAccVpcConfigUpdate = `
resource "tencentcloud_vpc" "foo" {
    name = "ci-temp-test-updated"
    cidr_block = "10.0.0.0/16"
	dns_servers=["119.29.29.29"]
	is_multicast=false
}
`

func testAccCheckVpcExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		logId := GetLogId(nil)
		ctx := context.WithValue(context.TODO(), "logId", logId)

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("resource %s is not found", r)
		}

		service := VpcService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}
		_, has, err := service.DescribeVpc(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}
		if has > 0 {
			return nil
		}

		return fmt.Errorf("redis not exists.")
	}
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	logId := GetLogId(nil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: testAccProvider.Meta().(*TencentCloudClient).apiV3Conn}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "tencentcloud_vpc" {
			continue
		}
		time.Sleep(5 * time.Second)
		_, has, err := service.DescribeVpc(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}
		if has == 0 {
			return nil
		}
		return fmt.Errorf("redis not delete ok")
	}
	return nil
}
