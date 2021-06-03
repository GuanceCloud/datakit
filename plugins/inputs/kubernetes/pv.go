package kubernetes

import (
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	corev1 "k8s.io/api/core/v1"
)

var pvMeasurement = "kube_pv"

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
		Name: pvMeasurement,
		Desc: "kubernet pv",
		Tags: map[string]interface{}{
			"pv_name":      &inputs.TagInfo{Desc: "pv name"},
			"phase":        &inputs.TagInfo{Desc: "Phase indicates if a volume is available, bound to a claim, or released by a claim"},
			"storageclass": &inputs.TagInfo{Desc: "Name of StorageClass to which this persistent volume belongs. Empty value means that this volume does not belong to any StorageClass"},
		},
		Fields: map[string]interface{}{
			"phase_type": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "phase type, bound:0, failed:1, pending:2, released:3, available:4, unknown: 5",
			},
		},
	}
}

func (i *Input) collectPersistentVolumes(collector string) error {
	list, err := i.client.getPersistentVolumes()
	if err != nil {
		return err
	}
	for _, pv := range list.Items {
		i.gatherPersistentVolume(collector, pv)
	}

	return nil
}

func (i *Input) gatherPersistentVolume(collector string, pv corev1.PersistentVolume) {
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
		name:   pvMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache[collector] = append(i.collectCache[collector], m)
}
