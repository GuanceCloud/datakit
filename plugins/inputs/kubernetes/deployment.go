package kubernetes

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	appsv1 "k8s.io/api/apps/v1"
)

const kubernetesDeploymentName = "kubernetes_deployments"

type deployment struct {
	client interface {
		getDeployments() (*appsv1.DeploymentList, error)
	}
	tags map[string]string
}

func (d *deployment) Gather() {
	start := time.Now()
	var pts []*io.Point

	list, err := d.client.getDeployments()
	if err != nil {
		l.Errorf("failed of get deployments resource: %s", err)
		return
	}

	for _, obj := range list.Items {
		tags := map[string]string{
			"name":            fmt.Sprintf("%v", obj.UID),
			"deployment_name": obj.Name,
			"cluster_name":    obj.ClusterName,
			"namespace":       obj.Namespace,
		}
		for k, v := range d.tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"age":         int64(time.Now().Sub(obj.CreationTimestamp.Time).Seconds()),
			"up_dated":    obj.Status.UpdatedReplicas,
			"ready":       obj.Status.ReadyReplicas,
			"available":   obj.Status.AvailableReplicas,
			"unavailable": obj.Status.UnavailableReplicas,
			"strategy":    fmt.Sprintf("%v", obj.Spec.Strategy.Type),
		}

		if obj.Spec.Strategy.RollingUpdate != nil && obj.Spec.Strategy.RollingUpdate.MaxSurge != nil {
			fields["max_surge"] = obj.Spec.Strategy.RollingUpdate.MaxSurge.IntValue()
		} else {
			fields["max_surge"] = defaultIntegerValue
		}

		if obj.Spec.Strategy.RollingUpdate != nil && obj.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
			fields["max_unavailable"] = obj.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
		} else {
			fields["max_unavailable"] = defaultIntegerValue
		}

		addMapToFields("annotations", obj.Annotations, fields)
		addLabelToFields(obj.Labels, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesDeploymentName, tags, fields, time.Now())
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

func (*deployment) Resource() { /*empty interface*/ }

func (*deployment) LineProto() (*io.Point, error) { return nil, nil }

func (*deployment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesDeploymentName,
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
