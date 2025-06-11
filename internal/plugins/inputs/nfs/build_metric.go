// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package nfs

import (
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/procfs"
	"golang.org/x/sys/unix"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) buildBaseMetric() ([]*point.Point, error) {
	clientRPCStats, err := getClientRPCStats()
	if err != nil {
		return []*point.Point{}, err
	}
	ms := []inputs.MeasurementV2{}
	m := &baseMeasurement{
		name: "nfs",
		ts:   ipt.ptsTime.UnixNano(),
		tags: map[string]string{},
	}

	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	m.fields = map[string]interface{}{
		"tcp_packets_total":                  clientRPCStats.Network.TCPCount,
		"udp_packets_total":                  clientRPCStats.Network.UDPCount,
		"connections_total":                  clientRPCStats.Network.TCPConnect,
		"rpcs_total":                         clientRPCStats.ClientRPC.RPCCount,
		"rpc_retransmissions_total":          clientRPCStats.ClientRPC.Retransmissions,
		"rpc_authentication_refreshes_total": clientRPCStats.ClientRPC.AuthRefreshes,
	}
	ms = append(ms, m)
	pts := getPointsFromMeasurement(ms)

	if requestv2StatsPts, err := collectNFSRequestsv2Stats(&clientRPCStats.V2Stats, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv2StatsPts...)
	}

	if requestv3StatsPts, err := collectNFSRequestsv3Stats(&clientRPCStats.V3Stats, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv3StatsPts...)
	}

	if requestv4StatsPts, err := collectNFSRequestsv4Stats(&clientRPCStats.ClientV4Stats, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv4StatsPts...)
	}

	return pts, nil
}

func (ipt *Input) buildNFSdMetric() ([]*point.Point, error) {
	serverRPCStats, err := getServerRPCStats()
	if err != nil {
		return []*point.Point{}, err
	}
	ms := []inputs.MeasurementV2{}
	m := &nfsdMeasurement{
		name: "nfsd",
		ts:   ipt.ptsTime.UnixNano(),
		tags: map[string]string{},
	}

	for key, value := range ipt.Tags {
		m.tags[key] = value
	}

	m.fields = map[string]interface{}{
		"tcp_packets_total":                serverRPCStats.Network.TCPCount,
		"udp_packets_total":                serverRPCStats.Network.UDPCount,
		"connections_total":                serverRPCStats.Network.TCPConnect,
		"reply_cache_hits_total":           serverRPCStats.ReplyCache.Hits,
		"reply_cache_misses_total":         serverRPCStats.ReplyCache.Misses,
		"reply_cache_nocache_total":        serverRPCStats.ReplyCache.NoCache,
		"file_handles_stale_total":         serverRPCStats.FileHandles.Stale,
		"disk_bytes_read_total":            serverRPCStats.InputOutput.Read,
		"disk_bytes_written_total":         serverRPCStats.InputOutput.Write,
		"server_threads":                   serverRPCStats.Threads.Threads,
		"read_ahead_cache_size_blocks":     serverRPCStats.ReadAheadCache.CacheSize,
		"read_ahead_cache_not_found_total": serverRPCStats.ReadAheadCache.NotFound,
		"server_rpcs_total":                serverRPCStats.ServerRPC.RPCCount,
	}
	ms = append(ms, m)
	pts := getPointsFromMeasurement(ms)

	if requestv2StatsPts, err := collectNFSdRequestsv2Stats(&serverRPCStats.V2Stats, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv2StatsPts...)
	}
	if requestv3StatsPts, err := collectNFSdRequestsv3Stats(&serverRPCStats.V3Stats, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv3StatsPts...)
	}
	if requestv4StatsPts, err := collectNFSdRequestsv4Stats(&serverRPCStats.V4Ops, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, requestv4StatsPts...)
	}

	if nfsdServerRPCStatsPts, err := collectNFSdServerRPCStats(serverRPCStats.ServerRPC, ipt.ptsTime.UnixNano()); err == nil {
		pts = append(pts, nfsdServerRPCStatsPts...)
	}

	return pts, nil
}

func (ipt *Input) buildMountStats() ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}
	mounts, err := getMountstatsInfo()
	if err != nil {
		return []*point.Point{}, err
	}
	for _, mount := range mounts {
		m := &mountstatsMeasurement{
			name: "nfs_mountstats",
			tags: map[string]string{
				"device":     mount.Device,
				"mountpoint": mount.Mount,
				"type":       mount.Type,
			},
			ts: ipt.ptsTime.UnixNano(),
		}

		for key, value := range ipt.Tags {
			m.tags[key] = value
		}

		if mount.Stats == nil {
			continue
		}

		mountStatsNfs := mount.Stats.(*procfs.MountStatsNFS)
		m.tags["protocol"] = mountStatsNfs.Transport.Protocol
		m.fields = map[string]interface{}{
			"age_seconds_total": mountStatsNfs.Age.Seconds(),
		}
		var stat unix.Statfs_t
		if err := unix.Statfs(mount.Mount, &stat); err == nil {
			// 计算总大小和已使用大小
			totalBytes := stat.Blocks * uint64(stat.Bsize)
			freeBytes := stat.Bavail * uint64(stat.Bsize)
			usedBytes := totalBytes - freeBytes

			m.fields["fs_size"] = totalBytes
			m.fields["fs_avail"] = freeBytes
			m.fields["fs_used"] = usedBytes
			m.fields["fs_used_percent"] = (float64(usedBytes) / float64(totalBytes)) * 100
		}

		if ipt.MountstatsMetric.Rw {
			// rw
			m.fields["read_bytes_total"] = mountStatsNfs.Bytes.Read
			m.fields["write_bytes_total"] = mountStatsNfs.Bytes.Write
			m.fields["direct_read_bytes_total"] = mountStatsNfs.Bytes.DirectRead
			m.fields["direct_write_bytes_total"] = mountStatsNfs.Bytes.DirectWrite
			m.fields["total_read_bytes_total"] = mountStatsNfs.Bytes.ReadTotal
			m.fields["total_write_bytes_total"] = mountStatsNfs.Bytes.WriteTotal
			m.fields["read_pages_total"] = mountStatsNfs.Bytes.ReadPages
			m.fields["write_pages_total"] = mountStatsNfs.Bytes.WritePages
		}

		if ipt.MountstatsMetric.Transport {
			// transport
			m.fields["transport_bind_total"] = mountStatsNfs.Transport.Bind
			m.fields["transport_connect_total"] = mountStatsNfs.Transport.Connect
			m.fields["transport_idle_time_seconds"] = mountStatsNfs.Transport.IdleTimeSeconds
			m.fields["transport_sends_total"] = mountStatsNfs.Transport.Sends
			m.fields["transport_receives_total"] = mountStatsNfs.Transport.Receives
			m.fields["transport_bad_transaction_ids_total"] = mountStatsNfs.Transport.BadTransactionIDs
			m.fields["transport_backlog_queue_total"] = mountStatsNfs.Transport.CumulativeBacklog
			m.fields["transport_maximum_rpc_slots"] = mountStatsNfs.Transport.MaximumRPCSlotsUsed
			m.fields["transport_sending_queue_total"] = mountStatsNfs.Transport.CumulativeSendingQueue
			m.fields["transport_pending_queue_total"] = mountStatsNfs.Transport.CumulativePendingQueue
		}

		if ipt.MountstatsMetric.Event {
			// event
			m.fields["event_inode_revalidate_total"] = mountStatsNfs.Events.InodeRevalidate
			m.fields["event_dnode_revalidate_total"] = mountStatsNfs.Events.DnodeRevalidate
			m.fields["event_data_invalidate_total"] = mountStatsNfs.Events.DataInvalidate
			m.fields["event_attribute_invalidate_total"] = mountStatsNfs.Events.AttributeInvalidate
			m.fields["event_vfs_open_total"] = mountStatsNfs.Events.VFSOpen
			m.fields["event_vfs_lookup_total"] = mountStatsNfs.Events.VFSLookup
			m.fields["event_vfs_access_total"] = mountStatsNfs.Events.VFSAccess
			m.fields["event_vfs_update_page_total"] = mountStatsNfs.Events.VFSUpdatePage
			m.fields["event_vfs_read_page_total"] = mountStatsNfs.Events.VFSReadPage
			m.fields["event_vfs_read_pages_total"] = mountStatsNfs.Events.VFSReadPages
			m.fields["event_vfs_write_page_total"] = mountStatsNfs.Events.VFSWritePage
			m.fields["event_vfs_write_pages_total"] = mountStatsNfs.Events.VFSWritePages
			m.fields["event_vfs_getdents_total"] = mountStatsNfs.Events.VFSGetdents
			m.fields["event_vfs_setattr_total"] = mountStatsNfs.Events.VFSSetattr
			m.fields["event_vfs_flush_total"] = mountStatsNfs.Events.VFSFlush
			m.fields["event_vfs_fsync_total"] = mountStatsNfs.Events.VFSFsync
			m.fields["event_vfs_lock_total"] = mountStatsNfs.Events.VFSLock
			m.fields["event_vfs_file_release_total"] = mountStatsNfs.Events.VFSFileRelease
			m.fields["event_truncation_total"] = mountStatsNfs.Events.Truncation
			m.fields["event_write_extension_total"] = mountStatsNfs.Events.WriteExtension
			m.fields["event_silly_rename_total"] = mountStatsNfs.Events.SillyRename
			m.fields["event_short_read_total"] = mountStatsNfs.Events.ShortRead
			m.fields["event_short_write_total"] = mountStatsNfs.Events.ShortWrite
			m.fields["event_jukebox_delay_total"] = mountStatsNfs.Events.JukeboxDelay
			m.fields["event_pnfs_read_total"] = mountStatsNfs.Events.PNFSRead
			m.fields["event_pnfs_write_total"] = mountStatsNfs.Events.PNFSWrite
		}

		ms = append(ms, m)

		// operations
		if ipt.MountstatsMetric.Operations {
			if opMs, err := collectMountStatsOperation(mount, mountStatsNfs.Operations, ipt.ptsTime.UnixNano()); err == nil {
				ms = append(ms, opMs...)
			}
		}
	}

	pts := getPointsFromMeasurement(ms)

	return pts, nil
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*point.Point {
	pts := []*point.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}
