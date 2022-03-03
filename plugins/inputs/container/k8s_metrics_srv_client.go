package container

import (
	"context"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"k8s.io/client-go/rest"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type k8sMetricsClientX interface {
	getPodMetrics() metricsv1beta1.PodMetricsInterface
	getPodMetricsForNamespace(namespace string) metricsv1beta1.PodMetricsInterface
	getNodeMetrics() metricsv1beta1.NodeMetricsInterface
}

type k8sMetricsClient struct {
	*metricsv1beta1.MetricsV1beta1Client
}

func newK8sMetricsClient(restConfig *rest.Config) (*k8sMetricsClient, error) {
	client, err := metricsv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &k8sMetricsClient{client}, nil
}

func (c *k8sMetricsClient) getPodMetrics() metricsv1beta1.PodMetricsInterface {
	return c.PodMetricses("")
}

func (c *k8sMetricsClient) getPodMetricsForNamespace(namespace string) metricsv1beta1.PodMetricsInterface {
	return c.PodMetricses(namespace)
}

func (c *k8sMetricsClient) getNodeMetrics() metricsv1beta1.NodeMetricsInterface {
	return c.NodeMetricses()
}

const (
	k8sPodMetricName = "kubelet_pod"
)

func gatherPodMetrics(client k8sMetricsClientX, extraTags map[string]string) ([]inputs.Measurement, error) {
	list, err := client.getPodMetrics().List(context.Background(), metaV1ListOption)
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}

	var res []inputs.Measurement

	for _, item := range list.Items {
		if len(item.Containers) == 0 {
			continue
		}
		obj := newPodMetric()
		obj.tags["pod_name"] = item.Name
		obj.tags.addValueIfNotEmpty("cluster_name", item.ClusterName)
		obj.tags.addValueIfNotEmpty("namespace", defaultNamespace(item.Namespace))
		obj.tags.append(extraTags)

		cpu := item.Containers[0].Usage["cpu"]
		mem := item.Containers[0].Usage["memory"]
		for i := 1; i < len(item.Containers); i++ {
			if c, ok := item.Containers[i].Usage["cpu"]; ok {
				cpu.Add(c)
			}
			if m, ok := item.Containers[i].Usage["memory"]; ok {
				mem.Add(m)
			}
		}

		cpuUsage, err := strconv.ParseFloat(cpu.AsDec().String(), 64)
		if err != nil {
			l.Debugf("k8s pod metrics, parsed cpu err: %w", err)
		}
		memUsage, _ := mem.AsInt64()

		obj.fields["cpu_usage"] = cpuUsage * 100 // percentage
		obj.fields["memory_usage_bytes"] = memUsage

		obj.time = time.Now()
		res = append(res, obj)
	}
	return res, nil
}

type podMetric struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newPodMetric() *podMetric {
	return &podMetric{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

func (p *podMetric) LineProto() (*io.Point, error) {
	return io.NewPoint(k8sPodMetricName, p.tags, p.fields, &io.PointOption{Time: p.time, Category: datakit.Metric})
}

//nolint:lll
func (*podMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: k8sPodMetricName,
		Desc: "Kubernetes pod 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"pod_name":     inputs.NewTagInfo(`pod name`),
			"cluster_name": inputs.NewTagInfo("The name of the cluster which the object belongs to."),
			"namespace":    inputs.NewTagInfo(`Namespace defines the space within each name must be unique.`),
		},
		Fields: map[string]interface{}{
			"cpu_usage":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Percent, Desc: "The percentage of cpu used"},
			"memory_usage_bytes": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "The number of memory used in bytes"},
		},
	}
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&podMetric{})
}
