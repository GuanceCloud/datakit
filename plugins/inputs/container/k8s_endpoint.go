// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
)

var _ k8sResourceMetricInterface = (*endpoint)(nil)

type endpoint struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.Endpoints
	host      string
}

func newEndpoint(client k8sClientX, extraTags map[string]string, host string) *endpoint {
	return &endpoint{
		client:    client,
		extraTags: extraTags,
		host:      host,
	}
}

func (e *endpoint) getHost() string {
	return e.host
}

func (e *endpoint) name() string {
	return "endpoint"
}

func (e *endpoint) pullItems() error {
	list, err := e.client.getEndpoints().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get endpoints resource: %w", err)
	}
	e.items = list.Items
	return nil
}

func (e *endpoint) metric(election bool) (inputsMeas, error) {
	if err := e.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range e.items {
		met := &endpointMetric{
			tags: map[string]string{
				"endpoint":  item.Name,
				"namespace": item.Namespace,
			},
			fields: map[string]interface{}{
				"address_available": 0,
				"address_not_ready": 0,
			},
			election: election,
		}
		if e.host != "" {
			met.tags["host"] = e.host
		}

		var available, notReady int
		for _, subset := range item.Subsets {
			available += len(subset.Addresses)
			notReady += len(subset.NotReadyAddresses)
		}

		met.fields["address_available"] = available
		met.fields["address_not_ready"] = notReady

		met.tags.append(e.extraTags)
		res = append(res, met)
	}

	count, _ := e.count()
	for ns, c := range count {
		met := &endpointMetric{
			tags:     map[string]string{"namespace": ns},
			fields:   map[string]interface{}{"count": c},
			election: election,
		}
		if e.host != "" {
			met.tags["host"] = e.host
		}
		met.tags.append(e.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (e *endpoint) count() (map[string]int, error) {
	if err := e.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range e.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

type endpointMetric struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (e *endpointMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_endpoint", e.tags, e.fields, point.MOptElectionV2(e.election))
}

//nolint:lll
func (*endpointMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_endpoint",
		Desc: "Kubernetes Endpoints 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"endpoint":  inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"count":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of endpoints"},
			"address_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses available in endpoint."},
			"address_not_ready": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of addresses not ready in endpoint."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string, host string) k8sResourceMetricInterface {
		return newEndpoint(c, m, host)
	})
	registerMeasurement(&endpointMetric{})
}
