// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cat input
package cat

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	_ inputs.InputV2   = &Input{}
	_ inputs.HTTPInput = &Input{}
)

const (
	inputName    = "cat"
	sampleConfig = `
[[inputs.cat]]
  ## tcp port
  tcp_port = "2280"

  ##native or plaintext, datakit only support native(NT1) !!!
  decode = "native"

  ## This is default cat-client Kvs configs.
  startTransactionTypes = "Cache.;Squirrel."
  MatchTransactionTypes = "SQL"
  block = "false"
  routers = "127.0.0.1:2280;"
  sample = "1.0"

  ## global tags.
  # [inputs.cat.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

`
)

var (
	log            = logger.DefaultSLogger(inputName)
	afterGatherRun itrace.AfterGatherHandler
	globalTags     map[string]string
)

type Input struct {
	TCPPort               string            `toml:"tcp_port"`
	Decode                string            `toml:"decode"`
	StartTransactionTypes string            `toml:"startTransactionTypes"`
	MatchTransactionTypes string            `toml:"MatchTransactionTypes"`
	Block                 string            `toml:"block"`
	Routers               string            `toml:"routers"`
	Sample                string            `toml:"sample"`
	Tags                  map[string]string `toml:"tags"`

	feeder   dkio.Feeder
	listener net.Listener
	semStop  *cliutils.Sem // start stop signal
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

//nolint:lll
var kvs = []byte(`{"kvs":{"startTransactionTypes":"Cache.;Squirrel.","block":"false","routers":"127.0.0.1:2280;","sample":"1.0","matchTransactionTypes":"SQL"}}`)

func router(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write(kvs)
}

func (ipt *Input) RegHTTPHandler() {
	catKvs := defaultKVS()
	if ipt.StartTransactionTypes != "" {
		catKvs.Kvs.StartTransactionTypes = ipt.StartTransactionTypes
	}
	if ipt.MatchTransactionTypes != "" {
		catKvs.Kvs.MatchTransactionTypes = ipt.MatchTransactionTypes
	}
	if ipt.Block != "" {
		catKvs.Kvs.Block = ipt.Block
	}
	if ipt.Routers != "" {
		catKvs.Kvs.Routers = ipt.Routers
	}
	if ipt.Sample != "" {
		catKvs.Kvs.Sample = ipt.Sample
	}

	bts, err := json.Marshal(catKvs)
	if err != nil {
		log.Errorf("json marshal ipt.catKvs is err=%v", err)
	} else {
		kvs = bts
	}

	httpapi.RegHTTPHandler("get", "/cat/s/router", router)
}

func (ipt *Input) Run() {
	if ipt.feeder == nil {
		ipt.feeder = dkio.DefaultFeeder()
	}

	log = logger.SLogger(inputName)

	if len(ipt.Tags) > 0 {
		globalTags = ipt.Tags
	}
	traceOpts = append(point.DefaultLoggingOptions(), point.WithExtraTags(datakit.DefaultGlobalTagger().HostTags()))
	afterGather := itrace.NewAfterGather(itrace.WithLogger(log),
		itrace.WithPointOptions(point.WithExtraTags(datakit.GlobalHostTags())),
		itrace.WithFeeder(ipt.feeder))

	afterGatherRun = afterGather

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_cat"})
	g.Go(func(ctx context.Context) error {
		ipt.dotcp(ipt.TCPPort)
		return nil
	})
	log.Infof("Cat input started")

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.Terminate()
			log.Infof("%s exit", inputName)
			return
		case <-ipt.semStop.Wait():
			ipt.Terminate()
			log.Infof("%s return", inputName)
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.listener != nil {
		_ = ipt.listener.Close()
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{semStop: cliutils.NewSem()}
	})
}
