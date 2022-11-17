// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var (
	_ k8sResourceObjectInterface = (*service)(nil)
	_ k8sResourceMetricInterface = (*service)(nil)
)

type service struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.Service
	host      string
}

func newService(client k8sClientX, extraTags map[string]string, host string) *service {
	return &service{
		client:    client,
		extraTags: extraTags,
		host:      host,
	}
}

func (s *service) getHost() string {
	return s.host
}

func (s *service) name() string {
	return "service"
}

func (s *service) count() (map[string]int, error) {
	if err := s.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range s.items {
		m[defaultNamespace(item.Namespace)]++
	}

	if len(m) == 0 {
		m["default"] = 0
	}

	return m, nil
}

func (s *service) pullItems() error {
	list, err := s.client.getServices().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get services resource: %w", err)
	}
	s.items = list.Items
	return nil
}

func (s *service) metric(election bool) (inputsMeas, error) {
	return nil, nil
}

func (s *service) object(election bool) (inputsMeas, error) {
	if err := s.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range s.items {
		obj := &serviceObject{
			tags: map[string]string{
				"name":         fmt.Sprintf("%v", item.UID),
				"service_name": item.Name,
				"type":         fmt.Sprintf("%v", item.Spec.Type),
				"namespace":    defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":                     int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"cluster_ip":              item.Spec.ClusterIP,
				"external_name":           item.Spec.ExternalName,
				"external_traffic_policy": fmt.Sprintf("%v", item.Spec.ExternalTrafficPolicy),
				"session_affinity":        fmt.Sprintf("%v", item.Spec.SessionAffinity),
			},
			election: election,
		}
		if s.host != "" {
			obj.tags["host"] = s.host
		}

		if y, err := yaml.Marshal(item); err != nil {
			l.Debugf("failed to get service yaml %s, namespace %s, name %s, ignored", err.Error(), item.Namespace, item.Name)
		} else {
			obj.fields["yaml"] = string(y)
		}

		obj.tags.append(s.extraTags)

		obj.fields.addSlice("external_ips", item.Spec.ExternalIPs)
		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")
		obj.fields.delete("yaml")

		res = append(res, obj)
	}

	return res, nil
}

type serviceMeta struct{ *v1.Service }

func (item *serviceMeta) servicePort(name string) int {
	for _, s := range item.Spec.Ports {
		if s.Name == name {
			return int(s.Port)
		}
	}
	return -1
}

type serviceObject struct {
	tags     tagsType
	fields   fieldsType
	election bool
}

func (s *serviceObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_services", s.tags, s.fields, point.OOptElectionV2(s.election))
}

//nolint:lll
func (*serviceObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_services",
		Desc: "Kubernetes service 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"service_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":    inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"type":         inputs.NewTagInfo("type determines how the Service is exposed. Defaults to ClusterIP. (ClusterIP/NodePort/LoadBalancer/ExternalName)"),
		},
		Fields: map[string]interface{}{
			"age":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"cluster_ip":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "clusterIP is the IP address of the service and is usually assigned randomly by the master."},
			"external_ips":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "externalIPs is a list of IP addresses for which nodes in the cluster will also accept traffic for this service."},
			"external_name":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "externalName is the external reference that kubedns or equivalent will return as a CNAME record for this service."},
			"external_traffic_policy": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "externalTrafficPolicy denotes if this Service desires to route external traffic to node-local or cluster-wide endpoints."},
			"session_affinity":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: `Supports "ClientIP" and "None".`},
			"message":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string, host string) k8sResourceMetricInterface {
		return newService(c, m, host)
	})
	registerK8sResourceObject(func(c k8sClientX, m map[string]string, host string) k8sResourceObjectInterface {
		return newService(c, m, host)
	})
	registerMeasurement(&serviceObject{})
}
