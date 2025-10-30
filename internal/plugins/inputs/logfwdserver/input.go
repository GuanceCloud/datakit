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
	"net"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/network/ws"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	gws "github.com/gobwas/ws"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "logfwdserver"

	sampleConfig = `
[inputs.logfwdserver]
  address = "0.0.0.0:9533"

  [inputs.logfwdserver.tags]
  # Add custom tags for log forwarding
  # service = "my-service"
  # environment = "production"
`
)

type Input struct {
	Address string            `toml:"address"`
	Tags    map[string]string `toml:"tags"`

	server  *ws.Server
	stopSem *cliutils.Sem
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
}

var (
	_ inputs.InputV2 = (*Input)(nil)

	log = logger.DefaultSLogger(inputName)
)

func (*Input) Catalog() string { return "log" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string { return []string{datakit.LabelK8s} }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (ipt *Input) Run() {
	log = logger.SLogger(inputName)

	if ipt.initializeServer() {
		return
	}

	ipt.startServer()
	ipt.waitForShutdown()
}

func (ipt *Input) Stop() {
	if ipt.server != nil {
		ipt.server.Stop()
	}
}

func (ipt *Input) Terminate() {
	if ipt.stopSem != nil {
		ipt.stopSem.Close()
	}
}

type logMessage struct {
	Source       string                 `json:"source"`
	StorageIndex string                 `json:"storage_index"`
	Pipeline     string                 `json:"pipeline"`
	Tags         map[string]string      `json:"tags"`
	Fields       map[string]interface{} `json:"fields"`
	Log          string                 `json:"log"`
}

func (ipt *Input) initializeServer() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			log.Info("logfwdserver exiting during initialization")
			return true
		default:
		}

		time.Sleep(time.Second)

		server, err := ws.NewServer(ipt.Address, "/logfwd")
		if err != nil {
			log.Error("failed to create websocket server: %v", err)
			continue
		}

		ipt.server = server
		ipt.setupMessageHandler()
		ipt.setupConnectionHandler()
		break
	}

	return false
}

func (ipt *Input) setupMessageHandler() {
	ipt.server.MsgHandler = func(s *ws.Server, c net.Conn, data []byte, op gws.OpCode) error {
		var msg logMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Error("failed to unmarshal message: %v", err)
			return err
		}

		return ipt.processLogMessage(msg)
	}
}

func (ipt *Input) setupConnectionHandler() {
	ipt.server.AddCli = func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := gws.UpgradeHTTP(r, w)
		if err != nil {
			log.Error("websocket upgrade failed: %v", err)
			return
		}

		if err := ipt.server.AddConnection(conn); err != nil {
			log.Error("failed to add connection: %v", err)
		}
	}
}

func (ipt *Input) processLogMessage(msg logMessage) error {
	feedName := ipt.buildFeedName(msg.Source, msg.StorageIndex)
	tags := ipt.mergeTags(msg.Tags)

	points := ipt.createLogPoints(msg.Source, msg.Log, tags, msg.Fields)
	if len(points) == 0 {
		return nil
	}

	err := ipt.feeder.Feed(point.Logging, points,
		dkio.WithStorageIndex(msg.StorageIndex),
		dkio.WithSource(feedName),
		dkio.WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{msg.Source: msg.Pipeline},
		}),
	)
	if err != nil {
		log.Error("failed to feed log message: %v", err)
		return err
	}

	return nil
}

func (ipt *Input) buildFeedName(source, storageIndex string) string {
	feedName := dkio.FeedSource("logfwd", source)
	if storageIndex != "" {
		feedName = dkio.FeedSource(feedName, storageIndex)
	}
	return feedName
}

func (ipt *Input) mergeTags(msgTags map[string]string) map[string]string {
	if msgTags == nil {
		msgTags = make(map[string]string)
	}
	for k, v := range ipt.Tags {
		if _, exists := msgTags[k]; !exists {
			msgTags[k] = v
		}
	}
	return msgTags
}

func (ipt *Input) createLogPoints(source, logContent string, tags map[string]string, fields map[string]interface{}) []*point.Point {
	return ipt.buildPoints(source, []string{logContent}, tags, fields)
}

func (ipt *Input) startServer() {
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_logfwdserver"})
	g.Go(func(ctx context.Context) error {
		ipt.server.Start()
		return nil
	})
}

func (ipt *Input) waitForShutdown() {
	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.Stop()
			log.Info("logfwdserver shutting down")
			return
		case <-ipt.stopSem.Wait():
			ipt.Stop()
			log.Info("logfwdserver terminated")
			return
		}
	}
}

func (ipt *Input) buildPoints(source string, logContents []string, tags map[string]string, fields map[string]interface{}) []*point.Point {
	if len(logContents) == 0 {
		return nil
	}

	points := make([]*point.Point, 0, len(logContents))
	now := ntp.Now()

	for _, content := range logContents {
		opts := point.DefaultLoggingOptions()
		opts = append(opts, point.WithTime(now))

		pointFields := map[string]interface{}{
			pipeline.FieldMessage: content,
			pipeline.FieldStatus:  pipeline.DefaultStatus,
		}
		for k, v := range fields {
			pointFields[k] = v
		}

		pt := point.NewPoint(
			source,
			append(point.NewTags(tags), point.NewKVs(pointFields)...),
			opts...,
		)
		points = append(points, pt)
	}
	return points
}

func newDefaultInput() *Input {
	return &Input{
		Tags:    make(map[string]string),
		stopSem: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
