package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
)

const k8sDeploymentName = "kubernetes_deployments"

func gatherDeployment(client k8sClientX, extraTags map[string]string) (k8sResourceStats, error) {
	list, err := client.getDeployments().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportDeployment(list.Items, extraTags), nil
}

func exportDeployment(items []v1.Deployment, extraTags tagsType) k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := newDeployment()
		obj.tags["name"] = fmt.Sprintf("%v", item.UID)
		obj.tags["deployment_name"] = item.Name

		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		obj.fields["age"] = int64(time.Since(item.CreationTimestamp.Time).Seconds())
		obj.fields["up_dated"] = item.Status.UpdatedReplicas
		obj.fields["ready"] = item.Status.ReadyReplicas
		obj.fields["available"] = item.Status.AvailableReplicas
		obj.fields["unavailable"] = item.Status.UnavailableReplicas
		obj.fields["strategy"] = fmt.Sprintf("%v", item.Spec.Strategy.Type)

		if item.Spec.Strategy.RollingUpdate != nil && item.Spec.Strategy.RollingUpdate.MaxSurge != nil {
			obj.fields["max_surge"] = item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue()
		} else {
			obj.fields["max_surge"] = 0
		}

		if item.Spec.Strategy.RollingUpdate != nil && item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
			obj.fields["max_unavailable"] = item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
		} else {
			obj.fields["max_unavailable"] = 0
		}

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)

		obj.time = time.Now()
		res.set(defaultNamespace(item.Namespace), obj)
	}
	return res
}

type deployment struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newDeployment() *deployment {
	return &deployment{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (d *deployment) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sDeploymentName, d.tags, d.fields, &io.PointOption{Time: d.time, Category: datakit.Object})
}

//nolint:lll
func (*deployment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sDeploymentName,
		Desc: "Kubernetes deployment 对象数据",
		Type: "object",
		Tags: map[string]interface{}{
			"name":            inputs.NewTagInfo("UID"),
			"deployment_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"cluster_name":    inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":       inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "age (seconds)"},
			"ready":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Total number of ready pods targeted by this deployment."},
			"available":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment."},
			"unavailable":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Total number of unavailable pods targeted by this deployment."},
			"max_surge":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be scheduled above the desired number of pods"},
			"max_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be unavailable during the update."},
			"up_dated":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Total number of non-terminated pods targeted by this deployment that have the desired template spec."},
			"strategy":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: `Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.`},
			"annotations":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "kubernetes annotations"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
			// TODO:
			// "selectors":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "condition":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "paused":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&deployment{})
}
