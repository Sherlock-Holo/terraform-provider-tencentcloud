package tencentcloud

import (
	"context"
	"fmt"

	"log"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/connectivity"
)

//VPC basic information
type VpcBasicInfo struct {
	vpcId       string
	name        string
	cidr        string
	isMulticast bool
	isDefault   bool
	dnsServers  []string
	createTime  string
}

//subnet basic information
type VpcSubnetBasicInfo struct {
	vpcId            string
	subnetId         string
	name             string
	cidr             string
	isMulticast      bool
	isDefault        bool
	zone             string
	availableIpCount int64
	createTime       string
}

type VpcService struct {
	client *connectivity.TencentCloudClient
}

/////////common
func (me *VpcService) fillFilter(ins []*vpc.Filter, key, value string) (outs []*vpc.Filter) {
	if ins == nil {
		ins = make([]*vpc.Filter, 0, 2)
	}

	var filter vpc.Filter = vpc.Filter{Name: &key, Values: []*string{&value}}
	ins = append(ins, &filter)
	outs = ins
	return
}

//////////api
func (me *VpcService) CreateVpc(ctx context.Context, name, cidr string,
	isMulticast bool, dnsServers []string) (vpcId string, isDefault bool, errRet error) {

	logId := GetLogId(ctx)
	request := vpc.NewCreateVpcRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	request.VpcName = &name
	request.CidrBlock = &cidr

	var enableMulticast = map[bool]string{true: "true", false: "false"}[isMulticast]
	request.EnableMulticast = &enableMulticast

	if len(dnsServers) > 0 {
		request.DnsServers = make([]*string, 0, len(dnsServers))
		for index, _ := range dnsServers {
			request.DnsServers = append(request.DnsServers, &dnsServers[index])
		}
	}
	response, err := me.client.UseVpcClient().CreateVpc(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
		vpcId, isDefault = *response.Response.Vpc.VpcId, *response.Response.Vpc.IsDefault
		return
	}

	errRet = err
	return
}

func (me *VpcService) DescribeVpc(ctx context.Context, vpcId string) (info VpcBasicInfo, has int, errRet error) {
	infos, err := me.DescribeVpcs(ctx, vpcId, "")
	if err != nil {
		errRet = err
		return
	}
	has = len(infos)
	if has > 0 {
		info = infos[0]
	}
	return
}

func (me *VpcService) DescribeVpcs(ctx context.Context, vpcId, name string) (infos []VpcBasicInfo, errRet error) {
	logId := GetLogId(ctx)
	request := vpc.NewDescribeVpcsRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	infos = make([]VpcBasicInfo, 0, 100)

	var offset = 0
	var limit = 100
	var total = -1
	var hasVpc = map[string]bool{}

	var filters []*vpc.Filter
	if vpcId != "" {
		filters = me.fillFilter(filters, "vpc-id", vpcId)
	}
	if name != "" {
		filters = me.fillFilter(filters, "vpc-name", name)
	}
	if len(filters) > 0 {
		request.Filters = filters
	}

getMoreData:

	if total >= 0 {
		if offset >= total {
			return
		}
	}
	var strLimit = fmt.Sprintf("%d", limit)
	request.Limit = &strLimit

	var strOffset = fmt.Sprintf("%d", offset)
	request.Offset = &strOffset

	response, err := me.client.UseVpcClient().DescribeVpcs(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if total < 0 {
		total = int(*response.Response.TotalCount)
	}

	if len(response.Response.VpcSet) > 0 {
		offset += limit
	} else {
		//get empty Vpcinfo,we're done
		return
	}
	for _, item := range response.Response.VpcSet {
		var basicInfo VpcBasicInfo
		basicInfo.cidr = *item.CidrBlock
		basicInfo.createTime = *item.CreatedTime
		basicInfo.dnsServers = make([]string, 0, len(item.DnsServerSet))

		for _, v := range item.DnsServerSet {
			basicInfo.dnsServers = append(basicInfo.dnsServers, *v)
		}
		basicInfo.isDefault = *item.IsDefault
		basicInfo.isMulticast = *item.EnableMulticast
		basicInfo.name = *item.VpcName
		basicInfo.vpcId = *item.VpcId

		if hasVpc[basicInfo.vpcId] {
			errRet = fmt.Errorf("get repeated vpc_id[%s] when doing DescribeVpcs", basicInfo.vpcId)
			return
		}
		hasVpc[basicInfo.vpcId] = true
		infos = append(infos, basicInfo)
	}
	goto getMoreData

}

func (me *VpcService) DescribeSubnets(ctx context.Context, subnet_id, vpc_id, subnet_name, zone string) (infos []VpcSubnetBasicInfo, errRet error) {

	logId := GetLogId(ctx)
	request := vpc.NewDescribeSubnetsRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	var offset = 0
	var limit = 100
	var total = -1
	var hasSubnet = map[string]bool{}

	var filters []*vpc.Filter
	if subnet_id != "" {
		filters = me.fillFilter(filters, "subnet-id", subnet_id)
	}
	if vpc_id != "" {
		filters = me.fillFilter(filters, "vpc-id", vpc_id)
	}
	if subnet_name != "" {
		filters = me.fillFilter(filters, "subnet-name", subnet_name)
	}
	if zone != "" {
		filters = me.fillFilter(filters, "zone", zone)
	}

	if len(filters) > 0 {
		request.Filters = filters
	}

getMoreData:
	if total >= 0 {
		if offset >= total {
			return
		}
	}
	var strLimit = fmt.Sprintf("%d", limit)
	request.Limit = &strLimit

	var strOffset = fmt.Sprintf("%d", offset)
	request.Offset = &strOffset

	response, err := me.client.UseVpcClient().DescribeSubnets(request)
	if err != nil {
		errRet = err
		return
	}
	log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	if total < 0 {
		total = int(*response.Response.TotalCount)
	}

	if len(response.Response.SubnetSet) > 0 {
		offset += limit
	} else {
		//get empty subnet ,we're done
		return
	}
	for _, item := range response.Response.SubnetSet {
		var basicInfo VpcSubnetBasicInfo

		basicInfo.cidr = *item.CidrBlock
		basicInfo.createTime = *item.CreatedTime
		basicInfo.vpcId = *item.VpcId
		basicInfo.subnetId = *item.SubnetId

		basicInfo.name = *item.SubnetName
		basicInfo.isDefault = *item.IsDefault
		basicInfo.isMulticast = *item.EnableBroadcast

		basicInfo.zone = *item.Zone
		basicInfo.availableIpCount = int64(*item.AvailableIpAddressCount)

		if hasSubnet[basicInfo.subnetId] {
			errRet = fmt.Errorf("get repeated subnetId[%s] when doing DescribeSubnets", basicInfo.subnetId)
			return
		}
		hasSubnet[basicInfo.subnetId] = true
		infos = append(infos, basicInfo)
	}
	goto getMoreData
	return

}

func (me *VpcService) ModifyVpcAttribute(ctx context.Context, vpcId, name string, isMulticast bool, dnsServers []string) (errRet error) {
	logId := GetLogId(ctx)
	request := vpc.NewModifyVpcAttributeRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()

	request.VpcId = &vpcId
	request.VpcName = &name

	if len(dnsServers) > 0 {
		request.DnsServers = make([]*string, 0, len(dnsServers))
		for index, _ := range dnsServers {
			request.DnsServers = append(request.DnsServers, &dnsServers[index])
		}
	}
	var enableMulticast = map[bool]string{true: "true", false: "false"}[isMulticast]
	request.EnableMulticast = &enableMulticast

	response, err := me.client.UseVpcClient().ModifyVpcAttribute(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	errRet = err
	return
}

func (me *VpcService) DeleteVpc(ctx context.Context, vpcId string) (errRet error) {
	logId := GetLogId(ctx)
	request := vpc.NewDeleteVpcRequest()
	defer func() {
		if errRet != nil {
			log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
				logId, request.GetAction(), request.ToJsonString(), errRet.Error())
		}
	}()
	if vpcId == "" {
		errRet = fmt.Errorf("DeleteVpc can not delete emputy vpc_id.")
		return
	}

	request.VpcId = &vpcId

	response, err := me.client.UseVpcClient().DeleteVpc(request)
	if err == nil {
		log.Printf("[DEBUG]%s api[%s] , request body [%s], response body[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())
	}
	errRet = err
	return

}
