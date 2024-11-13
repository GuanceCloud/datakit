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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	infov1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Service struct {
	role      Role
	clientset *kubernetes.Clientset
	informer  infov1.ServiceInformer
	queue     workqueue.DelayingInterface
	store     cache.Store

	instances  []*Instance
	svcTraits  map[string]string // It is safe from race conditions
	svcCancels map[string]context.CancelFunc
	scrape     scrapeManagerInterface
	feeder     dkio.Feeder
}

func NewService(
	clientset *kubernetes.Clientset,
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	scrape scrapeManagerInterface,
	feeder dkio.Feeder,
) (*Service, error) {
	informer := informerFactory.Core().V1().Services()
	if informer == nil {
		return nil, fmt.Errorf("cannot get service informer")
	}
	return &Service{
		role:      RoleService,
		clientset: clientset,
		informer:  informer,
		queue:     workqueue.NewNamedDelayingQueue(string(RoleService)),
		store:     informer.Informer().GetStore(),

		instances:  instances,
		svcTraits:  make(map[string]string),
		svcCancels: make(map[string]context.CancelFunc),
		scrape:     scrape,
		feeder:     feeder,
	}, nil
}

func (s *Service) Run(ctx context.Context) {
	defer s.queue.ShutDown()

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

	traits := serviceTraits(svc)
	if svcTraits, ok := s.svcTraits[key]; ok && svcTraits == traits {
		return true
	}

	klog.Infof("discovered Service %s", key)
	s.startScrape(ctx, key, traits, svc)
	return true
}

func (s *Service) startScrape(ctx context.Context, key, traits string, item *corev1.Service) {
	ctx, cancel := context.WithCancel(ctx)

	s.svcTraits[key] = traits
	s.svcCancels[key] = cancel

	for idx, ins := range s.instances {
		if !ins.validator.Matches(item.Namespace, item.Labels) {
			continue
		}

		pr := newServiceParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		idxKey := fmt.Sprintf("%s::index%d", key, idx)

		// record idxKey
		klog.Infof("added Service %s", idxKey)

		namespace := item.Namespace
		name := item.Name
		endpointsInstance := pr.transToEndpointsInstance(ins)

		managerGo.Go(func(_ context.Context) error {
			tick := time.NewTicker(time.Second * 20)
			defer tick.Stop()

			for {
				select {
				case <-ctx.Done():
					s.scrape.removeScrape(s.role, idxKey)
					klog.Infof("svc-ep %s exit", idxKey)
					return nil

				case <-tick.C:
					// Maybe the Service Name and Endpoint Name are not the same, so the Selector should be used here.
					ep, err := s.clientset.CoreV1().Endpoints(namespace).Get(context.Background(), name, metav1.GetOptions{})
					if err != nil {
						klog.Warnf("get endpoints fail %s", err)
					} else {
						tryCreateScrapeForEndpoints(ctx, s.role, idxKey, ep, endpointsInstance, s.scrape, s.feeder)
					}
				}
			}
		})
	}
}

func (s *Service) terminateScrape(key string) {
	delete(s.svcTraits, key)
	if cancel, ok := s.svcCancels[key]; ok {
		cancel()
	}
	delete(s.svcCancels, key)
	klog.Infof("%s for key %s was terminated", s.role, key)
}

func serviceTraits(item *corev1.Service) string {
	return item.Spec.ClusterIP
}

func shouldSkipService(_ *corev1.Service) bool {
	return false
}
