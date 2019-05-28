package tencentcloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceTencentCloudVpcSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTencentCloudVpcSubnetsRead,

		Schema: map[string]*schema.Schema{
			"subnet_id": {
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
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_default": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_multicast": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"route_table_id": {
							Type:     schema.TypeString,
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
				},
			},
		},
	}
}
func dataSourceTencentCloudVpcSubnetsRead(d *schema.ResourceData, meta interface{}) error {

	logId := GetLogId(nil)

	defer LogElapsed(logId + "data_source.tencentcloud_vpc_subnets.read")()

	ctx := context.WithValue(context.TODO(), "logId", logId)

	service := VpcService{client: meta.(*TencentCloudClient).apiV3Conn}

	var (
		subnetId string = ""
		name     string = ""
	)
	if temp, ok := d.GetOk("subnet_id"); ok {
		tempStr := temp.(string)
		if tempStr != "" {
			subnetId = tempStr
		}
	}
	if temp, ok := d.GetOk("name"); ok {
		tempStr := temp.(string)
		if tempStr != "" {
			name = tempStr
		}
	}

	var infos, err = service.DescribeSubnets(ctx, subnetId, "", name, "")
	if err != nil {
		return err
	}

	var infoList = make([]map[string]interface{}, 0, len(infos))

	for _, item := range infos {

		var infoMap = make(map[string]interface{})

		infoMap["availability_zone"] = item.zone
		infoMap["vpc_id"] = item.vpcId
		infoMap["subnet_id"] = item.subnetId
		infoMap["name"] = item.name
		infoMap["cidr_block"] = item.cidr
		infoMap["is_default"] = item.isDefault
		infoMap["is_multicast"] = item.isMulticast
		infoMap["route_table_id"] = item.routeTableId
		infoMap["available_ip_count"] = item.availableIpCount
		infoMap["create_time"] = item.createTime

		infoList = append(infoList, infoMap)
	}

	if err := d.Set("instance_list", infoList); err != nil {
		log.Printf("[CRITAL]%s provider set  subnet instances fail, reason:%s\n ", logId, err.Error())
		return err
	}

	d.SetId("vpc_subnet" + subnetId + "_" + name)

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), infoList); err != nil {
			log.Printf("[CRITAL]%s output file[%s] fail, reason[%s]\n",
				logId, output.(string), err.Error())
			return err
		}
	}
	return nil
}
