// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

var (
	inputName       = "windows_remote"
	objectInterval  = time.Minute * 5
	loggingInterval = time.Minute * 5
	metricInterval  = time.Second * 60

	l = logger.DefaultSLogger(inputName)

	cpuMeasurement                 = "cpu"
	memoryMeasurement              = "mem"
	systemMeasurement              = "system"
	netMeasurement                 = "net"
	diskMeasurement                = "disk"
	hostprocessesObjectMeasurement = "host_processes"
	hostobjectMeasurement          = "HOST"
)

const sampleCfg = `
[[inputs.windows_remote]]
  ## Network discovery configuration
  ip_list       = [ ]  # e.g. ["127.0.0.1"]
  cidrs         = [ ]  # e.g. ["10.100.1.0/24"]
  scan_interval = "10m"

  ## Set to 'true' to enable election.
  election = true
  ## Maximum number of workers. Default value is calculated as datakit.AvailableCPUs * 2 + 1.
  worker_num = 0

  ## Select mode 'wmi' or 'snmp'
  mode = "wmi"

  ## WMI Collection Module
  [inputs.windows_remote.wmi]
    port   = 135  # Port for WMI (DCOM 135)
    log_enable = true
    ## Authentication configuration
    [inputs.windows_remote.wmi.auth]
      username  = "user"
      password  = "password"

  ## SNMP Collection Module (Independent configuration)
  [inputs.windows_remote.snmp]
    ports     = [ 161 ]  # SNMP ports (default 161)
    community = "datakit"

  [inputs.windows_remote.tags]
  # "some_key" = "some_value"
`
