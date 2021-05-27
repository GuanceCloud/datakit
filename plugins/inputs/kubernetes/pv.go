package kubernetes

import (
	"context"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

var pvMeasurement = "kube_daemonset"

type pvM struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *pvM) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *pvM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: daemonSetMeasurement,
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

func (i *Input) collectPersistentVolumes(ctx context.Context) error {
	list, err := i.client.getPersistentVolumes(ctx)
	if err != nil {
		return err
	}
	for _, pv := range list.Items {
		i.gatherPersistentVolume(pv)
	}

	return nil
}

func (i *Input) gatherPersistentVolume(pv corev1.PersistentVolume) {
	phaseType := 5
	switch strings.ToLower(string(pv.Status.Phase)) {
	case "bound":
		phaseType = 0
	case "failed":
		phaseType = 1
	case "pending":
		phaseType = 2
	case "released":
		phaseType = 3
	case "available":
		phaseType = 4
	}
	fields := map[string]interface{}{
		"phase_type": phaseType,
	}
	tags := map[string]string{
		"pv_name":      pv.Name,
		"phase":        string(pv.Status.Phase),
		"storageclass": pv.Spec.StorageClassName,
	}

	m := &pvM{
		name:   daemonSetMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}
