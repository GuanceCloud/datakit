// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package goflowlib

import (
	flowpb "github.com/netsampler/goflow2/pb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
)

// ConvertFlow convert goflow flow structure to internal flow structure.
func ConvertFlow(srcFlow *flowpb.FlowMessage, namespace string) *common.Flow {
	return &common.Flow{
		Namespace:       namespace,
		FlowType:        convertFlowType(srcFlow.Type),
		SequenceNum:     srcFlow.SequenceNum,
		SamplingRate:    srcFlow.SamplingRate,
		Direction:       srcFlow.FlowDirection,
		ExporterAddr:    srcFlow.SamplerAddress, // Sampler is renamed to Exporter since it's a more commonly used
		StartTimestamp:  srcFlow.TimeFlowStart,
		EndTimestamp:    srcFlow.TimeFlowEnd,
		Bytes:           srcFlow.Bytes,
		Packets:         srcFlow.Packets,
		SrcAddr:         srcFlow.SrcAddr,
		DstAddr:         srcFlow.DstAddr,
		SrcMac:          srcFlow.SrcMac,
		DstMac:          srcFlow.DstMac,
		SrcMask:         srcFlow.SrcNet,
		DstMask:         srcFlow.DstNet,
		EtherType:       srcFlow.Etype,
		IPProtocol:      srcFlow.Proto,
		SrcPort:         int32(srcFlow.SrcPort),
		DstPort:         int32(srcFlow.DstPort),
		InputInterface:  srcFlow.InIf,
		OutputInterface: srcFlow.OutIf,
		Tos:             srcFlow.IpTos,
		NextHop:         srcFlow.NextHop,
		TCPFlags:        srcFlow.TcpFlags,
	}
}

func convertFlowType(flowType flowpb.FlowMessage_FlowType) common.FlowType {
	var flowTypeStr common.FlowType
	//nolint:exhaustive
	switch flowType {
	case flowpb.FlowMessage_SFLOW_5:
		flowTypeStr = common.TypeSFlow5
	case flowpb.FlowMessage_NETFLOW_V5:
		flowTypeStr = common.TypeNetFlow5
	case flowpb.FlowMessage_NETFLOW_V9:
		flowTypeStr = common.TypeNetFlow9
	case flowpb.FlowMessage_IPFIX:
		flowTypeStr = common.TypeIPFIX
	default:
		flowTypeStr = common.TypeUnknown
	}
	return flowTypeStr
}
