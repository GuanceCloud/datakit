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
		Desc: "kubernet endpoint",
		Tags: map[string]interface{}{
			"endpoint_name": &inputs.TagInfo{Desc: "endpoint name"},
			"namespace":     &inputs.TagInfo{Desc: "namespace"},
			"port_name":     &inputs.TagInfo{Desc: "The name of this port "},
			"port_protocol": &inputs.TagInfo{Desc: "The IP protocol for this port"},
		},
		Fields: map[string]interface{}{
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "created time",
			},
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "A sequence number representing a specific generation of the desired state",
			},
			"port": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "",
			},
			"ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "ready or not ready",
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

	for _, endpoint := range e.Subsets {
		for _, readyAddr := range endpoint.Addresses {
			tags := map[string]string{
				"endpoint_name": e.Name,
				"namespace":     e.Namespace,
				"hostname":      readyAddr.Hostname,
				// "node_name":     *readyAddr.NodeName,
			}

			fields := map[string]interface{}{
				"created":    e.GetCreationTimestamp().UnixNano(),
				"generation": e.Generation,
				"ready":      1,
			}

			if readyAddr.TargetRef != nil {
				tags[strings.ToLower(readyAddr.TargetRef.Kind)] = readyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				m := &endpointM{
					name:   endpointMeasurement,
					tags:   tags,
					fields: fields,
					ts:     time.Now(),
				}

				i.collectCache = append(i.collectCache, m)
			}
		}
		for _, notReadyAddr := range endpoint.NotReadyAddresses {
			tags := map[string]string{
				"endpoint_name": e.Name,
				"namespace":     e.Namespace,
				"hostname":      notReadyAddr.Hostname,
				// "node_name":     *notReadyAddr.NodeName,
			}

			fields := map[string]interface{}{
				"created":    e.GetCreationTimestamp().UnixNano(),
				"generation": e.Generation,
				"ready":      0,
			}

			if notReadyAddr.TargetRef != nil {
				tags[strings.ToLower(notReadyAddr.TargetRef.Kind)] = notReadyAddr.TargetRef.Name
			}

			for _, port := range endpoint.Ports {
				fields["port"] = port.Port

				tags["port_name"] = port.Name
				tags["port_protocol"] = string(port.Protocol)

				m := &endpointM{
					name:   endpointMeasurement,
					tags:   tags,
					fields: fields,
					ts:     time.Now(),
				}

				i.collectCache = append(i.collectCache, m)
			}
		}
	}
}
