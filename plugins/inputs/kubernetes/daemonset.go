package kubernetes

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
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
		Desc: "kubernet daemonSet",
		Tags: map[string]interface{}{
			"daemonset_name": &inputs.TagInfo{Desc: "pod name"},
			"namespace":      &inputs.TagInfo{Desc: "namespace"},
			"selector_*":     &inputs.TagInfo{Desc: "lab"},
		},
		Fields: map[string]interface{}{
			"generation": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "A sequence number representing a specific generation of the desired state",
			},
			"current_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes running at least one daemon pod and are supposed to",
			},
			"desired_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes that should be running the daemon pod",
			},
			"number_available": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available",
			},
			"number_misscheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes running a daemon pod but are not supposed to",
			},
			"number_ready": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready",
			},
			"number_unavailable": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available",
			},
			"updated_number_scheduled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "The total number of nodes that are running updated daemon pod",
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

	for key, val := range d.Spec.Selector.MatchLabels {
		tags["selector_"+key] = val
	}

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
