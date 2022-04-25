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

const k8sClusterRoleName = "kubernetes_cluster_roles"

func gatherCluster(client k8sClientX, extraTags map[string]string) (*k8sResourceStats, error) {
	list, err := client.getClusters().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get clusters resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportClusterRole(list.Items, extraTags), nil
}

func exportClusterRole(items []v1.ClusterRole, extraTags tagsType) *k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newClusterRole()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["cluster_name"] = defaultClusterName(item.ClusterName)

		obj.tags.addValueIfNotEmpty("cluster_role_name", item.Name)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["create_time"] = item.CreationTimestamp.Time.Unix()

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		delete(obj.fields, "annotations")

		obj.time = time.Now()
		res.meas = append(res.meas, obj)
	}
	return res
}

type clusterRole struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newClusterRole() *clusterRole {
	return &clusterRole{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (c *clusterRole) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sClusterRoleName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Object})
}

//nolint:lll
func (*clusterRole) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sClusterRoleName,
		Desc: "Kubernetes cluster role 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("UID"),
			"cluster_role_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name":      inputs.NewTagInfo("The name of the cluster which the object belongs to."),
		},
		Fields: map[string]interface{}{
			"age":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"create_time": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampSec, Desc: "CreationTimestamp is a timestamp representing the server time when this object was created.(second)"},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&clusterRole{})
}
