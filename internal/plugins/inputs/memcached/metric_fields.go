// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package memcached

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type inputMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *inputMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*inputMeasurement) Info() *inputs.MeasurementInfo {
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

type itemsMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (*itemsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "memcached_items",
		Fields: map[string]interface{}{
			"number":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in this slab class"},
			"number_hot":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the HOT LRU"},
			"number_warm":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the WARM LRU"},
			"number_cold":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the COLD LRU"},
			"number_noexp":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items presently stored in the `NOEXP` class"},
			"age":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Age of the oldest item in the LRU"},
			"evicted":           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of the items which had to be evicted from the LRU before expiring"},
			"evicted_nonzero":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of the `onzero` items which had an explicit expire time set had to be evicted from the LRU before expiring"},
			"expired_unfetched": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of the expired items reclaimed from the LRU which were never touched after being set"},
			"evicted_unfetched": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of the valid items evicted from the LRU which were never touched after being set"},
			"evicted_time":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Seconds since the last access for the most recent item evicted from this class"},
			"outofmemory":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of the underlying slab class which was unable to store a new item"},
			"tailrepairs":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many times memcache self-healed a slab with a `refcount` leak"},
			"moves_to_cold":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items which were moved from HOT or WARM into COLD"},
			"moves_to_warm":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items which were moved from COLD to WARM"},
			"moves_within_lru":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of active items which were bumped within HOT or WARM"},
			"reclaimed":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of entries which were stored using memory from an expired entry"},
			"crawler_reclaimed": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items which freed by the LRU Crawler"},
			"lrutail_reflocked": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of items which found to be `refcount` locked in the LRU tail"},
			"direct_reclaims":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of worker threads which had to directly pull LRU tails to find memory for a new item"},
		},
		Tags: map[string]interface{}{
			"server":  inputs.NewTagInfo("The host name from which metrics are gathered"),
			"slab_id": inputs.NewTagInfo("The id of the current slab"),
		},
	}
}

type slabsMeasurement struct {
	inputMeasurement
}

//nolint:lll
func (*slabsMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "memcached_slabs",
		Fields: map[string]interface{}{
			"chunk_size":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of space each chunk uses"},
			"chunks_per_page": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many chunks exist within one page"},
			"total_pages":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of pages allocated to the slab class"},
			"total_chunks":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of chunks allocated to the slab class"},
			"used_chunks":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "How many chunks have been allocated to items"},
			"free_chunks":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Chunks not yet allocated to items or freed via delete"},
			"free_chunks_end": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of free chunks at the end of the last allocated page"},
			"get_hits":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of get requests were serviced by this slab class"},
			"cmd_set":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of set requests stored data in this slab class"},
			"delete_hits":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of delete commands succeeded in this slab class"},
			"incr_hits":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of `incrs` commands modified this slab class"},
			"decr_hits":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of `decrs` commands modified this slab class"},
			"cas_hits":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of CAS commands modified this slab class"},
			"cas_badval":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of CAS commands failed to modify a value due to a bad CAS id"},
			"touch_hits":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of touches serviced by this slab class"},
			"active_slabs":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Total number of slab classes allocated"},
			"total_malloced":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Total amount of memory allocated to slab pages"},
		},
		Tags: map[string]interface{}{
			"server":  inputs.NewTagInfo("The host name from which metrics are gathered"),
			"slab_id": inputs.NewTagInfo("The id of the current slab"),
		},
	}
}
