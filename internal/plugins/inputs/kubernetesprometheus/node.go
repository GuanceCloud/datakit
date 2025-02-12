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

type Node struct {
	role     Role
	informer infov1.NodeInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scrape    scrapeManagerInterface
	feeder    dkio.Feeder
}

func NewNode(
	informerFactory informers.SharedInformerFactory,
	instances []*Instance,
	scrape scrapeManagerInterface,
	feeder dkio.Feeder,
) (*Node, error) {
	informer := informerFactory.Core().V1().Nodes()
	if informer == nil {
		return nil, fmt.Errorf("cannot get node informer")
	}

	return &Node{
		role:     RoleNode,
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RoleNode)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scrape:    scrape,
		feeder:    feeder,
	}, nil
}

func (n *Node) Run(ctx context.Context) {
	defer n.queue.ShutDown()

	n.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			n.enqueue(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			n.enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			n.enqueue(obj)
		},
	})

	managerGo.Go(func(_ context.Context) error {
		for n.process(ctx) {
		}
		return nil
	})

	<-ctx.Done()
}

func (n *Node) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}

	n.queue.Add(key)
}

func (n *Node) process(ctx context.Context) bool {
	keyObj, quit := n.queue.Get()
	if quit {
		return false
	}
	defer n.queue.Done(keyObj)
	key := keyObj.(string)

	obj, exists, err := n.store.GetByKey(key)
	if err != nil {
		return true
	}

	if !exists {
		klog.Infof("deleted Node %s", key)
		n.terminateScrape(key)
		return true
	}

	node, ok := obj.(*corev1.Node)
	if !ok {
		klog.Warnf("converting to Node object failed, %v", err)
		return true
	}

	nodeName, exists := nodeNameFrom(ctx)
	if exists && node.Name != nodeName {
		return true
	}

	if shouldSkipNode(node) {
		return true
	}

	traits := nodeTraits(node)
	if n.scrape.isTraitsExists(n.role, key, traits) {
		return true
	}

	klog.Infof("discovered Node %s", key)
	n.terminateScrape(key)
	n.startScrape(ctx, key, traits, node)
	return true
}

func (n *Node) startScrape(ctx context.Context, key, traits string, item *corev1.Node) {
	checkPausedFunc := func() bool {
		return checkPaused(ctx, false /* not use election */)
	}

	for _, ins := range n.instances {
		if ins.validator != nil && !ins.validator.Matches("", item.Labels) {
			continue
		}

		pr := newNodeParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Node %s", key)

		cfg, err := pr.parsePromConfig(ins)
		if err != nil {
			klog.Warnf("node %s has unexpected url, err %s", key, err)
			continue
		}

		opts := buildPromOptions(
			n.role, key,
			&ins.Auth,
			n.feeder,
			promscrape.WithHTTPHeader(ins.Headers),
			promscrape.WithMeasurement(cfg.measurement),
			promscrape.KeepExistMetricName(cfg.keepExistMetricName),
			promscrape.WithExtraTags(cfg.tags))

		prom, err := newPromScraper(RoleNode, key, cfg.urlstr, checkPausedFunc, opts)
		if err != nil {
			klog.Warnf("fail new prom %s for %s", cfg.urlstr, err)
			continue
		}

		n.scrape.registerScrape(n.role, key, traits, prom)
	}
}

func (n *Node) terminateScrape(key string) {
	n.scrape.removeScrape(n.role, key)
}

func nodeTraits(item *corev1.Node) string {
	internalIP := ""
	for _, address := range item.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			internalIP = address.Address
		}
	}
	return internalIP
}

func shouldSkipNode(item *corev1.Node) bool {
	for _, address := range item.Status.Addresses {
		if address.Type == corev1.NodeInternalIP && address.Address == "" {
			return true
		}
	}
	return false
}
