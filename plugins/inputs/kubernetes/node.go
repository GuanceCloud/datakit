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
		fields := map[string]interface{}{
			"age":             int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"kubelet_version": obj.Status.NodeInfo.KubeletVersion,
		}

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
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

func (*node) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*node) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesNodeName,
		Desc: kubernetesNodeName,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo(""),
			"node_name":    inputs.NewTagInfo(""),
			"cluster_name": inputs.NewTagInfo(""),
			"status":       inputs.NewTagInfo(""),
			"namespace":    inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			// "schedulability":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "role":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "taints":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pods":                   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pod_capacity":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pod_usage":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
