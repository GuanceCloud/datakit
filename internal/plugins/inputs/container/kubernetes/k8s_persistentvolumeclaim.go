// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	persistentvolumeclaimObjectMeasurement = "kubernetes_persistentvolumeclaims"
	persistentvolumeclaimChangeSource      = "kubernetes_persistentvolumeclaims"
	persistentvolumeclaimChangeSourceType  = "PersistentVolumeClaim"
)

//nolint:gochecknoinits
func init() {
	registerResource("persistentvolumeclaim", false, newPersistentvolumeclaim)
	registerMeasurements(&persistentvolumeclaimObject{})
}

type persistentvolumeclaim struct {
	client k8sClient
	cfg    *Config
}

func newPersistentvolumeclaim(client k8sClient, cfg *Config) resource {
	return &persistentvolumeclaim{client: client, cfg: cfg}
}

func (p *persistentvolumeclaim) gatherMetric(ctx context.Context, timestamp int64) { /* nil */ }

func (p *persistentvolumeclaim) gatherObject(ctx context.Context) {
	var continued string
	for {
		list, err := p.client.GetPersistentVolumeClaims(allNamespaces).List(ctx, newListOptions(emptyFieldSelector, continued))
		if err != nil {
			klog.Warn(err)
			break
		}
		continued = list.Continue

		pts := p.buildObjectPoints(list)
		feedObject("k8s-persistentvolumeclaim-object", p.cfg.Feeder, pts, true)

		if continued == "" {
			break
		}
	}
}

func (p *persistentvolumeclaim) addChangeInformer(informerFactory informers.SharedInformerFactory) {
	informer := informerFactory.Core().V1().PersistentVolumeClaims()
	if informer == nil {
		klog.Warn("cannot get persistentvolumeclaim informer")
		return
	}

	updateFunc := func(oldObj, newObj interface{}) {
		objectChangeCountVec.WithLabelValues(persistentvolumeclaimChangeSourceType, "update").Inc()

		oldPersistentVolumeClaimObj, ok := oldObj.(*apicorev1.PersistentVolumeClaim)
		if !ok {
			klog.Warnf("converting to PersistentVolumeClaim object failed, %v", oldObj)
			return
		}

		newPersistentVolumeClaimObj, ok := newObj.(*apicorev1.PersistentVolumeClaim)
		if !ok {
			klog.Warnf("converting to PersistentVolumeClaim object failed, %v", newObj)
			return
		}

		difftext, err := diffObject(oldPersistentVolumeClaimObj.Spec, newPersistentVolumeClaimObj.Spec)
		if err != nil {
			klog.Warnf("marshal failed, err: %s", err)
			return
		}

		if difftext != "" {
			objectChangeCountVec.WithLabelValues(persistentvolumeclaimChangeSourceType, "spec-changed").Inc()
			processChange(p.cfg.Feeder,
				persistentvolumeclaimChangeSource,
				persistentvolumeclaimChangeSourceType,
				difftext, newPersistentVolumeClaimObj)
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

func (p *persistentvolumeclaim) buildObjectPoints(list *apicorev1.PersistentVolumeClaimList) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for _, item := range list.Items {
		var kvs point.KVs

		kvs = kvs.AddTag("name", string(item.UID))
		kvs = kvs.AddTag("uid", string(item.UID))
		kvs = kvs.AddTag("persistentvolumeclaim_name", item.Name)
		kvs = kvs.AddTag("namespace", item.Namespace)
		kvs = kvs.AddTag("phase", string(item.Status.Phase))

		kvs = kvs.AddV2("volume_name", item.Spec.VolumeName, false)
		if item.Spec.VolumeMode != nil {
			kvs = kvs.AddV2("volume_mode", string(*item.Spec.VolumeMode), false)
		}
		if item.Spec.StorageClassName != nil {
			kvs = kvs.AddV2("storage_class_name", *item.Spec.StorageClassName, false)
		}

		kvs = kvs.AddV2("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3, false)

		if item.Spec.Resources.Requests != nil {
			storage, ok := item.Spec.Resources.Requests["storage"]
			if ok {
				kvs = kvs.AddV2("requests_storage", storage.String(), false)
			}
		}

		accessModes := []string{}
		for _, mode := range item.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}
		sort.Strings(accessModes)
		kvs = kvs.AddV2("access_modes", strings.Join(accessModes, ","), false)

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.AddV2("yaml", string(y), false)
		}
		kvs = kvs.AddV2("annotations", pointutil.MapToJSON(item.Annotations), false)
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		if item.Spec.Selector != nil {
			kvs = append(kvs, point.NewTags(item.Spec.Selector.MatchLabels)...)
		}

		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, p.cfg.LabelAsTagsForNonMetric.All, p.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(p.cfg.ExtraTags)...)
		pt := point.NewPointV2(persistentvolumeclaimObjectMeasurement, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type persistentvolumeclaimObject struct{}

//nolint:lll
func (*persistentvolumeclaimObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: persistentvolumeclaimObjectMeasurement,
		Desc: "The object of the Kubernetes PersistentVolumeClaim.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                       inputs.NewTagInfo("The UID of PersistentVolume."),
			"uid":                        inputs.NewTagInfo("The UID of PersistentVolume."),
			"persistentvolumeclaim_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":           inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"<all_selector_matchlabels>": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
		},
		Fields: map[string]interface{}{
			"age":                &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "Age (seconds)"},
			"phase":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The phase indicates if a volume is available, bound to a claim, or released by a claim.(Pending/Bound/Lost)"},
			"volume_name":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "VolumeName is the binding reference to the PersistentVolume backing this claim."},
			"volume_mode":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "VolumeMode defines what type of volume is required by the claim.(Block/Filesystem)"},
			"storage_class_name": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "StorageClassName is the name of the StorageClass required by the claim."},
			"access_modes":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "AccessModes contains the desired access modes the volume should have."},
			"requests_storage":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Specifies the maximum storage capacity of a PersistentVolume (PV), which Kubernetes uses for scheduling and resource allocation."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}
