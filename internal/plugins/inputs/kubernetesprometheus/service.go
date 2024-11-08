// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Service struct {
	clientset *kubernetes.Clientset
	informer  infov1.ServiceInformer
	queue     workqueue.DelayingInterface
	store     cache.Store

	instances []*Instance
	svcList   map[string]string // It is safe from race conditions
	scrape    *scrapeManager
	feeder    dkio.Feeder
}

func NewService(
	clientset *kubernetes.Clientset,
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	feeder dkio.Feeder,
) (*Service, error) {
	informer := informerFactory.Core().V1().Services()
	if informer == nil {
		return nil, fmt.Errorf("cannot get service informer")
	}
	return &Service{
		clientset: clientset,
		informer:  informer,
		queue:     workqueue.NewNamedDelayingQueue(string(RoleService)),
		store:     informer.Informer().GetStore(),

		instances: instances,
		svcList:   make(map[string]string),
		scrape:    newScrapeManager(RoleService),
		feeder:    feeder,
	}, nil
}

func (s *Service) Run(ctx context.Context) {
	defer s.queue.ShutDown()

	s.scrape.run(ctx, maxConcurrent(nodeLocalFrom(ctx)))

	s.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			s.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			s.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			s.enqueue(obj)
		},
	})

	managerGo.Go(func(_ context.Context) error {
		for s.process(ctx) {
		}
		return nil
	})

	<-ctx.Done()
}

func (s *Service) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}

	s.queue.Add(key)
}

func (s *Service) process(ctx context.Context) bool {
	keyObj, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(keyObj)
	key := keyObj.(string)

	obj, exists, err := s.store.GetByKey(key)
	if err != nil {
		return true
	}

	if !exists {
		klog.Infof("deleted Service %s", key)
		s.terminateScrape(key)
		return true
	}

	svc, ok := obj.(*corev1.Service)
	if !ok {
		klog.Warnf("converting to Service object failed, %v", obj)
		return true
	}

	if shouldSkipService(svc) {
		return true
	}

	if feature, ok := s.svcList[key]; ok && feature == serviceFeature(svc) {
		return true
	}

	klog.Infof("found new service %s", key)
	s.terminateScrape(key)
	s.startScrape(ctx, key, svc)
	return true
}

func (s *Service) startScrape(ctx context.Context, key string, item *corev1.Service) {
	svcFeature := serviceFeature(item)

	for _, ins := range s.instances {
		if !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newServiceParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Service %s", key)
		s.svcList[key] = svcFeature

		namespace := item.Namespace
		name := item.Name
		endpointsInstance := pr.transToEndpointsInstance(ins)

		managerGo.Go(func(_ context.Context) error {
			tick := time.NewTicker(time.Second * 20)
			defer tick.Stop()

			for {
				s.tryCreateScrapeForEndpoints(ctx, namespace, name, key, endpointsInstance)

				select {
				case <-ctx.Done():
					klog.Info("svc-ep exit")
					return nil

				case <-tick.C:
					// next
				}
			}
		})
	}
}

func (s *Service) tryCreateScrapeForEndpoints(ctx context.Context, namespace, name string, key string, endpointsInstance *Instance) {
	// Maybe the Service Name and Endpoint Name are not the same, so the Selector should be used here.
	ep, err := s.clientset.CoreV1().Endpoints(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		klog.Warn("get endpoints fail %s", err)
		return
	}

	endpointsFeature := endpointsFeature(ep)
	if s.scrape.matchesKey(key, endpointsFeature) {
		// no change
		return
	}

	s.scrape.terminateScrape(key)

	nodeName, nodeNameExist := nodeNameFrom(ctx)

	pr := newEndpointsParser(ep)
	cfgs, err := pr.parsePromConfig(endpointsInstance)
	if err != nil {
		klog.Warnf("svc-ep %s has unexpected url, err %s", key, err)
		return
	}

	for _, cfg := range cfgs {
		if s.scrape.existScrape(key, cfg.urlstr) {
			continue
		}

		if nodeNameExist && cfg.nodeName != "" && cfg.nodeName != nodeName {
			continue
		}

		opts := buildPromOptions(
			RoleService, key, s.feeder,
			promscrape.WithMeasurement(cfg.measurement),
			promscrape.WithExtraTags(cfg.tags))

		if tlsOpts, err := buildPromOptionsWithAuth(&endpointsInstance.Auth); err != nil {
			klog.Warnf("svc-ep %s has unexpected tls config %s", key, err)
		} else {
			opts = append(opts, tlsOpts...)
		}

		checkPausedFunc := func() bool {
			return checkPaused(ctx, cfg.nodeName == "")
		}

		prom, err := newPromScraper(RoleService, key, cfg.urlstr, checkPausedFunc, opts)
		if err != nil {
			klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
			continue
		}

		s.scrape.registerScrape(key, endpointsFeature, prom)
	}

	urlstrList := getURLstrListByPromConfigs(cfgs)
	s.scrape.tryCleanScrapes(key, urlstrList)
}

func (s *Service) terminateScrape(key string) {
	delete(s.svcList, key)
	s.scrape.terminateScrape(key)
}

func serviceFeature(item *corev1.Service) string {
	return item.Spec.ClusterIP
}

func shouldSkipService(_ *corev1.Service) bool {
	return false
}
