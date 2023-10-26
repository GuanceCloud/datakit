// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "ipmi"
	metricName  = inputName
)

var (
	_ inputs.ReadEnv = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)
	g                = datakit.G("inputs_" + inputName)

	_ inputs.ElectionInput = (*Input)(nil)
)

// be used for server drop warning.
type ipmiServer struct {
	server          string // ipmi server ip
	activeTimestamp int64  // alive timestamp nanosecond time.Now().UnixNano()
	isWarned        bool   // if be warned
}

type Input struct {
	Interval         time.Duration
	Tags             map[string]string
	BinPath          string        `toml:"bin_path"`           // the file path of "ipmitool"
	Envs             []string      `toml:"envs"`               // exec.Command ENV
	IpmiServers      []string      `toml:"ipmi_servers"`       // the ips of ipmi serverbools
	IpmiInterfaces   []string      `toml:"ipmi_interfaces"`    // the Interfaces of ipmi servers
	IpmiUsers        []string      `toml:"ipmi_users"`         // the users name of ipmi servers
	IpmiPasswords    []string      `toml:"ipmi_passwords"`     // the passwords of ipmi servers
	HexKeys          []string      `toml:"hex_keys"`           // provide the hex key for the IMPI connection.
	MetricVersions   []int         `toml:"metric_versions"`    // Schema Version: (defaults to version 1)
	RegexpCurrent    []string      `toml:"regexp_current"`     // regexp
	RegexpVoltage    []string      `toml:"regexp_voltage"`     // regexp
	RegexpPower      []string      `toml:"regexp_power"`       // regexp
	RegexpTemp       []string      `toml:"regexp_temp"`        // regexp
	RegexpFanSpeed   []string      `toml:"regexp_fan_speed"`   // regexp
	RegexpUsage      []string      `toml:"regexp_usage"`       // regexp
	RegexpCount      []string      `toml:"regexp_count"`       // regexp
	RegexpStatus     []string      `toml:"regexp_status"`      // regexp
	Timeout          time.Duration `toml:"timeout"`            // exec timeout
	DropWarningDelay time.Duration `toml:"drop_warning_delay"` // server drop warning delay, default:300s

	semStop    *cliutils.Sem
	feeder     dkio.Feeder
	platform   string
	mergedTags map[string]string
	tagger     datakit.GlobalTagger

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	servers   []ipmiServer // List of active ipmi servers, alarm after service failure
	serversMu sync.Mutex
	start     time.Time
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		if ipt.pause {
			l.Debug("%s election paused", inputName)
		} else {
			ipt.handleServerDropWarn()
			ipt.start = time.Now()
			if err := ipt.collect(); err != nil {
				l.Errorf("collect: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, "")
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	}
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	if err := ipt.checkConfig(); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for i := 0; i < len(ipt.IpmiServers); i++ {
		wg.Add(1)
		func(idx int) {
			g.Go(func(ctx context.Context) error {
				defer wg.Done()
				return ipt.doCollect(idx)
			})
		}(i)
	}

	wg.Wait()
	return nil
}

func (ipt *Input) doCollect(idx int) error {
	data, err := ipt.getBytes(idx)
	if err != nil {
		l.Errorf("getBytes %v", err)
		return err
	}

	// Got data, delay warning.  (server, isActive, isOnlyAdd)
	ipt.serverOnlineInfo(data.server, true, false)

	pts, err := ipt.getPoints(data.data, data.metricVersion, data.server)
	if err != nil {
		l.Errorf("getPoints %v", err)
		return err
	}

	if len(pts) > 0 {
		if err := ipt.feeder.Feed(metricName, point.Metric, pts,
			&dkio.Option{CollectCost: time.Since(ipt.start)}); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Metric),
			)
			l.Errorf("feed measurement: %s", err)
		}
	}

	return nil
}

func (ipt *Input) checkConfig() error {
	if ipt.BinPath == "" {
		return fmt.Errorf("BinPath is null")
	}

	if len(ipt.IpmiServers) == 0 {
		return fmt.Errorf("IpmiServers is null")
	}

	if len(ipt.HexKeys) == 0 && (len(ipt.IpmiUsers) == 0 || len(ipt.IpmiPasswords) == 0) {
		l.Errorf("No HexKeys && No IpmiUsers No IpmiPasswords")
		return fmt.Errorf("no HexKeys and No IpmiUsers No IpmiPasswords")
	}

	// Check if have same ip.
	arr := make([]string, len(ipt.IpmiServers))
	copy(arr, ipt.IpmiServers)
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] == arr[i+1] {
			return fmt.Errorf("IpmiServers have same ip: %v", arr[i])
		}
	}
	return nil
}

// Get parameter that exec need.
func (ipt *Input) getParameters(i int) (opts []string, metricVersion int, err error) {
	var (
		ipmiInterface string
		ipmiUser      string
		ipmiPassword  string
		hexKey        string
	)

	// Add ipmiInterface ipmiServer parameter.
	if len(ipt.IpmiInterfaces) < i+1 {
		ipmiInterface = ipt.IpmiInterfaces[0]
	} else {
		ipmiInterface = ipt.IpmiInterfaces[i]
	}
	opts = []string{"-I", ipmiInterface, "-H", ipt.IpmiServers[i]}

	if len(ipt.HexKeys) > 0 {
		// Add hexkey parameter.
		if len(ipt.HexKeys) < i+1 {
			hexKey = ipt.HexKeys[0]
		} else {
			hexKey = ipt.HexKeys[i]
		}
		opts = append(opts, "-y", hexKey)
	}

	// Add ipmiUser ipmiPassword parameter.
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

// Add or update server info.
func (ipt *Input) serverOnlineInfo(server string, isActive bool, isOnlyAdd bool) {
	ipt.serversMu.Lock()
	defer ipt.serversMu.Unlock()
	for i := 0; i < len(ipt.servers); i++ {
		if ipt.servers[i].server == server {
			if isActive && !isOnlyAdd {
				ipt.servers[i].isWarned = false                        // Set never warning
				ipt.servers[i].activeTimestamp = time.Now().UnixNano() // Reset timestamp
			}
			return
		}
	}
	// New, append, server not in ipt.servers.
	ipt.servers = append(ipt.servers, ipmiServer{server, time.Now().UnixNano(), false})
}

func (ipt *Input) handleServerDropWarn() {
	ipt.serversMu.Lock()
	defer ipt.serversMu.Unlock()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ipt.start))
	for i := 0; i < len(ipt.servers); i++ {
		if time.Now().UnixNano() > ipt.servers[i].activeTimestamp+ipt.DropWarningDelay.Nanoseconds() &&
			!ipt.servers[i].isWarned {
			// This server time stamp out of delay time && no alarm has been given.
			l.Warnf("before lost a server: %s", ipt.servers[i].server)

			// Send warning.
			var kvs point.KVs
			kvs = kvs.Add("host", ipt.servers[i].server, true, true)
			kvs = kvs.Add("warning", 1, false, true)
			for k, v := range ipt.mergedTags {
				kvs = kvs.AddTag(k, v)
			}
			pts := []*point.Point{point.NewPointV2(inputName, kvs, opts...)}
			if err := ipt.feeder.Feed(metricName, point.Metric, pts,
				&dkio.Option{CollectCost: time.Since(ipt.start)}); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}

			l.Warnf("lost a server: %s", ipt.servers[i].server)

			// Set warned state. Do not warn and warn.
			ipt.servers[i].isWarned = true
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return inputName }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
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
			l.Warnf("parse ENV_INPUT_IPMI_INTERVAL to time: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_TIMEOUT to time: %s, ignore", err)
		} else {
			ipt.Timeout = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_IPMI_DEOP_WARNING_DELAY"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_IPMI_DEOP_WARNING_DELAY to time: %s, ignore", err)
		} else {
			ipt.DropWarningDelay = config.ProtectedInterval(minInterval,
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

func newDefaultInput() *Input {
	ipt := &Input{
		platform:       runtime.GOOS,
		IpmiInterfaces: []string{"lan"},
		Interval:       time.Second * 10,
		semStop:        cliutils.NewSem(),

		Tags:             make(map[string]string),
		Timeout:          time.Second * 5,
		MetricVersions:   []int{1},
		DropWarningDelay: time.Second * 300,
		servers:          make([]ipmiServer, 0, 1),
		pauseCh:          make(chan bool, inputs.ElectionPauseChannelLength),
		Election:         true,
		feeder:           dkio.DefaultFeeder(),
		tagger:           datakit.DefaultGlobalTagger(),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
