// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/diff"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apiappsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	deploymentType              = "Deployment"
	deploymentMetricMeasurement = "kube_deployment"
	deploymentObjectClass       = "kubernetes_deployments"
	deploymentObjectResourceKey = "deployment_name"
)

//nolint:gochecknoinits
func init() {
	registerResource("deployment", false, newDeployment)
	registerMeasurements(&deploymentMetric{}, &deploymentObject{})
}

type deployment struct {
	client  k8sClient
	cfg     *Config
	counter map[string]int
}

func newDeployment(client k8sClient, cfg *Config) resource {
	return &deployment{client: client, cfg: cfg, counter: make(map[string]int)}
}

func (d *deployment) gatherMetric(ctx context.Context, timestamp int64) {
	var continued string
	for {
		list, err := d.client.GetDeployments(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := d.buildMetricPoints(list, timestamp)
		feedMetric("k8s-deployment-metric", d.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
	processCounter(d.cfg, "deployment", d.counter, timestamp)
}

func (d *deployment) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := d.client.GetDeployments(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := d.buildObjectPoints(list)
		feedObject("k8s-deployment-object", d.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (d *deployment) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Apps().V1().Deployments()
	if informer == nil {
		klog.Warn("cannot get deployment informer")
		return
	}

	addFunc := func(newObj interface{}) {
		obj, ok := newObj.(*apiappsv1.Deployment)
		if !ok {
			klog.Warnf("converting to Deployment object failed, %v", newObj)
			return
		}
		if obj.CreationTimestamp.After(controllerStartTime) {
			diffs := createNoChangedFieldDiffs(changes.DeploymentCreate, obj.Namespace, deploymentType, obj.Name)
			objectChangeCountVec.WithLabelValues(deploymentType, "create").Inc()
			processChange(d.cfg, deploymentObjectClass, deploymentObjectResourceKey, diffs, obj)
		}
	}

	deleteFunc := func(oldObj interface{}) {
		obj, ok := oldObj.(*apiappsv1.Deployment)
		if !ok {
			klog.Warnf("converting to Deployment object failed, %v", oldObj)
			return
		}

		diffs := createNoChangedFieldDiffs(changes.DeploymentDelete, obj.Namespace, deploymentType, obj.Name)
		objectChangeCountVec.WithLabelValues(deploymentType, "delete").Inc()
		processChange(d.cfg, deploymentObjectClass, deploymentObjectResourceKey, diffs, obj)
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(deploymentType, "update").Inc()

		oldDeploymentObj, ok := oldObj.(*apiappsv1.Deployment)
		if !ok {
			klog.Warnf("converting to Deployment object failed, %v", oldObj)
			return
		}

		newDeploymentObj, ok := newObj.(*apiappsv1.Deployment)
		if !ok {
			klog.Warnf("converting to Deployment object failed, %v", newObj)
			return
		}

		diffs := compareDeployment(oldDeploymentObj, newDeploymentObj)
		if len(diffs) != 0 {
			objectChangeCountVec.WithLabelValues(deploymentType, "spec-changed").Inc()
			processChange(d.cfg, deploymentObjectClass, deploymentObjectResourceKey, diffs, newDeploymentObj)
		}
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			addFunc(newObj)
		},
		DeleteFunc: func(oldObj interface{}) {
			deleteFunc(oldObj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateFunc(oldObj, newObj)
		},
	})
}

func (d *deployment) buildMetricPoints(list *apiappsv1.DeploymentList, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("deployment", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("replicas", item.Status.Replicas, false)
		kvs = kvs.AddV2("replicas_updated", item.Status.UpdatedReplicas, false)
		kvs = kvs.AddV2("replicas_ready", item.Status.ReadyReplicas, false)
		kvs = kvs.AddV2("replicas_available", item.Status.AvailableReplicas, false)
		kvs = kvs.AddV2("replicas_unavailable", item.Status.UnavailableReplicas, false)

		if item.Spec.Replicas != nil {
			kvs = kvs.AddV2("replicas_desired", *item.Spec.Replicas, true)
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				kvs = kvs.AddV2("rollingupdate_max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue(), false)
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				kvs = kvs.AddV2("rollingupdate_max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue(), false)
			}
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, d.cfg.LabelAsTagsForMetric.All, d.cfg.LabelAsTagsForMetric.Keys)...)
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(deploymentMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)

		d.counter[item.Namespace]++
	}

	return pts
}

func (d *deployment) buildObjectPoints(list *apiappsv1.DeploymentList) []*point.Point {
	var pts []*point.Point
	opts := append(point.DefaultObjectOptions(), point.WithTime(ntp.Now()))

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag(deploymentObjectResourceKey, item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)
		kvs = kvs.AddV2("paused", item.Spec.Paused, false)
		kvs = kvs.AddV2("replicas", item.Status.Replicas, false)
		kvs = kvs.AddV2("replicas_updated", item.Status.UpdatedReplicas, false)
		kvs = kvs.AddV2("replicas_ready", item.Status.ReadyReplicas, false)
		kvs = kvs.AddV2("replicas_available", item.Status.AvailableReplicas, false)
		kvs = kvs.AddV2("replicas_unavailable", item.Status.UnavailableReplicas, false)
		kvs = kvs.AddV2("strategy", string(item.Spec.Strategy.Type), false)

		if item.Spec.Replicas != nil {
			kvs = kvs.AddV2("replicas_desired", *item.Spec.Replicas, false)
		}

		if item.Spec.Strategy.RollingUpdate != nil {
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				kvs = kvs.AddV2("rollingupdate_max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue(), false)
				kvs = kvs.AddV2("max_unavailable", item.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue(), false) // Deprecated
			}
			if item.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				kvs = kvs.AddV2("rollingupdate_max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue(), false)
				kvs = kvs.AddV2("max_surge", item.Spec.Strategy.RollingUpdate.MaxSurge.IntValue(), false) // Deprecated
			}
		}

		// Deprecated
		kvs = kvs.AddV2("up_dated", item.Status.UpdatedReplicas, false)
		kvs = kvs.AddV2("ready", item.Status.ReadyReplicas, false)
		kvs = kvs.AddV2("available", item.Status.AvailableReplicas, false)
		kvs = kvs.AddV2("unavailable", item.Status.UnavailableReplicas, false)

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.AddV2("yaml", string(y), false)
		}
		kvs = kvs.AddV2("annotations", pointutil.MapToJSON(item.Annotations), false)
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		// message 不包含 Selector 和 Labels
		if item.Spec.Selector != nil {
			kvs = append(kvs, point.NewTags(item.Spec.Selector.MatchLabels)...)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, d.cfg.LabelAsTagsForNonMetric.All, d.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(deploymentObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

func compareDeployment(oldVal, newVal *apiappsv1.Deployment) []FieldDiff {
	res := comparePodTemplate(&(oldVal.Spec.Template), &(newVal.Spec.Template))
	res = append(res, compareLabels(changes.DeploymentLabels, &(oldVal.ObjectMeta), &(newVal.ObjectMeta))...)
	res = append(res, compareAnnotations(changes.DeploymentAnnotations, &(oldVal.ObjectMeta), &(newVal.ObjectMeta))...)

	if oldVal.Spec.Replicas != nil && newVal.Spec.Replicas != nil &&
		*oldVal.Spec.Replicas != *newVal.Spec.Replicas {
		oldReplicas := strconv.Itoa(int(*oldVal.Spec.Replicas))
		newReplicas := strconv.Itoa(int(*newVal.Spec.Replicas))
		res = append(res, FieldDiff{
			ChangeID: changes.DeploymentReplicas,
			OldValue: oldReplicas,
			NewValue: newReplicas,
			DiffText: formatAsDiffLines("replicas", oldReplicas, newReplicas),
		})
	}

	if equal, difftext := diff.Compare(oldVal.Spec.Strategy, newVal.Spec.Strategy); !equal {
		oldRollingUpdate := []string{}
		newRollingUpdate := []string{}

		if oldVal.Spec.Strategy.RollingUpdate != nil {
			if oldVal.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				oldRollingUpdate = append(oldRollingUpdate, fmt.Sprintf("MaxUnavailable=%s", oldVal.Spec.Strategy.RollingUpdate.MaxUnavailable))
			}
			if oldVal.Spec.Strategy.RollingUpdate.MaxSurge != nil {
				oldRollingUpdate = append(oldRollingUpdate, fmt.Sprintf("MaxSurge=%s", oldVal.Spec.Strategy.RollingUpdate.MaxSurge))
			}
		}
		if newVal.Spec.Strategy.RollingUpdate != nil {
			if newVal.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				oldRollingUpdate = append(oldRollingUpdate, fmt.Sprintf("MaxUnavailable=%s", newVal.Spec.Strategy.RollingUpdate.MaxUnavailable))
			}
			if newVal.Spec.Strategy.RollingUpdate.MaxSurge != nil {
				oldRollingUpdate = append(oldRollingUpdate, fmt.Sprintf("MaxSurge=%s", newVal.Spec.Strategy.RollingUpdate.MaxSurge))
			}
		}
		res = append(res, FieldDiff{
			ChangeID: changes.DeploymentStrategy,
			OldValue: fmt.Sprintf("Type=%s,{%s}", oldVal.Spec.Strategy.Type, strings.Join(oldRollingUpdate, ",")),
			NewValue: fmt.Sprintf("Type=%s,{%s}", newVal.Spec.Strategy.Type, strings.Join(newRollingUpdate, ",")),
			DiffText: difftext,
		})
	}

	fillOwnerInfoForDiffs(res, newVal.Namespace, deploymentType, newVal.Name)
	return res
}

type deploymentMetric struct{}

//nolint:lll
func (*deploymentMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: deploymentMetricMeasurement,
		Desc: "The metric of the Kubernetes Deployment.",
		Cat:  point.Metric,
		Tags: map[string]interface{}{
			"uid":              inputs.NewTagInfo("The UID of Deployment."),
			"deployment":       inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":        inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s": inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
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

type deploymentObject struct{}

//nolint:lll
func (*deploymentObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: deploymentObjectClass,
		Desc: "The object of the Kubernetes Deployment.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                       inputs.NewTagInfo("The UID of Deployment."),
			"uid":                        inputs.NewTagInfo("The UID of Deployment."),
			"deployment_name":            inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":           inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"<all_selector_matchlabels>": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
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
