// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"k8s.io/klog/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/discovery"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/kubernetes"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/option"
)

var l = logger.DefaultSLogger(inputName)

func getCollectorMeasurement() []inputs.Measurement {
	res := []inputs.Measurement{
		&containerMetric{},
		&containerObject{},
		&containerLog{},
	}
	res = append(res, kubernetes.Measurements()...)
	return res
}

func (ipt *Input) setup() {
	if ipt.DeprecatedDockerEndpoint != "" {
		ipt.Endpoints = append(ipt.Endpoints, ipt.DeprecatedDockerEndpoint)
	}
	if ipt.DeprecatedContainerdAddress != "" {
		ipt.Endpoints = append(ipt.Endpoints, "unix://"+ipt.DeprecatedContainerdAddress)
	}
	ipt.Endpoints = unique(ipt.Endpoints)
	l.Infof("endpoints: %v", ipt.Endpoints)
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("container input started")
	ipt.setup()

	if datakit.Docker && config.IsKubernetes() {
		ipt.startDiscovery()
	}

	ipt.runCollect()
}

func (ipt *Input) runCollect() {
	objectTick := time.NewTicker(objectInterval)
	defer objectTick.Stop()

	metricTick := time.NewTicker(metricInterval)
	defer metricTick.Stop()

	loggingInterval := metricInterval
	if ipt.LoggingSearchInterval > 0 {
		loggingInterval = ipt.LoggingSearchInterval
	}
	loggingTick := time.NewTicker(loggingInterval)
	defer loggingTick.Stop()

	collectors := ipt.newCollector()
	firstCollectElection := true

	// frist collect
	ipt.collectLogging(collectors)
	ipt.collectObject(collectors)
	time.Sleep(time.Second) // window time
	ipt.collectMetric(collectors)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("container exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("container terminate")
			return

		case <-metricTick.C:
			ipt.collectMetric(collectors)

		case <-objectTick.C:
			time.Sleep(time.Second) // window time
			ipt.collectObject(collectors)

		case <-loggingTick.C:
			ipt.collectLogging(collectors)

		case pause := <-ipt.chPause:
			ipt.pause.Store(pause)

			if !pause && firstCollectElection {
				l.Info("first collect election metrics and objects")

				ipt.collectMetric(collectors, option.WithOnlyElection(true))
				time.Sleep(time.Second) // window time
				ipt.collectObject(collectors, option.WithOnlyElection(true))

				firstCollectElection = false
			}
		}
	}
}

func (ipt *Input) collectMetric(collectors []Collector, opts ...option.CollectOption) {
	start := time.Now()

	g := goroutine.NewGroup(goroutine.Option{Name: "container/k8s-metric"})
	for idx := range collectors {
		func(c Collector) {
			fn := func(pts []*point.Point) error {
				if len(pts) == 0 {
					return nil
				}
				return ipt.Feeder.FeedV2(point.Metric, pts,
					dkio.WithBlocking(true),
					dkio.WithElection(c.Election()),
					dkio.WithInputName(c.Name()+"-metric"))
			}
			g.Go(func(ctx context.Context) error {
				c.Metric(fn, append(opts, option.WithPaused(ipt.pause.Load()))...)
				return nil
			})
		}(collectors[idx])
	}
	if err := g.Wait(); err != nil {
		klog.Errorf("unexpected collect error: %s", err)
	}

	totalCostVec.WithLabelValues("metric").Observe(time.Since(start).Seconds())
}

func (ipt *Input) collectObject(collectors []Collector, opts ...option.CollectOption) {
	start := time.Now()

	g := goroutine.NewGroup(goroutine.Option{Name: "container/k8s-object"})
	for idx := range collectors {
		func(c Collector) {
			fn := func(pts []*point.Point) error {
				if len(pts) == 0 {
					return nil
				}
				return ipt.Feeder.FeedV2(point.Object, pts,
					dkio.WithBlocking(true),
					dkio.WithElection(c.Election()),
					dkio.WithInputName(c.Name()+"-object"))
			}
			g.Go(func(ctx context.Context) error {
				c.Object(fn, append(opts, option.WithPaused(ipt.pause.Load()))...)
				return nil
			})
		}(collectors[idx])
	}
	if err := g.Wait(); err != nil {
		klog.Errorf("unexpected collect error: %s", err)
	}

	totalCostVec.WithLabelValues("object").Observe(time.Since(start).Seconds())
}

func (ipt *Input) collectLogging(collectors []Collector) {
	for _, c := range collectors {
		fn := func(pts []*point.Point) error {
			if len(pts) == 0 {
				return nil
			}
			return ipt.Feeder.FeedV2(point.Logging, pts,
				dkio.WithBlocking(true),
				dkio.WithElection(c.Election()),
				dkio.WithInputName(c.Name()+"-logging"))
		}
		c.Logging(fn)
	}
}

func (ipt *Input) startDiscovery() {
	discovery, err := newDiscovery(ipt)
	if err != nil {
		l.Errorf("init the auto-discovery fail, err: %s", err)
		return
	}

	g := datakit.G("k8s-discovery")
	g.Go(func(ctx context.Context) error {
		discovery.Run()
		return nil
	})
}

type Collector interface {
	Name() string
	Election() bool
	Metric(func(pts []*point.Point) error, ...option.CollectOption)
	Object(func(pts []*point.Point) error, ...option.CollectOption)
	Logging(func(pts []*point.Point) error)
}

func (ipt *Input) newCollector() []Collector {
	collectors := []Collector{}
	collectors = append(collectors, newCollectorsFromContainerEndpoints(ipt)...)

	if datakit.Docker && config.IsKubernetes() {
		k8sCollectors, err := newCollectorsFromKubernetes(ipt)
		if err != nil {
			l.Errorf("init the k8s fail, err: %s", err)
		} else {
			collectors = append(collectors, k8sCollectors)
		}
	}

	return collectors
}

func newCollectorsFromContainerEndpoints(ipt *Input) []Collector {
	var collectors []Collector

	if config.IsECSFargate() {
		var baseURL string

		v4 := config.ECSFargateBaseURIV4()
		if v4 != "" {
			baseURL = v4
			l.Infof("connect ecsfargate v4 with url %s", v4)
		} else {
			v3 := config.ECSFargateBaseURIV3()
			if v3 != "" {
				baseURL = v3
				l.Infof("connect ecsfargate v3 with url %s", v4)
			}
		}

		if baseURL != "" {
			collector, err := newECSFargate(ipt, baseURL)
			if err != nil {
				l.Errorf("unable to connect ecsfargate url %s, err: %s", baseURL, err)
			}
			collectors = append(collectors, collector)
		} else {
			l.Errorf("unexpected ecsfargate url, version only be v3 or v4")
		}

		return collectors
	}

	for _, endpoint := range ipt.Endpoints {
		if err := checkEndpoint(endpoint); err != nil {
			l.Warnf("%s, skip", err)
			continue
		}

		var client k8sclient.Client
		var err error

		if datakit.Docker && config.IsKubernetes() {
			client, err = newKubernetesClient(ipt)
			if err != nil {
				l.Warnf("unable to connect k8s client, err: %s, skip", err)
			}
		}

		collector, err := newContainer(ipt, endpoint, getMountPoint(), client)
		if err != nil {
			l.Warnf("cannot connect endpoint, err: %s", err)
			continue
		}

		l.Infof("connect runtime with %s", endpoint)
		collectors = append(collectors, collector)
	}

	return collectors
}

func newCollectorsFromKubernetes(ipt *Input) (Collector, error) {
	client, err := newKubernetesClient(ipt)
	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, "")
	if name := getClusterNameK8s(); name != "" {
		tags["cluster_name_k8s"] = name
	}

	optForNonMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys)
	optForMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2ForMetric, config.Cfg.Dataway.GlobalCustomerKeys)

	l.Infof("Use labels %s for k8s non-metric", optForNonMetric.keys)
	l.Infof("Use labels %s for k8s metric", optForMetric.keys)

	cfg := kubernetes.Config{
		NodeName:                      config.Cfg.Hostname,
		NodeLocal:                     ipt.EnableK8sNodeLocal,
		EnableK8sMetric:               ipt.EnableK8sMetric,
		EnableK8sObject:               true,
		EnablePodMetric:               ipt.EnablePodMetric,
		EnableK8sEvent:                ipt.EnableK8sEvent,
		EnableExtractK8sLabelAsTagsV1: ipt.DeprecatedEnableExtractK8sLabelAsTags,
		DisableCollectJob:             ipt.disableCollectK8sJob,
		LabelAsTagsForMetric: kubernetes.LabelsOption{
			All:  optForMetric.all,
			Keys: optForMetric.keys,
		},
		LabelAsTagsForNonMetric: kubernetes.LabelsOption{
			All:  optForNonMetric.all,
			Keys: optForNonMetric.keys,
		},
		ExtraTags: tags,
	}

	checkPaused := func() bool {
		return ipt.pause.Load()
	}

	return kubernetes.NewKubeCollector(client, &cfg, checkPaused, ipt.semStop.Wait())
}

func newDiscovery(ipt *Input) (*discovery.Discovery, error) {
	client, err := newKubernetesClient(ipt)
	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")
	opt := buildLabelsOption(nil, config.Cfg.Dataway.GlobalCustomerKeys)

	cfg := discovery.Config{
		EnablePrometheusPodAnnotations:     ipt.EnableAutoDiscoveryOfPrometheusPodAnnotations,
		EnablePrometheusServiceAnnotations: ipt.EnableAutoDiscoveryOfPrometheusServiceAnnotations,
		EnablePrometheusPodMonitors:        ipt.EnableAutoDiscoveryOfPrometheusPodMonitors,
		EnablePrometheusServiceMonitors:    ipt.EnableAutoDiscoveryOfPrometheusServiceMonitors,
		StreamSize:                         ipt.autoDiscoveryOfPromStreamSize,
		ExtraTags:                          tags,
		LabelAsTags:                        opt.keys,
		Feeder:                             ipt.Feeder,
	}

	return discovery.NewDiscovery(client, &cfg, ipt.semStop.Wait()), nil
}

func newKubernetesClient(ipt *Input) (k8sclient.Client, error) {
	if ipt.K8sBearerTokenString != "" {
		client, err := k8sclient.NewKubernetesClientFromBearerTokenString(ipt.K8sURL, ipt.K8sBearerTokenString)
		if err != nil {
			return nil, fmt.Errorf("new k8s client fails for the token string, err: %w", err)
		}
		return client, err
	}

	if ipt.K8sBearerToken != "" {
		client, err := k8sclient.NewKubernetesClientFromBearerToken(ipt.K8sURL, ipt.K8sBearerToken)
		if err != nil {
			return nil, fmt.Errorf("new k8s client fails for the token file, err: %w", err)
		}
		return client, err
	}

	return nil, fmt.Errorf("invalid token or token string, cannot be empty")
}

func checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s, err: %w", endpoint, err)
	}

	switch u.Scheme {
	case "unix":
		// nil
	default:
		return fmt.Errorf("using %s as endpoint is not supported protocol", endpoint)
	}

	info, err := os.Stat(u.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("endpoint %s does not exist, maybe it is not running", endpoint)
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("endpoint %s cannot be a directory", u.Path)
	}

	return nil
}

type labelsOption struct {
	all  bool
	keys []string
}

func buildLabelsOption(asTagKeys, customerKeys []string) labelsOption {
	// e.g. [""] (all)
	if len(asTagKeys) == 1 && asTagKeys[0] == "" {
		return labelsOption{all: true}
	}
	keys := unique(append(asTagKeys, customerKeys...))
	sort.Strings(keys)
	return labelsOption{keys: keys}
}

func getMountPoint() string {
	if !datakit.Docker {
		return ""
	}
	if n := os.Getenv("HOST_ROOT"); n != "" {
		return n
	}
	return "/rootfs"
}

func getClusterNameK8s() string {
	return os.Getenv("ENV_CLUSTER_NAME_K8S")
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
