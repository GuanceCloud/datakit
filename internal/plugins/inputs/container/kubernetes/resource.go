// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
)

var (
	maxMessageLength   int   = 256 * 1024 // 256KB
	queryLimit         int64 = 100
	allNamespaces            = ""
	emptyFieldSelector       = ""

	measurements []inputs.Measurement
)

func Measurements() []inputs.Measurement {
	return measurements
}

func registerMeasurements(meas ...inputs.Measurement) {
	measurements = append(measurements, meas...)
}

type (
	resourceConstructor func(client k8sClient, cfg *Config) resource

	resource interface {
		gatherMetric(ctx context.Context, timestamp int64 /*nanoseconds*/)
		gatherObject(ctx context.Context)
		addObjectChangeInformer(informerFactory informers.SharedInformerFactory)
	}
)

var (
	nodeLocalResources         []resourceConstructor // pod, dfpv
	nodeLocalResourcesNames    []string
	nonNodeLocalResources      []resourceConstructor // other
	nonNodeLocalResourcesNames []string
)

func registerResource(name string, nodeLocal bool, rc resourceConstructor) {
	name = "k8s-" + name
	if nodeLocal {
		nodeLocalResources = append(nodeLocalResources, rc)
		nodeLocalResourcesNames = append(nodeLocalResourcesNames, name)
		return
	}
	nonNodeLocalResources = append(nonNodeLocalResources, rc)
	nonNodeLocalResourcesNames = append(nonNodeLocalResourcesNames, name)
}

type LabelsOption struct {
	All  bool
	Keys []string
}

func newListOptions(fieldSelector, continued string) metav1.ListOptions {
	return metav1.ListOptions{
		Limit:         queryLimit,
		FieldSelector: fieldSelector,
		Continue:      continued,
	}
}
