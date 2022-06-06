// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sensors collect hardware sensor metrics.
package sensors

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	inputName = "sensors"

	sampleConfig = `
[[inputs.sensors]]
  ## Command path of 'sensor' usually under /usr/bin/sensors
  # path = "/usr/bin/sensors"

  ## Gathering interval
  # interval = "10s"

  ## Command timeout
  # timeout = "3s"

  ## Customer tags, if set will be seen with every metric.
  [inputs.sensors.tags]
    # "key1" = "value1"
    # "key2" = "value2"
`
	l = logger.DefaultSLogger(inputName)
)
