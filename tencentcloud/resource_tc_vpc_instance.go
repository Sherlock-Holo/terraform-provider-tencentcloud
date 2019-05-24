package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTencentCloudVpcInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcInstanceCreate,
		Read:   resourceTencentCloudVpcInstanceRead,
		Update: resourceTencentCloudVpcInstanceUpdate,
		Delete: resourceTencentCloudVpcInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringLengthInRange(1, 60),
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},

			"dns_servers": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_multicast": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// Computed values
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceTencentCloudVpcInstanceCreate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_vpc.create")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		name        string = ""
		cidrBlock   string = ""
		dnsServers         = make([]string, 0, 4)
		isMulticast bool   = true
	)
	if temp, ok := d.GetOk("name"); ok {
		name = temp.(string)
	}
	if temp, ok := d.GetOk("cidr_block"); ok {
		cidrBlock = temp.(string)
	}
	if temp, ok := d.GetOk("dns_servers"); ok {

		slice := temp.([]interface{})
		dnsServers = make([]string, 0, len(slice))
		for _, v := range slice {
			dnsServers = append(dnsServers, v.(string))
		}
		if len(dnsServers) < 1 {
			return fmt.Errorf("If dns_servers is set, then len(dns_servers) should be [1:4]")
		}
		if len(dnsServers) > 4 {
			return fmt.Errorf("If dns_servers is set, then len(dns_servers) should be [1:4]")
		}

	}
	if temp, ok := d.GetOk("is_multicast"); ok {
		isMulticast = temp.(bool)
	}

	vpcId, _, err := service.CreateVpc(ctx, name, cidrBlock, isMulticast, dnsServers)
	if err != nil {
		return err
	}
	d.SetId(vpcId)
	return resourceTencentCloudVpcInstanceRead(d, meta)
}

func resourceTencentCloudVpcInstanceRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_vpc.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	info, has, err := service.DescribeVpc(ctx, d.Id())
	if err != nil {
		return err
	}
	//deleted
	if has == 0 {
		d.SetId("")
		return nil
	}
	if has != 1 {
		errRet := fmt.Errorf("one vpc_id read get %d vpc info", has)
		log.Printf("[CRITAL]%s %s", logId, errRet.Error())
		return errRet
	}
	d.Set("name", info.name)
	d.Set("cidr_block", info.cidr)
	d.Set("dns_servers", info.dnsServers)
	d.Set("is_multicast", info.isMulticast)
	d.Set("create_time", info.createTime)
	d.Set("is_default", info.isDefault)
	return nil
}

func resourceTencentCloudVpcInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_vpc.update")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		name        string = ""
		dnsServers         = make([]string, 0, 4)
		slice              = make([]interface{}, 0, 4)
		isMulticast bool   = true
	)
	old, now := d.GetChange("name")
	if d.HasChange("name") {
		name = now.(string)
	} else {
		name = old.(string)
	}

	old, now = d.GetChange("dns_servers")

	if d.HasChange("dns_servers") {
		slice = now.([]interface{})
		if len(slice) < 1 {
			return fmt.Errorf("If dns_servers is set, then len(dns_servers) should be [1:4]")
		}
		if len(slice) > 4 {
			return fmt.Errorf("If dns_servers is set, then len(dns_servers) should be [1:4]")
		}
	} else {
		slice = old.([]interface{})
	}
	if len(slice) > 0 {
		for _, v := range slice {
			dnsServers = append(dnsServers, v.(string))
		}
	}

	old, now = d.GetChange("is_multicast")
	if d.HasChange("is_multicast") {
		isMulticast = now.(bool)
	} else {
		isMulticast = old.(bool)
	}

	if err := service.ModifyVpcAttribute(ctx, d.Id(), name, isMulticast, dnsServers); err != nil {
		return err
	}

	return resourceTencentCloudVpcInstanceRead(d, meta)
}

func resourceTencentCloudVpcInstanceDelete(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_vpc.delete")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	return service.DeleteVpc(ctx, d.Id())

}
