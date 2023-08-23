// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

// Package config contains netflow configuration.
package config

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
)

// NetflowConfig contains configuration for NetFlow collector.
type NetflowConfig struct {
	Listeners                     []ListenerConfig
	StopTimeout                   int
	AggregatorBufferSize          int
	AggregatorFlushInterval       int
	AggregatorFlowContextTTL      int
	AggregatorPortRollupThreshold int
	AggregatorPortRollupDisabled  bool

	// AggregatorRollupTrackerRefreshInterval is useful to speed up testing to avoid wait for 1h default
	AggregatorRollupTrackerRefreshInterval uint

	PrometheusListenerAddress string // Example `localhost:9090`
	PrometheusListenerEnabled bool
}

// ListenerConfig contains configuration for a single flow listener.
type ListenerConfig struct {
	FlowType  common.FlowType
	Port      uint16
	BindHost  string
	Workers   int
	Namespace string
}

// ReadConfig builds and returns configuration from Agent configuration.
//
//nolint:lll
func ReadConfig(flows []*common.FlowOpt, namespace string) (*NetflowConfig, error) {
	var mainConfig NetflowConfig

	for _, flow := range flows {
		mainConfig.Listeners = append(mainConfig.Listeners, ListenerConfig{
			FlowType:  flow.Type,
			Port:      flow.Port,
			Namespace: namespace,
		})
	}

	for i := range mainConfig.Listeners {
		listenerConfig := &mainConfig.Listeners[i]

		flowType, err := common.GetFlowTypeByName(listenerConfig.FlowType)
		if err != nil {
			return nil, fmt.Errorf("the provided flow type `%s` is not valid (valid flow types: %v)", listenerConfig.FlowType, common.GetAllFlowTypes())
		}

		if listenerConfig.Port == 0 {
			listenerConfig.Port = flowType.DefaultPort()
			if listenerConfig.Port == 0 {
				return nil, fmt.Errorf("no default port found for `%s`, a valid port must be set", listenerConfig.FlowType)
			}
		}
		if listenerConfig.BindHost == "" {
			listenerConfig.BindHost = common.DefaultBindHost
		}
		if listenerConfig.Workers == 0 {
			listenerConfig.Workers = 1
		}
		if listenerConfig.Namespace == "" {
			listenerConfig.Namespace = common.DefaultNamespace
		}
		normalizedNamespace, err := dkstring.NormalizeNamespace(listenerConfig.Namespace)
		if err != nil {
			return nil, fmt.Errorf("invalid namespace `%s` error: %w", listenerConfig.Namespace, err)
		}
		listenerConfig.Namespace = normalizedNamespace
	}

	if mainConfig.StopTimeout == 0 {
		mainConfig.StopTimeout = common.DefaultStopTimeout
	}
	if mainConfig.AggregatorFlushInterval == 0 {
		mainConfig.AggregatorFlushInterval = common.DefaultAggregatorFlushInterval
	}
	if mainConfig.AggregatorFlowContextTTL == 0 {
		// Set AggregatorFlowContextTTL to AggregatorFlushInterval to keep flow context around
		// for 1 flush-interval time after a flush.
		mainConfig.AggregatorFlowContextTTL = mainConfig.AggregatorFlushInterval
	}
	if mainConfig.AggregatorBufferSize == 0 {
		mainConfig.AggregatorBufferSize = common.DefaultAggregatorBufferSize
	}
	if mainConfig.AggregatorPortRollupThreshold == 0 {
		mainConfig.AggregatorPortRollupThreshold = common.DefaultAggregatorPortRollupThreshold
	}
	if mainConfig.AggregatorRollupTrackerRefreshInterval == 0 {
		mainConfig.AggregatorRollupTrackerRefreshInterval = common.DefaultAggregatorRollupTrackerRefreshInterval
	}

	if mainConfig.PrometheusListenerAddress == "" {
		mainConfig.PrometheusListenerAddress = common.DefaultPrometheusListenerAddress
	}

	return &mainConfig, nil
}

// Addr returns the host:port address to listen on.
func (c *ListenerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.BindHost, c.Port)
}
