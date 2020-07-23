package aliyunobjectecs

import (
	"context"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

const (
	sampleConfig = `
#[[ECSObject]]
# ## @param - aliyun authorization information - string - required
# region_id = ''
# access_key_id = ''
# access_key_secret = ''

# ## @param - collection interval - string - optional - default: 5m
# interval = '5m'
	
# ## @param - custom tags - [list of key:value element] - optional
#[ECSObject.tags]
# key1 = 'val1'
`
)

type AliyunCfg struct {
	RegionID        string            `toml:"region_id"`
	AccessKeyID     string            `toml:"access_key_id"`
	AccessKeySecret string            `toml:"access_key_secret"`
	Interval        internal.Duration `toml:"interval"`
	Tags            map[string]string `toml:"tags"`
}

type objectAgent struct {
	ECSObject []*AliyunCfg `toml:"ECSObject"`

	ctx       context.Context
	cancelFun context.CancelFunc

	wg sync.WaitGroup
}
