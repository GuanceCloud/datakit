// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dfpvMetricMeasurement = "kube_dfpv"
	dfpvObjectMeasurement = "kubernetes_dfpv"
)

//nolint:gochecknoinits
func init() {
	registerResource("dfpv", false, true, newDfpv)
	registerMeasurements(&dfpvMetric{}, &dfpvObject{})
}

type dfpv struct {
	client k8sClient
}

func newDfpv(client k8sClient) resource {
	return &dfpv{client: client}
}

func (c *dfpv) count() []pointV2 { return nil }

func (c *dfpv) hasNext() bool { return false }

func (c *dfpv) getMetadata(ctx context.Context, ns, _ string) (metadata, error) {
	data := newPodMetricsFromKubelet(c.client)
	pvcList, err := c.client.GetPersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		klog.Warnf("get pvc list fail, err: %s", err)
	}
	return &dfpvMetadata{data, pvcList}, nil
}

type dfpvMetadata struct {
	metricsData *podMetricsFromKubelet
	pvcList     *apicorev1.PersistentVolumeClaimList
}

func (m *dfpvMetadata) newMetric(conf *Config) pointKVs {
	var res pointKVs

	volumes, err := m.metricsData.GetPodsVolumeInfo(context.TODO())
	if err != nil {
		klog.Warnf("pod-volume info fail, err: %s, skip", err)
		return nil
	}

	for _, info := range volumes {
		met := typed.NewPointKV(dfpvMetricMeasurement)

		met.SetTag("name", info.pvcName+"/"+info.podName)
		met.SetTag("pvc_name", info.pvcName)
		met.SetTag("node_name", info.nodeName)
		met.SetTag("pod_name", info.podName)
		met.SetTag("namespace", info.namespace)
		met.SetTag("volume_mount_name", info.volumeMountName)

		met.SetField("available", info.available)
		met.SetField("capacity", info.capacity)
		met.SetField("used", info.used)
		met.SetField("inodes", info.inodes)
		met.SetField("inodes_used", info.inodesUsed)
		met.SetField("inodes_free", info.inodesFree)

		if info.capacity != 0 {
			met.SetField("used_percent", float64(info.used)/float64(info.capacity)*100) // percent
		}

		if m.pvcList != nil {
			for _, pvc := range m.pvcList.Items {
				if pvc.Namespace == info.namespace && pvc.Name == info.pvcName {
					met.SetLabelAsTags(pvc.Labels, conf.LabelAsTagsForNonMetric.All, conf.LabelAsTagsForMetric.Keys)
				}
			}
		}

		res = append(res, met)
	}

	return res
}

func (m *dfpvMetadata) newObject(conf *Config) pointKVs {
	var res pointKVs

	volumes, err := m.metricsData.GetPodsVolumeInfo(context.TODO())
	if err != nil {
		klog.Warnf("pod-volume info fail, err: %s, skip", err)
		return nil
	}

	for _, info := range volumes {
		obj := typed.NewPointKV(dfpvObjectMeasurement)

		obj.SetTag("name", info.pvcName+"/"+info.podName)
		obj.SetTag("pvc_name", info.pvcName)
		obj.SetTag("node_name", info.nodeName)
		obj.SetTag("pod_name", info.podName)
		obj.SetTag("namespace", info.namespace)
		obj.SetTag("volume_mount_name", info.volumeMountName)

		obj.SetField("available", info.available)
		obj.SetField("capacity", info.capacity)
		obj.SetField("used", info.used)
		obj.SetField("inodes", info.inodes)
		obj.SetField("inodes_used", info.inodesUsed)
		obj.SetField("inodes_free", info.inodesFree)

		if info.capacity != 0 {
			obj.SetField("used_percent", float64(info.used)/float64(info.capacity)*100) // percent
		}

		obj.SetField("message", typed.TrimString(obj.String(), maxMessageLength))

		if m.pvcList != nil {
			for _, pvc := range m.pvcList.Items {
				if pvc.Namespace == info.namespace && pvc.Name == info.pvcName {
					obj.SetLabelAsTags(pvc.Labels, conf.LabelAsTagsForNonMetric.All, conf.LabelAsTagsForNonMetric.Keys)
				}
			}
		}

		res = append(res, obj)
	}

	return res
}

type dfpvMetric struct{}

//nolint:lll
func (*dfpvMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dfpvMetricMeasurement,
		Desc: "The metric of the Kubernetes PersistentVolume.",
		Type: "metric",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("The dfpv name, consists of pvc name and pod name"),
			"pvc_name":          inputs.NewTagInfo("Reference to the PVC."),
			"node_name":         inputs.NewTagInfo("Reference to the Node."),
			"pod_name":          inputs.NewTagInfo("Reference to the Pod."),
			"namespace":         inputs.NewTagInfo("The namespace of Pod and PVC."),
			"volume_mount_name": inputs.NewTagInfo("The name given to the Volume."),
			"cluster_name_k8s":  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "AvailableBytes represents the storage space available (bytes) for the filesystem."},
			"capacity":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "CapacityBytes represents the total capacity (bytes) of the filesystems underlying storage."},
			"used":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "UsedBytes represents the bytes used for a specific task on the filesystem."},
			"inodes":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Inodes represents the total inodes in the filesystem."},
			"inodes_used": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "InodesUsed represents the inodes used by the filesystem."},
			"inodes_free": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "InodesFree represents the free inodes in the filesystem."},
		},
	}
}

type dfpvObject struct{}

//nolint:lll
func (*dfpvObject) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dfpvObjectMeasurement,
		Desc: "The object of the Kubernetes PersistentVolume.",
		Type: "object",
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("The dfpv name, consists of pvc name and pod name"),
			"pvc_name":          inputs.NewTagInfo("Reference to the PVC."),
			"node_name":         inputs.NewTagInfo("Reference to the Node."),
			"pod_name":          inputs.NewTagInfo("Reference to the Pod."),
			"namespace":         inputs.NewTagInfo("The namespace of Pod and PVC."),
			"volume_mount_name": inputs.NewTagInfo("The name given to the Volume."),
			"cluster_name_k8s":  inputs.NewTagInfo("K8s cluster name(default is `default`). We can rename it in datakit.yaml on ENV_CLUSTER_NAME_K8S."),
		},
		Fields: map[string]interface{}{
			"available":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "AvailableBytes represents the storage space available (bytes) for the filesystem."},
			"capacity":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "CapacityBytes represents the total capacity (bytes) of the filesystems underlying storage."},
			"used":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "UsedBytes represents the bytes used for a specific task on the filesystem."},
			"inodes":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "Inodes represents the total inodes in the filesystem."},
			"inodes_used": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "InodesUsed represents the inodes used by the filesystem."},
			"inodes_free": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "InodesFree represents the free inodes in the filesystem."},
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Object details"},
		},
	}
}
