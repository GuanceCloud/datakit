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
	role     Role
	informer infov1.EndpointsInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scrape    scrapeManagerInterface
	feeder    dkio.Feeder
}

func NewEndpoints(
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	scrape scrapeManagerInterface,
	feeder dkio.Feeder,
) (*Endpoints, error) {
	informer := informerFactory.Core().V1().Endpoints()
	if informer == nil {
		return nil, fmt.Errorf("cannot get endpoints informer")
	}

	return &Endpoints{
		role:     RoleEndpoints,
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RoleEndpoints)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scrape:    scrape,
		feeder:    feeder,
	}, nil
}

func (e *Endpoints) Run(ctx context.Context) {
	defer e.queue.ShutDown()

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

	traits := endpointsTraits(ep)
	if e.scrape.isTraitsExists(e.role, key, traits) {
		return true
	}

	klog.Infof("discovered Endpoints %s", key)
	e.startScrape(ctx, key, traits, ep)
	return true
}

func (e *Endpoints) startScrape(ctx context.Context, key, traits string, item *corev1.Endpoints) {
	nodeName, nodeNameExists := nodeNameFrom(ctx)
	urlstrList := []string{}

	for _, ins := range e.instances {
		if ins.validator != nil && !ins.validator.Matches(item.Namespace, item.Labels) {
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

		for _, cfg := range cfgs {
			if e.scrape.isScrapeExists(e.role, key, cfg.urlstr) {
				continue
			}

			if nodeNameExists && cfg.nodeName != "" && cfg.nodeName != nodeName {
				continue
			}

			opts := buildPromOptions(
				e.role, key,
				&ins.Auth,
				e.feeder,
				promscrape.WithHTTPHeader(ins.Headers),
				promscrape.WithMeasurement(cfg.measurement),
				promscrape.KeepExistMetricName(cfg.keepExistMetricName),
				promscrape.WithExtraTags(cfg.tags))

			checkPausedFunc := func() bool {
				return checkPaused(ctx, cfg.nodeName == "")
			}

			prom, err := newPromScraper(RoleEndpoints, key, cfg.urlstr, checkPausedFunc, opts)
			if err != nil {
				klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
				continue
			}

			e.scrape.registerScrape(e.role, key, traits, prom)
		}

		urlstrList = append(urlstrList, getURLstrListByPromConfigs(cfgs)...)
	}

	// clean urls
	e.scrape.tryCleanScrapes(e.role, key, urlstrList)
}

func (e *Endpoints) terminateScrape(key string) {
	e.scrape.removeScrape(e.role, key)
}

func endpointsTraits(item *corev1.Endpoints) string {
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

func tryCreateScrapeForEndpoints(
	ctx context.Context,
	role Role,
	key string,
	ep *corev1.Endpoints,
	endpointsInstance *Instance,
	scrapeManager scrapeManagerInterface,
	feeder dkio.Feeder,
) {
	epTraits := endpointsTraits(ep)
	if scrapeManager.isTraitsExists(role, key, epTraits) {
		return
	}

	nodeName, nodeNameExist := nodeNameFrom(ctx)

	pr := newEndpointsParser(ep)
	cfgs, err := pr.parsePromConfig(endpointsInstance)
	if err != nil {
		klog.Warnf("endpoints %s has unexpected url, err %s", key, err)
		return
	}

	for _, cfg := range cfgs {
		if scrapeManager.isScrapeExists(role, key, cfg.urlstr) {
			continue
		}

		if nodeNameExist && cfg.nodeName != "" && cfg.nodeName != nodeName {
			continue
		}

		opts := buildPromOptions(
			role, key,
			&endpointsInstance.Auth,
			feeder,
			promscrape.WithHTTPHeader(cfg.headers),
			promscrape.WithMeasurement(cfg.measurement),
			promscrape.KeepExistMetricName(cfg.keepExistMetricName),
			promscrape.WithExtraTags(cfg.tags))

		checkPausedFunc := func() bool {
			return checkPaused(ctx, cfg.nodeName == "")
		}

		prom, err := newPromScraper(role, key, cfg.urlstr, checkPausedFunc, opts)
		if err != nil {
			klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
			continue
		}

		scrapeManager.registerScrape(role, key, epTraits, prom)
	}

	urlstrList := getURLstrListByPromConfigs(cfgs)
	scrapeManager.tryCleanScrapes(role, key, urlstrList)
}
