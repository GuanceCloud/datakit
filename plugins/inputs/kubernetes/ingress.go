package kubernetes

import (
	// "context"
	"time"

	// netv1 "k8s.io/api/networking/v1"
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

// func (i *Input) collectIngress(ctx context.Context) error {
// 	list, err := i.client.getIngress(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	for _, i := range list.Items {
// 		i.gatherIngress(i)
// 	}

// 	return err
// }

// func (i *Input) gatherIngress(in netv1.Ingress) {
// 	if in.GetCreationTimestamp().Second() == 0 && in.GetCreationTimestamp().Nanosecond() == 0 {
// 		return
// 	}

// 	fields := map[string]interface{}{
// 		"created":    in.GetCreationTimestamp().UnixNano(),
// 		"generation": in.Generation,
// 	}

// 	tags := map[string]string{
// 		"ingress_name": in.Name,
// 		"namespace":    in.Namespace,
// 	}

// 	for _, ingress := range in.Status.LoadBalancer.Ingress {
// 		tags["hostname"] = ingress.Hostname
// 		tags["ip"] = ingress.IP

// 		for _, rule := range in.Spec.Rules {
// 			for _, path := range rule.IngressRuleValue.HTTP.Paths {
// 				fields["backend_service_port"] = path.Backend.Service.Port.Number
// 				fields["tls"] = in.Spec.TLS != nil

// 				tags["backend_service_name"] = path.Backend.Service.Name
// 				tags["path"] = path.Path
// 				tags["host"] = rule.Host

// 				m := &ingressM{
// 					name:   ingressMeasurement,
// 					tags:   tags,
// 					fields: fields,
// 					ts:     time.Now(),
// 				}

// 				i.collectCache = append(i.collectCache, m)
// 			}
// 		}
// 	}
// }
