package sensors

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	defCommand   = "sensors"
	defPath      = "/usr/bin/sensors"
	defInterval  = datakit.Duration{Duration: 10 * time.Second}
	defTimeout   = datakit.Duration{Duration: 3 * time.Second}
	inputName    = "sensors"
	sampleConfig = `
[[inputs.sensors]]
	## Command path of 'senssor' usually under /usr/bin/sensors
	# path = "/usr/bin/senssors"

	## Gathering interval
	# interval = "10s"

	## Command timeout
	# timeout = "3s"

	## Customer tags, if set will be seen with every metric.
	[inputs.sensors.tags]
		# "key1" = "value1"
		# "key2" = "value2"
`
	l = logger.SLogger(inputName)
)
