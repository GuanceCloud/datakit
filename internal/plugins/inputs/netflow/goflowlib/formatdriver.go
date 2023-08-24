// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package goflowlib

import (
	"context"
	"fmt"

	flowpb "github.com/netsampler/goflow2/pb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
)

// AggregatorFormatDriver is used as goflow formatter to forward flow data to aggregator/EP Forwarder.
type AggregatorFormatDriver struct {
	namespace string
	flowAggIn chan *common.Flow
}

// NewAggregatorFormatDriver returns a new AggregatorFormatDriver.
func NewAggregatorFormatDriver(flowAgg chan *common.Flow, namespace string) *AggregatorFormatDriver {
	return &AggregatorFormatDriver{
		namespace: namespace,
		flowAggIn: flowAgg,
	}
}

// Prepare desc.
func (d *AggregatorFormatDriver) Prepare() error {
	return nil
}

// Init desc.
func (d *AggregatorFormatDriver) Init(context.Context) error {
	return nil
}

// Format desc.
func (d *AggregatorFormatDriver) Format(data interface{}) ([]byte, []byte, error) {
	flow, ok := data.(*flowpb.FlowMessage)
	if !ok {
		return nil, nil, fmt.Errorf("message is not flowpb.FlowMessage")
	}
	d.flowAggIn <- ConvertFlow(flow, d.namespace)
	return nil, nil, nil
}
