package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/rbac/v1"
)

const k8sClusterName = "kubernetes_clusters"

func gatherCluster(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getClusters().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportCluster(list.Items, extraTags), nil
}

func exportCluster(items []v1.ClusterRole, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newCluster()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)

		obj.tags.addValueIfNotEmpty("cluster_name", item.Name)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["create_time"] = item.CreationTimestamp.Time.Unix()

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type cluster struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newCluster() *cluster {
	return &cluster{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (c *cluster) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sClusterName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

//nolint:lll
func (*cluster) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sClusterName,
		Desc: "Kubernetes cluster 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":         inputs.NewTagInfo("UID"),
			"cluster_name": inputs.NewTagInfo("Name must be unique within a namespace."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampSec, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(second)"},
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

//nolint:gochecknoinits
func init() {
	registerMeasurement(&cluster{})
}
