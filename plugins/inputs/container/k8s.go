package container

import (
	"fmt"
	"os"
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

type inputsMeas []inputs.Measurement

func (k *kubernetesInput) gatherResourceMetric() (inputsMeas, error) {
	var (
		res     inputsMeas
		lastErr error
	)

	{
		x := newDeployment(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newDaemonset(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newEndpoint(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newCronjob(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newJob(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newPod(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newNode(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newReplicaset(k.client, k.cfg.extraTags)
		if m, err := x.metric(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}

	return res, lastErr
}

func (k *kubernetesInput) gatherResourceObject() (inputsMeas, error) {
	var (
		res     inputsMeas
		lastErr error
	)

	{
		x := newCronjob(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newDeployment(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newJob(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newNode(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newPod(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newReplicaset(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newService(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}
	{
		x := newClusterRole(k.client, k.cfg.extraTags)
		if m, err := x.object(); err != nil {
			res = append(m)
		} else {
			lastErr = err
		}
	}

	return res, lastErr
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

type k8sResourceMetricInterface interface {
	metric() (inputsMeas, error)
	count() (map[string]int, error)
}

type k8sResourceObjectInterface interface {
	object() (inputsMeas, error)
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
			"cluster_role": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "RBAC cluster role count"},
			"deployment":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "deployment count"},
			"node":         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "node count"},
			"pod":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "pod count"},
			"cronjob":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "cronjob count"},
			"job":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "job count"},
			"service":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "service count"},
			"replica_set":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.UnknownUnit, Desc: "replica_set count"},
		},
	}
}

func defaultNamespace(ns string) string {
	if ns == "" {
		return "default"
	}
	return ns
}

func defaultClusterName(name string) string {
	if name != "" {
		return name
	}
	if e := os.Getenv("ENV_K8S_CLUSTER_NAME"); e != "" {
		return e
	}
	return "kubernetes"
}

//nolint:gochecknoinits
func init() {
	registerMeasurement(&count{})
}
