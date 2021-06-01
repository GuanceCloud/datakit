package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

var serviceMeasurement = "kube_service"

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
		Name: serviceMeasurement,
		Desc: "kubernetes service",
		Tags: map[string]interface{}{
			"service_name": &inputs.TagInfo{Desc: "service name"},
			"namespace":    &inputs.TagInfo{Desc: "namespace"},
			"type":         &inputs.TagInfo{Desc: "service type"},
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
			"cluster_ip": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "clusterIP is the IP address of the service",
			},
			"external_ip": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "externalIPs is a list of IP addresses for which nodes in the cluster",
			},
			"ports": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The list of ports that are exposed by this service",
			},
		},
	}
}

func (i *Input) collectServices() error {
	list, err := i.client.getServices()
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

	servicePorts := make([]string, 0, len(s.Spec.Ports))
	for _, p := range s.Spec.Ports {
		servicePorts = append(servicePorts, fmt.Sprintf("%d:%d/%s", p.Port, p.NodePort, p.Protocol))
	}

	externalIPs := make([]string, 0, len(s.Spec.ExternalIPs))
	for _, ip := range s.Spec.ExternalIPs {
		externalIPs = append(externalIPs, ip)
	}
	var externalIPsStr = "<none>"
	if len(externalIPs) > 0 {
		externalIPsStr = strings.Join(externalIPs, ",")
	}

	fields := map[string]interface{}{
		"created":     s.GetCreationTimestamp().UnixNano(),
		"generation":  s.Generation,
		"cluster_ip":  s.Spec.ClusterIP,
		"external_ip": externalIPsStr,
		"ports":       strings.Join(servicePorts, ","),
	}

	tags := map[string]string{
		"service_name": s.Name,
		"namespace":    s.Namespace,
		"type":         string(s.Spec.Type),
	}

	for key, val := range s.Spec.Selector {
		tags["selector_"+key] = val
	}

	m := &serviceM{
		name:   serviceMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}
