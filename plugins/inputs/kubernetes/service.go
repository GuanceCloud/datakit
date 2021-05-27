package kubernetes

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

var serviceMeasurement = "kube_deployment"

type serviceM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *serviceM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *serviceM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: deploymentMeasurement,
		Desc: "kubernet daemonSet 对象",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "pod name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"nodeName":  &inputs.TagInfo{Desc: "node name"},
		},
		Fields: map[string]interface{}{
			"ready": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "容器ready数/总数",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 状态",
			},
			"restarts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "重启次数",
			},
			"age": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod存活时长",
			},
			"podIp": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod ip",
			},
			"createTime": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 创建时间",
			},
			"label_xxx": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod lable",
			},
		},
	}
}

func (i *Input) collectServices(ctx context.Context) error {
	list, err := i.client.getServices(ctx)
	if err != nil {
		return err
	}
	for _, item := range list.Items {
		i.gatherService(item)
	}

	return nil
}

func (i *Input) gatherService(s corev1.Service) {
	if s.GetCreationTimestamp().Second() == 0 && s.GetCreationTimestamp().Nanosecond() == 0 {
		return
	}

	fields := map[string]interface{}{
		"created":    s.GetCreationTimestamp().UnixNano(),
		"generation": s.Generation,
	}

	tags := map[string]string{
		"service_name": s.Name,
		"namespace":    s.Namespace,
	}

	// for key, val := range s.Spec.Selector {
	// 	if i.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	var getPorts = func() {
		for _, port := range s.Spec.Ports {
			fields["port"] = port.Port
			fields["target_port"] = port.TargetPort.IntVal

			tags["port_name"] = port.Name
			tags["port_protocol"] = string(port.Protocol)

			if s.Spec.Type == "ExternalName" {
				tags["external_name"] = s.Spec.ExternalName
			} else {
				tags["cluster_ip"] = s.Spec.ClusterIP
			}

			m := &serviceM{
				name:   deploymentMeasurement,
				tags:   tags,
				fields: fields,
				ts:     time.Now(),
			}

			i.collectCache = append(i.collectCache, m)
		}
	}

	if externIPs := s.Spec.ExternalIPs; externIPs != nil {
		for _, ip := range externIPs {
			tags["ip"] = ip

			getPorts()
		}
	} else {
		getPorts()
	}
}
