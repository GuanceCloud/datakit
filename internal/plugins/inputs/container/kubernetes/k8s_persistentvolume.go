// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	persistentvolumeObjectMeasurement       = "kubernetes_persistentvolumes"
	persistentvolumeObjectChangeMeasurement = "kubernetes_persistentvolumes"
)

//nolint:gochecknoinits
func init() {
	registerResource("persistentvolume", false, newPersistentvolume)
	registerMeasurements(&persistentvolumeObject{})
}

type persistentvolume struct {
	client k8sClient
	cfg    *Config
}

func newPersistentvolume(client k8sClient, cfg *Config) resource {
	return &persistentvolume{client: client, cfg: cfg}
}

func (p *persistentvolume) gatherMetric(ctx context.Context, timestamp int64) {
	// nil
}

func (p *persistentvolume) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := p.client.GetPersistentVolumes().List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := p.buildObjectPoints(list)
		feedObject("k8s-persistentvolume-object", p.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (p *persistentvolume) addObjectChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Core().V1().PersistentVolumes()
	if informer == nil {
		klog.Warn("cannot get persistentvolume informer")
		return
	}

	updateFunc := func(oldObj, newObj interface{}) {
		oldPersistentVolumeObj, ok := oldObj.(*apicorev1.PersistentVolume)
		if !ok {
			klog.Warnf("converting to PersistentVolume object failed, %v", oldObj)
			return
		}

		newPersistentVolumeObj, ok := newObj.(*apicorev1.PersistentVolume)
		if !ok {
			klog.Warnf("converting to PersistentVolume object failed, %v", newObj)
			return
		}

		difftext, err := diffObject(oldPersistentVolumeObj.Spec, newPersistentVolumeObj.Spec)
		if err != nil {
			klog.Warnf("marshal failed, err: %s", err)
			return
		}

		if difftext != "" {
			processObjectChange(p.cfg.Feeder, persistentvolumeObjectChangeMeasurement, difftext, newPersistentVolumeObj)
		}
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { /* skip */ },
		DeleteFunc: func(_ interface{}) { /* skip */ },
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateFunc(oldObj, newObj)
		},
	})
}

func (p *persistentvolume) buildObjectPoints(list *apicorev1.PersistentVolumeList) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("persistentvolume_name", item.Name)
		kvs = kvs.AddTag("phase", string(item.Status.Phase))

		if item.Spec.ClaimRef != nil && item.Spec.ClaimRef.Kind == "PersistentVolumeClaim" {
			kvs = kvs.AddV2("claimRef_name", item.Spec.ClaimRef.Name, false)
			kvs = kvs.AddV2("claimRef_namespace", item.Spec.ClaimRef.Namespace, false)
		}

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.AddV2("yaml", string(y), false)
		}
		kvs = kvs.AddV2("annotations", pointutil.MapToJSON(item.Annotations), false)
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, p.cfg.LabelAsTagsForNonMetric.All, p.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(p.cfg.ExtraTags)...)
		pt := point.NewPointV2(persistentvolumeObjectMeasurement, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type persistentvolumeObject struct{}

//nolint:lll
func (*persistentvolumeObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: persistentvolumeObjectMeasurement,
		Desc: "The object of the Kubernetes PersistentVolume.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":                  inputs.NewTagInfo("The UID of PersistentVolume."),
			"uid":                   inputs.NewTagInfo("The UID of PersistentVolume."),
			"persistentvolume_name": inputs.NewTagInfo("The name of PersistentVolume"),
			"cluster_name_k8s":      inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"phase":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The phase indicates if a volume is available, bound to a claim, or released by a claim.(Pending/Available/Bound/Released/Failed)"},
			"claimRef_name":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Name of the bound PersistentVolumeClaim."},
			"claimRef_namespace": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Namespace of the PersistentVolumeClaim."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}
