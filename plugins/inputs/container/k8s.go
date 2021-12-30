package container

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//nolint:deadcode
// const k8sBearerToken = "/run/secrets/k8s.io/serviceaccount/token"

type kubernetesInput struct {
	client *k8sClient
	cfg    *kubernetesInputConfig
}

type kubernetesInputConfig struct {
	url               string
	bearerToken       string
	bearerTokenString string
	extraTags         map[string]string
}

func newKubernetesInput(cfg *kubernetesInputConfig) (*kubernetesInput, error) {
	k := &kubernetesInput{cfg: cfg}
	var err error

	//nolint:gocritic
	if cfg.bearerTokenString != "" {
		k.client, err = newK8sClientFromBearerTokenString(cfg.url, cfg.bearerTokenString)
	} else if cfg.bearerToken != "" {
		k.client, err = newK8sClientFromBearerToken(cfg.url, cfg.bearerToken)
	} else {
		err = fmt.Errorf("invalid bearerToken or bearerTokenString, cannot be empty")
	}
	if err != nil {
		return nil, err
	}
	return k, nil
}

var resourceList = []string{
	"cluster",
	"cronjob",
	"deployment",
	"job",
	"node",
	"pod",
	"replica_set",
	"service",
}

func (k *kubernetesInput) gather() (metrics, objects []inputs.Measurement, lastErr error) {
	resourceCount := make(map[string]map[string]int)

	must := func(res k8sResourceStats, err error) k8sResourceStats {
		lastErr = err
		return res
	}

	warpper := func(name string, res k8sResourceStats) {
		for namespace, v := range res {
			if x := resourceCount[namespace]; x == nil {
				resourceCount[namespace] = make(map[string]int)
			}
			resourceCount[namespace][name] += len(v)
			objects = append(objects, v...)
		}
	}

	warpper("cluster", must(gatherCluster(k.client, k.cfg.extraTags)))
	warpper("cronjob", must(gatherCronJob(k.client, k.cfg.extraTags)))
	warpper("deployment", must(gatherDeployment(k.client, k.cfg.extraTags)))
	warpper("job", must(gatherJob(k.client, k.cfg.extraTags)))
	warpper("node", must(gatherNode(k.client, k.cfg.extraTags)))
	warpper("pod", must(gatherPod(k.client, k.cfg.extraTags)))
	warpper("replica_set", must(gatherReplicaSet(k.client, k.cfg.extraTags)))
	warpper("service", must(gatherService(k.client, k.cfg.extraTags)))

	for namespace, resource := range resourceCount {
		c := newCount()
		c.tags["namespace"] = namespace
		for name, elem := range resource {
			c.fields[name] = elem
		}

		for _, r := range resourceList {
			if _, ok := c.fields[r]; !ok {
				c.fields[r] = 0
			}
		}

		c.time = time.Now()
		metrics = append(metrics, c)
	}
	return //nolint:nakedret
}

func (k *kubernetesInput) gatherPodMetrics() ([]inputs.Measurement, error) {
	if k.client.metricsClient == nil {
		return nil, nil
	}
	return gatherPodMetrics(k.client.metricsClient, k.cfg.extraTags)
}

func (k *kubernetesInput) watchingEventLog(stop <-chan interface{}) {
	watchingEvent(k.client, k.cfg.extraTags, stop)
}

type count struct {
	tags   tagsType
	fields fieldsType
	time   time.Time
}

func newCount() *count {
	return &count{
		tags:   make(tagsType),
		fields: make(fieldsType),
	}
}

const kubernetesMetricName = "kubernetes"

func (c *count) LineProto() (*io.Point, error) {
	return io.NewPoint(kubernetesMetricName, c.tags, c.fields, &io.PointOption{Time: c.time, Category: datakit.Metric})
}

//nolint:lll
func (*count) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: kubernetesMetricName,
		Desc: "Kubernetes count 指标数据",
		Type: "metric",
		Tags: map[string]interface{}{
			"namespace": &inputs.TagInfo{Desc: "namespace"},
		},
		Fields: map[string]interface{}{
			"cluster":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "cluster count"},
			"deployment":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "deployment count"},
			"node":        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "node count"},
			"pod":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "pod count"},
			"cronjob":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "cronjob count"},
			"job":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "job count"},
			"service":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "service count"},
			"replica_set": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "replica_set count"},
		},
	}
}

func defaultNamespace(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&count{})
}
