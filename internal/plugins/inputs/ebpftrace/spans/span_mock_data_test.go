// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package spans

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
)

func genMockTrace(direct string, appTrace bool, level, child int,
	tail_end, sampled, setTime bool, tn time.Time,
) ([]*point.Point, error) {
	var pts []*point.Point
	ulid, _ := NewULID()

	if child <= 0 {
		child = 1
	}

	if level <= 0 {
		level = 1
	}

	var etraceid string
	nextNetID := []ID128{}
	if direct == INCO {
		netTID, _ := ulid.ID()
		nextNetID = append(nextNetID, *netTID)
	}

	for i := 0; i < level; i++ {
		var netTID, entryTID *ID128
		var nexDirect string
		switch direct {
		case OUTG:
			netTID, _ = ulid.ID()
			entryTID, _ = ulid.ID()
			nexDirect = INCO
		case INCO:
			netTID = &nextNetID[0]
			entryTID, _ = ulid.ID()
			nexDirect = OUTG
		}
		if etraceid == "" {
			v := md5.Sum(netTID.Byte())
			etraceid = hex.EncodeToString(v[:])
		}

		// entry span
		{
			tn = tn.Add(time.Nanosecond * 50)

			spanid, _ := ulid.ID()
			fields := map[string]any{
				ReqSeq:       int64(netTID.Low),
				RespSeq:      int64(netTID.High),
				ThrTraceID:   int64(hash.Fnv1aU8Hash(entryTID.Byte())),
				Direction:    direct,
				EBPFSpanType: SpanTypEntry,

				SpanID: int64(hash.Fnv1aU8Hash(spanid.Byte())),

				AppSpanSampled:    true,
				"_" + EBPFTraceID: etraceid,
			}
			if setTime {
				fields[_SynTime] = tn.UnixNano()
			}
			if i == 0 && appTrace {
				id, _ := ulid.ID()
				fields[AppTraceIDL] = id.Low
				fields[AppTraceIDH] = id.High
				fields[AppParentIDL] = id.Low
			}

			kvs := point.NewKVs(fields)
			pt := point.NewPointV2("datakit-ebpf", kvs, point.WithTime(tn))
			pts = append(pts, pt)

			// first span is client span
			if direct == OUTG {
				direct = INCO
				nextNetID = append(nextNetID, *netTID)
				continue
			}
		}

		if (i == level-1) && tail_end {
			return pts, nil
		}
		nextNetID = []ID128{}
		for j := 0; j < child; j++ {
			tn = tn.Add(time.Nanosecond * 50)
			id, _ := ulid.ID()
			nextNetID = append(nextNetID, *id)
			spanid, _ := ulid.ID()
			fields := map[string]any{
				ReqSeq:            int64(id.Low),
				RespSeq:           int64(id.High),
				ThrTraceID:        int64(hash.Fnv1aU8Hash(entryTID.Byte())),
				SpanID:            int64(hash.Fnv1aU8Hash(spanid.Byte())),
				Direction:         nexDirect,
				EBPFSpanType:      "",
				AppSpanSampled:    true,
				"_" + EBPFTraceID: etraceid,
			}
			if setTime {
				fields[_SynTime] = tn.UnixNano()
			}
			kvs := point.NewKVs(fields)
			pt := point.NewPointV2("datakit-ebpf", kvs, point.WithTime(tn))
			pts = append(pts, pt)
		}
	}

	return pts, nil
}

func MockTraceData(trace, level, spans int, tn time.Time, sampled, setTime bool) []*point.Point {
	var pts []*point.Point
	for i := 0; i < trace; i++ {
		var dr string
		var bval bool
		switch i % 2 {
		case 0:
			dr = INCO
		default:
			dr = OUTG
		}

		bval = !bval
		ptss, err := genMockTrace(dr, false, level, spans, bval, sampled, setTime, tn)
		if err == nil {
			pts = append(pts, ptss...)
		}
	}
	return pts
}

func spanMeta(pts []*point.Point) (ESpans, bool) {
	spans := ESpans{}

	for _, pt := range pts {
		sp, ok := spanMetaData(pt)
		if !ok {
			return nil, false
		}
		spans = append(spans, sp)
	}
	return spans, true
}
