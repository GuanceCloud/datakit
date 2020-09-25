package tencentcost

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	inputName    = `tencentcost`
	sampleConfig = `
#[[inputs.tencentcost]]

# ##(required) authentic info
#access_key_id = ''
#access_key_secret = ''

# ##(required) do not less then one miniute
#transaction_interval = '5m'

# ##(required) do not less then one miniute
#order_interval = '5m'

# ##(required) do not less then one miniute
#bill_interval = '5m'

# ##(optional) whether collect the history data of last year
#collect_history_data = false

# ##(optional)
#[inputs.tencentcost.tags]
#key1 = "val1"
`
)

type subModule interface {
	run(context.Context)
	getName() string
}

type TencentCost struct {
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`

	TransactionInterval datakit.Duration `toml:"transaction_interval"`
	OrderInterval       datakit.Duration `toml:"order_interval"`
	BillInterval        datakit.Duration `toml:"bill_interval"`

	CollectHistoryData bool `toml:"collect_history_data "`

	Tags map[string]string `toml:"tags,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc

	subModules []subModule
}
