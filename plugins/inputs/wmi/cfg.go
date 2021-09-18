// +build windows

package wmi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	inputName = `wmi`

	sampleConfig = `
#[[inputs.wmi]]

# ##(optional) custom measurement name
# metric_name = 'WMI'

# ##(optional) global collect interval, default is 5min
# interval = '5m'

# ##(optional) custom tags
#[inputs.wmi.tags]
#key1 = "val1"

#[[inputs.wmi.query]]
#	##(required) the name of the WMI class. see: https://docs.microsoft.com/en-us/previous-versions//aa394084(v=vs.85)?redirectedfrom=MSDN
#	class = 'Win32_LogicalDisk'

#	##(optional) collect interval of this class，use global interval if not set
#	interval='1m'

#	##(required) property names of wmi class, you can optinally specify alias as field name.
#	metrics = [
#		['DeviceID'],
#		['FileSystem', 'disk_filesystem']
#	]

#[[inputs.wmi.query]]
#	class = 'Win32_OperatingSystem'
#	metrics = [
#		['NumberOfProcesses', 'system_proc_count'],
#		['NumberOfUsers']
#	]
`
)

type (
	ClassQuery struct {
		Class    string
		Interval datakit.Duration
		Metrics  [][]string

		lastTime time.Time
	}

	Instance struct {
		MetricName string `toml:"metric_name"`
		Interval   datakit.Duration
		Tags       map[string]string `toml:"tags"`
		Queries    []*ClassQuery     `toml:"query"`

		ctx       context.Context
		cancelFun context.CancelFunc

		mode string

		testError error
	}
)

func (ag *Instance) isTest() bool {
	return ag.mode == "test"
}

func (ag *Instance) isDebug() bool {
	return ag.mode == "debug"
}

func (c *ClassQuery) ToSql() (string, error) {
	sql := "SELECT "

	if len(c.Metrics) == 0 {
		// sql += "*"
		return "", fmt.Errorf("no metric found in class %s", c.Class)
	} else {
		fields := []string{}
		for _, ms := range c.Metrics {
			if len(ms) == 0 || ms[0] == "" {
				return "", fmt.Errorf("metric name cannot be empty in class %s", c.Class)
			}
			fields = append(fields, ms[0])
		}
		sql += strings.Join(fields, ",")
	}
	sql += " FROM " + c.Class

	return sql, nil
}
