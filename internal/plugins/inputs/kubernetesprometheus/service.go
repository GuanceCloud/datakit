// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Service struct {
	role              Role
	informer          infov1.ServiceInformer
	endpointsInformer infov1.EndpointsInformer
	queue             workqueue.DelayingInterface
	store             cache.Store
	endpointsStore    cache.Store

	instances []*Instance
	svcTraits map[string]string // It is safe from race conditions
	scrape    scrapeManagerInterface
	feeder    dkio.Feeder
}

func NewService(
	_ *kubernetes.Clientset,
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	scrape scrapeManagerInterface,
	feeder dkio.Feeder,
) (*Service, error) {
	informer := informerFactory.Core().V1().Services()
	if informer == nil {
		return nil, fmt.Errorf("cannot get service informer")
	}
	endpointsInformer := informerFactory.Core().V1().Endpoints()
	if endpointsInformer == nil {
		return nil, fmt.Errorf("cannot get endpoints informer")
	}

	return &Service{
		role:              RoleService,
		informer:          informer,
		endpointsInformer: endpointsInformer,
		queue:             workqueue.NewNamedDelayingQueue(string(RoleService)),
		store:             informer.Informer().GetStore(),
		endpointsStore:    endpointsInformer.Informer().GetStore(),

		instances: instances,
		svcTraits: make(map[string]string),
		scrape:    scrape,
		feeder:    feeder,
	}, nil
}

func (s *Service) Run(ctx context.Context) {
	defer s.queue.ShutDown()

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			s.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			s.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			s.enqueue(obj)
		},
	}
	s.informer.Informer().AddEventHandler(handler)
	s.endpointsInformer.Informer().AddEventHandler(handler)

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
		s.terminateScrape(key)
		return true
	}

	obj, exists, err = s.endpointsStore.GetByKey(key)
	if err != nil {
		return true
	}
	if !exists {
		s.terminateScrape(key)
		return true
	}
	ep, ok := obj.(*corev1.Endpoints)
	if !ok {
		klog.Warnf("converting to Endpoints object failed, %v", obj)
		return true
	}
	if shouldSkipEndpoints(ep) {
		s.terminateScrape(key)
		return true
	}

	traits := serviceTraits(svc) + "::" + endpointsTraits(ep)
	if svcTraits, ok := s.svcTraits[key]; ok && svcTraits == traits {
		return true
	}

	klog.Infof("discovered Service %s", key)
	s.terminateScrape(key)
	s.startScrape(ctx, key, traits, svc, ep)
	return true
}

func (s *Service) startScrape(
	ctx context.Context,
	key, traits string,
	item *corev1.Service,
	ep *corev1.Endpoints,
) {
	s.svcTraits[key] = traits

	for idx, ins := range s.instances {
		if !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newServiceParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		idxKey := fmt.Sprintf("%s::index%d", key, idx)
		klog.Infof("added Service %s", idxKey)

		endpointsInstance := pr.transToEndpointsInstance(ins)
		tryCreateScrapeForEndpoints(ctx, s.role, idxKey, ep, endpointsInstance, s.scrape, s.feeder)
	}
}

func (s *Service) terminateScrape(key string) {
	delete(s.svcTraits, key)
	for idx := range s.instances {
		idxKey := fmt.Sprintf("%s::index%d", key, idx)
		s.scrape.removeScrape(s.role, idxKey)
	}
	klog.Infof("%s for key %s was terminated", s.role, key)
}

func serviceTraits(item *corev1.Service) string {
	return item.Spec.ClusterIP
}

func shouldSkipService(_ *corev1.Service) bool {
	return false
}
