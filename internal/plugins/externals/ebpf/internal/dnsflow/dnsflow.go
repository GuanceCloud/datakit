//go:build linux
// +build linux

// Package dnsflow collects eBPF-network dnsflow metrics
package dnsflow

import (
	"context"
	"errors"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/google/gopacket/afpacket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

const (
	srcNameM   = "dnsflow"
	inputName  = "ebpf-net/dnsflow"
	DNSTIMEOUT = time.Second * 6
)

var l = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

var k8sNetInfo *cli.K8sInfo

func SetK8sNetInfo(n *cli.K8sInfo) {
	k8sNetInfo = n
}

func NewDNSFlowTracer() *DNSFlowTracer {
	return &DNSFlowTracer{
		statsMap: map[DNSQAKey]DNSStats{},
		pInfoCh:  make(chan *DNSPacketInfo, 1024),
	}
}

type DNSFlowTracer struct {
	statsMap map[DNSQAKey]DNSStats
	pInfoCh  chan *DNSPacketInfo

	lastTPacketStats tpacketStatsSnapshot
}

type tpacketStatsSnapshot struct {
	packets uint64
	drops   uint64
	freezes uint64
}

func (s *tpacketStatsSnapshot) observe(component string, packets, drops, freezes uint64) {
	var (
		deltaPackets uint64
		deltaDrops   uint64
		deltaFreezes uint64
	)
	if packets >= s.packets {
		deltaPackets = packets - s.packets
	}
	if drops >= s.drops {
		deltaDrops = drops - s.drops
	}
	if freezes >= s.freezes {
		deltaFreezes = freezes - s.freezes
	}
	exporter.AddTPacketStats(component, deltaPackets, deltaDrops, deltaFreezes)
	s.packets = packets
	s.drops = drops
	s.freezes = freezes
}

func (tracer *DNSFlowTracer) updateDNSStats(packetInfo *DNSPacketInfo, dnsRecord *DNSAnswerRecord) *DNSStats {
	stats, ok := tracer.statsMap[packetInfo.Key]

	if !ok {
		if !packetInfo.QR { // query
			tracer.statsMap[packetInfo.Key] = DNSStats{
				TS:          packetInfo.TS,
				Timeout:     false,
				Responded:   false,
				RCODE:       -1,
				QueryDomain: packetInfo.QueryDomain,
				QueryType:   packetInfo.QueryType,
			}
			return &stats
		}
	} else {
		if packetInfo.QR { // answer
			if stats.QueryDomain == "" {
				stats.QueryDomain = packetInfo.QueryDomain
			}
			if stats.QueryType == "" {
				stats.QueryType = packetInfo.QueryType
			}
			stats.RespTime = packetInfo.TS.Sub(stats.TS)
			stats.RCODE = int(packetInfo.RCODE)
			stats.Timeout = false
			delete(tracer.statsMap, packetInfo.Key)
			if dnsRecord != nil && !stats.Responded {
				stats.Responded = true
				dnsRecord.addRecord(packetInfo)
				return &stats
			}
		}
	}
	return nil
}

func (tracer *DNSFlowTracer) checkTimeoutDNSQuery() map[DNSQAKey]DNSStats {
	qaStats := map[DNSQAKey]DNSStats{}
	for k, v := range tracer.statsMap {
		if !v.Responded && time.Since(v.TS) > DNSTIMEOUT {
			v.Responded = true
			v.Timeout = true
			qaStats[k] = v
			delete(tracer.statsMap, k)
		}
	}
	return qaStats
}

func (tracer *DNSFlowTracer) readPacket(ctx context.Context, tp *afpacket.TPacket) {
	dnsParser := NewDNSParse()
	for {
		dnsParser.layers = dnsParser.layers[:0]

		d, ci, err := tp.ZeroCopyReadPacketData()
		ts := ci.Timestamp
		if err != nil {
			select {
			case <-ctx.Done():
				tp.Close()
				return
			default:
			}
			if errors.Is(err, afpacket.ErrTimeout) {
				continue
			}
			continue
		}

		if err := dnsParser.DecodeLayers(d, &dnsParser.layers); err != nil {
			continue
		}

		pinfo, err := ReadPacketInfoFromDNSParser(ts, &dnsParser)
		if err != nil {
			continue
		}

		select {
		case <-ctx.Done():
			tp.Close()
			return
		case tracer.pInfoCh <- pinfo:
		default:
			l.Debug("pinfoCh full")
		}
	}
}

func (tracer *DNSFlowTracer) Run(ctx context.Context, tp *afpacket.TPacket,
	gTag map[string]string, dnsRecord *DNSAnswerRecord,
) {
	mCh := make(chan []*point.Point, 256)
	agg := FlowAgg{}
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, s3, err := tp.SocketStats(); err == nil {
					tracer.lastTPacketStats.observe("dnsflow",
						uint64(s3.Packets()), uint64(s3.Drops()), uint64(s3.QueueFreezes()))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	go tracer.readPacket(ctx, tp)
	go func() {
		t := time.NewTicker(time.Second * 30)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				exporter.ObserveCacheEntries("dnsflow", "pending_queries", len(tracer.statsMap))
				exporter.ObserveCacheEntries("dnsflow", "packet_queue", len(tracer.pInfoCh))
				stats := tracer.checkTimeoutDNSQuery()
				for k, v := range stats {
					err := agg.Append(k, v)
					if err != nil {
						l.Debug(err)
					}
				}

				exporter.ObserveAggEntries("dnsflow", agg.Len())
				flushStart := time.Now()
				pts := agg.ToPoint(gTag, k8sNetInfo)
				agg.Clean()
				exporter.ObserveAggEntries("dnsflow", 0)
				select {
				case mCh <- pts:
					exporter.ObserveAggFlush("dnsflow", len(pts), time.Since(flushStart), "ok")
				default:
					l.Warn("mCh full, drop data")
					exporter.ObserveAggFlush("dnsflow", len(pts), time.Since(flushStart), "drop_channel")
				}
			case pinfo := <-tracer.pInfoCh:
				if stats := tracer.updateDNSStats(pinfo, dnsRecord); stats != nil {
					err := agg.Append(pinfo.Key, *stats)
					if err != nil {
						l.Debug(err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-mCh:
			exporter.ObserveCacheEntries("dnsflow", "flush_queue", len(mCh))
			if len(m) == 0 {
				l.Debug("dnsflow: no data")
			} else if err := exporter.FeedPoint(inputName, point.Network, m); err != nil {
				l.Error(err)
			}
		}
	}
}
