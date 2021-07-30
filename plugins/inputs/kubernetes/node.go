package kubernetes

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesNodeName = "kubernetes_nodes"

type node struct {
	client interface {
		getNodes() (*corev1.NodeList, error)
	}
	tags map[string]string
}

func (n node) Gather() {
	list, err := n.client.getNodes()
	if err != nil {
		l.Errorf("failed of get nodes resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":         fmt.Sprintf("%v", obj.UID),
			"node_name":    obj.Name,
			"cluster_name": obj.ClusterName,
			"namespace":    obj.Namespace,
			"status":       fmt.Sprintf("%v", obj.Status.Phase),
		}
		for k, v := range n.tags {
			tags[k] = v
		}
		fields := map[string]interface{}{
			"age":             int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"kubelet_version": obj.Status.NodeInfo.KubeletVersion,
		}

		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesNodeName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*node) Resource() { /*empty interface*/ }

func (*node) LineProto() (*io.Point, error) { return nil, nil }

func (*node) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesNodeName,
		Desc: kubernetesNodeName,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("node UID"),
			"node_name":    inputs.NewTagInfo("node 名称"),
			"cluster_name": inputs.NewTagInfo("所在 cluster"),
			"namespace":    inputs.NewTagInfo("所在命名空间"),
			"status":       inputs.NewTagInfo("当期状态，Pending/Running/Terminated"),
		},
		Fields: map[string]interface{}{
			"age":                    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"kubelet_version":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubelet 版本"},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "k8s annotations"},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
			// "schedulability":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "role":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "taints":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pods":                   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pod_capacity":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pod_usage":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
