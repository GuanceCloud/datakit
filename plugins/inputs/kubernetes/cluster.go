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
	tags map[string]string
}

func (c *cluster) Gather() {
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
		for k, v := range c.tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"create_time": obj.CreationTimestamp.Time.Unix(),
		}

		addMapToFields("annotations", obj.Annotations, fields)
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

func (*cluster) LineProto() (*io.Point, error) { return nil, nil }

func (*cluster) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesClusterName,
		Desc: fmt.Sprintf("%s 对象数据", kubernetesClusterName),
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("cluster UID"),
			"cluster_name": inputs.NewTagInfo("cluster 名称"),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "存活时长，单位为秒"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "创建时间戳，精度为秒"},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "详情数据"},
			// TODO:
			// "pod_capacity":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			// "pod_usage":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			// "namespaces":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "nodes":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "pods":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "versions":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "kubelet_version": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
