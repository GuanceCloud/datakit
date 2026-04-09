package aggregate

import (
	"fmt"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

// ptWrap implements fp.KVs.
type ptWrap struct {
	Point *point.Point
	pb    point.PBPoint
}

func (d *ptWrap) Reset(raw []byte) error {
	d.Point = nil
	d.pb.Reset()
	if len(raw) == 0 {
		return nil
	}

	if err := d.pb.Unmarshal(raw); err != nil {
		return fmt.Errorf("unmarshal pb point: %w", err)
	}

	return nil
}

func (d *ptWrap) Get(k string) (any, bool) {
	if d.Point != nil {
		v := d.Point.KVs().Get(k)
		if v == nil {
			return nil, false
		}

		switch x := v.GetVal().(type) {
		case *point.Field_F:
			return x.F, true
		case *point.Field_I:
			return x.I, true
		case *point.Field_U:
			return x.U, true
		case *point.Field_S:
			return x.S, true
		case *point.Field_D:
			return x.D, true
		case *point.Field_B:
			return x.B, true
		default: // other types are ignored
			return nil, true
		}
	}

	for _, v := range d.pb.Fields {
		if v == nil || v.Key != k {
			continue
		}

		switch x := v.GetVal().(type) {
		case *point.Field_F:
			return x.F, true
		case *point.Field_I:
			return x.I, true
		case *point.Field_U:
			return x.U, true
		case *point.Field_S:
			return x.S, true
		case *point.Field_D:
			return x.D, true
		case *point.Field_B:
			return x.B, true
		default: // other types are ignored
			return nil, true
		}
	}

	return nil, false
}

const (
	Seed1   = uint64(0x9E3779B97F4A7C15)
	Seed2   = uint64(0x6A09E667F3BCC908)
	SeedU32 = uint32(0x7F4A7C15)
)

func hash(pt *point.Point, sortedTagKeys []string) uint64 {
	h := Seed1

	// we always use measurement name and metric name for hash
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(pt.Name())))

	for _, k := range sortedTagKeys {
		if s, ok := pt.GetS(k); ok {
			h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(k)))
			h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(s)))
		}
	}

	for _, kv := range pt.KVs() {
		if kv.IsTag {
			continue
		}
		// NOTE: only get the first non-tag filed for hash, we should
		// make sure there only one field on each aggregate point.
		h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(kv.Key)))
		break
	}

	return h
}

func pickHash(pt *point.Point, sortedTagKeys []string) uint64 {
	h := Seed1

	// we always use measurement name and metric name for hash
	h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(pt.Name())))
	for _, k := range sortedTagKeys {
		h = HashCombine(h, xxhash.Sum64(cliutils.ToUnsafeBytes(k)))
	}
	return h
}

// HashCombine used to combine 2 u64 hash value, see https://zhuanlan.zhihu.com/p/574573421.
func HashCombine(seed, hash uint64) uint64 {
	return ((seed + 0x9e3779b9) ^ hash) * 0x517cc1b727220a95
}

func HashToken(token string, hash64 uint64) uint64 {
	return HashCombine(hash64, xxhash.Sum64(cliutils.ToUnsafeBytes(token)))
}

// pointAggrTags calculate point's aggregate tags.
func pointAggrTags(pt *point.Point, sortedKeys []string) [][2]string {
	kvs := [][2]string{}

	for _, k := range sortedKeys {
		if x := pt.Get(k); x != nil {
			if v, ok := x.(string); ok {
				kvs = append(kvs, [2]string{k, v})
			}
		}
	}

	return kvs
}

// getTime 原始单位是微妙，返回单位用纳秒.
func getTime(pt *point.Point) (startTime int64, duration int64) {
	var has bool
	startTime, has = pt.GetI("start_time")
	if !has {
		f64, _ := pt.GetF("start_time")
		if f64 != 0 {
			startTime = int64(f64)
		}
	}

	duration, has = pt.GetI("duration")
	if !has {
		f64, _ := pt.GetF("duration")
		if f64 != 0 {
			duration = int64(f64)
		}
	}

	return startTime * 1e3, duration * 1e3
}
