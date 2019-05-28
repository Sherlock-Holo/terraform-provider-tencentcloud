package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTencentCloudVpcSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcSubnetCreate,
		Read:   resourceTencentCloudVpcSubnetRead,
		Update: resourceTencentCloudVpcSubnetUpdate,
		Delete: resourceTencentCloudVpcSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"is_multicast": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			// Computed values
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"available_ip_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceTencentCloudVpcSubnetCreate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_subnet.create")()

	ctx := context.WithValue(context.TODO(), "logId", logId)
	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		vpcId            string = ""
		availabilityZone string = ""
		name             string = ""
		cidrBlock        string = ""
		isMulticast      bool   = true
		routeTableId     string = ""
	)
	if temp, ok := d.GetOk("vpc_id"); ok {
		vpcId = temp.(string)
		if len(vpcId) < 1 {
			return fmt.Errorf("vpc_id should be not empty string")
		}
	}
	if temp, ok := d.GetOk("availability_zone"); ok {
		availabilityZone = temp.(string)
		if len(availabilityZone) < 1 {
			return fmt.Errorf("availability_zone should be not empty string")
		}
	}
	if temp, ok := d.GetOk("name"); ok {
		name = temp.(string)
	}
	if temp, ok := d.GetOk("cidr_block"); ok {
		cidrBlock = temp.(string)
	}

	isMulticast = d.Get("is_multicast").(bool)

	if temp, ok := d.GetOk("route_table_id"); ok {
		routeTableId = temp.(string)
		if len(routeTableId) < 1 {
			return fmt.Errorf("route_table_id should be not empty string")
		}
	}

	if routeTableId != "" {
		_, has, err := service.IsRouteTableInVpc(ctx, routeTableId, vpcId)
		if err != nil {
			return err
		}
		if has != 1 {
			err = fmt.Errorf("error,routetable [%s]  not found in vpc [%s]", routeTableId, vpcId)
			log.Printf("[CRITAL]%s %s", logId, err.Error())
			return err
		}
	}

	subnetId, err := service.CreateSubnet(ctx, vpcId, name, cidrBlock, availabilityZone)
	if err != nil {
		return err
	}
	d.SetId(subnetId)

	err = service.ModifySubnetAttribute(ctx, subnetId, name, isMulticast)
	if err != nil {
		return err
	}

	if routeTableId != "" {
		err = service.ReplaceRouteTableAssociation(ctx, subnetId, routeTableId)
		if err != nil {
			return err
		}
	}

	return resourceTencentCloudVpcSubnetRead(d, meta)
}
func resourceTencentCloudVpcSubnetRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_subnet.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	info, has, err := service.DescribeSubnet(ctx, d.Id())
	if err != nil {
		return err
	}
	//deleted
	if has == 0 {
		d.SetId("")
		return nil
	}
	if has != 1 {
		errRet := fmt.Errorf("one subnet_id read get %d subnet info", has)
		log.Printf("[CRITAL]%s %s", logId, errRet.Error())
		return errRet
	}

	d.Set("vpc_id", info.vpcId)
	d.Set("availability_zone", info.zone)
	d.Set("name", info.name)
	d.Set("cidr_block", info.cidr)
	d.Set("is_multicast", info.isMulticast)
	d.Set("route_table_id", info.routeTableId)
	d.Set("is_default", info.isDefault)
	d.Set("available_ip_count", info.availableIpCount)
	d.Set("create_time", info.createTime)
	return nil
}
func resourceTencentCloudVpcSubnetUpdate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "resource.tencentcloud_subnet.update")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		name        string = ""
		isMulticast bool   = true
	)
	old, now := d.GetChange("name")
	if d.HasChange("name") {
		name = now.(string)
	} else {
		name = old.(string)
	}

	old, now = d.GetChange("is_multicast")
	if d.HasChange("is_multicast") {
		isMulticast = now.(bool)
	} else {
		isMulticast = old.(bool)
	}

	d.Partial(true)

	if err := service.ModifySubnetAttribute(ctx, d.Id(), name, isMulticast); err != nil {
		return err
	}
	d.SetPartial("name")
	d.SetPartial("is_multicast")

	if d.HasChange("route_table_id") {
		routeTableId := d.Get("route_table_id").(string)
		if len(routeTableId) < 1 {
			return fmt.Errorf("route_table_id should be not empty string")
		}

		_, has, err := service.IsRouteTableInVpc(ctx, routeTableId, d.Get("vpc_id").(string))
		if err != nil {
			return err
		}
		if has != 1 {
			err = fmt.Errorf("error,routetable [%s]  not found in vpc [%s]", routeTableId, d.Get("vpc_id").(string))
			log.Printf("[CRITAL]%s %s", logId, err.Error())
			return err
		}

		if err := service.ReplaceRouteTableAssociation(ctx, d.Id(), routeTableId); err != nil {
			return err
		}
		d.SetPartial("route_table_id")
	}

	return resourceTencentCloudVpcSubnetRead(d, meta)
}

func resourceTencentCloudVpcSubnetDelete(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "resource.tencentcloud_subnet.delete")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	return service.DeleteSubnet(ctx, d.Id())
}
