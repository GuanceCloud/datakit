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

var (
	_ k8sResourceMetricInterface = (*deployment)(nil)
	_ k8sResourceObjectInterface = (*deployment)(nil)
)

type deployment struct {
	client    k8sClientX
	extraTags map[string]string
	items     []v1.Deployment
}

func newDeployment(client k8sClientX, extraTags map[string]string) *deployment {
	return &deployment{
		client:    client,
		extraTags: extraTags,
	}
}

func (d *deployment) name() string {
	return "deployment"
}

func (d *deployment) pullItems() error {
	if len(d.items) != 0 {
		return nil
	}

	list, err := d.client.getDeployments().List(context.Background(), metaV1ListOption)
	if err != nil {
		return fmt.Errorf("failed to get deployments resource: %w", err)
	}

	d.items = list.Items
	return nil
}

func (d *deployment) metric() (inputsMeas, error) {
	if err := d.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range d.items {
		met := &deploymentMetric{
			tags: map[string]string{
				"deployment": item.Name,
				"namespace":  item.Namespace,
			},
			fields: map[string]interface{}{
				"paused":                        item.Spec.Paused,
				"condition":                     len(item.Status.Conditions),
				"replicas":                      item.Status.Replicas,
				"replicas_available":            item.Status.AvailableReplicas,
				"replicas_unavailable":          item.Status.UnavailableReplicas,
				"replicas_updated":              item.Status.UpdatedReplicas,
				"rollingupdate_max_unavailable": 0,
				"rollingupdate_max_surge":       0,
				// TODO:"replicas_desired"
			},
			time: time.Now(),
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.fields["rollingupdate_max_unavailable"] = item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.fields["rollingupdate_max_surge"] = item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue()
			}
		}

		met.tags.append(d.extraTags)
		res = append(res, met)
	}

	return res, nil
}

func (d *deployment) object() (inputsMeas, error) {
	if err := d.pullItems(); err != nil {
		return nil, err
	}
	var res inputsMeas

	for _, item := range d.items {
		obj := &deploymentObject{
			tags: map[string]string{
				"name":            fmt.Sprintf("%v", item.UID),
				"deployment_name": item.Name,
				"cluster_name":    defaultClusterName(item.ClusterName),
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

		obj.tags.append(d.extraTags)

		obj.fields.addMapWithJSON("annotations", item.Annotations)
		obj.fields.addLabel(item.Labels)
		obj.fields.mergeToMessage(obj.tags)
		obj.fields.delete("annotations")

		res = append(res, obj)
	}

	return res, nil
}

func (d *deployment) count() (map[string]int, error) {
	if err := d.pullItems(); err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, item := range d.items {
		m[defaultNamespace(item.Namespace)]++
	}

	return m, nil
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
		Type: "metric",
		Tags: map[string]interface{}{
			"deployment": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"count":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of deployments"},
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
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "object details"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerK8sResourceMetric(func(c k8sClientX, m map[string]string) k8sResourceMetricInterface { return newDeployment(c, m) })
	registerK8sResourceObject(func(c k8sClientX, m map[string]string) k8sResourceObjectInterface { return newDeployment(c, m) })
	registerMeasurement(&deploymentObject{})
	registerMeasurement(&deploymentMetric{})
}
