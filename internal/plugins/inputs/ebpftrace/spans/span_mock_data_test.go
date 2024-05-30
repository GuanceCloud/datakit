// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

func genMockTrace(direct string, appTrace bool, level, child int,
	tail_end, sampled bool, tn time.Time,
) ([]*point.Point, error) {
	var pts []*point.Point
	ulid, _ := espan.NewRandID()

	if child <= 0 {
		child = 1
	}

	if level <= 0 {
		level = 1
	}

	var etraceid string
	nextNetID := []espan.ID128{}
	if direct == espan.INCO {
		netTID, _ := ulid.ID()
		nextNetID = append(nextNetID, *netTID)
	}

	for i := 0; i < level; i++ {
		var netTID, entryTID *espan.ID128
		var nexDirect string
		switch direct {
		case espan.OUTG:
			netTID, _ = ulid.ID()
			entryTID, _ = ulid.ID()
			nexDirect = espan.INCO
		case espan.INCO:
			netTID = &nextNetID[0]
			entryTID, _ = ulid.ID()
			nexDirect = espan.OUTG
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
				espan.ReqSeq:       int64(netTID.Low),
				espan.RespSeq:      int64(netTID.High),
				espan.ThrTraceID:   int64(hash.Fnv1aU8Hash(entryTID.Byte())),
				espan.DirectionStr: direct,
				espan.EBPFSpanType: espan.SpanTypEntry,

				espan.SpanID: int64(hash.Fnv1aU8Hash(spanid.Byte())),

				espan.AppSpanSampled:    true,
				"_" + espan.EBPFTraceID: etraceid,
			}

			if i == 0 && appTrace {
				id, _ := ulid.ID()
				fields[espan.AppTraceIDL] = id.Low
				fields[espan.AppTraceIDH] = id.High
				fields[espan.AppParentIDL] = id.Low
			}

			kvs := point.NewKVs(fields)
			pt := point.NewPointV2("datakit-ebpf", kvs, point.WithTime(tn))
			pts = append(pts, pt)

			// first span is client span
			if direct == espan.OUTG {
				direct = espan.INCO
				nextNetID = append(nextNetID, *netTID)
				continue
			}
		}

		if (i == level-1) && tail_end {
			return pts, nil
		}
		nextNetID = []espan.ID128{}
		for j := 0; j < child; j++ {
			tn = tn.Add(time.Nanosecond * 50)
			id, _ := ulid.ID()
			nextNetID = append(nextNetID, *id)
			spanid, _ := ulid.ID()
			fields := map[string]any{
				espan.ReqSeq:            int64(id.Low),
				espan.RespSeq:           int64(id.High),
				espan.ThrTraceID:        int64(hash.Fnv1aU8Hash(entryTID.Byte())),
				espan.SpanID:            int64(hash.Fnv1aU8Hash(spanid.Byte())),
				espan.DirectionStr:      nexDirect,
				espan.EBPFSpanType:      "",
				espan.AppSpanSampled:    true,
				"_" + espan.EBPFTraceID: etraceid,
			}

			kvs := point.NewKVs(fields)
			pt := point.NewPointV2("datakit-ebpf", kvs, point.WithTime(tn))
			pts = append(pts, pt)
		}
	}

	return pts, nil
}

func MockTraceData(trace, level, spans int, tn time.Time, sampled bool) []*point.Point {
	var pts []*point.Point
	for i := 0; i < trace; i++ {
		var dr string
		var bval bool
		switch i % 2 {
		case 0:
			dr = espan.INCO
		default:
			dr = espan.OUTG
		}

		bval = !bval
		ptss, err := genMockTrace(dr, false, level, spans, bval, sampled, tn)
		if err == nil {
			pts = append(pts, ptss...)
		}
	}
	return pts
}
