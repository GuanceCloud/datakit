//go:build linux
// +build linux

package l4log

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"

type tpacketStatsSnapshot struct {
	packets uint64
	drops   uint64
	freezes uint64
}

func tpacketStatsDelta(last *tpacketStatsSnapshot, packets, drops, freezes uint64) tpacketStatsSnapshot {
	if last == nil {
		return tpacketStatsSnapshot{
			packets: packets,
			drops:   drops,
			freezes: freezes,
		}
	}

	delta := tpacketStatsSnapshot{}

	if packets >= last.packets {
		delta.packets = packets - last.packets
	}
	if drops >= last.drops {
		delta.drops = drops - last.drops
	}
	if freezes >= last.freezes {
		delta.freezes = freezes - last.freezes
	}

	last.packets = packets
	last.drops = drops
	last.freezes = freezes

	return delta
}

func observeTPacketStatsDelta(component string, last *tpacketStatsSnapshot, packets, drops, freezes uint64) {
	delta := tpacketStatsDelta(last, packets, drops, freezes)
	exporter.AddTPacketStats(component, delta.packets, delta.drops, delta.freezes)
}
