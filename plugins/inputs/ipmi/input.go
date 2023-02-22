// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

var valueTypes = [][2]string{
	{"current", "float"},
	{"voltage", "float"},
	{"power", "float"},
	{"temp", "float"},
	{"fan_speed", "int"},
	{"usage", "float"},
	{"count", "int"},
	{"status", "int"},
}

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName  = "ipmi"
	metricName = inputName
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
  ## (Example) regexp_power = ["pwr"]
  regexp_power = ["pwr"]

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
  ## (Example) regexp_status = ["fan","slot","drive"]
  regexp_status = ["fan","slot","drive"]

[inputs.ipmi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

var _ inputs.ElectionInput = (*Input)(nil)

type Input struct {
	Interval     datakit.Duration  // run every "Interval" seconds
	Tags         map[string]string // Indicator name
	collectCache []inputs.Measurement
	platform     string

	BinPath          string           `toml:"bin_path"`           // the file path of "ipmitool"
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

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool
}

type dataSrtuct struct {
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

// be used for server drop warning.
type ipmiServer struct {
	server          string // ipmi server
	activeTimestamp int64  // 活跃时间戳 nanosecond time.Now().UnixNano()
	isWarned        bool   // if alreday warned
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}

func newDefaultInput() *Input {
	ipt := &Input{
		platform:         runtime.GOOS,
		IpmiInterfaces:   []string{"lan"},
		Interval:         datakit.Duration{Duration: time.Second * 10},
		semStop:          cliutils.NewSem(),
		Tags:             make(map[string]string),
		Timeout:          datakit.Duration{Duration: time.Second * 5},
		MetricVersions:   []int{1},
		DropWarningDelay: datakit.Duration{Duration: time.Second * 300},
		servers:          make([]ipmiServer, 0, 1),
		pauseCh:          make(chan bool, inputs.ElectionPauseChannelLength),
		Election:         true,
	}
	return ipt
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("ipmi input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			l.Debugf("is leader, ipmi gathering...")
			// Collect() to get data
			if err := ipt.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
				io.FeedLastError(inputName, err.Error())
			}
			// If there is data in the cache, submit it
			if len(ipt.collectCache) > 0 {
				if err := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
					&io.Option{CollectCost: time.Since(start)}); err != nil {
					l.Errorf("FeedMeasurement: %s", err)
				}
			}
		}

		select {
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return
		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
		}
	}
}

// Collect Get, Aggregate Data.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)

	ipt.handleSreverDrop()

	if ipt.BinPath == "" {
		l.Errorf("BinPath is null.")
		return errors.New("BinPath is null")
	}

	if len(ipt.IpmiServers) == 0 {
		l.Errorf("IpmiServers is null.")
		return errors.New("IpmiServers is null")
	}

	if len(ipt.HexKeys) == 0 && (len(ipt.IpmiUsers) == 0 || len(ipt.IpmiPasswords) == 0) {
		l.Errorf("No HexKeys && No IpmiUsers No IpmiPasswords.")
		return errors.New("no HexKeys && no IpmiUsers no IpmiPasswords")
	}

	// use goroutine, chanel transfer data, with context timeout, with tag
	ipmiDataCh := make(chan dataSrtuct, 1)
	ctxNew, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()
	l = logger.SLogger("ipmi")
	g := datakit.G("ipmi")

	g.Go(func(ctx context.Context) error {
		return getDatas(ctxNew, ipmiDataCh, ipt)
	})

	handleDatas(ipmiDataCh, ipt)

	return nil
}

// use goroutine in goroutine,  parallel get ipmitools Data.
func getDatas(ctx context.Context, ch chan dataSrtuct, ipt *Input) error {
	doneCh := make(chan struct{})
	// cycle exec all ipmi servers
	go func(ch chan dataSrtuct) {
		var wg sync.WaitGroup
		for i := 0; i < len(ipt.IpmiServers); i++ {
			// only add a recorder,
			ipt.servrerOnlineInfo(ipt.IpmiServers[i], false, true)
			// get exec parameter
			opts, metricVersion, err := ipt.getParameters(i)
			if err != nil {
				continue // do not return, perhaps other ipmiServer be right.
			}

			wg.Add(1)
			go func(ch chan dataSrtuct, index int, metricVersion int, opts []string) {
				defer wg.Done()
				// get data from ipmiServer（[]byte）
				data, err := ipt.getBytes(ipt.BinPath, opts)
				if err != nil {
					l.Errorf("get bytes by binPath log: %s .", ipt.IpmiServers[index], err)
				} else {
					ch <- dataSrtuct{
						server:        ipt.IpmiServers[index],
						metricVersion: metricVersion,
						data:          data,
					}
					// Got data, Delay warning
					ipt.servrerOnlineInfo(ipt.IpmiServers[index], true, false)
				}
			}(ch, i, metricVersion, opts)
		}

		wg.Wait()
		// all finish, doneCh
		doneCh <- struct{}{}
	}(ch)

	// block until，perhaps timeout ，perhaps all finish
	select {
	case <-ctx.Done(): // block until timeout done
		l.Warnf("function ipmi.getDatas timeout : %s, ignore")
		close(ch)
		// close(doneCh) // need not close

		return fmt.Errorf("ipmi getDatas timeout")
	case <-doneCh: // block until finish all task
		close(ch)
		// close(doneCh) // need not close

		return nil
	}
}

// handle ipmitools Data. need not ctx, because getDatas has ctx timeout close ch.
func handleDatas(ch chan dataSrtuct, ipt *Input) {
	for datas := range ch {
		ipt.convert(datas.data, datas.metricVersion, datas.server)
	}
}

// get parameter that exec need.
func (ipt *Input) getParameters(i int) (opts []string, metricVersion int, err error) {
	var (
		ipmiInterface string
		ipmiUser      string
		ipmiPassword  string
		hexKey        string
	)

	// add ipmiInterface ipmiServer parameter
	if len(ipt.IpmiInterfaces) < i+1 {
		ipmiInterface = ipt.IpmiInterfaces[0]
	} else {
		ipmiInterface = ipt.IpmiInterfaces[i]
	}
	opts = []string{"-I", ipmiInterface, "-H", ipt.IpmiServers[i]}

	if len(ipt.HexKeys) > 0 {
		// add hexkey parameter
		if len(ipt.HexKeys) < i+1 {
			hexKey = ipt.HexKeys[0]
		} else {
			hexKey = ipt.HexKeys[i]
		}
		opts = append(opts, "-y", hexKey)
	}

	// add ipmiUser ipmiPassword parameter
	if len(ipt.IpmiUsers) == 0 || len(ipt.IpmiPasswords) == 0 {
		l.Error("have no hexKey && ipmiUser && ipmiPassword")
		err = errors.New("have no hexKey && ipmiUser && ipmiPassword")
		return
	}
	if len(ipt.IpmiUsers) < i+1 {
		ipmiUser = ipt.IpmiUsers[0]
	} else {
		ipmiUser = ipt.IpmiUsers[i]
	}
	if len(ipt.IpmiPasswords) < i+1 {
		ipmiPassword = ipt.IpmiPasswords[0]
	} else {
		ipmiPassword = ipt.IpmiPasswords[i]
	}
	opts = append(opts, "-U", ipmiUser, "-P", ipmiPassword)

	if len(ipt.MetricVersions) < i+1 {
		metricVersion = ipt.MetricVersions[0]
	} else {
		metricVersion = ipt.MetricVersions[i]
	}
	if metricVersion == 2 {
		opts = append(opts, "sdr", "elist")
	} else {
		opts = append(opts, "sdr")
	}

	return opts, metricVersion, err
}

// Get the result of binPath execution
// @binPath One of run bin files.
func (ipt *Input) getBytes(binPath string, opts []string) ([]byte, error) {
	c := exec.Command(binPath, opts...)
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := datakit.WaitTimeout(c, ipt.Timeout.Duration)
	return b.Bytes(), err
}

// covert to metric.
func (ipt *Input) convert(data []byte, metricVersion int, server string) {
	// data just like
	// V1
	// Temp             | 45 degrees C      | ok
	// NDC PG           | 0x00              | ok

	// V2
	// Inlet Temp       | 05h | ok  |  7.1 | 23 degrees C
	// NDC PG           | 08h | ok  |  7.1 | State Deasserted

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		// format the data
		strs := strings.Split(scanner.Text(), "|")
		if metricVersion == 2 {
			strs[1] = strs[4]
		}

		strs[0] = strings.Trim(strs[0], " ")
		strs[0] = strings.ToLower(strs[0])
		strs[0] = strings.ReplaceAll(strs[0], " ", "_")

		strs[1] = strings.Trim(strs[1], " ")
		strs[1] = strings.Split(strs[1], " ")[0]

		strs[2] = strings.Trim(strs[2], " ")
		strs[2] = strings.ToLower(strs[2])

		metricIndex, value, err := ipt.getMetric(strs)
		if err != nil {
			l.Errorf(" error: %v", err)
			continue
		}

		if metricIndex == -1 {
			// not error, only give up this data
			continue
		}

		tags := map[string]string{
			"host": server,
			"unit": strs[0],
		}

		fields := map[string]interface{}{
			valueTypes[metricIndex][0]: value,
		}

		// metric := metric{tags, fields}

		ipt.collectCache = append(ipt.collectCache, &ipmiMeasurement{
			name:     inputName,
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		})
	}
}

// get metric.
func (ipt *Input) getMetric(strs []string) (int, interface{}, error) {
	var expName *regexp.Regexp
	var expValue *regexp.Regexp
	var expStr string
	var expStrs []string
	var ok bool
	var err error

	// try if it's value, only like 1234 or 1234.56
	expValue = regexp.MustCompile(`^\d+(\.?\d+|\d*)$`)
	if expValue.MatchString(strs[1]) {
		// for try all metric kind from valueType
		for i := 0; i < len(valueTypes); i++ {
			if valueTypes[i][0] == "status" {
				continue
			}

			expStrs, err = ipt.getExpStrs(valueTypes[i][0])
			if err != nil {
				l.Errorf(" error: %v", err)
				continue
			}

			// for try all Regexp in this metric kind
			for _, expStr = range expStrs {
				expName = regexp.MustCompile(expStr)
				ok = expName.MatchString(strs[0])
				if ok && strs[2] == "ok" {
					// Match name
					switch valueTypes[i][1] {
					case "float":
						expValue = regexp.MustCompile(`^\d+(\.?\d+|\d*)$`)
					case "int":
						expValue = regexp.MustCompile(`^\d+$`)
					default:
						return i, "", fmt.Errorf(" error value type : %s", strs[1])
					}

					if expValue.MatchString(strs[1]) {
						// Matching value succeeded
						// get the value
						switch valueTypes[i][1] {
						case "float":
							value, err := strconv.ParseFloat(strs[1], 64)
							if err != nil {
								return i, "", fmt.Errorf(" error strconv.ParseFloat : %s", strs[1])
							}
							return i, value, nil
						case "int":
							value, err := strconv.Atoi(strs[1])
							if err != nil {
								return i, "", fmt.Errorf(" error strconv.Atoi : %s", strs[1])
							}
							return i, value, nil
						default:
							return i, "", fmt.Errorf(" error value type: %s", valueTypes[i][1])
						}
					} else {
						// Matching value failed
						return i, "", fmt.Errorf("can not matching metric value: %s", strs[0])
					}
				}
			}
		}
	}

	// try if it's status, only "ok" or "ns"
	// for try only status kind from valueType
	for i := 0; i < len(valueTypes); i++ {
		if valueTypes[i][0] != "status" {
			continue
		}

		expStrs, err = ipt.getExpStrs(valueTypes[i][0])
		if err != nil {
			l.Errorf(" error: %v", err)
			continue
		}

		// for try all Regexp in this metric kind
		for _, expStr = range expStrs {
			expName = regexp.MustCompile(expStr)
			ok = expName.MatchString(strs[0])
			if ok {
				if strs[2] == "ok" {
					return i, 1, nil
				} else {
					return i, 0, nil
				}
			}
		}
	}
	// not error, only givve uo this data
	return -1, "", nil
}

// get metric regexp list.
func (ipt *Input) getExpStrs(metricName string) (expStrs []string, err error) {
	switch metricName {
	case "current":
		return ipt.RegexpCurrent, nil
	case "voltage":
		return ipt.RegexpVoltage, nil
	case "power":
		return ipt.RegexpPower, nil
	case "temp":
		return ipt.RegexpTemp, nil
	case "fan_speed":
		return ipt.RegexpFanSpeed, nil
	case "usage":
		return ipt.RegexpUsage, nil
	case "count":
		return ipt.RegexpCount, nil
	case "status":
		return ipt.RegexpStatus, nil
	default:
		return nil, fmt.Errorf("not find metric name: %s", metricName)
	}
}

// Terminate Stop.
func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// Catalog Catalog.
func (*Input) Catalog() string {
	return inputName
}

// SampleConfig : conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs : OS support, reflected in the document.
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

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

// LineProto data formatting, submit through FeedMeasurement.
func (n *ipmiMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(n.name, n.tags, n.fields, point.MOptElectionV2(n.election))
}

// Info , reflected in the document
//nolint:lll
func (n *ipmiMeasurement) Info() *inputs.MeasurementInfo {
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

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ipmiMeasurement{},
	}
}

// Add a server or updata srever info.
func (ipt *Input) servrerOnlineInfo(server string, isActive bool, isOnlyAdd bool) {
	i := 0
	for i = 0; i < len(ipt.servers); i++ {
		if ipt.servers[i].server == server {
			if isActive && !isOnlyAdd {
				// here is multi thread but different recorder，so need not lock
				ipt.servers[i].isWarned = false                        // set never warning
				ipt.servers[i].activeTimestamp = time.Now().UnixNano() // reset timestamp
			}
			return
		}
	}
	if isOnlyAdd && i >= len(ipt.servers) {
		// here is single thread，so need not lock
		ipt.servers = append(ipt.servers, ipmiServer{server, time.Now().UnixNano(), false}) // 新的，加入队列
	}
}

// Handle server drop warning.
func (ipt *Input) handleSreverDrop() {
	for i := 0; i < len(ipt.servers); i++ {
		if time.Now().UnixNano() > ipt.servers[i].activeTimestamp+ipt.DropWarningDelay.Duration.Nanoseconds() &&
			!ipt.servers[i].isWarned {
			// this server time stamp out of delay time && no alarm has been given

			// send warning
			tags := map[string]string{
				"host": ipt.servers[i].server,
			}

			fields := map[string]interface{}{
				"warning": 1,
			}

			ipt.collectCache = append(ipt.collectCache, &ipmiMeasurement{
				name:     inputName,
				tags:     tags,
				fields:   fields,
				election: ipt.Election,
			})

			// set warned state
			ipt.servers[i].isWarned = true
		}
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}
