// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package flowaggregator

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/enrichment"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/payload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/portrollup"
)

func buildPayload(aggFlow *common.Flow, hostname string, flushTime time.Time) payload.FlowPayload {
	return payload.FlowPayload{
		// TODO: Implement Tos
		FlushTimestamp: flushTime.UnixMilli(),
		FlowType:       string(aggFlow.FlowType),
		SamplingRate:   aggFlow.SamplingRate,
		Direction:      enrichment.RemapDirection(aggFlow.Direction),
		Device: payload.Device{
			Namespace: aggFlow.Namespace,
		},
		Exporter: payload.Exporter{
			IP: common.IPBytesToString(aggFlow.ExporterAddr),
		},
		Start:      aggFlow.StartTimestamp,
		End:        aggFlow.EndTimestamp,
		Bytes:      aggFlow.Bytes,
		Packets:    aggFlow.Packets,
		EtherType:  enrichment.MapEtherType(aggFlow.EtherType),
		IPProtocol: enrichment.MapIPProtocol(aggFlow.IPProtocol),
		Source: payload.Endpoint{
			IP:   common.IPBytesToString(aggFlow.SrcAddr),
			Port: portrollup.PortToString(aggFlow.SrcPort),
			Mac:  enrichment.FormatMacAddress(aggFlow.SrcMac),
			Mask: enrichment.FormatMask(aggFlow.SrcAddr, aggFlow.SrcMask),
		},
		Destination: payload.Endpoint{
			IP:   common.IPBytesToString(aggFlow.DstAddr),
			Port: portrollup.PortToString(aggFlow.DstPort),
			Mac:  enrichment.FormatMacAddress(aggFlow.DstMac),
			Mask: enrichment.FormatMask(aggFlow.DstAddr, aggFlow.DstMask),
		},
		Ingress: payload.ObservationPoint{
			Interface: payload.Interface{
				Index: aggFlow.InputInterface,
			},
		},
		Egress: payload.ObservationPoint{
			Interface: payload.Interface{
				Index: aggFlow.OutputInterface,
			},
		},
		Host:     hostname,
		TCPFlags: enrichment.FormatFCPFlags(aggFlow.TCPFlags),
		NextHop: payload.NextHop{
			IP: common.IPBytesToString(aggFlow.NextHop),
		},
	}
}
