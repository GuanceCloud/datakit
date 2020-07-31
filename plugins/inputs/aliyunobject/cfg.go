package aliyunobject

import (
	"context"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
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

# ## @param - custom tags - [list of key:value element] - optional
#[inputs.aliyunobject.tags]
# key1 = 'val1'

`
)

type objectAgent struct {
	RegionID        string            `toml:"region_id"`
	AccessKeyID     string            `toml:"access_key_id"`
	AccessKeySecret string            `toml:"access_key_secret"`
	Interval        internal.Duration `toml:"interval"`
	Tags            map[string]string `toml:"tags,omitempty"`

	Ecs *Ecs `toml:"ecs,omitempty"`
	Slb *Slb `toml:"slb,omitempty"`
	Oss *Oss `toml:"oss,omitempty"`
	Rds *Rds `toml:"rds,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	wg sync.WaitGroup

	subModules []subModule
}

func (ag *objectAgent) addModule(m subModule) {

	if m == nil {
		return
	}
	ag.subModules = append(ag.subModules, m)
}
