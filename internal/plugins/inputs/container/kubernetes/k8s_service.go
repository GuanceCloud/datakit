// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

const (
	// serviceType              = "Service".
	serviceMetricMeasurement = "kube_service"
	serviceObjectClass       = "kubernetes_services"
	serviceObjectResourceKey = "service_name"
)

//nolint:gochecknoinits
func init() {
	registerResource("service", false, newService)
}

type service struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newService(client k8sClient, cfg *Config) resource {
	return &service{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (s *service) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := s.client.GetServices(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := s.buildMetricPoints(list, timestamp)
		feedMetric("k8s-service-metric", s.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(s.cfg, "service", s.counter, timestamp)
}

func (s *service) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := s.client.GetServices(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := s.buildObjectPoints(list)
		feedObject("k8s-service-object", s.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (*service) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (s *service) buildMetricPoints(list *apicorev1.ServiceList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("service", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.Add("ports", len(item.Spec.Ports))

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, s.cfg.LabelAsTagsForMetric.All, s.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(s.cfg.ExtraTags)...)
		pt := point.NewPoint(serviceMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		s.counter[item.Namespace]++
	}

	return pts
}

func (s *service) buildObjectPoints(list *apicorev1.ServiceList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag(serviceObjectResourceKey, item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)
		kvs = kvs.AddTag("type", string(item.Spec.Type))

		kvs = kvs.Add("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		kvs = kvs.Add("cluster_ip", item.Spec.ClusterIP)
		kvs = kvs.Add("external_name", item.Spec.ExternalName)
		kvs = kvs.Add("external_traffic_policy", string(item.Spec.ExternalTrafficPolicy))
		kvs = kvs.Add("session_affinity", string(item.Spec.SessionAffinity))
		kvs = kvs.Add("external_ips", strings.Join(item.Spec.ExternalIPs, ","))

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.Add("yaml", string(y))
		}
		kvs = kvs.Add("annotations", pointutil.MapToJSON(item.Annotations))
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.Add("message", pointutil.TrimString(msg, maxMessageLength))

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		kvs = append(kvs, point.NewTags(item.Spec.Selector)...)

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, s.cfg.LabelAsTagsForNonMetric.All, s.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(s.cfg.ExtraTags)...)
		pt := point.NewPoint(serviceObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type ServiceMetric struct{}

//nolint:lll
func (*ServiceMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: serviceMetricMeasurement,
		Desc: "The metric of the Kubernetes Service.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of Service"),
			"service":          inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"ports": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.Count, Desc: "Total number of ports that are exposed by this service."},
		},
	}
}

type ServiceObject struct{}

//nolint:lll
func (*ServiceObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: serviceObjectClass,
		Desc: "The object of the Kubernetes Service.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                 inputs.NewTagInfo("The UID of Service"),
			"uid":                  inputs.NewTagInfo("The UID of Service"),
			"service_name":         inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":            inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"type":                 inputs.NewTagInfo("Type determines how the Service is exposed. Defaults to ClusterIP. (ClusterIP/NodePort/LoadBalancer/ExternalName)"),
			"cluster_name_k8s":     inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"&lt;ALL-SELECTOR&gt;": inputs.NewTagInfo("Represents the selector for Kubernetes resources"),
		},
		Fields: map[string]interface{}{
			"age":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"cluster_ip":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "ClusterIP is the IP address of the service and is usually assigned randomly by the master."},
			"external_ips":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "ExternalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service."},
			"external_name":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "ExternalName is the external reference that kubedns or equivalent will return as a CNAME record for this service."},
			"external_traffic_policy": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "ExternalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints."},
			"session_affinity":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: `Supports "ClientIP" and "None".`},
			"message":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}
