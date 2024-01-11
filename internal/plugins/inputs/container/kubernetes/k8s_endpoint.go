// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	endpointMetricMeasurement = "kube_endpoint"
)

//nolint:gochecknoinits
func init() {
	registerResource("endpoint", true, false, newEndpoint)
	registerMeasurements(&endpointMetric{})
}

type endpoint struct {
	client    k8sClient
	continued string
	counter   map[string]int
}

func newEndpoint(client k8sClient) resource {
	return &endpoint{client: client, counter: make(map[string]int)}
}

func (e *endpoint) count() []pointV2 { return buildCountPoints("endpoint", e.counter) }

func (e *endpoint) hasNext() bool { return e.continued != "" }

func (e *endpoint) getMetadata(ctx context.Context, ns, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      e.continued,
		FieldSelector: fieldSelector,
	}

	list, err := e.client.GetEndpoints(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	e.continued = list.Continue
	return &endpointMetadata{e, list}, nil
}

type endpointMetadata struct {
	parent *endpoint
	list   *apicorev1.EndpointsList
}

func (m *endpointMetadata) newMetric(conf *Config) pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		met := typed.NewPointKV(endpointMetricMeasurement)

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

		met.SetLabelAsTags(item.Labels, conf.LabelAsTagsForMetric.All, conf.LabelAsTagsForMetric.Keys)
		res = append(res, met)

		m.parent.counter[item.Namespace]++
	}

	return res
}

func (m *endpointMetadata) newObject(conf *Config) pointKVs {
	return nil
}

type endpointMetric struct{}

//nolint:lll
func (*endpointMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: endpointMetricMeasurement,
		Desc: "The metric of the Kubernetes Endpoints.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of Endpoint."),
			"endpoint":         inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"address_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses available in endpoint."},
			"address_not_ready": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses not ready in endpoint."},
		},
	}
}
