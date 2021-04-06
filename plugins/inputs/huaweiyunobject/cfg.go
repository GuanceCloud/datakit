package huaweiyunobject

import (
	"context"

	rms "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	inputName = `huaweiyunobject`

	sampleConfig = `
#[[inputs.huaweiyunobject]]
# ##(required) authorization informations
# access_key_id = ''
# access_key_secret = ''

# ##(optional) collection interval, default is 15m
# interval = '15m'
`

	ecsPipelineConifg = `
json(_,id)
json(_,metadata.imageName,image)
json(_,flavor.name,flavor)
json(_,flavor.vcpus,vcpus)
json(_,hostStatus)
`

	rdsPipelineConfig = `
json(_,id)
json(_,engineName)
json(_,engineVersion)
json(_,flavorCode,flavor)
json(_,volumeType)
json(_,dataVolumeSizeInGBs)
`

	elbPipelineConfig = `
json(_,id)
json(_,vip_address)
json(_,admin_state_up)
json(_,provisioning_status)
`

	vpcPipelineConifg = `
json(_,id)
json(_,cidr)
json(_,status)
`

	evsPipelineConifg = `
json(_,id)
json(_,shareable)
json(_,volumeType)
json(_,size)
json(_,status)
`

	imsPipelineConifg = `
json(_,id)
json(_,platform)
json(_,osVersion)
json(_,imageSize)
json(_,imageType)
json(_,diskFormat)
json(_,status)
`
)

type agent struct {
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`

	RegionID  string `toml:"region_id"`  //deprated
	ProjectID string `toml:"project_id"` //deprated

	Interval datakit.Duration  `toml:"interval"`
	Tags     map[string]string `toml:"tags,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	limiter *rate.Limiter

	apiClient *rms.RmsClient

	mode string

	testError error
}

func (ag *agent) IsDebug() bool {
	return ag.mode == "debug"
}
