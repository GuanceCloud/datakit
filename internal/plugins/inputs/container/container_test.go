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
	in := []string{"image:pubrepo.guance.com/kodo*", "namespace:kube-system"}
	ex := []string{"image:datakit*", "namespace:nginx"}

	filter, err := filter.NewFilter(in, ex)
	assert.Nil(t, err)

	c := &container{logFilter: filter}

	cases := []struct {
		in     *logInstance
		should bool
	}{
		{
			in: &logInstance{
				image:        "pubrepo.guance.com/kodo:1.11",
				podNamespace: "kube-system",
			},
			should: true,
		},
		{
			in: &logInstance{
				image:        "pubrepo.guance.com/kodo:1.12",
				podNamespace: "faker",
			},
			should: true,
		},
		{
			in: &logInstance{
				image:        "k8s.io/etcd:1.21",
				podNamespace: "kube-system",
			},
			should: true,
		},
		{
			in: &logInstance{
				image:        "pubrepo.guance.com/faker:1.11",
				podNamespace: "nginx",
			},
			should: false,
		},
		{
			in: &logInstance{
				image:        "nginx:1.12",
				podNamespace: "middleware",
			},
			should: false,
		},
		{
			in: &logInstance{
				image:        "datakit:1.12",
				podNamespace: "datakit",
			},
			should: false,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
			res := c.shouldPullContainerLog(tc.in)
			assert.Equal(t, tc.should, res)
		})
	}
}
