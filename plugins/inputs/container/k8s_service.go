package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
)

const k8sServiceName = "kubernetes_services"

func gatherService(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getServices().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get services resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportService(list.Items, extraTags), nil
}

func exportService(items []v1.Service, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newService()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["service_name"] = item.Name
		obj.tags["type"] = fmt.Sprintf("%v", item.Spec.Type)

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["cluster_ip"] = item.Spec.ClusterIP
		obj.fields["external_name"] = item.Spec.ExternalName
		obj.fields["external_traffic_policy"] = fmt.Sprintf("%v", item.Spec.ExternalTrafficPolicy)
		obj.fields["session_affinity"] = fmt.Sprintf("%v", item.Spec.SessionAffinity)

		// obj.fields.addSlice("selectors", item.Spec.Selector)
		// obj.fields.addSlice("load_balancer_ingress", item.Status.LoadBalancer)
		obj.fields.addSlice("external_ips", item.Spec.ExternalIPs)
		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type service struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newService() *service {
	return &service{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (s *service) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sServiceName, s.tags, s.fields, &io.PointOption{Time: s.time, Category: datakit.Object})
}

//nolint:lll
func (*service) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sServiceName,
		Desc: "Kubernetes service 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"service_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
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
			"annotations":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
			// TODO:
			// "load_balancer_ingress":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "selectors":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "ip_family":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "ports":                   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&service{})
}
