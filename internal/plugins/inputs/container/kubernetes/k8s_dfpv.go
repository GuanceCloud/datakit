// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"k8s.io/client-go/informers"
)

const (
	dfpvMetricMeasurement = "kube_dfpv"
	dfpvObjectClass       = "kubernetes_dfpv"
)

//nolint:gochecknoinits
func init() {
	registerResource("dfpv", true, newDfpv)
	registerMeasurements(&dfpvMetric{}, &dfpvObject{})
}

type dfpv struct {
	client k8sClient
	cfg    *Config
}

func newDfpv(client k8sClient, cfg *Config) resource {
	return &dfpv{client: client, cfg: cfg}
}

func (d *dfpv) gatherMetric(ctx context.Context, timestamp int64) {
	if !d.cfg.NodeLocal {
		return
	}

	list, err := newPodMetricsFromKubelet(d.client).GetPodsVolumeInfo(context.TODO())
	if err != nil {
		klog.Warnf("query for pod-volume failed, err: %s", err)
		return
	}

	pts := d.buildMetricPoints(list, timestamp)
	feedMetric("k8s-dfpv-metric", d.cfg.Feeder, pts, false)
}

func (d *dfpv) gatherObject(ctx context.Context) {
	if !d.cfg.NodeLocal {
		return
	}

	list, err := newPodMetricsFromKubelet(d.client).GetPodsVolumeInfo(context.TODO())
	if err != nil {
		klog.Warnf("query for pod-volume failed, err: %s", err)
		return
	}

	pts := d.buildObjectPoints(list)
	feedObject("k8s-dfpv-object", d.cfg.Feeder, pts, false)
}

func (d *dfpv) addChangeInformer(_ informers.SharedInformerFactory) { /* nil */ }

func (d *dfpv) buildMetricPoints(list []*podVolumeInfo, timestamp int64) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, item := range list {
		var kvs point.KVs

		kvs = kvs.AddTag("name", item.pvcName+"/"+item.podName)
		kvs = kvs.AddTag("pvc_name", item.pvcName)
		kvs = kvs.AddTag("node_name", config.RenameNode(item.nodeName))
		kvs = kvs.AddTag("pod_name", item.podName)
		kvs = kvs.AddTag("namespace", item.namespace)
		kvs = kvs.AddTag("volume_mount_name", item.volumeMountName)

		kvs = kvs.AddV2("available", item.available, false)
		kvs = kvs.AddV2("capacity", item.capacity, false)
		kvs = kvs.AddV2("used", item.used, false)
		kvs = kvs.AddV2("inodes", item.inodes, false)
		kvs = kvs.AddV2("inodes_used", item.inodesUsed, false)
		kvs = kvs.AddV2("inodes_free", item.inodesFree, false)

		if item.capacity != 0 {
			kvs = kvs.AddV2("used_percent", float64(item.used)/float64(item.capacity)*100, false) // percent
		}

		//
		// dfpv 不使用 LabelAsTagsForMetric
		//
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(dfpvMetricMeasurement, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

func (d *dfpv) buildObjectPoints(list []*podVolumeInfo) []*point.Point {
	var pts []*point.Point
	opts := point.DefaultObjectOptions()

	for _, item := range list {
		var kvs point.KVs

		kvs = kvs.AddTag("name", item.pvcName+"/"+item.podName)
		kvs = kvs.AddTag("pvc_name", item.pvcName)
		kvs = kvs.AddTag("node_name", config.RenameNode(item.nodeName))
		kvs = kvs.AddTag("pod_name", item.podName)
		kvs = kvs.AddTag("namespace", item.namespace)
		kvs = kvs.AddTag("volume_mount_name", item.volumeMountName)

		kvs = kvs.AddV2("available", item.available, false)
		kvs = kvs.AddV2("capacity", item.capacity, false)
		kvs = kvs.AddV2("used", item.used, false)
		kvs = kvs.AddV2("inodes", item.inodes, false)
		kvs = kvs.AddV2("inodes_used", item.inodesUsed, false)
		kvs = kvs.AddV2("inodes_free", item.inodesFree, false)

		if item.capacity != 0 {
			kvs = kvs.AddV2("used_percent", float64(item.used)/float64(item.capacity)*100, false) // percent
		}

		msg := pointutil.PointKVsToJSON(kvs)
		kvs = kvs.AddV2("message", pointutil.TrimString(msg, maxMessageLength), false)

		//
		// dfpv 不使用 LabelAsTagsForNonMetric
		//
		kvs = append(kvs, point.NewTags(d.cfg.ExtraTags)...)
		pt := point.NewPointV2(dfpvObjectClass, kvs, opts...)
		pts = append(pts, pt)
	}

	return pts
}

type dfpvMetric struct{}

//nolint:lll
func (*dfpvMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dfpvMetricMeasurement,
		Desc: "The metric of the Kubernetes PersistentVolume.",
		Cat:  point.Metric,
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
		Name: dfpvObjectClass,
		Desc: "The object of the Kubernetes PersistentVolume.",
		Cat:  point.Object,
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
