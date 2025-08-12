// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:deadcode,unused
package nfs

import (
	"os"
	"reflect"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/procfs/nfs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type nfsdMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

//nolint:lll
func (m *nfsdMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name: "nfsd",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"tcp_packets_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total NFSd network TCP packets (sent+received) by protocol type.",
			},
			"udp_packets_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total NFSd network UDP packets (sent+received) by protocol type.",
			},
			"connections_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd TCP connections.",
			},
			"reply_cache_hits_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd Reply Cache hits (client lost server response).",
			},
			"reply_cache_misses_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd Reply Cache an operation that requires caching (idempotent).",
			},
			"reply_cache_nocache_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd Reply Cache non-idempotent operations (rename/delete/…).",
			},
			"file_handles_stale_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd stale file handles",
			},
			"disk_bytes_read_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total NFSd bytes read.",
			},
			"disk_bytes_written_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total NFSd bytes written.",
			},
			"server_threads": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd kernel threads that are running.",
			},
			"read_ahead_cache_size_blocks": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "How large the read ahead cache is in blocks.",
			},
			"read_ahead_cache_not_found_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd read ahead cache not found.",
			},

			"server_rpcs_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd RPCs.",
			},
			"rpc_errors_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of NFSd RPC errors by error type.",
			},
		},
		Tags: map[string]interface{}{
			"protocol": &inputs.TagInfo{Desc: "Protocol type."},
			"method":   &inputs.TagInfo{Desc: "Invoked method."},
			"error":    &inputs.TagInfo{Desc: "Error type."},
		},
	}
}

// Point implement MeasurementV2.
func (m *nfsdMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// ServerRPCStats retrieves NFS daemon RPC statistics
// from proc/net/rpc/nfsd.
func getServerRPCStats() (*nfs.ServerRPCStats, error) {
	f, err := os.Open(hostProc("net/rpc/nfsd"))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return parseServerRPCStats(f)
}

func collectNFSdRequestsv2Stats(s *nfs.V2Stats, ts int64) ([]*point.Point, error) {
	const proto = "2"
	ms := []inputs.MeasurementV2{}
	m := &nfsdMeasurement{
		name: "nfsd",
		ts:   ts,
		tags: map[string]string{
			"proto": proto,
		},
		fields: map[string]interface{}{},
	}
	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		m.tags["method"] = v.Type().Field(i).Name
		m.fields["requests_total"] = field.Uint()

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*point.Point{}, nil
}

func collectNFSdRequestsv3Stats(s *nfs.V3Stats, ts int64) ([]*point.Point, error) {
	const proto = "3"
	ms := []inputs.MeasurementV2{}
	m := &nfsdMeasurement{
		name: "nfsd",
		ts:   ts,
		tags: map[string]string{
			"proto": proto,
		},
		fields: map[string]interface{}{},
	}

	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		m.tags["method"] = v.Type().Field(i).Name
		m.fields["requests_total"] = field.Uint()

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*point.Point{}, nil
}

func collectNFSdRequestsv4Stats(s *nfs.V4Ops, ts int64) ([]*point.Point, error) {
	const proto = "4"
	ms := []inputs.MeasurementV2{}
	m := &nfsdMeasurement{
		name: "nfsd",
		ts:   ts,
		tags: map[string]string{
			"proto": proto,
		},
		fields: map[string]interface{}{},
	}

	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		m.tags["method"] = v.Type().Field(i).Name
		m.fields["requests_total"] = field.Uint()

		ms = append(ms, m)
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*point.Point{}, nil
}

func collectNFSdServerRPCStats(s nfs.ServerRPC, ts int64) ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}
	m := &nfsdMeasurement{
		name:   "nfsd",
		ts:     ts,
		fields: map[string]interface{}{},
		tags:   map[string]string{},
	}
	m.tags["error"] = "fmt"
	m.fields["rpc_errors_total"] = s.BadFmt
	ms = append(ms, m)

	m.tags["error"] = "auth"
	m.fields["rpc_errors_total"] = s.BadAuth
	ms = append(ms, m)

	m.tags["error"] = "cInt"
	m.fields["rpc_errors_total"] = s.BadCnt
	ms = append(ms, m)

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*point.Point{}, nil
}
