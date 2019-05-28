package tencentcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTencentCloudVpcRouteEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcRouteEntryCreate,
		Read:   resourceTencentCloudVpcRouteEntryRead,
		Delete: resourceTencentCloudVpcRouteEntryDelete,

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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
		description          = ""
		routeTableId         = ""
		destinationCidrBlock = ""
		nextType             = ""
		nextHub              = ""
	)

	if temp, ok := d.GetOk("description"); ok == true {
		description = temp.(string)
	}
	if temp, ok := d.GetOk("route_table_id"); ok == true {
		routeTableId = temp.(string)
	}
	if temp, ok := d.GetOk("destination_cidr_block"); ok == true {
		destinationCidrBlock = temp.(string)
	}
	if temp, ok := d.GetOk("next_type"); ok == true {
		nextType = temp.(string)
	}
	if temp, ok := d.GetOk("next_hub"); ok == true {
		nextHub = temp.(string)
	}

	if routeTableId == "" || destinationCidrBlock == "" || nextType == "" || nextHub == "" {
		return fmt.Errorf("some needed fields is empty string")
	}

	if nextType == GATE_WAY_TYPE_EIP && nextHub != "0" {
		return fmt.Errorf("if next_type is %s, next_hub can only be \"0\" ", GATE_WAY_TYPE_EIP)
	}

	entryId, err := service.CreateRoutes(ctx, routeTableId, destinationCidrBlock, nextType, nextHub, description)

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d.%s", entryId, routeTableId))

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
		err = fmt.Errorf("one routeTable id get %d routeTable infos", has)
		return err
	}
	for _, v := range info.entryInfos {
		if fmt.Sprintf("%d", v.routeEntryId) == items[0] {
			d.Set("description", v.description)
			d.Set("route_table_id", v.routeEntryId)
			d.Set("destination_cidr_block", v.destinationCidr)
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

	items := strings.Split(d.Id(), ".")

	if len(items) != 2 {
		return fmt.Errorf("entry id be destroyed, we can not get route table id")
	}
	routeTableId := items[1]

	entryId, err := strconv.ParseUint(items[0], 10, 64)
	if err != nil {
		return fmt.Errorf("entry id be destroyed, we can not get route entry id")
	}
	return service.DeleteRoutes(ctx, routeTableId, entryId)
}
