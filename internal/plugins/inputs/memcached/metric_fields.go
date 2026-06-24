// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package memcached

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type inputMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *inputMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts), m.ipt.opt)

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*inputMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "Memcached server-wide statistics returned by the `stats` command, including current cache state, memory usage, connection state, and cumulative command counters.",
		DescZh: "通过 `stats` 命令采集的 Memcached 服务级统计，包括当前缓存状态、内存使用、连接状态以及累计命令计数。",
		Fields: memFields,
		Tags:   map[string]interface{}{"server": inputs.NewTagInfo("The memcached server address from which metrics are gathered.")},
	}
}

//nolint:lll
var memFields = map[string]interface{}{
	"accepting_conns":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Whether the server is currently accepting new connections. `1` means accepting, `0` means not accepting."},
	"auth_cmds":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of authentication commands handled, successful or failed."},
	"auth_errors":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of failed authentications."},
	"bytes":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Current number of bytes used to store items"},
	"bytes_read":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes read by this server from the network."},
	"bytes_written":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes written by this server to the network."},
	"cas_badval":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative CAS requests for which a key was found but the CAS value did not match."},
	"cas_hits":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative successful CAS requests."},
	"cas_misses":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative CAS requests against missing keys."},
	"cmd_flush":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of flush requests."},
	"cmd_get":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of retrieval requests."},
	"cmd_set":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of storage requests."},
	"cmd_touch":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of touch requests."},
	"conn_yields":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of times a connection yielded to another connection after hitting the `-R` limit."},
	"connection_structures": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of connection structures allocated by the server"},
	"curr_connections":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of open connections"},
	"curr_items":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of items stored"},
	"decr_hits":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative successful `decr` requests."},
	"decr_misses":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative `decr` requests against missing keys."},
	"delete_hits":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative delete requests that removed an item."},
	"delete_misses":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative delete requests for missing keys."},
	"evicted_unfetched":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items evicted from the LRU that were never fetched after being set."},
	"evictions":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative valid items removed from cache to free memory for new items."},
	"expired_unfetched":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative expired items reclaimed from the LRU that were never fetched after being set."},
	"get_hits":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative keys requested and found present."},
	"get_misses":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative requested items that were not found."},
	"hash_bytes":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Bytes currently used by hash tables"},
	"hash_is_expanding":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Whether the hash table is currently being grown to a new size. `1` means expanding, `0` means not expanding."},
	"hash_power_level":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current size multiplier for hash table"},
	"incr_hits":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative successful `incr` requests."},
	"incr_misses":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative `incr` requests against missing keys."},
	"limit_maxbytes":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Number of bytes this server is allowed to use for storage"},
	"listen_disabled_num":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of times the server stopped accepting new connections after reaching `maxconns`."},
	"reclaimed":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative entries stored using memory reclaimed from an expired entry."},
	"threads":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of worker threads requested"},
	"total_connections":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative connections opened since the server started."},
	"total_items":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items stored since the server started."},
	"touch_hits":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative keys touched with a new expiration time."},
	"touch_misses":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative touch requests for items that were not found."},
	"uptime":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Seconds since the server started."},
}

type itemsMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (*itemsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "memcached_items",
		Cat:    point.Metric,
		Desc:   "Per-slab-class Memcached item statistics returned by `stats items`, including current LRU item counts and cumulative item movement, eviction, reclaim, and out-of-memory counters.",
		DescZh: "通过 `stats items` 命令采集的 Memcached slab class 级 item 统计，包括当前 LRU item 数量以及累计迁移、驱逐、回收和内存不足计数。",
		Fields: map[string]interface{}{
			"number":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in this slab class"},
			"number_hot":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the HOT LRU"},
			"number_warm":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the WARM LRU"},
			"number_cold":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the COLD LRU"},
			"number_noexp":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the `NOEXP` class"},
			"age":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Age in seconds of the oldest item in the LRU for this slab class."},
			"evicted":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items evicted from this slab class before expiring."},
			"evicted_nonzero":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative evicted items from this slab class that had an explicit non-zero expiration time."},
			"expired_unfetched": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative expired items reclaimed from this slab class that were never fetched after being set."},
			"evicted_unfetched": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative valid items evicted from this slab class that were never fetched after being set."},
			"evicted_time":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Seconds since the last access for the most recent item evicted from this class"},
			"outofmemory":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative times this slab class could not store a new item due to memory pressure."},
			"tailrepairs":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative self-healing repairs for slab items with leaked `refcount` references."},
			"moves_to_cold":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items moved from HOT or WARM into COLD."},
			"moves_to_warm":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items moved from COLD to WARM."},
			"moves_within_lru":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative active items bumped within HOT or WARM."},
			"reclaimed":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative entries stored using memory reclaimed from an expired entry."},
			"crawler_reclaimed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items freed by the LRU crawler."},
			"lrutail_reflocked": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative items found to be `refcount` locked in the LRU tail."},
			"direct_reclaims":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative worker threads that directly pulled LRU tails to find memory for a new item."},
		},
		Tags: map[string]interface{}{
			"server":  inputs.NewTagInfo("The memcached server address from which metrics are gathered."),
			"slab_id": inputs.NewTagInfo("Slab class ID from the `stats items` key."),
		},
	}
}

type slabsMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (*slabsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "memcached_slabs",
		Cat:    point.Metric,
		Desc:   "Per-slab-class Memcached slab allocator statistics returned by `stats slabs`, including current chunk/page allocation and cumulative command counters served by each slab class.",
		DescZh: "通过 `stats slabs` 命令采集的 Memcached slab 分配器统计，包括每个 slab class 当前 chunk/page 分配情况以及累计命令计数。",
		Fields: map[string]interface{}{
			"chunk_size":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of space each chunk uses"},
			"chunks_per_page": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many chunks exist within one page"},
			"total_pages":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of pages allocated to the slab class"},
			"total_chunks":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of chunks allocated to the slab class"},
			"used_chunks":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many chunks have been allocated to items"},
			"free_chunks":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Chunks not yet allocated to items or freed via delete"},
			"free_chunks_end": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of free chunks at the end of the last allocated page"},
			"get_hits":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative get requests serviced by this slab class."},
			"cmd_set":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative set requests that stored data in this slab class."},
			"delete_hits":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative successful delete commands in this slab class."},
			"incr_hits":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative `incr` commands that modified this slab class."},
			"decr_hits":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative `decr` commands that modified this slab class."},
			"cas_hits":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative CAS commands that modified this slab class."},
			"cas_badval":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative CAS commands that failed to modify a value due to a bad CAS ID."},
			"touch_hits":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative touch commands serviced by this slab class."},
			"active_slabs":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of slab classes allocated"},
			"total_malloced":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory allocated to slab pages"},
		},
		Tags: map[string]interface{}{
			"server":  inputs.NewTagInfo("The memcached server address from which metrics are gathered."),
			"slab_id": inputs.NewTagInfo("Slab class ID from the `stats slabs` key."),
		},
	}
}
