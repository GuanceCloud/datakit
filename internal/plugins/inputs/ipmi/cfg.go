// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

type metricType struct {
	dataName string
	dataType string
}

var metricTypes = []metricType{
	{"status", "int"},
	{"current", "float"},
	{"voltage", "float"},
	{"power", "float"},
	{"temp", "float"},
	{"fan_speed", "int"},
	{"usage", "float"},
	{"count", "int"},
}

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "ipmi"
	metricName  = inputName
	// conf File samples, reflected in the document.
	sampleCfg = `
[[inputs.ipmi]]
  ## If you have so many servers that 10 seconds can't finish the job.
  ## You can start multiple collectors.

  ## (Optional) Collect interval: (defaults to "10s").
  interval = "10s"

  ## Set true to enable election
  election = true

  ## The binPath of ipmitool
  ## (Example) bin_path = "/usr/bin/ipmitool"
  bin_path = "/usr/bin/ipmitool"

  ## (Optional) The envs of LD_LIBRARY_PATH
  ## (Example) envs = [ "LD_LIBRARY_PATH=XXXX:$LD_LIBRARY_PATH" ]

  ## The ips of ipmi servers
  ## (Example) ipmi_servers = ["192.168.1.1"]
  ipmi_servers = ["192.168.1.1"]

  ## The interfaces of ipmi servers: (defaults to []string{"lan"}).
  ## If len(ipmi_users)<len(ipmi_ips), will use ipmi_users[0].
  ## (Example) ipmi_interfaces = ["lanplus"]
  ipmi_interfaces = ["lanplus"]

  ## The users name of ipmi servers: (defaults to []string{}).
  ## If len(ipmi_users)<len(ipmi_ips), will use ipmi_users[0].
  ## (Example) ipmi_users = ["root"]
  ## (Warning!) You'd better use hex_keys, it's more secure.
  ipmi_users = ["root"]

  ## The passwords of ipmi servers: (defaults to []string{}).
  ## If len(ipmi_passwords)<len(ipmi_ips), will use ipmi_passwords[0].
  ## (Example) ipmi_passwords = ["calvin"]
  ## (Warning!) You'd better use hex_keys, it's more secure.
  ipmi_passwords = ["calvin"]

  ## (Optional) Provide the hex key for the IMPI connection: (defaults to []string{}).
  ## If len(hex_keys)<len(ipmi_ips), will use hex_keys[0].
  ## (Example) hex_keys = ["XXXX"]
  # hex_keys = []

  ## (Optional) Schema Version: (defaults to [1]).input.go
  ## If len(metric_versions)<len(ipmi_ips), will use metric_versions[0].
  ## (Example) metric_versions = [2]
  metric_versions = [2]

  ## (Optional) Exec ipmitool timeout: (defaults to "5s").
  timeout = "5s"

  ## (Optional) Ipmi server drop warning delay: (defaults to "300s").
  ## (Example) drop_warning_delay = "300s"
  drop_warning_delay = "300s"

  ## Key words of current.
  ## (Example) regexp_current = ["current"]
  regexp_current = ["current"]

  ## Key words of voltage.
  ## (Example) regexp_voltage = ["voltage"]
  regexp_voltage = ["voltage"]

  ## Key words of power.
  ## (Example) regexp_power = ["pwr","power"]
  regexp_power = ["pwr","power"]

  ## Key words of temp.
  ## (Example) regexp_temp = ["temp"]
  regexp_temp = ["temp"]

  ## Key words of fan speed.
  ## (Example) regexp_fan_speed = ["fan"]
  regexp_fan_speed = ["fan"]

  ## Key words of usage.
  ## (Example) regexp_usage = ["usage"]
  regexp_usage = ["usage"]

  ## Key words of usage.
  ## (Example) regexp_count = []
  # regexp_count = []

  ## Key words of status.
  ## (Example) regexp_status = ["fan"]
  regexp_status = ["fan"]

[inputs.ipmi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

var _ inputs.ElectionInput = (*Input)(nil)

type Input struct {
	Interval datakit.Duration  // run every "Interval" seconds
	Tags     map[string]string // Indicator name
	// collectCache []inputs.Measurement
	collectCache []*point.Point
	platform     string

	BinPath          string           `toml:"bin_path"`           // the file path of "ipmitool"
	Envs             []string         `toml:"envs"`               // exec.Command ENV
	IpmiServers      []string         `toml:"ipmi_servers"`       // the ips of ipmi serverbools
	IpmiInterfaces   []string         `toml:"ipmi_interfaces"`    // the Interfaces of ipmi servers
	IpmiUsers        []string         `toml:"ipmi_users"`         // the users name of ipmi servers
	IpmiPasswords    []string         `toml:"ipmi_passwords"`     // the passwords of ipmi servers
	HexKeys          []string         `toml:"hex_keys"`           // provide the hex key for the IMPI connection.
	MetricVersions   []int            `toml:"metric_versions"`    // Schema Version: (defaults to version 1)
	RegexpCurrent    []string         `toml:"regexp_current"`     // regexp
	RegexpVoltage    []string         `toml:"regexp_voltage"`     // regexp
	RegexpPower      []string         `toml:"regexp_power"`       // regexp
	RegexpTemp       []string         `toml:"regexp_temp"`        // regexp
	RegexpFanSpeed   []string         `toml:"regexp_fan_speed"`   // regexp
	RegexpUsage      []string         `toml:"regexp_usage"`       // regexp
	RegexpCount      []string         `toml:"regexp_count"`       // regexp
	RegexpStatus     []string         `toml:"regexp_status"`      // regexp
	Timeout          datakit.Duration `toml:"timeout"`            // exec timeout
	DropWarningDelay datakit.Duration `toml:"drop_warning_delay"` // server drop warning delay, defaut:300s
	servers          []ipmiServer     // List of active ipmi servers, alarm after service failure
	semStop          *cliutils.Sem    // start stop signal
	Feeder           io.Feeder

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	ipmiDataCh chan dataStruct
}

// The bytes got from ipmi server.
type dataStruct struct {
	server        string
	metricVersion int
	data          []byte
}

// Measurement structure.
type ipmiMeasurement struct {
	name     string                 // Indicator set name
	tags     map[string]string      // Indicator name
	fields   map[string]interface{} // Indicator measurement results
	election bool
}

// LineProto data formatting, submit through FeedMeasurement.
func (m *ipmiMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(m.name, m.tags, m.fields, dkpt.MOptElectionV2(m.election))
}

// Info , reflected in the document
//
//nolint:lll
func (m *ipmiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"current":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Ampere, Desc: "Current."},
			"fan_speed":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.RotationRete, Desc: "Fan speed."},
			"power_consumption": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Watt, Desc: "Power consumption."},
			"temp":              &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Celsius, Desc: "Temperature."},
			"usage":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Usage."},
			"voltage":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Volt, Desc: "Voltage."},
			"count":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count, Desc: "Count."},
			"status":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Status of the unit."},
			"warning":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Warning on/off."},
		},

		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Monitored host name"},
			"unit": &inputs.TagInfo{Desc: "Unit name in the host"},
		},
	}
}

// be used for server drop warning.
type ipmiServer struct {
	server          string // ipmi server ip
	activeTimestamp int64  // alive timestamp nanosecond time.Now().UnixNano()
	isWarned        bool   // if be warned
}

// ReadEnv support envs：only for K8S.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if tagsStr, ok := envs["ENV_INPUT_IPMI_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_TIMEOUT to time.Duration: %s, ignore", err)
		} else {
			ipt.Timeout.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_DEOP_WARNING_DELAY"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_DEOP_WARNING_DELAY to time.Duration: %s, ignore", err)
		} else {
			ipt.DropWarningDelay.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_BIN_PATH"]; ok {
		str = strings.Trim(str, "\"")
		str = strings.Trim(str, " ")
		ipt.BinPath = str
	}

	if str, ok := envs["ENV_INPUT_IPMI_ENVS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_ENVS: %s, ignore", err)
		} else {
			ipt.Envs = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_SERVERS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_SERVERS: %s, ignore", err)
		} else {
			ipt.IpmiServers = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_INTERFACES"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_INTERFACES: %s, ignore", err)
		} else {
			ipt.IpmiInterfaces = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_USERS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_USERS: %s, ignore", err)
		} else {
			ipt.IpmiUsers = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_PASSWORDS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_PASSWORDS: %s, ignore", err)
		} else {
			ipt.IpmiPasswords = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_HEX_KEYS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_HEX_KEYS: %s, ignore", err)
		} else {
			ipt.HexKeys = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_METRIC_VERSIONS"]; ok {
		var ints []int
		err := json.Unmarshal([]byte(str), &ints)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_METRIC_VERSIONS: %s, ignore", err)
		} else {
			ipt.MetricVersions = ints
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_HEX_KEYS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_HEX_KEYS: %s, ignore", err)
		} else {
			ipt.HexKeys = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_CURRENT"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_CURRENT: %s, ignore", err)
		} else {
			ipt.RegexpCurrent = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_VOLTAGE"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_VOLTAGE: %s, ignore", err)
		} else {
			ipt.RegexpVoltage = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_POWER"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_POWER: %s, ignore", err)
		} else {
			ipt.RegexpPower = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_TEMP"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_TEMP: %s, ignore", err)
		} else {
			ipt.RegexpTemp = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_FAN_SPEED"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_FAN_SPEED: %s, ignore", err)
		} else {
			ipt.RegexpFanSpeed = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_USAGE"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_USAGE: %s, ignore", err)
		} else {
			ipt.RegexpUsage = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_COUNT"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_COUNT: %s, ignore", err)
		} else {
			ipt.RegexpCount = strs
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_REGEXP_STATUS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_REGEXP_STATUS: %s, ignore", err)
		} else {
			ipt.RegexpStatus = strs
		}
	}
}
