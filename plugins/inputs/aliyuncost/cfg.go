package aliyuncost

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	aliyuncostConfigSample = `
#[[inputs.aliyuncost]]
#  ## Aliyun Region (required)
#  ## See: https://www.alibabacloud.com/help/zh/doc-detail/40654.htm
#  region_id = 'cn-hangzhou'

#  ## Aliyun Credentials (required)
#  access_key_id = ''
#  access_key_secret = ''

#  account_interval = "24h"
#  bill_interval = "1h"
#  order_interval = "1h"

#  ##history data for last year
#  collect_history_data = false
`
)

// var (
// 	Cfg AliyunBoaCfg
// )

type (
	CostCfg struct {
		AccessKeyID        string           `toml:"access_key_id"`
		AccessKeySecret    string           `toml:"access_key_secret"`
		RegionID           string           `toml:"region_id"`
		AccountInterval    datakit.Duration `toml:"account_interval"`
		BiilInterval       datakit.Duration `toml:"bill_interval"`
		OrdertInterval     datakit.Duration `toml:"order_interval"`
		CollectHistoryData bool             `toml:"collect_history_data "`

		ctx       context.Context
		cancelFun context.CancelFunc
	}
)

func unixTimeStr(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02T15:04:05Z`)
	return s
}
