// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package netflow collect flow records from your NetFlow-enabled devices.
package netflow

import (
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/flowaggregator"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/goflowlib"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/metrics"
)

////////////////////////////////////////////////////////////////////////////////

const (
	configSample = `
[[inputs.netflow]]
    source    = "netflow"
    namespace = "namespace"

    #[[inputs.netflow.listeners]]
    #    flow_type = "netflow9"
    #    port      = 2055

    [[inputs.netflow.listeners]]
        flow_type = "netflow5"
        port      = 2056

    #[[inputs.netflow.listeners]]
    #    flow_type = "ipfix"
    #    port      = 4739

    #[[inputs.netflow.listeners]]
    #    flow_type = "sflow5"
    #    port      = 6343

    [inputs.netflow.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var (
	l = logger.DefaultSLogger(common.InputName)

	_ inputs.InputV2   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

type Input struct {
	Source    string            `toml:"source"`
	Namespace string            `toml:"namespace"`
	Listeners []common.FlowOpt  `toml:"listeners,omitempty"`
	Tags      map[string]string `toml:"tags"`

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
}

func setLogger() {
	l = logger.SLogger(common.InputName)

	flowaggregator.SetLogger(l)
	goflowlib.SetLogger(l)
}

func (ipt *Input) Run() {
	setLogger()
	if len(ipt.Source) == 0 {
		ipt.Source = metrics.DefaultSource
	}

	if err := StartServer(ipt); err != nil {
		l.Errorf("Failed to start NetFlow server: %s", err)
		return
	}

	select {
	case <-datakit.Exit.Wait():
		ipt.exit()
		l.Info(common.InputName + " input exit")
		return
	case <-ipt.semStop.Wait():
		ipt.exit()
		l.Info(common.InputName + " input return")
		return
	}
}

////////////////////////////////////////////////////////////////////////////////

// netflowListener contains state of goflow listener and the related netflow config
// flowState can be of type *utils.StateNetFlow/StateSFlow/StateNFLegacy.
type netflowListener struct {
	flowState *goflowlib.FlowStateWrapper
	config    config.ListenerConfig
}

// Shutdown will close the goflow listener state.
func (l *netflowListener) shutdown() {
	l.flowState.Shutdown()
}

//nolint:lll
func startFlowListener(listenerConfig config.ListenerConfig, flowAgg *flowaggregator.FlowAggregator) (*netflowListener, error) {
	flowState, err := goflowlib.StartFlowRoutine(listenerConfig.FlowType, listenerConfig.BindHost, listenerConfig.Port, listenerConfig.Workers, listenerConfig.Namespace, flowAgg.GetFlowInChan())
	if err != nil {
		return nil, err
	}
	return &netflowListener{
		flowState: flowState,
		config:    listenerConfig,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

var serverInstance *Server

const (
	stopTimeout = 2
)

// Server manages netflow listeners.
type Server struct {
	Addr      string
	listeners []*netflowListener
	flowAgg   *flowaggregator.FlowAggregator
}

func getFlows(ipt *Input) []*common.FlowOpt {
	var out []*common.FlowOpt

	for _, v := range ipt.Listeners {
		if _, err := common.GetFlowTypeByName(v.Type); err != nil {
			continue
		}
		if v.Port == 0 {
			continue
		}

		out = append(out, &common.FlowOpt{
			Type: v.Type,
			Port: v.Port,
		})
	}

	return out
}

// NewNetflowServer configures and returns a running SNMP traps server.
//
//nolint:lll
func NewNetflowServer(ipt *Input) (*Server, error) {
	var listeners []*netflowListener

	flows := getFlows(ipt)

	mainConfig, err := config.ReadConfig(flows, ipt.Namespace)
	if err != nil {
		return nil, err
	}

	combinedTags := internal.CopyMapString(ipt.Tags)
	// global tags(host/election tags) got the lowest priority.
	for k, v := range ipt.tagger.HostTags() {
		if _, ok := combinedTags[k]; !ok {
			combinedTags[k] = v
		}
	}

	flowAgg := flowaggregator.NewFlowAggregator(mainConfig, combinedTags, ipt.feeder, ipt.Source)
	go flowAgg.Start()

	l.Debugf("NetFlow Server configs (aggregator_buffer_size=%d, aggregator_flush_interval=%d, aggregator_flow_context_ttl=%d)", mainConfig.AggregatorBufferSize, mainConfig.AggregatorFlushInterval, mainConfig.AggregatorFlowContextTTL)
	for _, listenerConfig := range mainConfig.Listeners {
		l.Infof("Starting Netflow listener for flow type %s on %s", listenerConfig.FlowType, listenerConfig.Addr())
		listener, err := startFlowListener(listenerConfig, flowAgg)
		if err != nil {
			l.Warnf("Error starting listener for config (flow_type:%s, bind_Host:%s, port:%d): %s", listenerConfig.FlowType, listenerConfig.BindHost, listenerConfig.Port, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return &Server{
		listeners: listeners,
		flowAgg:   flowAgg,
	}, nil
}

// Stop stops the Server.
func (svr *Server) stop() {
	l.Infof("Stop NetFlow Server")

	svr.flowAgg.Stop()

	for _, listener := range svr.listeners {
		stopped := make(chan interface{})

		go func() {
			l.Infof("Listener `%s` shutting down", listener.config.Addr())
			listener.shutdown()
			close(stopped)
		}()

		select {
		case <-stopped:
			l.Infof("Listener `%s` stopped", listener.config.Addr())
		case <-time.After(stopTimeout * time.Second):
			l.Errorf("Stopping listener `%s`. Timeout after %d seconds", listener.config.Addr(), stopTimeout)
		}
	}
}

// StartServer starts the global NetFlow collector.
func StartServer(ipt *Input) error {
	server, err := NewNetflowServer(ipt)
	if err != nil {
		return err
	}
	serverInstance = server
	return nil
}

// StopServer stops the netflow server, if it is running.
func StopServer() {
	if serverInstance != nil {
		serverInstance.stop()
		serverInstance = nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func (*Input) Catalog() string {
	return common.InputName
}

func (*Input) SampleConfig() string {
	return configSample
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&metrics.NetflowMeasurement{},
	}
}

func (*Input) Singleton() {}

func (*Input) exit() {
	StopServer()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		feeder:  dkio.DefaultFeeder(),
		semStop: cliutils.NewSem(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(common.InputName, func() inputs.Input {
		return defaultInput()
	})
}

////////////////////////////////////////////////////////////////////////////////
