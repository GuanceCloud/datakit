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

type baseMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

//nolint:lll
func (m *baseMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name: "nfs",
		Type: "metric",
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
			"rpcs_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Total number of RPCs performed.",
			},
			"rpc_retransmissions_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of RPC transmissions performed.",
			},
			"rpc_authentication_refreshes_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of RPC authentication refreshes performed.",
			},
			"requests_total": &inputs.FieldInfo{
				Type:     inputs.Count,
				DataType: inputs.Int,
				Unit:     inputs.NCount,
				Desc:     "Number of NFS procedures invoked.",
			},
		},
		Tags: map[string]interface{}{
			"protocol": &inputs.TagInfo{Desc: "Protocol type."},
			"method":   &inputs.TagInfo{Desc: "Invoked method."},
		},
	}
}

// Point implement MeasurementV2.
func (m *baseMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

// ClientRPCStats retrieves NFS client RPC statistics
// from proc/net/rpc/nfs.
func getClientRPCStats() (*nfs.ClientRPCStats, error) {
	f, err := os.Open(hostProc("net/rpc/nfs"))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return nfs.ParseClientRPCStats(f)
}

func collectNFSRequestsv2Stats(s *nfs.V2Stats, ts int64) ([]*point.Point, error) {
	const proto = "2"
	ms := []inputs.MeasurementV2{}
	m := &baseMeasurement{
		name: "nfs",
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

func collectNFSRequestsv3Stats(s *nfs.V3Stats, ts int64) ([]*point.Point, error) {
	const proto = "3"
	ms := []inputs.MeasurementV2{}
	m := &baseMeasurement{
		name: "nfs",
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

func collectNFSRequestsv4Stats(s *nfs.ClientV4Stats, ts int64) ([]*point.Point, error) {
	const proto = "4"
	ms := []inputs.MeasurementV2{}
	m := &baseMeasurement{
		name: "nfs",
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
