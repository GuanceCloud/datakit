// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

func TestMRTrace(t *testing.T) {
	ptsSrc := MockTraceData(1, 2, 2, time.Now(), true)

	t.Log(len(ptsSrc))

	tn := time.Now()

	metaLi := espan.ESpans{}
	uid, _ := espan.NewRandID()

	id1, _ := uid.ID()
	id2, _ := uid.ID()
	pts := []*point.Point{
		point.NewPointV2("dketrace", point.NewKVs(map[string]any{
			espan.DirectionStr: espan.OUTG,
			"__id__":           "1_1",
			espan.ReqSeq:       1,
			espan.RespSeq:      2,
			espan.ThrTraceID:   11111,
			espan.SpanID:       int64(hash.Fnv1aU8Hash(id1.Byte())),
			espan.EBPFSpanType: "exit",
		})),
		point.NewPointV2("dketrace", point.NewKVs(map[string]any{
			espan.DirectionStr: espan.INCO,
			"__id__":           "2_1",
			espan.ReqSeq:       1,
			espan.RespSeq:      2,
			espan.ThrTraceID:   111112,
			espan.SpanID:       int64(hash.Fnv1aU8Hash(id2.Byte())),
			espan.EBPFSpanType: espan.SpanTypEntry,
		})),
	}

	for _, pt := range pts {
		meta, ok := espan.SpanMetaData(pt)
		if !ok {
			t.Error("!ok")
		} else {
			pts = append(pts, pt)
			metaLi = append(metaLi, &espan.EBPFSpan{Meta: meta})
		}
	}

	mrr := MRRunner{}
	metadata := MRMetaData{
		firstHalf: metaLi,
	}
	mrr.connectSpans(&metadata)
	info, _ := mrr.linkAndgatherTrace(&metadata)

	for _, v := range info {
		t.Log(*v)
	}
	t.Log(time.Since(tn))
}
