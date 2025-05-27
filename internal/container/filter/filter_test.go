// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
)

func TestSplitRules(t *testing.T) {
	cases := []struct {
		in  []string
		out [][]string
	}{
		{
			in: []string{"image:*"},
			out: [][]string{
				{"**"}, // filterImage
				nil,    // filterImageName
				nil,    // filterImageShortName
				nil,    // filterNamespace
			},
		},
		{
			in: []string{" image: pubrepo.guance.com/datakit:1.18.0 "},
			out: [][]string{
				{"pubrepo.guance.com/datakit:1.18.0"},
				nil,
				nil,
				nil,
			},
		},
		{
			in: []string{"image:pubrepo.guance.com/datakit:1.18.0", "image:*"},
			out: [][]string{
				{"pubrepo.guance.com/datakit:1.18.0", "**"},
				nil,
				nil,
				nil,
			},
		},
		{
			in: []string{"image_name:pubrepo.guance.com/datakit"},
			out: [][]string{
				nil,
				{"pubrepo.guance.com/datakit"},
				nil,
				nil,
			},
		},
		{
			in: []string{"image_short_name:datakit"},
			out: [][]string{
				nil,
				nil,
				{"datakit"},
				nil,
			},
		},
		{
			in: []string{"image_short_name:datakit", "image_short_name:kodo*"},
			out: [][]string{
				nil,
				nil,
				{"datakit", "kodo*"},
				nil,
			},
		},
		{
			in: []string{"namespace:datakit-ns"},
			out: [][]string{
				nil,
				nil,
				nil,
				{"datakit-ns"},
			},
		},
		{
			in: []string{"image_short_name:datakit", "namespace:datakit-ns"},
			out: [][]string{
				nil,
				nil,
				{"datakit"},
				{"datakit-ns"},
			},
		},
	}

	for _, tc := range cases {
		res := splitRules(tc.in)
		assert.Equal(t, tc.out, res)
	}
}

func TestNewFilter(t *testing.T) {
	cases := []struct {
		include []string
		exclude []string
		filter  Filter
	}{
		{
			include: []string{"image:pubrepo.guance.com/kodo", "image_name:kodo"},
			exclude: []string{"image_name:datakit*", "namespace:testns"},
			filter: []filter.Filter{
				func() filter.Filter {
					x, _ := filter.NewIncludeExcludeFilter([]string{"pubrepo.guance.com/kodo"}, nil)
					return x
				}(),
				// filterImageName
				func() filter.Filter {
					x, _ := filter.NewIncludeExcludeFilter([]string{"kodo"}, []string{"datakit*"})
					return x
				}(),
				// filterImageShortName
				nil,
				// filterNamespace
				func() filter.Filter {
					x, _ := filter.NewIncludeExcludeFilter(nil, []string{"testns"})
					return x
				}(),
			},
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			filter, err := NewFilter(tc.include, tc.exclude)
			assert.Nil(t, err)
			assert.Equal(t, tc.filter, filter)
		})
	}
}

func TestMatchFilter(t *testing.T) {
	in := []string{"image:pubrepo.guance.com/kodo*", "namespace:kube-system"}
	ex := []string{"image:datakit*"}

	filter, err := NewFilter(in, ex)
	assert.Nil(t, err)

	cases := []struct {
		inType  FilterType
		inField string
		matched bool
	}{
		{
			inType:  FilterImage,
			inField: "pubrepo.guance.com/kodo:1.11",
			matched: true,
		},
		{
			inType:  FilterImage,
			inField: "pubrepo.guance.com/datakit:1.11",
			matched: false,
		},
		{
			inType:  FilterImageName,
			inField: "nginx",
			matched: true,
		},
		{
			inType:  FilterNamespace,
			inField: "kube-system",
			matched: true,
		},
		{
			inType:  FilterNamespace,
			inField: "datakit-ns",
			matched: false,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			res := filter.Match(tc.inType, tc.inField)
			assert.Equal(t, tc.matched, res)
		})
	}
}
