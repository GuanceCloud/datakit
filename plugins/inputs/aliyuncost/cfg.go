package aliyuncost

import (
	"context"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"
	"golang.org/x/time/rate"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	sampleConfig = `
#[[inputs.aliyuncost]]
#  ## Aliyun Region (required)
#  ## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
#  region_id = 'cn-hangzhou'

#  ## Aliyun Credentials (required)
#  access_key_id = ''
#  access_key_secret = ''

# ## collect interval, will not collect if not set
  account_interval = "24h"

  bill_interval = "1h"
# ##(optional) collect bill by instance
  by_instance = true

  order_interval = "1h"

#  ##history data for last year (optional)
#  collect_history_data = false
`
)

type (
	agent struct {
		AccessKeyID        string           `toml:"access_key_id"`
		AccessKeySecret    string           `toml:"access_key_secret"`
		RegionID           string           `toml:"region_id"`
		AccountInterval    datakit.Duration `toml:"account_interval"`
		BiilInterval       datakit.Duration `toml:"bill_interval"`
		ByInstance         bool             `toml:"by_instance"`
		OrdertInterval     datakit.Duration `toml:"order_interval"`
		CollectHistoryData bool             `toml:"collect_history_data "`

		client *bssopenapi.Client

		subModules []subModule

		rateLimiter *rate.Limiter

		ctx       context.Context
		cancelFun context.CancelFunc

		accountName string
		accountID   string

		mode string

		testError error
	}

	subModule interface {
		getInterval() time.Duration
		getName() string
		run(context.Context)
	}
)

func (a *agent) isTest() bool {
	return a.mode == "test"
}

func (a *agent) isDebug() bool {
	return a.mode == "debug"
}

func unixTimeStr(t time.Time) string {
	s := t.Format(`2006-01-02T15:04:05Z`)
	return s
}
