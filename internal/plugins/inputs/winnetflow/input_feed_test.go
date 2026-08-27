// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"errors"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type feedMetadataRecorder struct {
	sources    []string
	categories []point.Category
	feedErr    error
	lastErrors []metrics.LastError
}

func (r *feedMetadataRecorder) Feed(cat point.Category, _ []*point.Point, opts ...dkio.FeedOption) error {
	fd := dkio.GetFeedData()
	for _, opt := range opts {
		if opt != nil {
			opt(fd)
		}
	}
	r.sources = append(r.sources, fd.GetFeedSource())
	r.categories = append(r.categories, cat)
	return r.feedErr
}

func (r *feedMetadataRecorder) FeedLastError(_ string, opts ...metrics.LastErrorOption) {
	le := metrics.NewLastError()
	for _, opt := range opts {
		if opt != nil {
			opt(le)
		}
	}
	r.lastErrors = append(r.lastErrors, *le)
}

func TestFeedPointsMetadata(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "netflow", source: netflowSource},
		{name: "httpflow", source: httpflowSource},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &feedMetadataRecorder{}
			ipt := &Input{feeder: recorder}
			ipt.feedPoints(tc.source, []*point.Point{{}})

			if len(recorder.sources) != 1 || recorder.sources[0] != tc.source {
				t.Fatalf("feed sources = %v, want [%s]", recorder.sources, tc.source)
			}
			if len(recorder.categories) != 1 || recorder.categories[0] != point.Network {
				t.Fatalf("feed categories = %v, want [%s]", recorder.categories, point.Network)
			}
		})
	}
}

func TestFeedPointsErrorMetadata(t *testing.T) {
	recorder := &feedMetadataRecorder{feedErr: errors.New("feed failed")}
	ipt := &Input{feeder: recorder}
	ipt.feedPoints(netflowSource, []*point.Point{{}})

	if len(recorder.lastErrors) != 1 {
		t.Fatalf("last errors = %d, want 1", len(recorder.lastErrors))
	}
	le := recorder.lastErrors[0]
	if le.Input != inputName || le.Source != netflowSource {
		t.Fatalf("last error input/source = %q/%q, want %q/%q",
			le.Input, le.Source, inputName, netflowSource)
	}
	if len(le.Categories) != 1 || le.Categories[0] != point.Network {
		t.Fatalf("last error categories = %v, want [%s]", le.Categories, point.Network)
	}
}
