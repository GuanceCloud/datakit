// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package discovery collect prom metric from kubernetes.
package discovery

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var klog = logger.DefaultSLogger("k8s-discovery")

type Config struct {
	ExtraTags   map[string]string
	LabelAsTags []string
	Feeder      io.Feeder
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
	return d.newPromFromPodAnnotationExport()
}

const allNamespaces = ""

func (d *Discovery) newPromFromPodAnnotationExport() []*promRunner {
	var res []*promRunner

	pods := d.getLocalPodsFromLabelSelector("pod-export-config", allNamespaces, nil)

	for _, pod := range pods {
		cfgStr, ok := pod.Annotations[annotationPromExport]
		if !ok {
			continue
		}

		runners, err := newPromRunnersForPod(d, pod, cfgStr)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s export-config, err: %s, skip", pod.Name, err)
			continue
		}

		for _, runner := range runners {
			if runner.conf == nil {
				continue
			}
			withTag("namespace", pod.Namespace)(runner.conf)
			withTag("pod_name", pod.Name)(runner.conf)
			withLabelAsTags(pod.Labels, d.cfg.LabelAsTags)(runner.conf)

			klog.Infof("created prom runner of pod-export-config %s, urls %s", pod.Name, runner.conf.URLs)
			res = append(res, runner)
		}
	}

	return res
}

func (d *Discovery) getLocalPodsFromLabelSelector(source, namespace string, selector *metav1.LabelSelector) (res []*apicorev1.Pod) {
	opt := metav1.ListOptions{
		ResourceVersion: "0",
		FieldSelector:   "spec.nodeName=" + d.localNodeName,
	}
	if selector != nil {
		opt.LabelSelector = newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String()
	}

	list, err := d.client.GetPods(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods from namespace '%s' by %s, err: %s, skip", namespace, source, err)
		return
	}
	for idx := range list.Items {
		if list.Items[idx].Status.Phase == apicorev1.PodRunning {
			res = append(res, &list.Items[idx])
		}
	}
	return
}

func newPromRunnersForPod(discovery *Discovery, pod *apicorev1.Pod, inputConfig string) ([]*promRunner, error) {
	cfg := completePromConfig(inputConfig, pod)
	return newPromRunnerWithTomlConfig(discovery, cfg)
}
