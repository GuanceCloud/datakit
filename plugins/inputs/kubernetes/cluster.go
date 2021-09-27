package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	rbacv1 "k8s.io/api/rbac/v1"
)

const kubernetesClusterName = "kubernetes_clusters"

type cluster struct {
	client interface {
		getClusters() (*rbacv1.ClusterRoleList, error)
	}
	tags map[string]string
}

func (c *cluster) Gather() {
	start := time.Now()
	var pts []*io.Point

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
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesClusterName, tags, fields, time.Now())
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

func (*cluster) LineProto() (*io.Point, error) { return nil, nil }

func (*cluster) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesClusterName,
		Desc: "Kubernetes cluster 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"cluster_name": inputs.NewTagInfo("Name must be unique within a namespace."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(second)"},
			"annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
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
