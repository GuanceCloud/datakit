// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	in := []string{"image:pubrepo.guance.com/kodo", "image_name:kodo"}
	ex := []string{"image_name:*", "namespace:datakit-ns"}

	filters, err := newFilters(in, ex)
	assert.Nil(t, err)

	c := &container{loggingFilters: filters}

	cases := []struct {
		inType  filterType
		inField string
		ignored bool
	}{
		{
			inType:  filterImage,
			inField: "nginx:v1.22",
			ignored: true,
		},
		{
			inType:  filterImage,
			inField: "pubrepo.guance.com/kodo",
			ignored: false,
		},
		{
			inType:  filterImageName,
			inField: "kodo",
			ignored: true,
		},
		{
			inType:  filterImageName,
			inField: "nginx",
			ignored: true,
		},
		{
			inType:  filterNamespace,
			inField: "datakit-ns",
			ignored: true,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			res := c.ignoreContainerLogging(tc.inType, tc.inField)
			assert.Equal(t, tc.ignored, res)
		})
	}
}
