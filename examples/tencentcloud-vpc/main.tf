resource "tencentcloud_vpc" "main" {
  name        = "test_vpc_instance"
  cidr_block  = "10.1.0.0/21"
  dns_servers = ["8.8.8.8","114.114.114.114","119.29.29.29"]
  is_multicast = false
}
