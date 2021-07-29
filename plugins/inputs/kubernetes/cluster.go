package kubernetes

import (
	"fmt"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesClusterName = "kubernetes_clusters"

type cluster struct {
	client interface {
		getClusters() (*rbacv1.ClusterRoleList, error)
	}
}

func (c cluster) Gather() {
	list, err := c.client.getClusters()
	if err != nil {
		l.Errorf("failed of get clusters resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":         fmt.Sprintf("%v", obj.UID),
			"cluster_name": obj.Name,
		}
		fields := map[string]interface{}{
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"create_time": obj.CreationTimestamp.Time,
		}

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesClusterName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*cluster) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*cluster) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesClusterName,
		Desc: kubernetesClusterName,
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo(""),
			"cluster_name": inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"pod_capacity":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			"pod_usage":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			"namespaces":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"nodes":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"pods":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"versions":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
