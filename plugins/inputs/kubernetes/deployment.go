package kubernetes

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const kubernetesDeploymentName = "kubernetes_deployments"

type deployment struct {
	client interface {
		getDeployments() (*appsv1.DeploymentList, error)
	}
}

func (d deployment) Gather() {
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
		}
		if obj.Spec.Strategy.RollingUpdate != nil && obj.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
			fields["max_unavailable"] = obj.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue()
		}

		addJSONStringToMap("kubernetes_labels", obj.Labels, fields)
		addJSONStringToMap("kubernetes_annotations", obj.Annotations, fields)
		addMessageToFields(tags, fields)

		pt, err := io.MakePoint(kubernetesDeploymentName, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else {
			if err := io.Feed(inputName, datakit.Object, []*io.Point{pt}, nil); err != nil {
				l.Error(err)
			}
		}
	}
}

func (*deployment) LineProto() (*io.Point, error) {
	return nil, nil
}

func (*deployment) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesDeploymentName,
		Desc: kubernetesDeploymentName,
		Tags: map[string]interface{}{
			"name":            inputs.NewTagInfo(""),
			"deployment_name": inputs.NewTagInfo(""),
			"cluster_name":    inputs.NewTagInfo(""),
			"namespace":       inputs.NewTagInfo(""),
		},
		Fields: map[string]interface{}{
			"age":             &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"max_surge":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			"max_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: ""},
			"up_dated":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "updated replicas"},
			"ready":           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"available":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			"unavailable":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: ""},
			//"condition":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"strategy":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"selectors": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "paused":                 &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			// "current/desired":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_labels":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"kubernetes_annotations": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
			"message":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: ""},
		},
	}
}
