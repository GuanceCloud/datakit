package kubernetes

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var k8sMeasurement = "kubernetes"

type kubernetesMetric struct {
	client *client
	tags   map[string]string
}

func (m *kubernetesMetric) LineProto() (*io.Point, error) { return nil, nil }

//nolint:lll
func (m *kubernetesMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sMeasurement,
		Desc: "Kubernetes 指标数据",
		Type: "metric",
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

func (m *kubernetesMetric) Gather() {
	list, err := m.client.getNamespaces().List(context.Background(), metav1ListOption)
	if err != nil {
		l.Error(err)
		return
	}

	defer func() {
		m.client.namespace = ""
	}()

	for _, item := range list.Items {
		tags := map[string]string{}
		fields := map[string]interface{}{}

		ns := item.Name
		m.client.namespace = ns

		tags["namespace"] = ns

		for key, value := range m.tags {
			tags[key] = value
		}

		// DaemonSets
		if list, err := m.client.getDaemonSets().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["daemonSet"] = len(list.Items)
		}

		// deployment
		if list, err := m.client.getDeployments().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["deployment"] = len(list.Items)
		}

		// endpoint
		if list, err := m.client.getEndpoints().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["endpoint"] = len(list.Items)
		}

		// node
		if list, err := m.client.getNodes().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["node"] = len(list.Items)
		}

		// service
		if list, err := m.client.getServices().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["service"] = len(list.Items)
		}

		// statefulSets
		if list, err := m.client.getStatefulSets().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["statefulSets"] = len(list.Items)
		}

		// ingress
		if list, err := m.client.getIngress().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["ingress"] = len(list.Items)
		}

		if list, err := m.client.getPods("").List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["pod"] = len(list.Items)
			containerCnt := 0
			for _, p := range list.Items {
				containerCnt += len(p.Spec.Containers)
			}
			fields["container"] = containerCnt
		}

		if list, err := m.client.getJobs().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["job"] = len(list.Items)
		}

		if list, err := m.client.getCronJobs().List(context.Background(), metav1ListOption); err != nil {
			l.Error(err)
		} else {
			fields["cronJob"] = len(list.Items)
		}

		pt, err := io.MakePoint(k8sMeasurement, tags, fields, time.Now())
		if err != nil {
			l.Error(err)
		} else if err := io.Feed(inputName, datakit.Metric, []*io.Point{pt}, nil); err != nil {
			l.Error(err)
		}
	}
}
