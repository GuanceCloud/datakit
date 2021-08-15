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
	tags map[string]string
}

func (s *service) Gather() {
	var start = time.Now()
	var pts []*io.Point

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
		for k, v := range s.tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"age":                     int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"cluster_ip":              obj.Spec.ClusterIP,
			"external_name":           obj.Spec.ExternalName,
			"external_traffic_policy": fmt.Sprintf("%v", obj.Spec.ExternalTrafficPolicy),
			"session_affinity":        fmt.Sprintf("%v", obj.Spec.SessionAffinity),
		}

		// addSliceToFields("selectors", obj.Spec.Selector, fields)
		// addSliceToFields("load_balancer_ingress", obj.Status.LoadBalancer, fields)
		addSliceToFields("external_ips", obj.Spec.ExternalIPs, fields)

		addMapToFields("annotations", obj.Annotations, fields)
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesServiceName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			pts = append(pts, pt)
		}
	}

	if err := io.Feed(inputName, datakit.Object, pts, &io.Option{CollectCost: time.Since(start)}); err != nil {
		l.Error(err)
	}
}

func (*service) Resource() { /*empty interface*/ }

func (*service) LineProto() (*io.Point, error) { return nil, nil }

func (*service) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesServiceName,
		Desc: "Kubernetes service 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("service UID"),
			"service_name": inputs.NewTagInfo("service 名称"),
			"cluster_name": inputs.NewTagInfo("所在 cluster"),
			"namespace":    inputs.NewTagInfo("所在命名空间"),
			"type":         inputs.NewTagInfo("服务类型，ClusterIP/NodePort/LoadBalancer/ExternalName"),
		},
		Fields: map[string]interface{}{
			"age":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"cluster_ip":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "cluster IP"},
			"external_ips":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "external IP 列表，内容是以英文逗号拼接的字符串"},
			"external_name":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "external 名称"},
			"external_traffic_policy": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "external 负载均衡"},
			"session_affinity":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "session关联性"},
			"annotations":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
			// TODO:
			// "load_balancer_ingress":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "selectors":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "ip_family":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "ports":                   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
