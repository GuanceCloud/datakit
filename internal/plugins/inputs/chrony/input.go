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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/getdatassh"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "chrony"
	metricName  = inputName

	defaultInterval = time.Second * 10
	defaultTimeout  = time.Second * 5
	minTimeout      = time.Second * 5
	maxTimeout      = time.Second * 30
)

var (
	_ inputs.ReadEnv = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)

	_ inputs.ElectionInput = (*Input)(nil)
)

type (
	urlTags map[string]string
	Input   struct {
		Interval time.Duration `toml:"interval"`
		Timeout  time.Duration `toml:"timeout"`
		BinPath  string        `toml:"bin_path"`
		getdatassh.SSHServers
		Tags map[string]string `toml:"tags"`

		semStop      *cliutils.Sem
		collectCache []*point.Point
		platform     string
		feeder       dkio.Feeder
		mergedTags   map[string]urlTags
		tagger       datakit.GlobalTagger

		Election bool `toml:"election"`
		pause    bool
		pauseCh  chan bool
	}
)

// Run Start the process of timing acquisition.
// If this indicator is included in the list to be collected, it will only be called once.
// The for{} loops every tick.
func (ipt *Input) Run() {
	if err := ipt.setup(); err != nil {
		l.Errorf("setup err: %v", err)
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()

		if ipt.pause {
			l.Debugf("not leader, %s skipped", inputName)
		} else {
			if err := ipt.collect(); err != nil {
				l.Errorf("collect: %s", err)
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
			}

			if len(ipt.collectCache) > 0 {
				if err := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
					&dkio.Option{CollectCost: time.Since(start)}); err != nil {
					ipt.feeder.FeedLastError(err.Error(),
						dkio.WithLastErrorInput(inputName),
						dkio.WithLastErrorCategory(point.Metric),
					)
					l.Errorf("feed measurement: %s", err)
				}
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
		case ipt.pause = <-ipt.pauseCh:
		}
	}
}

func (ipt *Input) setup() error {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.Timeout = config.ProtectedInterval(minInterval, maxInterval, ipt.Timeout)

	if err := ipt.checkConf(); err != nil {
		return err
	}

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
		for _, u := range ipt.RemoteAddrs {
			uu := u
			if !strings.HasPrefix(u, "http") {
				u = "http://" + u
			}
			_, err := url.Parse(u)
			if err != nil {
				return fmt.Errorf("parse remote_addrs : %s, error : %w", u, err)
			}

			if ipt.Election {
				ipt.mergedTags[uu] = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, u)
			} else {
				ipt.mergedTags[uu] = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, u)
			}
		}
		l.Debugf("merged tags: %+#v", ipt.mergedTags)
	} else {
		if ipt.Election {
			ipt.mergedTags["localhost"] = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, "")
		} else {
			ipt.mergedTags["localhost"] = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
		}
	}
	return nil
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)

	data, err := ipt.getData()
	if err != nil {
		return err
	}

	if err = ipt.getPts(data); err != nil {
		return err
	}

	return nil
}

func (ipt *Input) getData() ([]*getdatassh.SSHData, error) {
	// data := make([]*getdatassh.SSHData, 0)
	if len(ipt.SSHServers.RemoteAddrs) > 0 {
		// Remote servers
		return getdatassh.GetDataSSH(&ipt.SSHServers, ipt.Timeout)
	} else if ipt.BinPath != "" {
		// Local server
		ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout)
		defer cancel()
		//nolint:gosec
		c := exec.CommandContext(ctx, ipt.BinPath, "-n", "tracking")

		var b bytes.Buffer
		c.Stdout = &b
		c.Stderr = &b
		if err := c.Start(); err != nil {
			return nil, fmt.Errorf("c.Start(): %w, %v", err, b.String())
		}
		err := c.Wait()
		if err != nil {
			return nil, fmt.Errorf("c.Wait(): %s, %w, %v", inputName, err, b.String())
		}

		bytes := b.Bytes()
		l.Debugf("get bytes len: %v.", len(bytes))

		return []*getdatassh.SSHData{{
			Server: "localhost",
			Data:   bytes,
		}}, err
	}

	return nil, fmt.Errorf("%s got no data", inputName)
}

func (ipt *Input) getPts(data []*getdatassh.SSHData) error {
	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	for _, sshData := range data {
		fields, tags, err := getFields(string(sshData.Data))
		if err != nil {
			return err
		}

		var kvs point.KVs

		for k, v := range fields {
			kvs = kvs.Add(k, v, false, true)
		}

		for k, v := range tags {
			kvs = kvs.AddTag(k, v)
		}

		for k, v := range ipt.mergedTags[sshData.Server] {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))
	}

	return nil
}

// getFields get fields and tags from data.
//
// Data like:
//
//	Reference ID    : CA760182 (202.118.1.130)
//	Stratum         : 2
//	Ref time (UTC)  : Wed Jun 07 06:22:16 2023
//	System time     : 0.000000000 seconds slow of NTP time
//	Last offset     : -0.000291720 seconds
//	RMS offset      : 0.004762660 seconds
//	Frequency       : 1.452 ppm slow
//	Residual freq   : -0.094 ppm
//	Skew            : 4.524 ppm
//	Root delay      : 0.041327540 seconds
//	Root dispersion : 0.003143095 seconds
//	Update interval : 65.3 seconds
//	Leap status     : Normal
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

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Timeout", Default: `8s`},
		{FieldName: "BinPath", Type: doc.String, Default: "`chronyc`", Desc: "The path of Chrony", DescZh: "Chrony 的路径"},
		{FieldName: "RemoteAddrs", Type: doc.JSON, Example: `["192.168.1.1:22","192.168.1.2:22"]`, Desc: "If use remote Chrony servers", DescZh: "可以使用远程 Chrony 服务器"},
		{FieldName: "RemoteUsers", Type: doc.JSON, Example: `["user_1","user_2"]`, Desc: "Remote login name", DescZh: "远程登录名"},
		{FieldName: "RemotePasswords", Type: doc.JSON, Example: `["pass_1","pass_2"]`, Desc: "Remote password", DescZh: "远程登录密码"},
		{FieldName: "RemoteRsaPaths", Type: doc.JSON, Example: `["/home/your_name/.ssh/id_rsa"]`, Desc: "Remote rsa paths", DescZh: "秘钥文件路径"},
		{FieldName: "RemoteCommand", Type: doc.String, Example: "\"`chronyc -n tracking`\"", Desc: "Remote command", DescZh: "执行指令"},
		{FieldName: "Election"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_CHRONY_", infos)
}

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

func newDefaultInput() *Input {
	ipt := &Input{
		Interval: defaultInterval,
		Tags:     make(map[string]string),
		Timeout:  defaultTimeout,
		Election: true,

		platform: runtime.GOOS,
		feeder:   dkio.DefaultFeeder(),

		semStop:    cliutils.NewSem(),
		pauseCh:    make(chan bool, inputs.ElectionPauseChannelLength),
		tagger:     datakit.DefaultGlobalTagger(),
		mergedTags: make(map[string]urlTags),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	// To adapt to the variable name nvidia_smi --> chrony in the .conf file
	inputs.Add("chrony", func() inputs.Input {
		return newDefaultInput()
	})
}
