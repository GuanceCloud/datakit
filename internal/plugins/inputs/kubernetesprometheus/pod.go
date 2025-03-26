// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Pod struct {
	role     Role
	informer infov1.PodInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scrape    scrapeManagerInterface
	feeder    dkio.Feeder
}

func NewPod(
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	scrape scrapeManagerInterface,
	feeder dkio.Feeder,
) (*Pod, error) {
	informer := informerFactory.Core().V1().Pods()
	if informer == nil {
		return nil, fmt.Errorf("cannot get pod informer")
	}

	return &Pod{
		role:     RolePod,
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RolePod)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scrape:    scrape,
		feeder:    feeder,
	}, nil
}

func (p *Pod) Run(ctx context.Context) {
	defer p.queue.ShutDown()

	p.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			p.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			p.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			p.enqueue(obj)
		},
	})

	managerGo.Go(func(_ context.Context) error {
		for p.process(ctx) {
		}
		return nil
	})

	<-ctx.Done()
}

func (p *Pod) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}

	p.queue.Add(key)
}

func (p *Pod) process(ctx context.Context) bool {
	keyObj, quit := p.queue.Get()
	if quit {
		return false
	}
	defer p.queue.Done(keyObj)
	key := keyObj.(string)

	obj, exists, err := p.store.GetByKey(key)
	if err != nil {
		return true
	}

	if !exists {
		klog.Infof("deleted Pod %s", key)
		p.terminateScrape(key)
		return true
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Warnf("converting to Pod object failed, %v", obj)
		return true
	}

	nodeName, exists := nodeNameFrom(ctx)
	if exists && pod.Spec.NodeName != nodeName {
		return true
	}

	if shouldSkipPod(pod) {
		return true
	}

	traits := podTraits(pod)
	if p.scrape.isTraitsExists(p.role, key, traits) {
		return true
	}

	klog.Infof("discovered Pod %s", key)
	p.terminateScrape(key)
	p.startScrape(ctx, key, traits, pod)
	return true
}

func (p *Pod) startScrape(ctx context.Context, key, traits string, item *corev1.Pod) {
	checkPausedFunc := func() bool {
		return checkPaused(ctx, false /* not use election */)
	}

	for _, ins := range p.instances {
		if ins.validator != nil && !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newPodParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Pod %s", key)

		cfg, err := pr.parsePromConfig(ins)
		if err != nil {
			klog.Warnf("pod %s has unexpected url, err %s", key, err)
			continue
		}

		opts := buildPromOptions(
			p.role, key,
			&ins.Auth,
			p.feeder,
			promscrape.WithHTTPHeader(ins.Headers),
			promscrape.WithMeasurement(cfg.measurement),
			promscrape.KeepExistMetricName(cfg.keepExistMetricName),
			promscrape.WithExtraTags(cfg.tags))

		prom, err := newPromScraper(p.role, key, cfg.urlstr, cfg.measurement, p.feeder, checkPausedFunc, opts)
		if err != nil {
			klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
			continue
		}

		p.scrape.registerScrape(p.role, key, traits, prom)
	}
}

func (p *Pod) terminateScrape(key string) {
	p.scrape.removeScrape(p.role, key)
}

func podTraits(item *corev1.Pod) string {
	return item.Status.HostIP + "/" + item.Status.PodIP
}

func shouldSkipPod(item *corev1.Pod) bool {
	return item.Status.PodIP == "" || item.Status.Phase != corev1.PodRunning
}
