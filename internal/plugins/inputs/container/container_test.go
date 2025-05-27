// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
)

func TestBuildLabelsOption(t *testing.T) {
	cases := []struct {
		inAsTagKeys    []string
		inCustomerKeys []string
		out            labelsOption
	}{
		{
			inAsTagKeys:    []string{"app"},
			inCustomerKeys: []string{},
			out:            labelsOption{all: false, keys: []string{"app"}},
		},
		{
			inAsTagKeys:    []string{"app", "name"},
			inCustomerKeys: []string{"sink-project", "name"},
			out:            labelsOption{all: false, keys: []string{"app", "name", "sink-project"}},
		},
		{
			inAsTagKeys:    []string{""},
			inCustomerKeys: []string{"sink-project", "name"},
			out:            labelsOption{all: true, keys: nil},
		},
	}

	for _, tc := range cases {
		res := buildLabelsOption(tc.inAsTagKeys, tc.inCustomerKeys)
		assert.Equal(t, tc.out, res)
	}
}

func TestShouldPullLogs(t *testing.T) {
	cases := []struct {
		include, exclude []string
		in               *logInstance
		should           bool
	}{
		{
			include: []string{"image:redis*", "namespace:kube-system"},
			in: &logInstance{
				image:        "redis:1.23",
				podNamespace: "kube-system",
			},
			should: true,
		},
		{
			include: []string{"image:redis*", "namespace:kube-system"},
			in: &logInstance{
				image:        "redis:1.23",
				podNamespace: "middleware",
			},
			should: false,
		},
		{
			// exclude
			exclude: []string{"image:redis*", "namespace:kube-system"},
			in: &logInstance{
				image:        "redis:1.23",
				podNamespace: "middleware",
			},
			should: false,
		},
		{
			// exclude
			exclude: []string{"image:redis*", "namespace:kube-system"},
			in: &logInstance{
				image:        "<invalid>:1.23",
				podNamespace: "kube-system",
			},
			should: false,
		},
		{
			// exclude
			exclude: []string{"image:redis*", "namespace:kube-system"},
			in: &logInstance{
				image:        "nginx:1.23",
				podNamespace: "middleware",
			},
			should: true,
		},
		{
			include: []string{"image:redis*", "namespace:kube-system"},
			exclude: []string{"image:nginx*", "namespace:middleware"},
			in: &logInstance{
				image:        "redis:1.23",
				podNamespace: "kube-system",
			},
			should: true,
		},
		{
			include: []string{"image:redis*", "namespace:kube-system"},
			exclude: []string{"image:nginx*", "namespace:middleware"},
			in: &logInstance{
				image:        "redis:1.23",
				podNamespace: "middleware",
			},
			should: false,
		},
		{
			include: []string{"image:redis*", "namespace:kube-system"},
			exclude: []string{"image:nginx*", "namespace:middleware"},
			in: &logInstance{
				image:        "nginx:1.12",
				podNamespace: "kube-system",
			},
			should: false,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			filter, err := filter.NewFilter(tc.include, tc.exclude)
			assert.Nil(t, err)

			c := &container{logFilter: filter}

			res := c.shouldPullContainerLog(tc.in)
			assert.Equal(t, tc.should, res)
		})
	}
}
