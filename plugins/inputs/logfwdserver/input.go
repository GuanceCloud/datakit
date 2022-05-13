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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/scriptstore"
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
		if pts := runPipeline(msg.Source, msg.Pipeline, []string{msg.Log}, tags, plCallback); len(pts) > 0 {
			if err := io.Feed(name, datakit.Logging, pts, nil); err != nil {
				l.Errorf("logfwd failed to feed log, pod_name:%s filename:%s, err: %w", tags["pod_name"], tags["filename"], err)
				return err
			}
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

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			semStop: cliutils.NewSem(),
		}
	})
}

func plCallback(res *pipeline.Result) (*pipeline.Result, error) {
	res.CheckFieldValLen(0)
	return pipeline.ResultUtilsLoggingProcessor(res, false, nil), nil
}

func runPipeline(source, scriptName string, cnt []string, tags map[string]string,
	callback func(*pipeline.Result) (*pipeline.Result, error)) []*io.Point {
	ret := []*io.Point{}

	for _, cnt := range cnt {
		pt, err := io.MakePoint(source, tags,
			map[string]interface{}{pipeline.PipelineMessageField: cnt},
			time.Now(),
		)
		if err != nil {
			l.Error(err)
			continue
		}
		drop := false
		if script, ok := scriptstore.QueryScript(scriptName); ok {
			if ptRet, dropRet, err := pipeline.RunScript(pt, script, callback); err != nil {
				l.Error(err)
			} else {
				pt = ptRet
				drop = dropRet
			}
		}
		if !drop {
			ret = append(ret, pt)
		}
	}
	return ret
}
