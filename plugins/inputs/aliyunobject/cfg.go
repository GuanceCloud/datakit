package aliyunobject

import (
	"context"
	"reflect"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#[[inputs.aliyunobject]]
# ## @param - aliyun authorization informations - string - required
# region_id = ''
# access_key_id = ''
# access_key_secret = ''

# ## @param - collection interval - string - optional - default: 5m
# interval = '5m'

`
)

type objectAgent struct {
	RegionID        string           `toml:"region_id"`
	AccessKeyID     string           `toml:"access_key_id"`
	AccessKeySecret string           `toml:"access_key_secret"`
	Interval        datakit.Duration `toml:"interval"`

	Ecs      *Ecs           `toml:"ecs,omitempty"`
	Slb      *Slb           `toml:"slb,omitempty"`
	Oss      *Oss           `toml:"oss,omitempty"`
	Rds      *Rds           `toml:"rds,omitempty"`
	Ons      *Ons           `toml:"rocketmq,omitempty"`
	Domain   *Domain        `toml:"domain,omitempty"`
	Dds      *Dds           `toml:"mongodb,omitempty"`
	Redis    *Redis         `toml:"redis,omitempty"`
	Cdn      *Cdn           `toml:"cdn,omitempty"`
	Waf      *Waf           `toml:"waf,omitempty"`
	Es       *Elasticsearch `toml:"elasticsearch,omitempty"`
	InfluxDB *InfluxDB      `toml:"influxdb,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	wg sync.WaitGroup

	subModules []subModule

	mode string

	testError error
}

func (ag *objectAgent) addModule(m subModule) {
	if m == nil {
		return
	}
	v := reflect.ValueOf(m)
	if v.IsNil() {
		return
	}
	if m.disabled() {
		return
	}
	ag.subModules = append(ag.subModules, m)
}

func (ag *objectAgent) isTest() bool {
	return ag.mode == "test"
}

func (ag *objectAgent) isDebug() bool {
	return ag.mode == "debug"
}
