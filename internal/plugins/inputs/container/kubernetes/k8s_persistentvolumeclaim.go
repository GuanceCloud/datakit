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
)

const (
	persistentvolumeclaimObjectClass = "kubernetes_persistentvolumeclaims"
)

//nolint:gochecknoinits
func init() {
	registerResource("persistentvolumeclaim", false, newPersistentvolumeclaim)
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

func (*persistentvolumeclaim) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

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

		kvs = kvs.Add("volume_name", item.Spec.VolumeName)
		if item.Spec.VolumeMode != nil {
			kvs = kvs.Add("volume_mode", string(*item.Spec.VolumeMode))
		}
		if item.Spec.StorageClassName != nil {
			kvs = kvs.Add("storage_class_name", *item.Spec.StorageClassName)
		}

		kvs = kvs.Add("age", time.Since(item.CreationTimestamp.Time).Milliseconds()/1e3)

		if item.Spec.Resources.Requests != nil {
			storage, ok := item.Spec.Resources.Requests["storage"]
			if ok {
				kvs = kvs.Add("requests_storage", storage.String())
			}
		}

		accessModes := []string{}
		for _, mode := range item.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}
		sort.Strings(accessModes)
		kvs = kvs.Add("access_modes", strings.Join(accessModes, ","))

		if y, err := yaml.Marshal(item); err == nil {
			kvs = kvs.Add("yaml", string(y))
		}
		kvs = kvs.Add("annotations", pointutil.MapToJSON(item.Annotations))
		kvs = append(kvs, pointutil.ConvertDFLabels(item.Labels)...)

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.Add("message", pointutil.TrimString(msg, maxMessageLength))

		kvs = kvs.Del("annotations")
		kvs = kvs.Del("yaml")

		if item.Spec.Selector != nil {
			kvs = append(kvs, point.NewTags(item.Spec.Selector.MatchLabels)...)
		}

		kvs = append(kvs, pointutil.ExtractSourceCodeFromAnnotations(item.Annotations)...) // add source_code
		kvs = append(kvs, pointutil.LabelsToPointKVs(item.Labels, p.cfg.LabelAsTagsForNonMetric.All, p.cfg.LabelAsTagsForNonMetric.Keys)...)
		kvs = append(kvs, point.NewTags(p.cfg.ExtraTags)...)
		pt := point.NewPoint(persistentvolumeclaimObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type PersistentvolumeclaimObject struct{}

//nolint:lll
func (*PersistentvolumeclaimObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: persistentvolumeclaimObjectClass,
		Desc: "The object of the Kubernetes PersistentVolumeClaim.",
		Cat:  point.Object,
		Tags: map[string]interface{}{
			"name":                              inputs.NewTagInfo("The UID of PersistentVolume."),
			"uid":                               inputs.NewTagInfo("The UID of PersistentVolume."),
			"persistentvolumeclaim_name":        inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                         inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":                  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
			"&lt;ALL-SELECTOR-MATCH-LABELS&gt;": inputs.NewTagInfo("Represents the selector.matchLabels for Kubernetes resources"),
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
