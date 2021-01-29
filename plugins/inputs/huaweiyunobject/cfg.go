package huaweiyunobject

import (
	"context"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = `huaweiyunobject`

	sampleConfig = `
#[[inputs.huaweiyunobject]]
# ##(required) authorization informations
# access_key_id = ''
# access_key_secret = ''
# project_id = ''
# region_id=''

# ##(optional) collection interval, default is 6h
# interval = '6h'

# ##(optional) custom tags
#[inputs.huaweiyunobject.tags]
# key1 = 'val1'
# key2 = 'val2'
`
)

type objectAgent struct {
	RegionID        string `toml:"region_id"`
	ProjectID       string `toml:"project_id"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`

	Interval datakit.Duration  `toml:"interval"`
	Tags     map[string]string `toml:"tags,omitempty"`

	Ecs *Ecs `toml:"ecs,omitempty"`
	Elb *Elb `toml:"elb,omitempty"`
	Obs *Obs `toml:"obs,omitempty"`
	Rds *Rds `toml:"rds,omitempty"`
	Vpc *Vpc `toml:"vpc,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	wg sync.WaitGroup

	subModules []subModule

	mode string

	testResult *inputs.TestResult
	testError  error
}

func (ag *objectAgent) addModule(m subModule) {
	ag.subModules = append(ag.subModules, m)
}

func (ag *objectAgent) IsDebug() bool {
	return ag.mode == "debug"
}
