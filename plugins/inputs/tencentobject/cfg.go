package tencentobject

import (
	"context"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#[[inputs.tencentobject]]
# ## @param - tencent cloud authorization informations - string - required
# region_id = ''
# access_key_id = ''
# access_key_secret = ''

# ## @param - collection interval - string - optional - default: 5m
# interval = '5m'

# ## @param - custom tags - [list of key:value element] - optional
#[inputs.tencentobject.tags]
# key1 = 'val1'

`
)

type objectAgent struct {
	RegionID        string            `toml:"region_id"`
	AccessKeyID     string            `toml:"access_key_id"`
	AccessKeySecret string            `toml:"access_key_secret"`
	Interval        datakit.Duration  `toml:"interval"`
	Tags            map[string]string `toml:"tags,omitempty"`

	Cvm *Cvm `toml:"cvm,omitempty"`
	Cos *Cos `toml:"cos,omitempty"`
	Clb *Clb `toml:"clb,omitempty"`
	Cdb *Cdb `toml:"cdb,omitempty"`
	// Redis *Redis `toml:"redis,omitempty"`
	// Cdn *Cdn `toml:"cdn,omitempty"`
	// Waf *Waf `toml:"waf,omitempty"`
	// Es *Elasticsearch `toml:"elasticsearch,omitempty"`
	// InfluxDB *InfluxDB `toml:"influxdb,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	wg sync.WaitGroup

	subModules []subModule

	mode string

	testError error
}

func (ag *objectAgent) isTest() bool {
	return ag.mode == "test"
}

func (ag *objectAgent) isDebug() bool {
	return ag.mode == "debug"
}

func (ag *objectAgent) addModule(m subModule) {
	ag.subModules = append(ag.subModules, m)
}
