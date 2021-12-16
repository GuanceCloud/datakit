//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package dnsflow

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket/afpacket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkfeed "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/feed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName  = "dnsflow"
	DNSTIMEOUT = time.Second * 6
)

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewDNSFlowTracer() *DNSFlowTracer {
	return &DNSFlowTracer{
		statsMap: map[DNSQAKey]DNSStats{},
		pInfoCh:  make(chan *DNSPacketInfo, 128),
	}
}

type DNSFlowTracer struct {
	statsMap map[DNSQAKey]DNSStats
	pInfoCh  chan *DNSPacketInfo
}

func (tracer *DNSFlowTracer) updateDNSStats(packetInfo *DNSPacketInfo, dnsRecord *DNSAnswerRecord) *DNSStats {
	stats, ok := tracer.statsMap[packetInfo.Key]

	if !ok {
		if !packetInfo.QR { // query
			tracer.statsMap[packetInfo.Key] = DNSStats{
				TS:        packetInfo.TS,
				Timeout:   false,
				Responded: false,
			}
		}
		return nil
	} else {
		if packetInfo.QR { // answer
			stats.Responded = true
			stats.RespTime = packetInfo.TS.Sub(stats.TS)
			stats.RCODE = packetInfo.RCODE
			stats.Timeout = false
			delete(tracer.statsMap, packetInfo.Key)
			if dnsRecord != nil && stats.RCODE == 0 {
				dnsRecord.addRecord(packetInfo)
			}
		} else {
			return nil
		}
	}

	return &stats
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
	for {
		dnsParser := NewDNSParse()

		d, ci, err := tp.ZeroCopyReadPacketData()
		ts := ci.Timestamp
		if err != nil {
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

func (tracer *DNSFlowTracer) Run(ctx context.Context, tp *afpacket.TPacket, gTag map[string]string,
	dnsRecord *DNSAnswerRecord, feedAddr string) {
	mCh := make(chan []inputs.Measurement)
	finishedStatsM := []inputs.Measurement{}
	go tracer.readPacket(ctx, tp)
	go func() {
		t := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-t.C:
				stats := tracer.checkTimeoutDNSQuery()
				for k, v := range stats {
					finishedStatsM = append(finishedStatsM, conv2M(k, v, gTag))
				}
				collectData := finishedStatsM
				finishedStatsM = make([]inputs.Measurement, 0)
				select {
				case mCh <- collectData:
				default:
					l.Warn("mCh full, drop data")
				}
			case pinfo := <-tracer.pInfoCh:
				serverIP := net.ParseIP(pinfo.Key.ServerIP)
				if serverIP == nil {
					continue
				} else if serverIP.IsLoopback() {
					continue
				}

				if stats := tracer.updateDNSStats(pinfo, dnsRecord); stats != nil {
					finishedStatsM = append(finishedStatsM, conv2M(pinfo.Key, *stats, gTag))
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
			if len(m) == 0 {
				l.Debug("dnsflow: no data")
			} else if err := dkfeed.FeedMeasurement(m, feedAddr); err != nil {
				l.Error(err)
			}
		}
	}
}

func conv2M(key DNSQAKey, stats DNSStats, gTags map[string]string) *measurement {
	m := measurement{
		ts:     stats.TS,
		name:   inputName,
		tags:   map[string]string{},
		fields: map[string]interface{}{},
	}
	for k, v := range gTags {
		m.tags[k] = v
	}
	m.tags["source"] = "dnsflow"
	m.tags["src_ip"] = key.ClientIP
	m.tags["src_port"] = fmt.Sprintf("%d", key.ClientPort)
	m.tags["dst_ip"] = key.ServerIP
	m.tags["dst_port"] = fmt.Sprintf("%d", key.ServerPort)
	if key.IsUDP {
		m.tags["transport"] = "udp"
	} else {
		m.tags["transport"] = "tcp"
	}
	if key.IsV4 {
		m.tags["family"] = "IPv4"
	} else {
		m.tags["family"] = "IPv6"
	}
	m.fields["timeout"] = stats.Timeout
	m.fields["rcode"] = int64(stats.RCODE)
	m.fields["resp_time"] = stats.RespTime.Nanoseconds()
	return &m
}
