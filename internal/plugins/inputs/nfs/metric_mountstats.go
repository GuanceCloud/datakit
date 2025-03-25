// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:deadcode,unused
package nfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/procfs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type MountstatsMetric struct {
	Rw         bool `toml:"rw"`
	Transport  bool `toml:"transport"`
	Event      bool `toml:"event"`
	Operations bool `toml:"operations"`
}

type mountstatsMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

//nolint:lll
func (m *mountstatsMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name: "nfs_mountstats",
		Type: "metric",
		Fields: map[string]interface{}{
			// base
			"fs_avail": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Available space on the filesystem.",
			},
			"fs_size": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Total size of the filesystem.",
			},
			"fs_used": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Used space on the filesystem.",
			},
			"fs_used_percent": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.Percent,
				Desc:     "Percentage of used space on the filesystem.",
			},
			"age_seconds_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.DurationSecond,
				Desc:     "The age of the NFS mount in seconds.",
			},
			// read and write total statistics
			"read_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes read using the read() syscall.",
			},
			"write_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes written using the write() syscall.",
			},
			"direct_read_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes read using the read() syscall in O_DIRECT mode.",
			},
			"direct_write_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes written using the write() syscall in O_DIRECT mode.",
			},
			"total_read_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes read from the NFS server, in total.",
			},
			"total_write_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes written to the NFS server, in total.",
			},
			"read_pages_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read directly via mmap()'d files.",
			},
			"write_pages_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of pages written directly via mmap()'d files.",
			},

			// transport
			"transport_bind_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the client has had to establish a connection from scratch to the NFS server.",
			},
			"transport_connect_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the client has made a TCP connection to the NFS server.",
			},
			"transport_idle_time_seconds": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.DurationSecond,
				Desc:     "Duration since the NFS mount last saw any RPC traffic, in seconds.",
			},
			"transport_sends_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of RPC requests for this mount sent to the NFS server.",
			},
			"transport_receives_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of RPC responses for this mount received from the NFS server.",
			},
			"transport_bad_transaction_ids_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the NFS server sent a response with a transaction ID unknown to this client.",
			},
			"transport_backlog_queue_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of items added to the RPC backlog queue.",
			},
			"transport_maximum_rpc_slots": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Maximum number of simultaneously active RPC requests ever used.",
			},
			"transport_sending_queue_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of items added to the RPC transmission sending queue.",
			},
			"transport_pending_queue_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of items added to the RPC transmission pending queue.",
			},

			// event
			"event_inode_revalidate_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times cached inode attributes are re-validated from the server.",
			},
			"event_dnode_revalidate_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times cached dentry nodes are re-validated from the server.",
			},
			"event_data_invalidate_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times an inode cache is cleared.",
			},
			"event_attribute_invalidate_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times cached inode attributes are invalidated.",
			},
			"event_vfs_open_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times cached inode attributes are invalidated.",
			},
			"event_vfs_lookup_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a directory lookup has occurred.",
			},
			"event_vfs_access_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times permissions have been checked.",
			},
			"event_vfs_update_page_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of updates (and potential writes) to pages.",
			},
			"event_vfs_read_page_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of pages read directly via mmap()'d files.",
			},
			"event_vfs_read_pages_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a group of pages have been read.",
			},
			"event_vfs_write_page_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of pages written directly via mmap()'d files.",
			},
			"event_vfs_write_pages_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a group of pages have been written.",
			},
			"event_vfs_getdents_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times directory entries have been read with getdents().",
			},
			"event_vfs_setattr_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times directory entries have been read with getdents().",
			},
			"event_vfs_flush_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of pending writes that have been forcefully flushed to the server.",
			},
			"event_vfs_fsync_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times fsync() has been called on directories and files.",
			},
			"event_vfs_lock_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times locking has been attempted on a file.",
			},
			"event_vfs_file_release_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times files have been closed and released.",
			},
			"event_truncation_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times files have been truncated.",
			},
			"event_write_extension_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a file has been grown due to writes beyond its existing end.",
			},
			"event_silly_rename_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a file was removed while still open by another process.",
			},
			"event_short_read_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the NFS server gave less data than expected while reading.",
			},
			"event_short_write_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the NFS server wrote less data than expected while writing.",
			},
			"event_jukebox_delay_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times the NFS server indicated EJUKEBOX; retrieving data from offline storage.",
			},
			"event_pnfs_read_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of NFS v4.1+ pNFS reads.",
			},
			"event_pnfs_write_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of NFS v4.1+ pNFS writes.",
			},

			// operations
			"operations_requests_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of requests performed for a given operation.",
			},
			"operations_transmissions_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times an actual RPC request has been transmitted for a given operation.",
			},
			"operations_major_timeouts_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of times a request has had a major timeout for a given operation.",
			},
			"operations_sent_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes sent for a given operation, including RPC headers and payload.",
			},
			"operations_received_bytes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.SizeByte,
				Desc:     "Number of bytes received for a given operation, including RPC headers and payload.",
			},
			"operations_queue_time_seconds_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.DurationSecond,
				Unit:     inputs.NCount,
				Desc:     "Duration all requests spent queued for transmission for a given operation before they were sent, in seconds.",
			},
			"operations_response_time_seconds_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.DurationSecond,
				Unit:     inputs.NCount,
				Desc:     "Duration all requests took to get a reply back after a request for a given operation was transmitted, in seconds.",
			},
			"operations_request_time_seconds_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.DurationSecond,
				Unit:     inputs.NCount,
				Desc:     "Duration all requests took from when a request was enqueued to when it was completely handled for a given operation, in seconds.",
			},
			"operations_latency_seconds": &inputs.FieldInfo{
				Type:     inputs.Gauge,
				DataType: inputs.Float,
				Unit:     inputs.DurationSecond,
				Desc:     "Average RPC latency (RTT) for a given operation, in seconds.",
			},
		},
		Tags: map[string]interface{}{
			"device":     &inputs.TagInfo{Desc: "Device name."},
			"mountpoint": &inputs.TagInfo{Desc: "Where the device is mounted."},
			"type":       &inputs.TagInfo{Desc: "Device type."},
		},
	}
}

// Point implement MeasurementV2.
func (m *mountstatsMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func getMountstatsInfo() ([]*procfs.Mount, error) {
	pid, err := datakit.PID()
	var filePath string
	if err != nil {
		filePath = hostProc("self/mountstats")
	} else {
		filePath = hostProc(strconv.Itoa(pid), "mountstats")
	}
	f, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return parseMountStats(f)
}

// parseMountStats parses a /proc/[pid]/mountstats file and returns a slice
// of Mount structures containing detailed information about each mount.
func parseMountStats(r io.Reader) ([]*procfs.Mount, error) {
	const (
		device            = "device"
		statVersionPrefix = "statvers="

		nfs3Type = "nfs"
		nfs4Type = "nfs4"
	)

	var mounts []*procfs.Mount

	s := bufio.NewScanner(r)
	for s.Scan() {
		// Only look for device entries in this function
		ss := strings.Fields(string(s.Bytes()))
		if len(ss) == 0 || ss[0] != device {
			continue
		}

		m, err := parseMount(ss)
		if err != nil {
			return nil, err
		}

		if len(ss) > deviceEntryLen {
			// Only NFSv3 and v4 are supported for parsing statistics
			if m.Type != nfs3Type && m.Type != nfs4Type {
				return nil, fmt.Errorf("%w: Cannot parse MountStats for %q", ErrFileParse, m.Type)
			}

			statVersion := strings.TrimPrefix(ss[8], statVersionPrefix)

			stats, err := parseMountStatsNFS(s, statVersion)
			if err != nil {
				return nil, err
			}

			m.Stats = stats
			mounts = append(mounts, m)
		}
	}

	return mounts, s.Err()
}

func parseMount(ss []string) (*procfs.Mount, error) {
	if len(ss) < deviceEntryLen {
		return nil, fmt.Errorf("%w: Invalid device %q", ErrFileParse, ss)
	}

	format := []struct {
		i int
		s string
	}{
		{i: 0, s: "device"},
		{i: 2, s: "mounted"},
		{i: 3, s: "on"},
		{i: 5, s: "with"},
		{i: 6, s: "fstype"},
	}

	for _, f := range format {
		if ss[f.i] != f.s {
			return nil, fmt.Errorf("%w: Invalid device %q", ErrFileParse, ss)
		}
	}

	return &procfs.Mount{
		Device: ss[1],
		Mount:  ss[4],
		Type:   ss[7],
	}, nil
}

func parseMountStatsNFS(s *bufio.Scanner, statVersion string) (*procfs.MountStatsNFS, error) {
	// Field indicators for parsing specific types of data
	const (
		fieldOpts       = "opts:"
		fieldAge        = "age:"
		fieldBytes      = "bytes:"
		fieldEvents     = "events:"
		fieldPerOpStats = "per-op"
		fieldTransport  = "xprt:"
	)

	stats := &procfs.MountStatsNFS{
		StatVersion: statVersion,
	}

	for s.Scan() {
		ss := strings.Fields(string(s.Bytes()))
		if len(ss) == 0 {
			break
		}

		switch ss[0] {
		case fieldOpts:
			if len(ss) < 2 {
				return nil, fmt.Errorf("%w: Incomplete information for NFS stats: %v", ErrFileParse, ss)
			}
			if stats.Opts == nil {
				stats.Opts = map[string]string{}
			}
			for _, opt := range strings.Split(ss[1], ",") {
				split := strings.Split(opt, "=")
				if len(split) == 2 {
					stats.Opts[split[0]] = split[1]
				} else {
					stats.Opts[opt] = ""
				}
			}
		case fieldAge:
			if len(ss) < 2 {
				return nil, fmt.Errorf("%w: Incomplete information for NFS stats: %v", ErrFileParse, ss)
			}
			// Age integer is in seconds
			d, err := time.ParseDuration(ss[1] + "s")
			if err != nil {
				return nil, err
			}

			stats.Age = d
		case fieldBytes:
			if len(ss) < 2 {
				return nil, fmt.Errorf("%w: Incomplete information for NFS stats: %v", ErrFileParse, ss)
			}
			bstats, err := parseNFSBytesStats(ss[1:])
			if err != nil {
				return nil, err
			}

			stats.Bytes = *bstats
		case fieldEvents:
			if len(ss) < 2 {
				return nil, fmt.Errorf("%w: Incomplete information for NFS events: %v", ErrFileParse, ss)
			}
			estats, err := parseNFSEventsStats(ss[1:])
			if err != nil {
				return nil, err
			}

			stats.Events = *estats
		case fieldTransport:
			if len(ss) < 3 {
				return nil, fmt.Errorf("%w: Incomplete information for NFS transport stats: %v", ErrFileParse, ss)
			}

			tstats, err := parseNFSTransportStats(ss[1:], statVersion)
			if err != nil {
				return nil, err
			}

			stats.Transport = *tstats
		}

		if ss[0] == fieldPerOpStats {
			break
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	perOpStats, err := parseNFSOperationStats(s)
	if err != nil {
		return nil, err
	}

	stats.Operations = perOpStats

	return stats, nil
}

func parseNFSBytesStats(ss []string) (*procfs.NFSBytesStats, error) {
	if len(ss) != fieldBytesLen {
		return nil, fmt.Errorf("%w: Invalid NFS bytes stats: %v", ErrFileParse, ss)
	}

	ns := make([]uint64, 0, fieldBytesLen)
	for _, s := range ss {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}

		ns = append(ns, n)
	}

	return &procfs.NFSBytesStats{
		Read:        ns[0],
		Write:       ns[1],
		DirectRead:  ns[2],
		DirectWrite: ns[3],
		ReadTotal:   ns[4],
		WriteTotal:  ns[5],
		ReadPages:   ns[6],
		WritePages:  ns[7],
	}, nil
}

func parseNFSEventsStats(ss []string) (*procfs.NFSEventsStats, error) {
	if len(ss) != fieldEventsLen {
		return nil, fmt.Errorf("%w: invalid NFS events stats: %v", ErrFileParse, ss)
	}

	ns := make([]uint64, 0, fieldEventsLen)
	for _, s := range ss {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}

		ns = append(ns, n)
	}

	return &procfs.NFSEventsStats{
		InodeRevalidate:     ns[0],
		DnodeRevalidate:     ns[1],
		DataInvalidate:      ns[2],
		AttributeInvalidate: ns[3],
		VFSOpen:             ns[4],
		VFSLookup:           ns[5],
		VFSAccess:           ns[6],
		VFSUpdatePage:       ns[7],
		VFSReadPage:         ns[8],
		VFSReadPages:        ns[9],
		VFSWritePage:        ns[10],
		VFSWritePages:       ns[11],
		VFSGetdents:         ns[12],
		VFSSetattr:          ns[13],
		VFSFlush:            ns[14],
		VFSFsync:            ns[15],
		VFSLock:             ns[16],
		VFSFileRelease:      ns[17],
		CongestionWait:      ns[18],
		Truncation:          ns[19],
		WriteExtension:      ns[20],
		SillyRename:         ns[21],
		ShortRead:           ns[22],
		ShortWrite:          ns[23],
		JukeboxDelay:        ns[24],
		PNFSRead:            ns[25],
		PNFSWrite:           ns[26],
	}, nil
}

func parseNFSOperationStats(s *bufio.Scanner) ([]procfs.NFSOperationStats, error) {
	const (
		// Minimum number of expected fields in each per-operation statistics set
		minFields = 9
	)

	var ops []procfs.NFSOperationStats

	for s.Scan() {
		ss := strings.Fields(string(s.Bytes()))
		if len(ss) == 0 {
			break
		}

		if len(ss) < minFields {
			return nil, fmt.Errorf("%w: invalid NFS per-operations stats: %v", ErrFileParse, ss)
		}

		// Skip string operation name for integers
		ns := make([]uint64, 0, minFields-1)
		for _, st := range ss[1:] {
			n, err := strconv.ParseUint(st, 10, 64)
			if err != nil {
				return nil, err
			}

			ns = append(ns, n)
		}
		opStats := procfs.NFSOperationStats{
			Operation:                           strings.TrimSuffix(ss[0], ":"),
			Requests:                            ns[0],
			Transmissions:                       ns[1],
			MajorTimeouts:                       ns[2],
			BytesSent:                           ns[3],
			BytesReceived:                       ns[4],
			CumulativeQueueMilliseconds:         ns[5],
			CumulativeTotalResponseMilliseconds: ns[6],
			CumulativeTotalRequestMilliseconds:  ns[7],
		}
		if ns[0] != 0 {
			opStats.AverageRTTMilliseconds = float64(ns[6]) / float64(ns[0])
		}

		if len(ns) > 8 {
			opStats.Errors = ns[8]
		}

		ops = append(ops, opStats)
	}

	return ops, s.Err()
}

func parseNFSTransportStats(ss []string, statVersion string) (*procfs.NFSTransportStats, error) {
	protocol := ss[0]
	ss = ss[1:]

	switch statVersion {
	case statVersion10:
		var expectedLength int
		switch protocol {
		case "tcp":
			expectedLength = fieldTransport10TCPLen
		case "udp":
			expectedLength = fieldTransport10UDPLen
		default:
			return nil, fmt.Errorf("%w: Invalid NFS protocol \"%s\" in stats 1.0 statement: %v", ErrFileParse, protocol, ss)
		}
		if len(ss) != expectedLength {
			return nil, fmt.Errorf("%w: Invalid NFS transport stats 1.0 statement: %v", ErrFileParse, ss)
		}
	case statVersion11:
		var expectedLength int
		switch protocol {
		case "tcp":
			expectedLength = fieldTransport11TCPLen
		case "udp":
			expectedLength = fieldTransport11UDPLen
		default:
			return nil, fmt.Errorf("%w: Invalid NFS protocol \"%s\" in stats 1.1 statement: %v", ErrFileParse, protocol, ss)
		}
		if len(ss) != expectedLength {
			return nil, fmt.Errorf("%w: Invalid NFS transport stats 1.1 statement: %v", ErrFileParse, ss)
		}
	default:
		return nil, fmt.Errorf("%w: Unrecognized NFS transport stats version: %q", ErrFileParse, statVersion)
	}

	ns := make([]uint64, fieldTransport11TCPLen)
	for i, s := range ss {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}

		ns[i] = n
	}

	if protocol == "udp" {
		ns = append(ns[:2], append(make([]uint64, 3), ns[2:]...)...)
	}

	return &procfs.NFSTransportStats{
		Protocol:                 protocol,
		Port:                     ns[0],
		Bind:                     ns[1],
		Connect:                  ns[2],
		ConnectIdleTime:          ns[3],
		IdleTimeSeconds:          ns[4],
		Sends:                    ns[5],
		Receives:                 ns[6],
		BadTransactionIDs:        ns[7],
		CumulativeActiveRequests: ns[8],
		CumulativeBacklog:        ns[9],
		MaximumRPCSlotsUsed:      ns[10],
		CumulativeSendingQueue:   ns[11],
		CumulativePendingQueue:   ns[12],
	}, nil
}

func collectMountStatsOperation(mount *procfs.Mount, operations []procfs.NFSOperationStats, ts int64) ([]inputs.MeasurementV2, error) {
	ms := []inputs.MeasurementV2{}
	for _, op := range operations {
		m := &mountstatsMeasurement{
			name: "nfs_mountstats",
			tags: map[string]string{
				"device":     mount.Device,
				"mountpoint": mount.Mount,
				"type":       mount.Type,
				"operation":  op.Operation,
			},
			fields: map[string]interface{}{
				"operations_requests_total":              op.Requests,
				"operations_transmissions_total":         op.Transmissions,
				"operations_major_timeouts_total":        op.MajorTimeouts,
				"operations_sent_bytes_total":            op.BytesSent,
				"operations_received_bytes_total":        op.BytesReceived,
				"operations_queue_time_seconds_total":    float64(op.CumulativeQueueMilliseconds%float64Mantissa) / 1000.0,
				"operations_response_time_seconds_total": float64(op.CumulativeTotalResponseMilliseconds%float64Mantissa) / 1000.0,
				"operations_request_time_seconds_total":  float64(op.CumulativeTotalRequestMilliseconds%float64Mantissa) / 1000.0,
				"operations_latency_seconds":             op.AverageRTTMilliseconds / 1000.0,
			},
			ts: ts,
		}

		ms = append(ms, m)
	}

	return ms, nil
}
