// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

// Package logfwdserver implement logfwd websocket server
package logfwdserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/network/ws"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/point"
	gws "github.com/gobwas/ws"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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

	feeder dkio.Feeder
	Tagger datakit.GlobalTagger
}

var (
	_ inputs.InputV2 = (*Input)(nil)
	l                = logger.DefaultSLogger(inputName)
)

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
	return []string{datakit.LabelK8s}
}

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.setup() {
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_logfwdserver"})
	g.Go(func(ctx context.Context) error {
		ipt.srv.Start()
		return nil
	})

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

type message struct {
	Source   string            `json:"source"`
	Pipeline string            `json:"pipeline"`
	Tags     map[string]string `json:"tags"`
	Log      string            `json:"log"`
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

		pts := makePts(msg.Source, []string{msg.Log}, tags)
		if len(pts) == 0 {
			return nil
		}

		err := ipt.feeder.Feed(
			name,
			point.Logging,
			pts,
			&dkio.Option{
				PlOption: &plmanager.Option{
					ScriptMap: map[string]string{msg.Source: msg.Pipeline},
				},
			},
		)
		if err != nil {
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

func makePts(source string, cnt []string, tags map[string]string) []*point.Point {
	pts := []*point.Point{}

	now := time.Now()
	for _, cnt := range cnt {
		opts := point.DefaultLoggingOptions()
		opts = append(opts, point.WithTime(now))

		fields := map[string]interface{}{
			pipeline.FieldMessage: cnt,
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}

		pt := point.NewPointV2(
			source,
			append(point.NewTags(tags), point.NewKVs(fields)...),
			opts...,
		)
		pts = append(pts, pt)
	}
	return pts
}

func defaultInput() *Input {
	return &Input{
		Tags:    make(map[string]string),
		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
