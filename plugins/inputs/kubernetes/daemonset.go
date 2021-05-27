package kubernetes

import (
	"context"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
	"time"
)

var daemonSetMeasurement = "kube_daemonset"

type daemonSet struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *daemonSet) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *daemonSet) Info() *inputs.MeasurementInfo {
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

func (i *Input) collectDaemonSets(ctx context.Context) error {
	list, err := i.client.getDaemonSets(ctx)
	if err != nil {
		return err
	}

	for _, d := range list.Items {
		i.gatherDaemonSet(d)
	}

	return nil
}

func (i *Input) gatherDaemonSet(d v1.DaemonSet) {
	fields := map[string]interface{}{
		"generation":               d.Generation,
		"current_number_scheduled": d.Status.CurrentNumberScheduled,
		"desired_number_scheduled": d.Status.DesiredNumberScheduled,
		"number_available":         d.Status.NumberAvailable,
		"number_misscheduled":      d.Status.NumberMisscheduled,
		"number_ready":             d.Status.NumberReady,
		"number_unavailable":       d.Status.NumberUnavailable,
		"updated_number_scheduled": d.Status.UpdatedNumberScheduled,
	}

	tags := map[string]string{
		"daemonset_name": d.Name,
		"namespace":      d.Namespace,
	}
	// for key, val := range d.Spec.Selector.MatchLabels {
	// 	if i.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	if d.GetCreationTimestamp().Second() != 0 {
		fields["created"] = d.GetCreationTimestamp().UnixNano()
	}

	m := &daemonSet{
		name:   daemonSetMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)
}
