package aliyuncms

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type aliyunResource struct {
	RegionID        string
	AccessKeyID     string
	AccessKeySecret string
}

func (r *aliyunResource) checkHaveECS() {

	client, err := ecs.NewClientWithAccessKey(r.RegionID, r.AccessKeyID, r.AccessKeySecret)

	request := ecs.CreateDescribeInstancesRequest()
	request.Scheme = "https"

	response, err := client.DescribeInstances(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n", response)
}
