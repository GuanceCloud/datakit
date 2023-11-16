// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package discovery collect prom metric from kubernetes.
package discovery

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
)

var klog = logger.DefaultSLogger("k8s-discovery")

type Config struct {
	EnablePrometheusPodAnnotations     bool
	EnablePrometheusServiceAnnotations bool
	EnablePrometheusPodMonitors        bool
	EnablePrometheusServiceMonitors    bool
	StreamSize                         int
	ExtraTags                          map[string]string
	CustomerKeys                       []string
	Feeder                             io.Feeder
}

type Discovery struct {
	client        client.Client
	cfg           *Config
	localNodeName string

	done <-chan interface{}
}

func NewDiscovery(client client.Client, cfg *Config, done <-chan interface{}) *Discovery {
	return &Discovery{
		client: client,
		cfg:    cfg,
		done:   done,
	}
}

func (d *Discovery) Run() {
	klog = logger.SLogger("k8s-discovery")
	klog.Info("start")

	if d.cfg.StreamSize == 0 {
		d.cfg.StreamSize = 10
	}
	klog.Infof("stream size: %d", d.cfg.StreamSize)

	d.start()
}

func (d *Discovery) start() {
	if d.client == nil {
		klog.Info("unreachable, invalid k8s client, exit")
		return
	}

	localNodeName, err := getLocalNodeName()
	if err != nil {
		klog.Infof("unable to get node name, err: %s, exit", err)
		return
	}

	d.localNodeName = localNodeName
	klog.Infof("node name is %s", localNodeName)

	updateTicker := time.NewTicker(time.Minute * 3)
	defer updateTicker.Stop()

	collectTicker := time.NewTicker(time.Second * 1)
	defer collectTicker.Stop()

	runners := d.getRunners()

	for {
		for _, r := range runners {
			r.runOnce()
		}

		select {
		case <-datakit.Exit.Wait():
			klog.Info("exit")
			return

		case <-d.done:
			klog.Info("terminated")
			return

		case <-updateTicker.C:
			runners = d.getRunners()

		case <-collectTicker.C:
			// nil
		}
	}
}

func (d *Discovery) getRunners() []*promRunner {
	runners := []*promRunner{}
	runners = append(runners, d.newPromFromPodAnnotationExport()...)
	runners = append(runners, d.newPromFromDatakitCRD()...)

	if d.cfg.EnablePrometheusPodAnnotations {
		runners = append(runners, d.newPromFromPodAnnotations()...)
	}
	if d.cfg.EnablePrometheusServiceAnnotations {
		runners = append(runners, d.newPromFromServiceAnnotations()...)
	}

	if d.cfg.EnablePrometheusPodMonitors {
		runners = append(runners, d.newPromForPodMonitors()...)
	}
	if d.cfg.EnablePrometheusServiceMonitors {
		runners = append(runners, d.newPromForServiceMonitors()...)
	}

	klog.Infof("fetch prom input list, len %d", len(runners))
	return runners
}
