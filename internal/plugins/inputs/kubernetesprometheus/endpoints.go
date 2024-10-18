// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Endpoints struct {
	informer infov1.EndpointsInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scrape    *scrapeManager
	feeder    dkio.Feeder
}

func NewEndpoints(informerFactory informers.SharedInformerFactory, instances []*Instance, feeder dkio.Feeder) (*Endpoints, error) {
	informer := informerFactory.Core().V1().Endpoints()
	if informer == nil {
		return nil, fmt.Errorf("cannot get endpoints informer")
	}
	return &Endpoints{
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RoleEndpoints)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scrape:    newScrapeManager(RoleEndpoints),
		feeder:    feeder,
	}, nil
}

func (e *Endpoints) Run(ctx context.Context) {
	defer e.queue.ShutDown()

	e.scrape.run(ctx, maxConcurrent(nodeLocalFrom(ctx)))

	e.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			e.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			e.enqueue(obj)
		},
	})

	managerGo.Go(func(_ context.Context) error {
		for e.process(ctx) {
		}
		return nil
	})

	<-ctx.Done()
}

func (e *Endpoints) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}

	e.queue.Add(key)
}

func (e *Endpoints) process(ctx context.Context) bool {
	keyObj, quit := e.queue.Get()
	if quit {
		return false
	}
	defer e.queue.Done(keyObj)
	key := keyObj.(string)

	obj, exists, err := e.store.GetByKey(key)
	if err != nil {
		return true
	}

	if !exists {
		klog.Infof("deleted Endpoints %s", key)
		e.terminateScrape(key)
		return true
	}

	ep, ok := obj.(*corev1.Endpoints)
	if !ok {
		klog.Warnf("converting to Endpoints object failed, %v", obj)
		return true
	}

	if shouldSkipEndpoints(ep) {
		return true
	}

	if e.scrape.matchesKey(key, endpointsFeature(ep)) {
		return true
	}

	klog.Infof("found new endpoints %s", key)
	e.terminateScrape(key)
	e.startScrape(ctx, key, ep)
	return true
}

func (e *Endpoints) startScrape(ctx context.Context, key string, item *corev1.Endpoints) {
	nodeName, nodeNameExists := nodeNameFrom(ctx)
	feature := endpointsFeature(item)

	for _, ins := range e.instances {
		if !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newEndpointsParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Endpoints %s", key)

		cfgs, err := pr.parsePromConfig(ins)
		if err != nil {
			klog.Warnf("endpoints %s has unexpected url, err %s", key, err)
			continue
		}
		interval := ins.Interval

		for _, cfg := range cfgs {
			if nodeNameExists && cfg.nodeName != "" && cfg.nodeName != nodeName {
				continue
			}

			opts := buildPromOptions(
				RoleEndpoints, key, e.feeder,
				promscrape.WithMeasurement(cfg.measurement),
				promscrape.WithExtraTags(cfg.tags))

			if tlsOpts, err := buildPromOptionsWithAuth(&ins.Auth); err != nil {
				klog.Warnf("endpoints %s has unexpected tls config %s", key, err)
			} else {
				opts = append(opts, tlsOpts...)
			}

			urlstr := cfg.urlstr
			checkPausedFunc := func() bool {
				return checkPaused(ctx, cfg.nodeName == "")
			}

			prom, err := newPromScraper(RoleEndpoints, key, urlstr, interval, checkPausedFunc, opts)
			if err != nil {
				klog.Warnf("fail new prom %s for %s", urlstr, err)
				continue
			}

			e.scrape.registerScrape(key, feature, prom)
		}
	}
}

func (e *Endpoints) terminateScrape(key string) {
	e.scrape.terminateScrape(key)
}

func endpointsFeature(item *corev1.Endpoints) string {
	var ips []string
	for _, sub := range item.Subsets {
		for _, address := range sub.Addresses {
			ips = append(ips, address.IP)
		}
	}
	return strconv.Itoa(len(ips)) + "::" + strings.Join(ips, ",")
}

func shouldSkipEndpoints(item *corev1.Endpoints) bool {
	return len(item.Subsets) == 0 || len(item.Subsets[0].Addresses) == 0
}
