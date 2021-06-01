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
		Desc: "kubernet deployment",
		Tags: map[string]interface{}{
			"deployment_name": &inputs.TagInfo{Desc: "deployment name"},
			"namespace":       &inputs.TagInfo{Desc: "namespace"},
			"selector_*":      &inputs.TagInfo{Desc: "lab"},
		},
		Fields: map[string]interface{}{
			"replicas_available": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Total number of available pods (ready for at least minReadySeconds) targeted by this deployment",
			},
			"replicas_unavailable": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "Total number of unavailable pods targeted by this deployment",
			},
			"created": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "create time",
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

	for key, val := range d.Spec.Selector.MatchLabels {
		tags["selector_"+key] = val
	}

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
