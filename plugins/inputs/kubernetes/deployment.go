package kubernetes

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/apps/v1"
)

var deploymentMeasurement = "kube_deployment"

type deployment struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *deployment) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *deployment) Info() *inputs.MeasurementInfo {
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

func (i *Input) collectDeployments(ctx context.Context) error {
	list, err := i.client.getDeployments(ctx)
	if err != nil {
		return err
	}
	for _, d := range list.Items {
		i.gatherDeployment(d)
	}

	return nil
}

func (i *Input) gatherDeployment(d v1.Deployment) {
	fields := map[string]interface{}{
		"replicas_available":   d.Status.AvailableReplicas,
		"replicas_unavailable": d.Status.UnavailableReplicas,
		"created":              d.GetCreationTimestamp().UnixNano(),
	}
	tags := map[string]string{
		"deployment_name": d.Name,
		"namespace":       d.Namespace,
	}

	// for key, val := range d.Spec.Selector.MatchLabels {
	// 	if ki.selectorFilter.Match(key) {
	// 		tags["selector_"+key] = val
	// 	}
	// }

	if d.GetCreationTimestamp().Second() != 0 {
		fields["created"] = d.GetCreationTimestamp().UnixNano()
	}

	m := &deployment{
		name:   deploymentMeasurement,
		tags:   tags,
		fields: fields,
		ts:     time.Now(),
	}

	i.collectCache = append(i.collectCache, m)

}
