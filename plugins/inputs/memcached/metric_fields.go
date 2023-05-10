// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package memcached

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type inputMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

// Point implement MeasurementV2.
func (m *inputMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	} else {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalHostTags()))
	}

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m inputMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
	return nil, fmt.Errorf("not implement")
}

//nolint:lll
func (m inputMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Fields: memFields,
		Tags:   map[string]interface{}{"server": inputs.NewTagInfo("The host name from which metrics are gathered")},
	}
}

//nolint:lll
var memFields = map[string]interface{}{
	"accepting_conns":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Whether or not server is accepting conns"},
	"auth_cmds":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of authentication commands handled, success or failure"},
	"auth_errors":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of failed authentications"},
	"bytes":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current number of bytes used to store items"},
	"bytes_read":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes read by this server from network"},
	"bytes_written":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total number of bytes sent by this server to network"},
	"cas_badval":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of CAS  for which a key was found, but the CAS value did not match"},
	"cas_hits":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of successful CAS requests"},
	"cas_misses":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of CAS requests against missing keys"},
	"cmd_flush":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative number of flush requests"},
	"cmd_get":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative number of retrieval requests"},
	"cmd_set":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative number of storage requests"},
	"cmd_touch":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Cumulative number of touch requests"},
	"conn_yields":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of times any connection yielded to another due to hitting the -R limit"},
	"connection_structures": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of connection structures allocated by the server"},
	"curr_connections":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of open connections"},
	"curr_items":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of items stored"},
	"decr_hits":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of successful `decr` requests"},
	"decr_misses":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of `decr` requests against missing keys"},
	"delete_hits":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of deletion requests resulting in an item being removed"},
	"delete_misses":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "umber of deletions requests for missing keys"},
	"evicted_unfetched":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Items evicted from LRU that were never touched by get/incr/append/etc"},
	"evictions":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of valid items removed from cache to free memory for new items"},
	"expired_unfetched":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Items pulled from LRU that were never touched by get/incr/append/etc before expiring"},
	"get_hits":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys that have been requested and found present"},
	"get_misses":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items that have been requested and not found"},
	"hash_bytes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes currently used by hash tables"},
	"hash_is_expanding":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Indicates if the hash table is being grown to a new size"},
	"hash_power_level":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current size multiplier for hash table"},
	"incr_hits":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of successful incr requests"},
	"incr_misses":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of incr requests against missing keys"},
	"limit_maxbytes":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes this server is allowed to use for storage"},
	"listen_disabled_num":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of times server has stopped accepting new connections (`maxconns`)"},
	"reclaimed":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of times an entry was stored using memory from an expired entry"},
	"threads":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of worker threads requested"},
	"total_connections":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of connections opened since the server started running"},
	"total_items":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of items stored since the server started"},
	"touch_hits":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of keys that have been touched with a new expiration time"},
	"touch_misses":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items that have been touched and not found"},
	"uptime":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of secs since the server started"},
}
