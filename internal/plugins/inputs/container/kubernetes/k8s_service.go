// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	serviceObjectMeasurement = "kubernetes_services"
)

//nolint:gochecknoinits
func init() {
	registerResource("service", true, newService)
	registerMeasurements(&serviceObject{})
}

type service struct {
	client    k8sClient
	continued string
}

func newService(client k8sClient) resource {
	return &service{client: client}
}

func (s *service) hasNext() bool { return s.continued != "" }

func (s *service) getMetadata(ctx context.Context, ns string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:    queryLimit,
		Continue: s.continued,
	}

	list, err := s.client.GetServices(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	s.continued = list.Continue
	return &serviceMetadata{list}, nil
}

type serviceMetadata struct {
	list *apicorev1.ServiceList
}

func (m *serviceMetadata) transformMetric() pointKVs {
	return nil
}

func (m *serviceMetadata) transformObject() pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(serviceObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("service_name", item.Name)

		obj.SetTag("namespace", item.Namespace)
		obj.SetTag("type", fmt.Sprintf("%v", item.Spec.Type))

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("cluster_ip", item.Spec.ClusterIP)
		obj.SetField("external_name", item.Spec.ExternalName)
		obj.SetField("external_traffic_policy", fmt.Sprintf("%v", item.Spec.ExternalTrafficPolicy))
		obj.SetField("session_affinity", fmt.Sprintf("%v", item.Spec.SessionAffinity))
		obj.SetField("external_ips", strings.Join(item.Spec.ExternalIPs, ","))

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, obj)
	}

	return res
}

type serviceObject struct{}

//nolint:lll
func (*serviceObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: serviceObjectMeasurement,
		Desc: "The object of the Kubernetes Service.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("The UID of Service"),
			"uid":          inputs.NewTagInfo("The UID of Service"),
			"service_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"type":         inputs.NewTagInfo("type determines how the Service is exposed. Defaults to ClusterIP. (ClusterIP/NodePort/LoadBalancer/ExternalName)"),
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
