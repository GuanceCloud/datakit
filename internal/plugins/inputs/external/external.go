// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package external wraps all external command to collect various metrics
package external

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	configSample = `
[[inputs.external]]

    # Collector's name.
    name = 'some-external-inputs'  # required

    # Whether or not to run the external program in the background.
    daemon = false

    # If the external program running in a Non-daemon mode,
    #     runs it in every this interval time.
    #interval = '10s'

    # The environment variables running the external program.
    #envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

    # The external program' full path. Filling in absolute path whenever possible.
    cmd = "python" # required

    # Filling "true" if this collecor is involved in the election.
    # Note: The external program must running in a daemon mode if involving the election.
    election = false
    args = []

    [[inputs.external.tags]]
        # tag1 = "val1"
        # tag2 = "val2"
`
)

var (
	inputName                = "external"
	l                        = logger.DefaultSLogger(inputName)
	_         inputs.InputV2 = (*Input)(nil)
)

type Input struct {
	Name     string            `toml:"name"`
	Daemon   bool              `toml:"daemon"`
	Election bool              `toml:"election"`
	Interval string            `toml:"interval"`
	Envs     []string          `toml:"envs"`
	Cmd      string            `toml:"cmd"`
	Args     []string          `toml:"args"`
	Tags     map[string]string `toml:"tags"`

	cmd      *exec.Cmd      `toml:"-"`
	duration time.Duration  `toml:"-"`
	Query    []*customQuery `toml:"custom_queries"`

	semStop        *cliutils.Sem // start stop signal
	semStopProcess *cliutils.Sem
	Tagger         datakit.GlobalTagger
	procExitReply  chan struct{}

	daemonStarted bool

	pauseCh chan bool
	pause   bool
}

// customQuery contains custom sql query info.
type customQuery struct {
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	MD5Hash string
}

func NewInput() *Input {
	return &Input{
		semStop:        cliutils.NewSem(),
		semStopProcess: cliutils.NewSem(),
		Tagger:         datakit.DefaultGlobalTagger(),
		Election:       true,
		pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (*Input) Catalog() string {
	return "external"
}

func (*Input) SampleConfig() string {
	return configSample
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (ipt *Input) precheck() error {
	ipt.duration = time.Second * 10
	if ipt.Interval != "" {
		du, err := time.ParseDuration(ipt.Interval)
		if err != nil {
			l.Errorf("parse external input %s interval failed: %s", ipt.Name, err.Error())
			return err
		}

		ipt.duration = du
	}

	// TODO: check ex.Cmd is ok

	return nil
}

func (ipt *Input) start() error {
	ipt.getCustomQuery()

	l.Debugf("starting %s cmd %s %s, envs: %+#v", ipt.Name, ipt.Cmd, strings.Join(ipt.Args, " "), ipt.Envs)
	ipt.cmd = exec.Command(ipt.Cmd, ipt.Args...) //nolint:gosec
	if ipt.Envs != nil {
		ipt.cmd.Env = ipt.Envs
	}

	if err := ipt.cmd.Start(); err != nil {
		l.Errorf("start external input %s failed: %s", ipt.Name, err.Error())
		return err
	}

	return nil
}

// NeedElectionFlag decides if an external input start-up needs flag 'election = T/F'.
func NeedElectionFlag(name string) bool {
	list := []string{"oracle", "db2", "oceanbase"}
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Infof("starting external input %s...", ipt.Name)

	tagsStr := ""
	arr := []string{}
	for tagKey, tagVal := range ipt.Tags {
		arr = append(arr, fmt.Sprintf("%s=%s", tagKey, tagVal))
	}
	if len(arr) > 0 {
		tagsStr = strings.Join(arr, ";")
	}

	if tagsStr != "" {
		ipt.Args = append(ipt.Args, []string{"--tags", tagsStr}...)
	}

	if NeedElectionFlag(ipt.Name) && config.Cfg.Election.Enable && ipt.Election {
		ipt.Args = append(ipt.Args, "--election")
	}

	for {
		if err := ipt.precheck(); err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	tick := time.NewTicker(ipt.duration)
	defer tick.Stop()

	for {
		if ipt.pause {
			l.Debugf("%s not leader, skipped", ipt.Name)
		} else {
			if ipt.Daemon {
				ipt.daemonRun()
			} else {
				// run as new process
				l.Debug("non-daemon starting")
				_ = ipt.start() //nolint:errcheck
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Infof("external input %s exiting", ipt.Name)
			ipt.semStopProcess.Close()
			return

		case <-ipt.semStop.Wait():
			l.Infof("external input %s stopped", ipt.Name)
			ipt.semStopProcess.Close()
			return

		case ipt.pause = <-ipt.pauseCh:
			if ipt.pause {
				l.Infof("%s paused", ipt.Name)
				if ipt.Daemon && ipt.daemonStarted { // stop the daemon running process
					ipt.semStopProcess.Close() // trigger the daemon process exit
					<-ipt.procExitReply        // sync with goroutine monitoring external input process
					ipt.daemonStarted = false
					ipt.semStopProcess = cliutils.NewSem() // reopen the sem
				}
			}

		case <-tick.C:
		}
	}
}

func (ipt *Input) queryToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(ipt.Query); err != nil {
		l.Errorf("Encode() error: %v", err)
		return nil
	}
	return buffer.Bytes()
}

func (ipt *Input) getCustomQuery() {
	if len(ipt.Query) == 0 {
		l.Debug("ipt.Query empty")
		return
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "custom_query")
	if err != nil {
		l.Errorf("os.CreateTemp() failed: %v", err)
		return
	}

	bys := ipt.queryToBytes()
	if len(bys) == 0 {
		l.Debug("bytes empty")
		return
	}

	cnt, err := tmpFile.Write(bys)
	if err != nil {
		l.Errorf("Write() failed: %v", err)
		return
	}
	l.Infof("Wrote file %s, wrote %d bytes.", tmpFile.Name(), cnt)

	ipt.Args = append(ipt.Args, "--custom-query", tmpFile.Name())
}

func (ipt *Input) daemonRun() {
	if ipt.daemonStarted {
		return
	}

	// start failed, retry
	for {
		l.Debug("daemon starting")
		if err := ipt.start(); err != nil {
			time.Sleep(time.Second)
			continue
		}
		ipt.daemonStarted = true
		break
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_external"})

	ipt.procExitReply = make(chan struct{})

	func(process *os.Process, name string, semStopProcess *cliutils.Sem, procExitReply chan struct{}) {
		g.Go(func(ctx context.Context) error {
			if err := datakit.MonitProc(process, name, semStopProcess); err != nil { // blocking here...
				l.Errorf("datakit.MonitProc: %s", err.Error())
			}
			close(procExitReply)
			return nil
		})
	}(ipt.cmd.Process, ipt.Name, ipt.semStopProcess, ipt.procExitReply)

	// We must not modify ex.cmd.Process.Pid beyong this point,
	// because pid is needed by MonitProc to kill the process when signaled.
	// TODO: refactor to make cmd private to the goroutine above.
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", ipt.Name)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", ipt.Name)
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
