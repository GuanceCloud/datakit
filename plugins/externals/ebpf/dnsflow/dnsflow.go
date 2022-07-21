//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package dnsflow

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket/afpacket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/output"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	srcNameM   = "dnsflow"
	DNSTIMEOUT = time.Second * 6
)

var l = logger.DefaultSLogger("ebpf")

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
	dnsRecord *DNSAnswerRecord, feedAddr string,
) {
	mCh := make(chan []*point.Point)
	finishedStatsM := []*point.Point{}
	go tracer.readPacket(ctx, tp)
	go func() {
		t := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-t.C:
				stats := tracer.checkTimeoutDNSQuery()
				for k, v := range stats {
					if pt, err := conv2M(k, v, gTag); err != nil {
						l.Error(err)
					} else {
						finishedStatsM = append(finishedStatsM, pt)
					}
				}
				collectData := finishedStatsM
				finishedStatsM = []*point.Point{}
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
					if pt, err := conv2M(pinfo.Key, *stats, gTag); err != nil {
						l.Error(err)
					} else {
						finishedStatsM = append(finishedStatsM, pt)
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
			if len(m) == 0 {
				l.Debug("dnsflow: no data")
			} else if err := dkout.FeedMeasurement(feedAddr, m); err != nil {
				l.Error(err)
			}
		}
	}
}

func conv2M(key DNSQAKey, stats DNSStats, gTags map[string]string) (*point.Point, error) {
	mTags := map[string]string{}
	// ts:     stats.TS,

	for k, v := range gTags {
		mTags[k] = v
	}
	mTags["src_ip"] = key.ClientIP
	mTags["src_port"] = fmt.Sprintf("%d", key.ClientPort)
	mTags["dst_ip"] = key.ServerIP
	mTags["dst_port"] = fmt.Sprintf("%d", key.ServerPort)
	if key.IsUDP {
		mTags["transport"] = "udp"
	} else {
		mTags["transport"] = "tcp"
	}
	if key.IsV4 {
		mTags["family"] = "IPv4"
	} else {
		mTags["family"] = "IPv6"
	}
	mFields := map[string]interface{}{
		"timeout":   stats.Timeout,
		"rcode":     int64(stats.RCODE),
		"resp_time": stats.RespTime.Nanoseconds(),
	}

	return point.NewPoint(srcNameM, mTags, mFields, inputs.OptNetwork)
}
