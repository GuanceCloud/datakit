package kubernetes

import (
	"context"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

var endpointMeasurement = "kube_endpoint"

type endpointM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *endpointM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *endpointM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: endpointMeasurement,
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

func (i *Input) collectEndpoints(ctx context.Context) error {
	list, err := i.client.getEndpoints(ctx)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		i.gatherEndpoint(item)
	}

	return nil
}

func (i *Input) gatherEndpoint(e corev1.Endpoints) {
	if e.GetCreationTimestamp().Second() == 0 && e.GetCreationTimestamp().Nanosecond() == 0 {
		return
	}

	fields := map[string]interface{}{
		"created":    e.GetCreationTimestamp().UnixNano(),
		"generation": e.Generation,
	}

	tags := map[string]string{
		"endpoint_name": e.Name,
		"namespace":     e.Namespace,
	}

	for _, endpoint := range e.Subsets {
		for _, readyAddr := range endpoint.Addresses {
			fields["ready"] = true

			tags["hostname"] = readyAddr.Hostname
			// tags["node_name"] = *readyAddr.NodeName
			if readyAddr.TargetRef != nil {
				tags[strings.ToLower(readyAddr.TargetRef.Kind)] = readyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				m := &endpointM{
					name:   deploymentMeasurement,
					tags:   tags,
					fields: fields,
					ts:     time.Now(),
				}

				i.collectCache = append(i.collectCache, m)
			}
		}

		for _, notReadyAddr := range endpoint.NotReadyAddresses {
			fields["ready"] = false

			tags["hostname"] = notReadyAddr.Hostname
			tags["node_name"] = *notReadyAddr.NodeName
			if notReadyAddr.TargetRef != nil {
				tags[strings.ToLower(notReadyAddr.TargetRef.Kind)] = notReadyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				m := &endpointM{
					name:   deploymentMeasurement,
					tags:   tags,
					fields: fields,
					ts:     time.Now(),
				}

				i.collectCache = append(i.collectCache, m)
			}
		}
	}
}
