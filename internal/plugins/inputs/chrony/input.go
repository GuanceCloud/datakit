// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package chrony collects chrony metrics.
package chrony

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
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

const (
	defaultInterval = time.Second * 10
	minInterval     = time.Second * 10
	maxInterval     = time.Minute
	defaultTimeout  = time.Second * 5
	minTimeout      = time.Second * 5
	maxTimeout      = time.Second * 30

	inputName  = "chrony"
	metricName = "chrony"
	sampleCfg  = `
[[inputs.chrony]]
  ## (Optional) Collect interval, default is 30 seconds
  # interval = "30s"

  ## (Optional) Exec chronyc timeout, default is 8 seconds
  # timeout = "8s"

  ## (Optional) The binPath of chrony
  bin_path = "chronyc"
 
  ## (Optional) Remote chrony servers
  ## If use remote chrony servers, election must be true
  ## If use remote chrony servers, bin_paths should be shielded
  # remote_addrs = ["<ip>:22"]
  # remote_users = ["<remote_login_name>"]
  # remote_passwords = ["<remote_login_password>"]
  ## If use remote_rsa_path, remote_passwords should be shielded
  # remote_rsa_paths = ["/home/<your_name>/.ssh/id_rsa"]
  # remote_command = "chronyc -n tracking"

  ## Set true to enable election
  election = false

[inputs.chrony.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval time.Duration `toml:"interval"`
	Timeout  time.Duration `toml:"timeout"`
	BinPath  string        `toml:"bin_path"`
	datakit.SSHServers
	Tags     map[string]string
	Election bool `toml:"election"`

	collectCache []*point.Point
	platform     string
	feeder       io.Feeder

	semStop *cliutils.Sem
	pause   bool
	pauseCh chan bool
}

type ChronyMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// LineProto data formatting, submit through FeedMeasurement.
func (n *ChronyMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(n.name, n.tags, n.fields, dkpt.MOptElectionV2(n.election))
}

// Info for docs and integrate testing.
// nolint:lll
func (n *ChronyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Fields: map[string]interface{}{
			"system_time":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the current offset between the NTP clock and system clock."},
			"last_offset":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the estimated local offset on the last clock update."},
			"rms_offset":      &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is a long-term average of the offset value."},
			"frequency":       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This is the rate by which the system clock would be wrong if chronyd was not correcting it."},
			"residual_freq":   &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This shows the residual frequency for the currently selected reference source."},
			"skew":            &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PartPerMillion, Desc: "This is the estimated error bound on the frequency."},
			"root_delay":      &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the total of the network path delays to the stratum-1 computer from which the computer is ultimately synchronized."},
			"root_dispersion": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the total dispersion accumulated through all the computers back to the stratum-1 computer from which the computer is ultimately synchronized."},
			"update_interval": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampSec, Desc: "This is the interval between the last two clock updates."},
		},

		Tags: map[string]interface{}{
			"host":         &inputs.TagInfo{Desc: "Host name"},
			"reference_id": &inputs.TagInfo{Desc: "This is the reference ID and name (or IP address) of the server to which the computer is currently synchronized."},
			"stratum":      &inputs.TagInfo{Desc: "The stratum indicates how many hops away from a computer with an attached reference clock we are."},
			"leap_status":  &inputs.TagInfo{Desc: "This is the leap status, which can be Normal, Insert second, Delete second or Not synchronized."},
		},
	}
}

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("chrony input started")
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.Timeout = config.ProtectedInterval(minInterval, maxInterval, ipt.Timeout)
	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()

		if ipt.pause {
			l.Debugf("not leader, chrony skipped")
		} else {
			l.Debugf("is leader, chrony gathering...")

			if err := ipt.Collect(); err != nil {
				l.Errorf("Collect: %s", err)
				ipt.feeder.FeedLastError(inputName, err.Error())
			}

			if len(ipt.collectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache, &io.Option{CollectCost: time.Since(start)})
				if err != nil {
					l.Errorf("FeedMeasurement: %s", err.Error())
					ipt.feeder.FeedLastError(inputName, err.Error())
				}
				ipt.collectCache = ipt.collectCache[:0]
			}
		}

		select {
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

// Collect Get, Aggregate Data.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)

	if err := ipt.checkConf(); err != nil {
		return err
	}

	data, err := ipt.getData()
	if err != nil {
		return err
	}

	pts, err := ipt.getPts(data)
	if err != nil {
		return err
	}
	ipt.collectCache = pts

	return nil
}

// checkConf check binPath & datakit.SSHServers.
func (ipt *Input) checkConf() error {
	if len(ipt.BinPath) == 0 && len(ipt.RemoteAddrs) == 0 {
		return fmt.Errorf("remote_addrs & bin_path all be null")
	}
	if len(ipt.RemoteAddrs) > 0 {
		if len(ipt.RemoteCommand) == 0 {
			return fmt.Errorf("remote_command is null")
		}
		if (len(ipt.RemoteUsers) == 0 || len(ipt.RemotePasswords) == 0) && len(ipt.RemoteRsaPaths) == 0 {
			return fmt.Errorf("remote_users & remote_passwords & remote_rsa_paths all be null")
		}
		for _, v := range ipt.RemoteAddrs {
			if !strings.HasPrefix(v, "http") {
				v = "http://" + v
			}
			_, err := url.Parse(v)
			if err != nil {
				return fmt.Errorf("parse remote_addrs : %s, error : %w", v, err)
			}
		}
	}
	return nil
}

func (ipt *Input) getData() ([]datakit.SSHData, error) {
	data := make([]datakit.SSHData, 0)
	if len(ipt.SSHServers.RemoteAddrs) > 0 {
		// Remote servers
		// use goroutine, send data through dataCh, with timeout
		dataCh := make(chan datakit.SSHData, 1)
		g := datakit.G("chrony")

		g.Go(func(ctx context.Context) error {
			return datakit.SSHGetData(dataCh, &ipt.SSHServers, ipt.Timeout)
		})

		// collect data
		for v := range dataCh {
			data = append(data, v)
		}
	} else if ipt.BinPath != "" {
		// local server
		v, err := ipt.getLocalBytes(ipt.BinPath)
		if err != nil {
			return nil, fmt.Errorf("get local data error : %w", err)
		}
		data = append(data, datakit.SSHData{Server: "localhost", Data: v})
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("got no data")
	} else {
		return data, nil
	}
}

// getBytes Get the result of binPath execution.
func (ipt *Input) getLocalBytes(binPath string) ([]byte, error) {
	c := exec.Command(binPath, "-n", "tracking")

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return nil, err
	}
	err := datakit.WaitTimeout(c, ipt.Timeout)
	return b.Bytes(), err
}

func (ipt *Input) getPts(data []datakit.SSHData) ([]*point.Point, error) {
	pts := make([]*point.Point, 0)

	opts := point.DefaultMetricOptions()
	if ipt.Election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	} else {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalHostTags()))
	}

	for _, v := range data {
		fields, tags, err := getFields(string(v.Data))
		if err != nil {
			return nil, err
		}

		if v.Server != "localhost" {
			tags["host"] = v.Server
		}

		pt := point.NewPointV2([]byte(metricName), append(point.NewTags(tags), point.NewKVs(fields)...), opts...)
		pts = append(pts, pt)
	}

	return pts, nil
}

// getFields get fields and tags from data.
//
// Data like:
//  Reference ID    : CA760182 (202.118.1.130)
//  Stratum         : 2
//  Ref time (UTC)  : Wed Jun 07 06:22:16 2023
//  System time     : 0.000000000 seconds slow of NTP time
//  Last offset     : -0.000291720 seconds
//  RMS offset      : 0.004762660 seconds
//  Frequency       : 1.452 ppm slow
//  Residual freq   : -0.094 ppm
//  Skew            : 4.524 ppm
//  Root delay      : 0.041327540 seconds
//  Root dispersion : 0.003143095 seconds
//  Update interval : 65.3 seconds
//  Leap status     : Normal
func getFields(out string) (map[string]interface{}, map[string]string, error) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			return nil, nil, fmt.Errorf("unexpected output from chronyc, expected ':' in %s", out)
		}
		name := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(stats[0]), " ", "_"))
		// ignore reference time
		if strings.Contains(name, "ref_time") {
			continue
		}
		valueFields := strings.Fields(stats[1])
		if len(valueFields) == 0 {
			return nil, nil, fmt.Errorf("unexpected output from chronyc: %s", out)
		}
		if strings.Contains(strings.ToLower(name), "stratum") {
			tags["stratum"] = valueFields[0]
			continue
		}
		if strings.Contains(strings.ToLower(name), "reference_id") {
			tags["reference_id"] = valueFields[0]
			continue
		}
		value, err := strconv.ParseFloat(valueFields[0], 64)
		if err != nil {
			tags[name] = strings.ToLower(strings.Join(valueFields, " "))
			continue
		}
		if strings.Contains(stats[1], "slow") {
			value = -value
		}
		fields[name] = value
	}

	return fields, tags, nil
}

// Terminate Stop.
// TODO 请示 这种方法生成点，如何指定是Gauge？
func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// Catalog Catalog.
func (*Input) Catalog() string {
	return "chrony"
}

// SampleConfig : conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs : OS support, reflected in the document.
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) AvailableArchsDCGM() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelK8s}
}

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ChronyMeasurement{},
	}
}

// CHRONY

// ReadEnv support envs：only for K8S.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if str, ok := envs["ENV_INPUT_CHRONY_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, da)
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_TIMEOUT"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_TIMEOUT to time.Duration: %s, ignore", err)
		} else {
			ipt.Timeout = config.ProtectedInterval(minTimeout, maxTimeout, da)
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_BIN_PATH"]; ok {
		ipt.BinPath = str
	}

	if str, ok := envs["ENV_INPUT_CHRONY_REMOTE_ADDRS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_REMOTE_ADDRS: %s, ignore", err)
		} else {
			ipt.RemoteAddrs = strs
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_REMOTE_USERS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_REMOTE_USERS: %s, ignore", err)
		} else {
			ipt.RemoteUsers = strs
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_REMOTE_PASSWORDS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_REMOTE_PASSWORDS: %s, ignore", err)
		} else {
			ipt.RemotePasswords = strs
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_REMOTE_RSA_PATHS"]; ok {
		var strs []string
		err := json.Unmarshal([]byte(str), &strs)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_REMOTE_RSA_PATHS: %s, ignore", err)
		} else {
			ipt.RemoteRsaPaths = strs
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_REMOTE_COMMAND"]; ok {
		ipt.RemoteCommand = str
	}

	if tagsStr, ok := envs["ENV_INPUT_CHRONY_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if str, ok := envs["ENV_INPUT_CHRONY_ELECTION"]; ok {
		election, err := strconv.ParseBool(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_CHRONY_ELECTION: %s, ignore", err)
		} else {
			ipt.Election = election
		}
	}
}

// ElectionEnabled election.
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

func newDefaultInput() *Input {
	ipt := &Input{
		Interval: defaultInterval,
		Tags:     make(map[string]string),
		Timeout:  defaultTimeout,
		Election: true,

		platform: runtime.GOOS,
		feeder:   io.DefaultFeeder(),

		semStop: cliutils.NewSem(),
		pauseCh: make(chan bool, inputs.ElectionPauseChannelLength),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	// To adapt to the variable name nvidia_smi --> chrony in the .conf file
	inputs.Add("chrony", func() inputs.Input {
		return newDefaultInput()
	})
}
