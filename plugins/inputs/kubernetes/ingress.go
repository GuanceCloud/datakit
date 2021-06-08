package kubernetes

import (
	// "context"
	"strconv"
	"strings"
	"time"

	v1beta1 "k8s.io/api/extensions/v1beta1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var ingressMeasurement = "kube_ingress"

type ingressM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ingressM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *ingressM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: ingressMeasurement,
		Desc: "kubernet ingress",
		Tags: map[string]interface{}{
			"ingress_name": &inputs.TagInfo{Desc: "pod name"},
			"namespace":    &inputs.TagInfo{Desc: "namespace"},
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
			"ingress_hosts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "A list of host rules used to configure the Ingress",
			},
			"ingress_address": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "LoadBalancerStatus represents the status of a load-balancer",
			},
			"service_Ports": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Specifies the port of the referenced service",
			},
		},
	}
}

func (i *Input) collectIngress(collector string) error {
	list, err := i.client.getIngress()
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		i.gatherIngress(collector, item)
	}

	return err
}

func (i *Input) gatherIngress(collector string, in v1beta1.Ingress) {
	if in.GetCreationTimestamp().Second() == 0 && in.GetCreationTimestamp().Nanosecond() == 0 {
		return
	}

	ingressHosts := make([]string, 0, len(in.Spec.Rules))
	for _, rule := range in.Spec.Rules {
		ingressHosts = append(ingressHosts, rule.Host)
	}
	// 获取 Address
	ingressAddress := make([]string, 0, len(in.Status.LoadBalancer.Ingress))
	for _, ingStatus := range in.Status.LoadBalancer.Ingress {
		ingressAddress = append(ingressAddress, ingStatus.IP)
	}
	// 获取 Service 的端口
	servicePortsSet := make(map[int]struct{})
	for _, rule := range in.Spec.Rules {
		for _, path := range rule.IngressRuleValue.HTTP.Paths {
			servicePortsSet[path.Backend.ServicePort.IntValue()] = struct{}{}
		}
	}
	servicePorts := make([]string, 0, len(servicePortsSet))
	for port := range servicePortsSet {
		servicePorts = append(servicePorts, strconv.Itoa(port))
	}

	fields := map[string]interface{}{
		"created":         in.GetCreationTimestamp().UnixNano(),
		"generation":      in.Generation,
		"ingress_hosts":   strings.Join(ingressHosts, ","),
		"ingress_address": strings.Join(ingressAddress, ","),
		"service_ports":   strings.Join(servicePorts, ","),
	}

	tags := map[string]string{
		"ingress_name": in.Name,
		"namespace":    in.Namespace,
	}

	for key, value := range i.Tags {
		tags[key] = value
	}

	m := &ingressM{
		name:   ingressMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache[collector] = append(i.collectCache[collector], m)
}
