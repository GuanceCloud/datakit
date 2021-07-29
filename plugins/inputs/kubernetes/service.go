package kubernetes

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesServiceName = "kubernetes_services"

type service struct {
	client interface {
		getServices() (*corev1.ServiceList, error)
	}
}

func (s service) Gather() {
	list, err := s.client.getServices()
	if err != nil {
		l.Errorf("failed of get services resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":         fmt.Sprintf("%v", obj.UID),
			"service_name": obj.Name,
			"cluster_name": obj.ClusterName,
			"namespace":    obj.Namespace,
			"type":         fmt.Sprintf("%v", obj.Spec.Type),
		}
		fields := map[string]interface{}{
			"age":                     int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"cluster_ip":              obj.Spec.ClusterIP,
			"external_name":           obj.Spec.ExternalName,
			"external_traffic_policy": fmt.Sprintf("%v", obj.Spec.ExternalTrafficPolicy),
			"session_affinity":        fmt.Sprintf("%v", obj.Spec.SessionAffinity),
		}

		addJSONStringToMap("external_ips", obj.Spec.ExternalIPs, fields)
		addJSONStringToMap("selectors", obj.Spec.Selector, fields)
		addJSONStringToMap("load_balancer_ingress", obj.Status.LoadBalancer, fields)

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesServiceName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*service) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*service) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesServiceName,
		Desc: kubernetesServiceName,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo(""),
			"service_name": inputs.NewTagInfo(""),
			"cluster_name": inputs.NewTagInfo(""),
			"namespace":    inputs.NewTagInfo(""),
			"type":         inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"cluster_ip":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"external_ips":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"ports":                   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"external_name":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"external_traffic_policy": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			//"ip_family":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"load_balancer_ingress":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"session_affinity":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"selectors":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
