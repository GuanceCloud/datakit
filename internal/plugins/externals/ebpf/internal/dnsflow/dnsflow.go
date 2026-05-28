//go:build linux
// +build linux

// Package dnsflow collects eBPF-network dnsflow metrics
package dnsflow

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
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

const (
	defaultPendingQueryLimit = 65536
	maxPendingQueryLimit     = 1_000_000
	pendingQueryLimitEnv     = "DK_EBPF_DNSFLOW_PENDING_QUERY_LIMIT"
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
		statsMap:          map[DNSQAKey]DNSStats{},
		pInfoCh:           make(chan *DNSPacketInfo, 1024),
		pendingQueryLimit: dnsPendingQueryLimit(),
	}
}

type DNSFlowTracer struct {
	statsMap          map[DNSQAKey]DNSStats
	pInfoCh           chan *DNSPacketInfo
	pendingQueryLimit int

	lastTPacketStats tpacketStatsSnapshot
}

func dnsPendingQueryLimit() int {
	raw := strings.TrimSpace(os.Getenv(pendingQueryLimitEnv))
	if raw == "" {
		return defaultPendingQueryLimit
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		l.Warnf("invalid %s=%q, use default %d", pendingQueryLimitEnv, raw, defaultPendingQueryLimit)
		return defaultPendingQueryLimit
	}
	if limit > maxPendingQueryLimit {
		return maxPendingQueryLimit
	}
	return limit
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
	if tracer == nil || packetInfo == nil {
		return nil
	}
	if tracer.statsMap == nil {
		tracer.statsMap = map[DNSQAKey]DNSStats{}
	}

	stats, ok := tracer.statsMap[packetInfo.Key]

	if !ok {
		if !packetInfo.QR { // query
			if tracer.pendingQueryLimit <= 0 {
				tracer.pendingQueryLimit = defaultPendingQueryLimit
			}
			if len(tracer.statsMap) >= tracer.pendingQueryLimit {
				timeoutStats := tracer.checkTimeoutDNSQuery()
				if len(timeoutStats) > 0 {
					exporter.AddCacheEvictions("dnsflow", "pending_queries", "timeout_on_limit", len(timeoutStats))
				}
			}
			if len(tracer.statsMap) >= tracer.pendingQueryLimit {
				exporter.IncBPFEventDrop("dnsflow", "query", "pending_query_limit")
				return nil
			}
			tracer.statsMap[packetInfo.Key] = DNSStats{
				TS:          packetInfo.TS,
				Timeout:     false,
				Responded:   false,
				RCODE:       -1,
				QueryDomain: packetInfo.QueryDomain,
				QueryType:   packetInfo.QueryType,
			}
			return nil
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
			if !stats.Responded {
				stats.Responded = true
				if dnsRecord != nil {
					dnsRecord.addRecord(packetInfo)
				}
				return &stats
			}
		}
	}
	return nil
}

func (tracer *DNSFlowTracer) checkTimeoutDNSQuery() map[DNSQAKey]DNSStats {
	qaStats := map[DNSQAKey]DNSStats{}
	if tracer == nil {
		return qaStats
	}
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

func (tracer *DNSFlowTracer) appendTimeoutDNSStats(agg *FlowAgg) {
	if tracer == nil || agg == nil {
		return
	}
	stats := tracer.checkTimeoutDNSQuery()
	for k, v := range stats {
		err := agg.Append(k, v)
		if err != nil {
			l.Debug(err)
		}
	}
}

func (tracer *DNSFlowTracer) readPacket(ctx context.Context, tp *afpacket.TPacket) {
	if tracer == nil || tp == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if tracer.pInfoCh == nil {
		tracer.pInfoCh = make(chan *DNSPacketInfo, 1024)
	}

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
			exporter.IncBPFEventDrop("dnsflow", "packet", "queue_full")
		}
	}
}

func (tracer *DNSFlowTracer) Run(ctx context.Context, tp *afpacket.TPacket,
	gTag map[string]string, dnsRecord *DNSAnswerRecord,
) {
	if tracer == nil || tp == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if tracer.pInfoCh == nil {
		tracer.pInfoCh = make(chan *DNSPacketInfo, 1024)
	}
	if tracer.statsMap == nil {
		tracer.statsMap = map[DNSQAKey]DNSStats{}
	}

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
		flushTicker := time.NewTicker(time.Second * 30)
		timeoutTicker := time.NewTicker(DNSTIMEOUT)
		defer flushTicker.Stop()
		defer timeoutTicker.Stop()
		for {
			select {
			case <-timeoutTicker.C:
				tracer.appendTimeoutDNSStats(&agg)
				exporter.ObserveAggEntries("dnsflow", agg.Len())
			case <-flushTicker.C:
				exporter.ObserveCacheEntries("dnsflow", "pending_queries", len(tracer.statsMap))
				exporter.ObserveCacheEntries("dnsflow", "packet_queue", len(tracer.pInfoCh))
				tracer.appendTimeoutDNSStats(&agg)

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
				if pinfo == nil {
					exporter.IncBPFEventDrop("dnsflow", "packet", "nil_info")
					continue
				}
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
