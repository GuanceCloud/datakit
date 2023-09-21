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
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/discovery"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/kubernetes"
)

var (
	l                     = logger.DefaultSLogger(inputName)
	getGlobalCustomerKeys = func() []string { return nil }
)

func getCollectorMeasurement() []inputs.Measurement {
	res := []inputs.Measurement{
		&containerMetric{},
		&containerObject{},
		&containerLog{},
	}
	res = append(res, kubernetes.Measurements()...)
	return res
}

func (i *Input) setup() {
	if i.DeprecatedDockerEndpoint != "" {
		i.Endpoints = append(i.Endpoints, i.DeprecatedDockerEndpoint)
	}
	if i.DeprecatedContainerdAddress != "" {
		i.Endpoints = append(i.Endpoints, "unix://"+i.DeprecatedContainerdAddress)
	}
	i.Endpoints = unique(i.Endpoints)
	l.Infof("endpoints: %v", i.Endpoints)

	getGlobalCustomerKeys = func() []string { return config.Cfg.Dataway.GlobalCustomerKeys }
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("container input started")
	i.setup()

	if datakit.Docker {
		i.startDiscovery()
	}

	i.runCollect()
}

func (i *Input) runCollect() {
	objectTick := time.NewTicker(objectInterval)
	defer objectTick.Stop()

	metricTick := time.NewTicker(metricInterval)
	defer metricTick.Stop()

	loggingInterval := metricInterval
	if i.LoggingSearchInterval > 0 {
		loggingInterval = i.LoggingSearchInterval
	}
	loggingTick := time.NewTicker(loggingInterval)
	defer loggingTick.Stop()

	collectors := i.newCollector()

	// first collect
	i.collectLogging(collectors)
	i.collectObject(collectors)
	i.collectMetric(collectors)

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("container exit")
			return

		case <-i.semStop.Wait():
			l.Info("container terminate")
			return

		case <-metricTick.C:
			i.collectMetric(collectors)

		case <-objectTick.C:
			i.collectObject(collectors)

		case <-loggingTick.C:
			i.collectLogging(collectors)

		case pause := <-i.chPause:
			i.pause.Store(pause)
		}
	}
}

func (i *Input) collectMetric(collectors []Collector) {
	for _, c := range collectors {
		if i.pause.Load() && c.Name() == kubernetes.Name() {
			continue
		}
		fn := func(pts []*point.Point) error {
			return i.Feeder.Feed(c.Name()+"-metric", point.Metric, pts, &io.Option{Blocking: true})
		}
		c.Metric(fn)
	}
}

func (i *Input) collectObject(collectors []Collector) {
	for _, c := range collectors {
		if i.pause.Load() && c.Name() == kubernetes.Name() {
			continue
		}
		fn := func(pts []*point.Point) error {
			return i.Feeder.Feed(c.Name()+"-object", point.Object, pts, &io.Option{Blocking: true})
		}
		c.Object(fn)
	}
}

func (i *Input) collectLogging(collectors []Collector) {
	for _, c := range collectors {
		fn := func(pts []*point.Point) error {
			return i.Feeder.Feed(c.Name()+"-logging", point.Logging, pts, &io.Option{Blocking: true})
		}
		c.Logging(fn)
	}
}

func (i *Input) startDiscovery() {
	discovery, err := newDiscovery(i)
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

func (i *Input) newCollector() []Collector {
	collectors := []Collector{}
	collectors = append(collectors, newCollectorsFromContainerEndpoints(i)...)

	if datakit.Docker {
		k8sCollectors, err := newCollectorsFromKubernetes(i)
		if err != nil {
			l.Errorf("init the k8s fail, err: %s", err)
		} else {
			collectors = append(collectors, k8sCollectors)
		}
	}

	return collectors
}

type Collector interface {
	Name() string
	Metric(func(pts []*point.Point) error)
	Object(func(pts []*point.Point) error)
	Logging(func(pts []*point.Point) error)
}

func newCollectorsFromContainerEndpoints(ipt *Input) []Collector {
	var collectors []Collector
	for _, endpoint := range ipt.Endpoints {
		if err := checkEndpoint(endpoint); err != nil {
			l.Warnf("%s, skip", err)
			continue
		}

		var client k8sclient.Client
		var err error

		if datakit.Docker {
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

	cfg := kubernetes.Config{
		NodeName:                    config.Cfg.Hostname,
		EnableK8sMetric:             ipt.EnableK8sMetric,
		EnableK8sObject:             true,
		EnablePodMetric:             ipt.EnablePodMetric,
		EnableK8sEvent:              ipt.EnableK8sEvent,
		EnableExtractK8sLabelAsTags: ipt.EnableExtractK8sLabelAsTags,
		ExtraTags:                   tags,
		GlobalCustomerKeys:          getGlobalCustomerKeys(),
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

	cfg := discovery.Config{
		EnablePrometheusPodAnnotations:     ipt.EnableAutoDiscoveryOfPrometheusPodAnnotations,
		EnablePrometheusServiceAnnotations: ipt.EnableAutoDiscoveryOfPrometheusServiceAnnotations,
		EnablePrometheusPodMonitors:        ipt.EnableAutoDiscoveryOfPrometheusPodMonitors,
		EnablePrometheusServiceMonitors:    ipt.EnableAutoDiscoveryOfPrometheusServiceMonitors,
		ExtraTags:                          tags,
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

func getMountPoint() string {
	if !datakit.Docker {
		return ""
	}
	if n := os.Getenv("HOST_ROOT"); n != "" {
		return n
	}
	return "/rootfs"
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
		return fmt.Errorf("endpoint cannot be a directory")
	}

	return nil
}
