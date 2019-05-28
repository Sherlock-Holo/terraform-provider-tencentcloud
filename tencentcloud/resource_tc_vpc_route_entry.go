package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTencentCloudVpcRouteEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcRouteEntryCreate,
		Read:   resourceTencentCloudVpcRouteEntryRead,
		Delete: resourceTencentCloudVpcRouteEntryDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"next_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAllowedStringValue(ALL_GATE_WAY_TYPES),
			},
			"next_hub": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTencentCloudVpcRouteEntryCreate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "resource.tencentcloud_vpc_route_entry.create")()

	ctx := context.WithValue(context.TODO(), "logId", logId)
	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		vpcId                = d.Get("vpc_id").(string)
		description          = d.Get("description").(string)
		routetableId         = d.Get("route_table_id").(string)
		destinationCidrBlock = d.Get("cidr_block").(string)
		nextType             = d.Get("next_type").(string)
		nextHub              = d.Get("next_hub").(string)
	)

	if routetableId == "" || destinationCidrBlock == "" || nextType == "" || nextHub == "" {
		return fmt.Errorf("some needed fields is empty string")
	}

	_, has, err := service.IsRouteTableInVpc(ctx, routetableId, vpcId)
	if err != nil {
		return err
	}
	if has != 1 {
		err = fmt.Errorf("error,routetable [%s]  not found in vpc [%s]", routetableId, vpcId)
		log.Printf("[CRITAL]%s %s", logId, err.Error())
		return err
	}
	entryId, err := service.CreateRoutes(ctx, routetableId, destinationCidrBlock, nextType, nextHub, description)

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d.%s", entryId, routetableId))

	return nil
}

func resourceTencentCloudVpcRouteEntryRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "resource.tencentcloud_vpc_route_entry.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	items := strings.Split(d.Id(), ".")

	if len(items) != 2 {
		return fmt.Errorf("entry id be destroyed, we can not get route table id")
	}

	info, has, err := service.DescribeRouteTable(ctx, items[1])

	if err != nil {
		return err
	}
	if has == 0 {
		d.SetId("")
		return nil
	}

	if has != 1 {
		err = fmt.Errorf("one routetable id get %d routetable infos", has)
		return nil
	}
	for _, v := range info.entryInfos {
		if fmt.Sprintf("%d", v.routeEntryId) == items[0] {
			d.Set("description", v.description)
			d.Set("route_table_id", v.routeEntryId)
			d.Set("cidr_block", v.destinationCidr)
			d.Set("next_type", v.nextType)
			d.Set("next_hub", v.nextBub)
			return nil
		}
	}

	d.SetId("")
	return nil
}
func resourceTencentCloudVpcRouteEntryDelete(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "resource.tencentcloud_vpc_route_entry.delete")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		routetableId         = d.Get("route_table_id").(string)
		destinationCidrBlock = d.Get("cidr_block").(string)
		nextType             = d.Get("next_type").(string)
		nextHub              = d.Get("next_hub").(string)
	)

	return service.DeleteRoutes(ctx, routetableId, destinationCidrBlock, nextType, nextHub)
}
