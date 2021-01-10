package huaweiyunobject

import (
	"context"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	sampleConfig = `
#[[inputs.huaweiyunobject]]
# ## @param - huaweiyun authorization informations - string - required
# access_key_id = ''
# access_key_secret = ''
# project_id = ''
# region_id=''

# ## @param - collection interval - string - optional - default: 6h  1h <= interval <= 24h
# interval = '6h'

# ## @param - custom tags - [list of key:value element] - optional
#[inputs.huaweiyunobject.tags]
# key1 = 'val1'

`
)

type objectAgent struct {
	// EndPoint        string `toml:"endpoint"`
	RegionID        string `toml:"region_id"`
	ProjectID       string `toml:"project_id"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`

	Interval datakit.Duration  `toml:"interval"`
	Tags     map[string]string `toml:"tags,omitempty"`

	Ecs   *Ecs   `toml:"ecs,omitempty"`
	Elb   *Elb   `toml:"elb,omitempty"`
	Obs   *Obs   `toml:"obs,omitempty"`
	Mysql *Mysql `toml:"mysql,omitempty"`

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
