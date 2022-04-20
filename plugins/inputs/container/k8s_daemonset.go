package container

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
)

func gatherDaemonsetMetric(client k8sClientX, extraTags map[string]string) (*k8sResourceStats, error) {
	list, err := client.getDaemonSets().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonSets resource: %w", err)
	}
	if len(list.Items) == 0 {
		return nil, nil
	}
	return exportDaemonsetMetric(list.Items, extraTags), nil
}

func exportDaemonsetMetric(items []v1.DaemonSet, extraTags tagsType) *k8sResourceStats {
	res := newK8sResourceStats()
	for _, item := range items {
		met := &daemonsetMetric{
			tags: map[string]string{
				"daemonset": item.Name,
				"namespace": item.Namespace,
			},
			fields: map[string]interface{}{
				"count":               -1,
				"scheduled":           item.Status.CurrentNumberScheduled,
				"desired":             item.Status.DesiredNumberScheduled,
				"misscheduled":        item.Status.NumberMisscheduled,
				"ready":               item.Status.NumberReady,
				"updated":             item.Status.UpdatedNumberScheduled,
				"daemons_unavailable": item.Status.NumberUnavailable,
			},
			time: time.Now(),
		}
		met.tags.append(extraTags)
		res.meas = append(res.meas, met)
		res.namespaceList[item.Namespace]++
	}
	return res
}

type daemonsetMetric struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func (d *daemonsetMetric) LineProto() (*io.Point, error) {
	return io.NewPoint("kube_daemonset", d.tags, d.fields, &io.PointOption{Time: d.time, Category: datakit.Metric})
}

//nolint:lll
func (*daemonsetMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "kube_daemonset",
		Desc: "Kubernetes DaemonSet 指标数据",
		Type: "object",
		Tags: map[string]interface{}{
			"daemonset": inputs.NewTagInfo("Name must be unique within a namespace."),
			"namespace": inputs.NewTagInfo("Namespace defines the space within each name must be unique."),
		},
		Fields: map[string]interface{}{
			"scheduled":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running at least one daemon pod and are supposed to run the daemon pod."},
			"desired":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that should be running the daemon pod (including nodes correctly running the daemon pod)."},
			"misscheduled": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that are running the daemon pod, but are not supposed to run the daemon pod."},
			"ready":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready."},

			"updated": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The total number of nodes that are running updated daemon pod."},

			"daemons_unavailable": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of nodes that should be running the daemon pod and have none of the daemon pod running and available (ready for at least spec.minReadySeconds)."},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&daemonsetMetric{})
}
