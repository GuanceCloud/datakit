// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ebpftrace

import (
	"net/http"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/spans"
)

func apiBPFTracing(ulid *spans.ULID, mrr MRRunnerInterface) httpapi.APIHandler {
	return func(w http.ResponseWriter, req *http.Request, x ...interface{}) (interface{}, error) {
		if ulid == nil || mrr == nil {
			return nil, nil
		}
		var body []byte
		var err error

		opts := []point.Option{
			point.WithPrecision(point.PrecNS),
			point.WithTime(time.Now()),
		}

		body, err = uhttp.ReadBody(req)
		if err != nil {
			return nil, err
		}

		ct := httpapi.GetPointEncoding(req.Header)

		pts, err := httpapi.HandleWriteBody(body, ct, opts...)
		if err != nil {
			return nil, err
		}

		if len(pts) == 0 {
			return nil, httpapi.ErrNoPoints
		}

		for _, pt := range pts {
			id, _ := ulid.ID()
			pt.Add(spans.SpanID, int64(hash.Fnv1aU8Hash(id.Byte())))
		}

		if mrr != nil {
			mrr.InsertSpans(pts)
		}
		return nil, nil
	}
}
