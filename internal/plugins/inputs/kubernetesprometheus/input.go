// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package kubernetesprometheus wraps prometheus collect on kubernetes.
package kubernetesprometheus

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
)

type Input struct {
	NodeLocal                                     bool              `toml:"node_local"`
	ScrapeInterval                                time.Duration     `toml:"scrape_interval"`
	KeepExistMetricName                           bool              `toml:"keep_exist_metric_name"`
	HonorTimestamps                               bool              `toml:"honor_timestamps"`
	EnableDiscoveryOfPrometheusPodAnnotations     bool              `toml:"enable_discovery_of_prometheus_pod_annotations"`
	EnableDiscoveryOfPrometheusServiceAnnotations bool              `toml:"enable_discovery_of_prometheus_service_annotations"`
	EnableDiscoveryOfPrometheusPodMonitors        bool              `toml:"enable_discovery_of_prometheus_pod_monitors"`
	EnableDiscoveryOfPrometheusServiceMonitors    bool              `toml:"enable_discovery_of_prometheus_service_monitors"`
	GlobalTags                                    map[string]string `toml:"global_tags"`
	InstanceManager

	nodeName string
	chPause  chan bool
	pause    *atomic.Bool
	feeder   dkio.Feeder
	cancel   context.CancelFunc

	semStop *cliutils.Sem // start stop signal
	runOnce sync.Once
}

func (*Input) SampleConfig() string                    { return sampleConfig }
func (*Input) Catalog() string                         { return inputName }
func (*Input) AvailableArchs() []string                { return []string{datakit.LabelK8s} }
func (*Input) Singleton()                              { /*nil*/ }
func (*Input) SampleMeasurement() []inputs.Measurement { return nil /* no measurement docs exported */ }

func (ipt *Input) Run() {
	klog = logger.SLogger("kubernetesprometheus")
	if err := ipt.setup(); err != nil {
		klog.Warnf("failed to setup error %s, exit", err)
		return
	}

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for {
		// enable nodeLocal model or election success
		if ipt.NodeLocal || !ipt.pause.Load() {
			ipt.runOnce.Do(
				func() {
					managerGo.Go(func(_ context.Context) error {
						if err := ipt.start(); err != nil {
							klog.Warn(err)
						}
						return nil
					})
				})
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.stop()
			klog.Info("exit")
			return

		case <-ipt.semStop.Wait():
			ipt.stop()
			klog.Info("exit")
			return

		case pause := <-ipt.chPause:
			ipt.pause.Store(pause)

			// disable nodeLocal model and election defeat
			if !ipt.NodeLocal && pause {
				ipt.stop()
				ipt.runOnce = sync.Once{} // reset runOnce
			}

		case <-tick.C:
			// next
		}
	}
}

func (ipt *Input) setup() error {
	ipt.ScrapeInterval = config.ProtectedInterval(time.Second, time.Minute*5, ipt.ScrapeInterval)

	if str := os.Getenv("ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_ANNOTATIONS"); isTrue(str) {
		ipt.EnableDiscoveryOfPrometheusPodAnnotations = true
		klog.Info("enable pod annotations")
	}
	if str := os.Getenv("ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_ANNOTATIONS"); isTrue(str) {
		ipt.EnableDiscoveryOfPrometheusServiceAnnotations = true
		klog.Info("enable service annotations")
	}
	if str := os.Getenv("ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_POD_MONITORS"); isTrue(str) {
		ipt.EnableDiscoveryOfPrometheusPodMonitors = true
		klog.Info("enable pod monitor")
	}
	if str := os.Getenv("ENV_INPUT_CONTAINER_ENABLE_AUTO_DISCOVERY_OF_PROMETHEUS_SERVICE_MONITORS"); isTrue(str) {
		ipt.EnableDiscoveryOfPrometheusServiceMonitors = true
		klog.Info("enable service monitor")
	}

	for _, ins := range ipt.Instances {
		// set default values
		ins.setDefault(ipt)
	}

	ipt.applyPredefinedInstances()

	var err error
	ipt.nodeName, err = getLocalNodeName()

	return err
}

func (ipt *Input) start() error {
	klog.Info("start")

	ctx, cancel := context.WithCancel(context.Background())
	ipt.cancel = cancel

	ctx = withPause(ctx, ipt.pause)
	if ipt.NodeLocal {
		ctx = withNodeName(ctx, ipt.nodeName)
		ctx = withNodeLocal(ctx, ipt.NodeLocal)
	}

	client, err := k8sclient.NewKubernetesClientInCluster()
	if err != nil {
		return err
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		client.KubernetesClientset(), 0,
		informers.WithTweakListOptions(func(v *metav1.ListOptions) { v.Limit = 50 }),
	)

	scrapeManager := newScrapeManager()
	scrapeManager.runWorker(ctx, maxConcurrent(nodeLocalFrom(ctx))*2, ipt.ScrapeInterval)

	ipt.applyCRDs(ctx, client, scrapeManager)

	ipt.InstanceManager.Run(ctx, client.KubernetesClientset(), informerFactory, scrapeManager, ipt.feeder)
	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	<-ctx.Done()
	klog.Info("end")
	return nil
}

func (ipt *Input) stop() {
	if ipt.cancel != nil {
		ipt.cancel()
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func newPauseVar() *atomic.Bool {
	b := &atomic.Bool{}
	b.Store(true)
	return b
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			NodeLocal:           true,
			ScrapeInterval:      time.Second * 30,
			KeepExistMetricName: true,
			HonorTimestamps:     true,
			chPause:             make(chan bool, inputs.ElectionPauseChannelLength),
			pause:               newPauseVar(),
			feeder:              dkio.DefaultFeeder(),
			semStop:             cliutils.NewSem(),
		}
	})
}
