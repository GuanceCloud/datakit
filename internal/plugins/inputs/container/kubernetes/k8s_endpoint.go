// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	apicorev1 "k8s.io/api/core/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("endpoint", gatherEndpointMetric)
	registerMeasurement(&endpointMetric{})
}

func gatherEndpointMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetEndpoints().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeEndpointMetric(list), nil
}

func composeEndpointMetric(list *apicorev1.EndpointsList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("endpoint", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("address_available", 0)
		met.SetField("address_not_ready", 0)

		var available, notReady int
		for _, subset := range item.Subsets {
			available += len(subset.Addresses)
			notReady += len(subset.NotReadyAddresses)
		}

		met.SetField("address_available", available)
		met.SetField("address_not_ready", notReady)

		res = append(res, &endpointMetric{met})
	}

	return res
}

type endpointMetric struct{ typed.PointKV }

func (e *endpointMetric) namespace() string { return e.GetTag("namespace") }

func (e *endpointMetric) addExtraTags(m map[string]string) { e.SetTags(m) }

func (e *endpointMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_endpoint", e.Tags(), e.Fields(), metricOpt)
}

//nolint:lll
func (*endpointMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_endpoint",
		Desc: "The metric of the Kubernetes Endpoints.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":       inputs.NewTagInfo("The UID of endpoint."),
			"endpoint":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"address_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses available in endpoint."},
			"address_not_ready": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses not ready in endpoint."},
		},
	}
}
