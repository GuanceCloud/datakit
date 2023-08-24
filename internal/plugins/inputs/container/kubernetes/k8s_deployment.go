// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apiappsv1 "k8s.io/api/apps/v1"
)

//nolint:gochecknoinits
func init() {
	registerMetricResource("deployment", gatherDeploymentMetric)
	registerObjectResource("deployment", gatherDeploymentObject)
	registerMeasurement(&deploymentMetric{})
	registerMeasurement(&deploymentObject{})
}

func gatherDeploymentMetric(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetDeployments().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeDeploymentMetric(list), nil
}

func composeDeploymentMetric(list *apiappsv1.DeploymentList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		met := typed.NewPointKV()

		met.SetTag("uid", fmt.Sprintf("%v", item.UID))
		met.SetTag("deployment", item.Name)
		met.SetTag("namespace", item.Namespace)

		met.SetField("replicas", item.Status.Replicas)
		met.SetField("replicas_updated", item.Status.UpdatedReplicas)
		met.SetField("replicas_ready", item.Status.ReadyReplicas)
		met.SetField("replicas_available", item.Status.AvailableReplicas)
		met.SetField("replicas_unavailable", item.Status.UnavailableReplicas)
		met.SetField("rollingupdate_max_unavailable", 0)
		met.SetField("rollingupdate_max_surge", 0)

		if item.Spec.Replicas != nil {
			met.SetField("replicas_desired", *item.Spec.Replicas)
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.SetField("rollingupdate_max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue())
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				met.SetField("rollingupdate_max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue())
			}
		}

		met.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, &deploymentMetric{met})
	}

	return res
}

func gatherDeploymentObject(ctx context.Context, client k8sClient) ([]measurement, error) {
	list, err := client.GetDeployments().List(ctx, metaV1ListOption)
	if err != nil {
		return nil, err
	}
	return composeDeploymentObject(list), nil
}

func composeDeploymentObject(list *apiappsv1.DeploymentList) []measurement {
	var res []measurement

	for _, item := range list.Items {
		obj := typed.NewPointKV()

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("deployment_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)
		obj.SetField("paused", item.Spec.Paused)
		obj.SetField("replicas", item.Status.Replicas)
		obj.SetField("replicas_updated", item.Status.UpdatedReplicas)
		obj.SetField("replicas_ready", item.Status.ReadyReplicas)
		obj.SetField("replicas_available", item.Status.AvailableReplicas)
		obj.SetField("replicas_unavailable", item.Status.UnavailableReplicas)
		obj.SetField("strategy", fmt.Sprintf("%v", item.Spec.Strategy.Type))
		obj.SetField("rollingupdate_max_unavailable", 0)
		obj.SetField("rollingupdate_max_surge", 0)

		if item.Spec.Replicas != nil {
			obj.SetField("replicas_desired", *item.Spec.Replicas)
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				obj.SetField("rollingupdate_max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue())
				// Deprecated
				obj.SetField("max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue())
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				obj.SetField("rollingupdate_max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue())
				// Deprecated
				obj.SetField("max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue())
			}
		}

		// Deprecated
		obj.SetField("up_dated", item.Status.UpdatedReplicas)
		obj.SetField("ready", item.Status.ReadyReplicas)
		obj.SetField("available", item.Status.AvailableReplicas)
		obj.SetField("unavailable", item.Status.UnavailableReplicas)

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		obj.SetCustomerTags(item.Labels, getGlobalCustomerKeys())
		res = append(res, &deploymentObject{obj})
	}

	return res
}

type deploymentMetric struct{ typed.PointKV }

func (d *deploymentMetric) namespace() string { return d.GetTag("namespace") }

func (d *deploymentMetric) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *deploymentMetric) LineProto() (*point.Point, error) {
	return point.NewPoint("kube_deployment", d.Tags(), d.Fields(), metricOpt)
}

//nolint:lll
func (*deploymentMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_deployment",
		Desc: "The metric of the Kubernetes Deployment.",
		Type: "metric",
		Tags: map[string]interface{}{
			"uid":        inputs.NewTagInfo("The UID of Deployment."),
			"deployment": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"replicas":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment (their labels match the selector)."},
			"replicas_desired":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of desired pods for a Deployment."},
			"replicas_ready":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods targeted by this Deployment with a Ready Condition."},
			"replicas_available":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment."},
			"replicas_unavailable":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unavailable pods targeted by this deployment."},
			"replicas_updated":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment that have the desired template spec."},
			"rollingupdate_max_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be unavailable during the update."},
			"rollingupdate_max_surge":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be scheduled above the desired number of pods. "},
		},
	}
}

type deploymentObject struct{ typed.PointKV }

func (d *deploymentObject) namespace() string { return d.GetTag("namespace") }

func (d *deploymentObject) addExtraTags(m map[string]string) { d.SetTags(m) }

func (d *deploymentObject) LineProto() (*point.Point, error) {
	return point.NewPoint("kubernetes_deployments", d.Tags(), d.Fields(), objectOpt)
}

//nolint:lll
func (*deploymentObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kubernetes_deployments",
		Desc: "The object of the Kubernetes Deployment.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":            inputs.NewTagInfo("The UID of Deployment."),
			"uid":             inputs.NewTagInfo("The UID of Deployment."),
			"deployment_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":       inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"age":                           &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"paused":                        &inputs.FieldInfo{DataType: inputs.Bool, Unit: inputs.UnknownUnit, Desc: "Indicates that the deployment is paused (true or false)."},
			"replicas":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment (their labels match the selector)."},
			"replicas_desired":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of desired pods for a Deployment."},
			"replicas_ready":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods targeted by this Deployment with a Ready Condition."},
			"replicas_available":            &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment."},
			"replicas_unavailable":          &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unavailable pods targeted by this deployment."},
			"replicas_updated":              &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment that have the desired template spec."},
			"rollingupdate_max_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be unavailable during the update."},
			"rollingupdate_max_surge":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be scheduled above the desired number of pods. "},
			"strategy":                      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: `Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.`},
			"message":                       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
			"ready":                         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of pods targeted by this Deployment with a Ready Condition. (Deprecated)"},
			"available":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment. (Deprecated)"},
			"unavailable":                   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of unavailable pods targeted by this deployment. (Deprecated)"},
			"up_dated":                      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of non-terminated pods targeted by this deployment that have the desired template spec. (Deprecated)"},
			"max_unavailable":               &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be unavailable during the update. (Deprecated)"},
			"max_surge":                     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The maximum number of pods that can be scheduled above the desired number of pods. (Deprecated)"},
		},
	}
}
