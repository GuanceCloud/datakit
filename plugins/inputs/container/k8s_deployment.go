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

func gatherDeploymentObject(client k8sClientX, extraTags map[string]string) (*k8sResourceStats, error) {
	list, err := client.getDeployments().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments resource: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportDeploymentObject(list.Items, extraTags), nil
}

func exportDeploymentObject(items []v1.Deployment, extraTags tagsType) *k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		obj := &deploymentObject{
			tags: map[string]string{
				"name":            fmt.Sprintf("%v", item.UID),
				"deployment_name": item.Name,
				"cluster_name":    item.ClusterName,
				"namespace":       defaultNamespace(item.Namespace),
			},
			fields: map[string]interface{}{
				"age":             int64(time.Since(item.CreationTimestamp.Time).Seconds()),
				"up_dated":        item.Status.UpdatedReplicas,
				"ready":           item.Status.ReadyReplicas,
				"available":       item.Status.AvailableReplicas,
				"unavailable":     item.Status.UnavailableReplicas,
				"strategy":        fmt.Sprintf("%v", item.Spec.Strategy.Type),
				"max_surge":       0,
				"max_unavailable": 0,
			},
			time: time.Now(),
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				obj.fields["max_unavailable"] = item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				obj.fields["max_surge"] = item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue()
			}
		}

		obj.tags.append(extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")

		res.meas = append(res.meas, obj)
	}

	return res
}

type deploymentObject struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (d *deploymentObject) LineProto() (*io.Point, error) {
	return io.NewPoint("kubernetes_deployments", d.tags, d.fields, &io.PointOption{Time: d.time, Category: datakit.Object})
}

//nolint:lll
func (*deploymentObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_deployments",
		Desc: "Kubernetes Deployment 对象数据",
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
		},
	}
}

func gatherDeploymentMetric(client k8sClientX, extraTags map[string]string) (*k8sResourceStats, error) {
	list, err := client.getDeployments().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments resource: %w", err)
	}
	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportDeploymentMetric(list.Items, extraTags), nil
}

func exportDeploymentMetric(items []v1.Deployment, extraTags tagsType) *k8sResourceStats {
	res := newK8sResourceStats()

	for _, item := range items {
		met := &deploymentMetric{
			tags: map[string]string{
				"deployment": item.Name,
				"namespace":  item.Namespace,
			},
			fields: map[string]interface{}{
				"count":                         -1,
				"paused":                        item.Spec.Paused,
				"condition":                     "",
				"replicas":                      item.Status.Replicas,
				"replicas_available":            item.Status.AvailableReplicas,
				"replicas_unavailable":          item.Status.UnavailableReplicas,
				"replicas_updated":              item.Status.UpdatedReplicas,
				"rollingupdate_max_unavailable": 0,
				"rollingupdate_max_surge":       0,
			},
			time: time.Now(),
		}
		met.fields["replicas_desired"] = ""

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.fields["rollingupdate_max_unavailable"] = item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.fields["rollingupdate_max_surge"] = item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue()
			}
		}

		met.tags.append(extraTags)
		res.meas = append(res.meas, met)
		res.namespaceList[item.Namespace]++
	}
	return res
}

type deploymentMetric struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (d *deploymentMetric) LineProto() (*io.Point, error) {
	return io.NewPoint("kube_deployment", d.tags, d.fields, &io.PointOption{Time: d.time, Category: datakit.Metric})
}

//nolint:lll
func (*deploymentMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_deployment",
		Desc: "Kubernetes Deployment 指标数据",
		Type: "object",
		Tags: map[string]interface{}{
			"deployment": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"paused":             &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "Indicates that the deployment is paused (true or false)."},
			"condition":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The current status conditions of a deployment"},
			"replicas":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment (their labels match the selector)."},
			"replicas_available": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment."},

			"replicas_unavailable":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unavailable pods targeted by this deployment."},
			"replicas_updated":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment that have the desired template spec."},
			"rollingupdate_max_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be unavailable during the update."},
			"rollingupdate_max_surge":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be scheduled above the desired number of pods. "},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&deploymentObject{})
	registerMeasurement(&deploymentMetric{})
}
