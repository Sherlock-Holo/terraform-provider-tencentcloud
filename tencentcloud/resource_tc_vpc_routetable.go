package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceTencentCloudVpcRoutetable() *schema.Resource {
	return &schema.Resource{
		Create: resourceTencentCloudVpcRoutetableCreate,
		Read:   resourceTencentCloudVpcRoutetableRead,
		Update: resourceTencentCloudVpcRoutetableUpdate,
		Delete: resourceTencentCloudVpcRoutetableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateStringLengthInRange(1, 60),
			},
			// Computed values
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"route_entry_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
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

func resourceTencentCloudVpcRoutetableCreate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_routetable.create")()

	ctx := context.WithValue(context.TODO(), "logId", logId)
	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		vpcId string = ""
		name  string = ""
	)
	if temp, ok := d.GetOk("vpc_id"); ok {
		vpcId = temp.(string)
		if len(vpcId) < 1 {
			return fmt.Errorf("vpc_id should be not empty string")
		}
	}
	if temp, ok := d.GetOk("name"); ok {
		name = temp.(string)
	}

	routetableId, err := service.CreateRouteTable(ctx, name, vpcId)
	if err != nil {
		return err
	}
	d.SetId(routetableId)

	return resourceTencentCloudVpcRoutetableRead(d, meta)
}
func resourceTencentCloudVpcRoutetableRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_routetable.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	info, has, err := service.DescribeRouteTable(ctx, d.Id())
	if err != nil {
		return err
	}
	//deleted
	if has == 0 {
		d.SetId("")
		return nil
	}
	if has != 1 {
		errRet := fmt.Errorf("one routetable_id read get %d routetable info", has)
		log.Printf("[CRITAL]%s %s", logId, errRet.Error())
		return errRet
	}

	routeEntryIds := make([]string, 0, len(info.entryInfos))
	for _, v := range info.entryInfos {
		tfRouteEntryId := fmt.Sprintf("%d.%s", v.routeEntryId, d.Id())
		routeEntryIds = append(routeEntryIds, tfRouteEntryId)
	}

	d.Set("vpc_id", info.vpcId)
	d.Set("name", info.name)
	d.Set("subnet_ids", info.subnetIds)
	d.Set("route_entry_ids", routeEntryIds)
	d.Set("is_default", info.isDefault)
	d.Set("create_time", info.createTime)
	return nil
}
func resourceTencentCloudVpcRoutetableUpdate(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_routetable.update")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}
	var (
		name string = ""
	)

	if temp, ok := d.GetOk("name"); ok {
		name = temp.(string)
	}

	err := service.ModifyRouteTableAttribute(ctx, d.Id(), name)

	if err != nil {
		return err
	}

	return resourceTencentCloudVpcRoutetableRead(d, meta)
}

func resourceTencentCloudVpcRoutetableDelete(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)
	defer LogElapsed(logId + "resource.tencentcloud_routetable.delete")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	return service.DeleteRouteTable(ctx, d.Id())
}
