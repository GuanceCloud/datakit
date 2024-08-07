// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Pod struct {
	informer infov1.PodInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scraper   *scraper
	keys      map[string]string
	feeder    dkio.Feeder
}

func NewPod(informerFactory informers.SharedInformerFactory, instances []*Instance, feeder dkio.Feeder) (*Pod, error) {
	informer := informerFactory.Core().V1().Pods()
	if informer == nil {
		return nil, fmt.Errorf("cannot get pod informer")
	}
	return &Pod{
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RolePod)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scraper:   newScraper(),
		keys:      make(map[string]string),
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

	go func() {
		for p.process(ctx) {
		}
	}()

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
		p.terminateProms(key)
		return true
	}

	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Warnf("converting to Pod object failed, %v", obj)
		return true
	}

	info, exists := p.keys[key]
	if exists && info == joinPodInfo(pod) {
		return true
	}

	klog.Infof("found pod %s", key)

	p.terminateProms(key)
	p.runProm(ctx, key, pod)
	return true
}

func (p *Pod) runProm(ctx context.Context, key string, item *corev1.Pod) {
	if shouldSkipPod(item) {
		return
	}

	for _, ins := range p.instances {
		if !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newPodParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Pod %s", key)
		p.keys[key] = joinPodInfo(item)

		cfg, err := pr.parsePromConfig(ins)
		if err != nil {
			klog.Warnf("pod %s has unexpected url, err %s", key, err)
			continue
		}

		interval := ins.Interval
		urlstr := cfg.urlstr

		opts := buildPromOptions(
			RolePod, key, p.feeder,
			iprom.WithMeasurementName(cfg.measurement),
			iprom.WithTags(cfg.tags))

		if tlsOpts, err := buildPromOptionsWithAuth(&ins.Auth); err != nil {
			klog.Warnf("pod %s has unexpected tls config %s", key, err)
		} else {
			opts = append(opts, tlsOpts...)
		}

		workerInc(RolePod, key)
		workerGo.Go(func(_ context.Context) error {
			defer func() {
				workerInc(RolePod, key)
			}()

			p.scraper.runProm(ctx, key, urlstr, interval, opts)
			return nil
		})
	}
}

func (p *Pod) terminateProms(key string) {
	p.scraper.terminateProms(key)
	delete(p.keys, key)
}

func joinPodInfo(item *corev1.Pod) string {
	return item.Status.HostIP + ":" + item.Status.PodIP
}

func shouldSkipPod(item *corev1.Pod) bool {
	return maxedOutClients() || item.Status.PodIP == "" || item.Status.Phase != corev1.PodRunning
}
