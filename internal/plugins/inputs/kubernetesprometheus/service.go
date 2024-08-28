// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
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
	scrape    *scrapeWorker
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
		scrape:    newScrapeWorker(RoleService),
		feeder:    feeder,
	}, nil
}

func (s *Service) Run(ctx context.Context) {
	defer s.queue.ShutDown()

	s.scrape.startWorker(ctx, maxConcurrent(nodeLocalFrom(ctx)))

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

	if s.scrape.matchesKey(key, serviceFeature(svc)) {
		return true
	}

	klog.Infof("found service %s", key)

	s.terminateScrape(key)
	s.startScrape(ctx, key, svc)
	return true
}

func (s *Service) startScrape(ctx context.Context, key string, item *corev1.Service) {
	if shouldSkipService(item) {
		return
	}

	nodeName, nodeNameExist := nodeNameFrom(ctx)
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
		s.scrape.registerKey(key, svcFeature)

		epIns := pr.transToEndpointsInstance(ins)
		interval := ins.Interval

		managerGo.Go(func(_ context.Context) error {
			tick := time.NewTicker(time.Second * 20)
			defer tick.Stop()

			var oldEndpointsFeature string

			for {
				select {
				case <-datakit.Exit.Wait():
					klog.Info("svc-ep prom exit")
					return nil

				case <-ctx.Done():
					klog.Info("svc-ep return")
					return nil

				case <-tick.C:
					// next
				}

				// Maybe the Service Name and Endpoint Name are not the same, so the Selector should be used here.
				ep, err := s.clientset.CoreV1().Endpoints(item.Namespace).Get(context.Background(), item.Name, metav1.GetOptions{})
				if err != nil {
					klog.Warn("get endpoints fail %s", err)
					continue
				}

				newEndpointsFeature := endpointsFeature(ep)
				if newEndpointsFeature == oldEndpointsFeature {
					// no change
					continue
				}
				// reset oldEndpointsFeature
				oldEndpointsFeature = newEndpointsFeature
				s.scrape.terminate(key)
				s.scrape.registerKey(key, svcFeature)

				pr := newEndpointsParser(ep)
				cfgs, err := pr.parsePromConfig(epIns)
				if err != nil {
					klog.Warnf("svc-ep %s has unexpected url, err %s", key, err)
					continue
				}

				for _, cfg := range cfgs {
					if nodeNameExist && cfg.nodeName != "" && cfg.nodeName != nodeName {
						continue
					}

					opts := buildPromOptions(
						RoleService, key, s.feeder,
						iprom.WithMeasurementName(cfg.measurement),
						iprom.WithTags(cfg.tags))

					if tlsOpts, err := buildPromOptionsWithAuth(&epIns.Auth); err != nil {
						klog.Warnf("svc-ep %s has unexpected tls config %s", key, err)
					} else {
						opts = append(opts, tlsOpts...)
					}

					urlstr := cfg.urlstr
					election := cfg.nodeName == ""

					prom, err := newPromTarget(ctx, urlstr, interval, election, opts)
					if err != nil {
						klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
						continue
					}

					s.scrape.registerTarget(key, prom)
				}
			}
		})
	}
}

func (s *Service) terminateScrape(key string) {
	s.scrape.terminate(key)
}

func serviceFeature(item *corev1.Service) string {
	return item.Spec.ClusterIP
}

func shouldSkipService(_ *corev1.Service) bool {
	return false
}
