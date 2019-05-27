package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceTencentCloudVpcRoutetables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudVpcRoutetablesRead,

		Schema: map[string]*schema.Schema{
			"routetable_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"result_output_file": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "Used to save results.",
			},

			// Computed values
			"instance_list": {Type: schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"routetable_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_ids": {
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
						"route_entry_infos": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"route_entry_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"destination_cidr_block": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"next_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"next_hub": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
func dataSourceTencentCloudVpcRoutetablesRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_routetables.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		routetableId string = ""
		name         string = ""
	)
	if temp, ok := d.GetOk("routetable_id"); ok {
		tempStr := temp.(string)
		if tempStr != "" {
			routetableId = tempStr
		}
	}
	if temp, ok := d.GetOk("name"); ok {
		tempStr := temp.(string)
		if tempStr != "" {
			name = tempStr
		}
	}

	var infos, err = service.DescribeRouteTables(ctx, routetableId, name, "")
	if err != nil {
		return err
	}

	var infoList = make([]map[string]interface{}, 0, len(infos))

	for _, item := range infos {

		routeEntryInfos := make([]map[string]string, len(item.entryInfos))

		for _, v := range item.entryInfos {
			routeEntryInfo := make(map[string]string)
			routeEntryInfo["route_entry_id"] = fmt.Sprintf("%d.%s",
				v.routeEntryId, item.routetableId)
			routeEntryInfo["description"] = v.description
			routeEntryInfo["destination_cidr_block"] = v.destinationCidr
			routeEntryInfo["next_type"] = v.nextType
			routeEntryInfo["next_hub"] = v.nextBub
			routeEntryInfos = append(routeEntryInfos, routeEntryInfo)
		}

		var infoMap = make(map[string]interface{})

		infoMap["routetable_id"] = item.routetableId
		infoMap["name"] = item.name
		infoMap["vpc_id"] = item.vpcId
		infoMap["is_default"] = item.isDefault
		infoMap["subnet_ids"] = item.subnetIds
		infoMap["route_entry_infos"] = routeEntryInfos
		infoMap["create_time"] = item.createTime

		infoList = append(infoList, infoMap)
	}

	if err := d.Set("instance_list", infoList); err != nil {
		log.Printf("[CRITAL]%s provider set  routetable instances fail, reason:%s\n ", logId, err.Error())
		return err
	}

	d.SetId("vpc_routetable" + routetableId + "_" + name)

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), infoList); err != nil {
			log.Printf("[CRITAL]%s output file[%s] fail, reason[%s]\n",
				logId, output.(string), err.Error())
			return err
		}
	}
	return nil
}
