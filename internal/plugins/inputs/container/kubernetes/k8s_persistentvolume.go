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
	persistentvolumeObjectMeasurement = "kubernetes_persistentvolumes"
)

//nolint:gochecknoinits
func init() {
	registerResource("persistentvolume", false, false, newPersistentvolume)
	registerMeasurements(&persistentvolumeObject{})
}

type persistentvolume struct {
	client    k8sClient
	continued string
}

func newPersistentvolume(client k8sClient) resource {
	return &persistentvolume{client: client}
}

func (c *persistentvolume) count() []pointV2 { return nil }

func (c *persistentvolume) hasNext() bool { return c.continued != "" }

func (c *persistentvolume) getMetadata(ctx context.Context, _, fieldSelector string) (metadata, error) {
	opt := metav1.ListOptions{
		Limit:         queryLimit,
		Continue:      c.continued,
		FieldSelector: fieldSelector,
	}

	list, err := c.client.GetPersistentVolumes().List(ctx, opt)
	if err != nil {
		return nil, err
	}

	c.continued = list.Continue
	return &persistentvolumeMetadata{c, list}, nil
}

type persistentvolumeMetadata struct {
	parent *persistentvolume
	list   *apicorev1.PersistentVolumeList
}

func (m *persistentvolumeMetadata) newMetric(conf *Config) pointKVs {
	return nil
}

func (m *persistentvolumeMetadata) newObject(conf *Config) pointKVs {
	var res pointKVs

	for _, item := range m.list.Items {
		obj := typed.NewPointKV(persistentvolumeObjectMeasurement)

		obj.SetTag("name", fmt.Sprintf("%v", item.UID))
		obj.SetTag("uid", fmt.Sprintf("%v", item.UID))
		obj.SetTag("persistentvolume_name", item.Name)

		obj.SetField("phase", string(item.Status.Phase))

		if item.Spec.ClaimRef != nil && item.Spec.ClaimRef.Kind == "PersistentVolumeClaim" {
			obj.SetField("claimRef_name", item.Spec.ClaimRef.Name)
			obj.SetField("claimRef_namespace", item.Spec.ClaimRef.Namespace)
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
