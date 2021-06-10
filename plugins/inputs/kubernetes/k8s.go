package kubernetes

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var k8sMeasurement = "kubernetes"

type kubernetesMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *kubernetesMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *kubernetesMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sMeasurement,
		Desc: "kubernetes metric",
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"daemonset": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "daemonset count",
			},
			"deployment": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "deployment count",
			},
			"endpoint": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "endpoint count",
			},
			"node": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "node count",
			},
			"container": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "container count",
			},
			"pod": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "pod count",
			},
			"statefulSet": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "statefulSet count",
			},
			"cronjob": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "cronjob count",
			},
			"job": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "job count",
			},
			"service": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.UnknownUnit,
				Desc:     "service count",
			},
		},
	}
}

func (i *Input) collectKubernetes(collector string) error {
	list, err := i.client.getNamespaces()
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		ns := item.Name

		i.client.namespace = ns

		tags["namespace"] = ns

		for key, value := range i.Tags {
			tags[key] = value
		}

		// DaemonSets
		if list, err := i.client.getDaemonSets(); err != nil {
			l.Error(err)
		} else {
			fields["daemonSet"] = len(list.Items)
		}

		// deployment
		if list, err := i.client.getDeployments(); err != nil {
			l.Error(err)
		} else {
			fields["deployment"] = len(list.Items)
		}

		// endpoint
		if list, err := i.client.getEndpoints(); err != nil {
			l.Error(err)
		} else {
			fields["endpoint"] = len(list.Items)
		}

		// node
		if list, err := i.client.getNodes(); err != nil {
			l.Error(err)
		} else {
			fields["node"] = len(list.Items)
		}

		// service
		if list, err := i.client.getServices(); err != nil {
			l.Error(err)
		} else {
			fields["service"] = len(list.Items)
		}

		// statefulSets
		if list, err := i.client.getStatefulSets(); err != nil {
			l.Error(err)
		} else {
			fields["statefulSets"] = len(list.Items)
		}

		// ingress
		if list, err := i.client.getIngress(); err != nil {
			l.Error(err)
		} else {
			fields["ingress"] = len(list.Items)
		}

		if list, err := i.client.getPods(); err != nil {
			l.Error(err)
		} else {
			fields["pod"] = len(list.Items)
			containerCnt := 0
			for _, p := range list.Items {
				containerCnt += len(p.Spec.Containers)
			}
			fields["container"] = containerCnt
		}

		if list, err := i.client.getJobs(); err != nil {
			l.Error(err)
		} else {
			fields["job"] = len(list.Items)
		}

		if list, err := i.client.getCronJobs(); err != nil {
			l.Error(err)
		} else {
			fields["cronJob"] = len(list.Items)
		}

		m := &kubernetesMetric{
			name:   k8sMeasurement,
			tags:   tags,
			fields: fields,
			ts:     time.Now(),
		}

		i.collectCache[collector] = append(i.collectCache[collector], m)
		i.client.namespace = ""
	}

	return nil
}
