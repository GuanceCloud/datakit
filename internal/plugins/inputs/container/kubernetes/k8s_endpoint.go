// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

const (
	endpointMetricMeasurement = "kube_endpoint"
)

//nolint:gochecknoinits
func init() {
	registerResource("endpoint", false, newEndpoint)
	registerMeasurements(&endpointMetric{})
}

type endpoint struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newEndpoint(client k8sClient, cfg *Config) resource {
	return &endpoint{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (e *endpoint) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := e.client.GetEndpoints(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := e.buildMetricPoints(list, timestamp)
		feedMetric("k8s-endpoint-metric", e.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(e.cfg, "endpoint", e.counter, timestamp)
}

func (*endpoint) gatherObject(_ context.Context)                      { /* nil */ }
func (*endpoint) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (e *endpoint) buildMetricPoints(list *apicorev1.EndpointsList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("endpoint", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		var available, notReady int
		for _, subset := range item.Subsets {
			available += len(subset.Addresses)
			notReady += len(subset.NotReadyAddresses)
		}

		kvs = kvs.AddV2("address_available", available, false)
		kvs = kvs.AddV2("address_not_ready", notReady, false)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, e.cfg.LabelAsTagsForMetric.All, e.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(e.cfg.ExtraTags)...)
		pt := point.NewPointV2(endpointMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		e.counter[item.Namespace]++
	}

	return pts
}

type endpointMetric struct{}

//nolint:lll
func (*endpointMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: endpointMetricMeasurement,
		Desc: "The metric of the Kubernetes Endpoints.",
		Cat:  point.Metric,
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
