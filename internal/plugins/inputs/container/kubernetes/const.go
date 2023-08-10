// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	globalName = "kubernetes"

	maxMessageLength = 256 * 1024 // 256KB
	metaV1ListOption = metav1.ListOptions{}
	metaV1GetOption  = metav1.GetOptions{}

	metricOpt  = &point.PointOption{Category: datakit.Metric, GlobalElectionTags: true}
	objectOpt  = &point.PointOption{Category: datakit.Object, GlobalElectionTags: true}
	loggingOpt = &point.PointOption{Category: datakit.Logging, GlobalElectionTags: true}

	metricResourceList = map[string]resourceHandle{}
	objectResourceList = map[string]resourceHandle{}
)

type measurement interface {
	inputs.Measurement
	namespace() string
	addExtraTags(map[string]string)
}

type resourceHandle func(context.Context, k8sClient) ([]measurement, error)

func registerMetricResource(name string, handle resourceHandle) {
	metricResourceList[name] = handle
}

func registerObjectResource(name string, handle resourceHandle) {
	objectResourceList[name] = handle
}

var pointMeasurements []inputs.Measurement

func PointMeasurement() []inputs.Measurement {
	return pointMeasurements
}

func registerMeasurement(mea inputs.Measurement) {
	pointMeasurements = append(pointMeasurements, mea)
}

func Name() string {
	return globalName
}

type contextKeyType string

const (
	canCollectPodMetricsKey   contextKeyType = "canCollectPodMetrics"
	setExtraK8sLabelAsTagsKey contextKeyType = "setExtraK8sLabelAsTags"
)
