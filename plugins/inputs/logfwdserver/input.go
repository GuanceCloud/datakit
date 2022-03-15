// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwdserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	gws "github.com/gobwas/ws"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/ws"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "logfwdserver"

	sampleCfg = `
[inputs.logfwdserver]
  address = "0.0.0.0:9533"

  [inputs.logfwdserver.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	Address string            `toml:"address"`
	Tags    map[string]string `toml:"tags"`

	srv     *ws.Server
	semStop *cliutils.Sem // start stop signal
}

var l = logger.DefaultSLogger(inputName)

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.setup() {
		return
	}

	go ipt.srv.Start()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.Stop()
			l.Infof("%s exit", inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.Stop()
			l.Infof("%s return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() bool {
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit", inputName)
			return true
		default:
			// nil
		}

		time.Sleep(time.Second)

		ipt.srv, err = ws.NewServer(ipt.Address, "/logfwd")
		if err != nil {
			l.Error(err)
			continue
		}

		break
	}

	ipt.srv.MsgHandler = func(s *ws.Server, c net.Conn, data []byte, op gws.OpCode) error {
		var msg message
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}

		name := "logfwd/" + msg.Source
		tags := msg.Tags
		if tags == nil {
			tags = make(map[string]string)
		}
		for k, v := range ipt.Tags {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}
		if tags["pod_name"] != "" {
			name += fmt.Sprintf("(podname:%s)", tags["pod_name"])
		}
		task := &worker.Task{
			TaskName:   name,
			Source:     msg.Source,
			ScriptName: msg.Pipeline,
			Data: []worker.TaskData{
				&taskData{
					tags: tags,
					log:  msg.Log,
				},
			},
			TS: time.Now(),
		}

		if err := worker.FeedPipelineTaskBlock(task); err != nil {
			l.Errorf("logfwd failed to feed log, pod_name:%s filename:%s, err: %w", tags["pod_name"], tags["filename"], err)
			return err
		}

		return nil
	}

	// add-cli callback
	ipt.srv.AddCli = func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := gws.UpgradeHTTP(r, w)
		if err != nil {
			l.Error("ws.UpgradeHTTP error: %s", err.Error())
			return
		}

		if err := ipt.srv.AddConnection(conn); err != nil {
			l.Error(err)
		}
	}

	return false
}

func (ipt *Input) Stop() {
	if ipt.srv != nil {
		ipt.srv.Stop()
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string {
	return "log"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

type message struct {
	Source   string            `json:"source"`
	Pipeline string            `json:"pipeline"`
	Tags     map[string]string `json:"tags"`
	Log      string            `json:"log"`
}

type taskData struct {
	tags map[string]string
	log  string
}

func (t *taskData) GetContent() string {
	return t.log
}

func (t *taskData) Handler(r *worker.Result) error {
	for k, v := range t.tags {
		if _, err := r.GetTag(k); err != nil {
			r.SetTag(k, v)
		}
	}
	return nil
}

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			semStop: cliutils.NewSem(),
		}
	})
}
