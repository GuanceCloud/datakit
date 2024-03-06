// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"sigs.k8s.io/yaml"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	persistentvolumeclaimObjectMeasurement = "kubernetes_persistentvolumeclaims"
)

//nolint:gochecknoinits
func init() {
	registerResource("persistentvolumeclaim", false, false, newPersistentvolumeclaim)
	registerMeasurements(&persistentvolumeclaimObject{})
}

type persistentvolumeclaim struct {
	client    k8sClient
	continued string
}

func newPersistentvolumeclaim(client k8sClient) resource {
	return &persistentvolumeclaim{client: client}
}

func (c *persistentvolumeclaim) count() []pointV2 { return nil }

func (c *persistentvolumeclaim) hasNext() bool { return c.continued != "" }

func (c *persistentvolumeclaim) getMetadata(ctx context.Context, ns, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      c.continued,
		FieldSelector: fieldSelector,
	}

	list, err := c.client.GetPersistentVolumeClaims(ns).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	c.continued = list.Continue
	return &persistentvolumeclaimMetadata{c, list}, nil
}

type persistentvolumeclaimMetadata struct {
	parent *persistentvolumeclaim
	list   *apicorev1.PersistentVolumeClaimList
}

func (m *persistentvolumeclaimMetadata) newMetric(conf *Config) pointKVs {
	return nil
}

func (m *persistentvolumeclaimMetadata) newObject(conf *Config) pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(persistentvolumeclaimObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("persistentvolumeclaim_name", item.Name)
		obj.SetTag("namespace", item.Namespace)

		obj.SetField("phase", string(item.Status.Phase))
		obj.SetField("volume_name", item.Spec.VolumeName)
		if item.Spec.VolumeMode != nil {
			obj.SetField("volume_mode", string(*item.Spec.VolumeMode))
		}
		if item.Spec.StorageClassName != nil {
			obj.SetField("storage_class_name", *item.Spec.StorageClassName)
		}

		if y, err := yaml.Marshal(item); err == nil {
			obj.SetField("yaml", string(y))
		}

		obj.SetFields(transLabels(item.Labels))
		obj.SetField("annotations", typed.MapToJSON(item.Annotations))
		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))
		obj.DeleteField("annotations")
		obj.DeleteField("yaml")

		obj.SetLabelAsTags(item.Labels, conf.LabelAsTagsForNonMetric.All, conf.LabelAsTagsForNonMetric.Keys)
		res = append(res, obj)
	}

	return res
}

type persistentvolumeclaimObject struct{}

//nolint:lll
func (*persistentvolumeclaimObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: persistentvolumeclaimObjectMeasurement,
		Desc: "The object of the Kubernetes PersistentVolumeClaim.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":                       inputs.NewTagInfo("The UID of PersistentVolume."),
			"uid":                        inputs.NewTagInfo("The UID of PersistentVolume."),
			"persistentvolumeclaim_name": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace":                  inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
			"cluster_name_k8s":           inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"phase":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The phase indicates if a volume is available, bound to a claim, or released by a claim.(Pending/Bound/Lost)"},
			"volume_name":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "VolumeName is the binding reference to the PersistentVolume backing this claim."},
			"volume_mode":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "VolumeMode defines what type of volume is required by the claim.(Block/Filesystem)"},
			"storage_class_name": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "StorageClassName is the name of the StorageClass required by the claim."},
			"message":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}
