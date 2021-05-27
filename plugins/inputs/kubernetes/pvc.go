package kubernetes

import (
	"context"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

var pvcMeasurement = "kube_pvc"

type pvcM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *pvcM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *pvcM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: deploymentMeasurement,
		Desc: "kubernet daemonSet 对象",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "pod name"},
			"namespace": &inputs.TagInfo{Desc: "namespace"},
			"nodeName":  &inputs.TagInfo{Desc: "node name"},
		},
		Fields: map[string]interface{}{
			"ready": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "容器ready数/总数",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 状态",
			},
			"restarts": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "重启次数",
			},
			"age": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod存活时长",
			},
			"podIp": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod ip",
			},
			"createTime": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod 创建时间",
			},
			"label_xxx": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod lable",
			},
		},
	}
}

func (i *Input) collectPersistentVolumeClaims(ctx context.Context) error {
	list, err := i.client.getPersistentVolumeClaims(ctx)
	if err != nil {
		return err
	}
	for _, pvc := range list.Items {
		i.gatherPersistentVolumeClaim(pvc)
	}

	return err
}

func (i *Input) gatherPersistentVolumeClaim(pvc corev1.PersistentVolumeClaim) {
	phaseType := 3
	switch strings.ToLower(string(pvc.Status.Phase)) {
	case "bound":
		phaseType = 0
	case "lost":
		phaseType = 1
	case "pending":
		phaseType = 2
	}
	fields := map[string]interface{}{
		"phase_type": phaseType,
	}
	tags := map[string]string{
		"pvc_name":     pvc.Name,
		"namespace":    pvc.Namespace,
		"phase":        string(pvc.Status.Phase),
		"storageclass": *pvc.Spec.StorageClassName,
	}

	// for key, val := range pvc.Spec.Selector.MatchLabels {
	// 	if i.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	m := &pvcM{
		name:   daemonSetMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}
