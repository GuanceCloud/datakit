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

type Node struct {
	informer infov1.NodeInformer
	queue    workqueue.DelayingInterface
	store    cache.Store

	instances []*Instance
	scraper   *scraper
	keys      map[string]string
	feeder    dkio.Feeder
}

func NewNode(informerFactory informers.SharedInformerFactory, instances []*Instance, feeder dkio.Feeder) (*Node, error) {
	informer := informerFactory.Core().V1().Nodes()
	if informer == nil {
		return nil, fmt.Errorf("cannot get node informer")
	}
	return &Node{
		informer: informer,
		queue:    workqueue.NewNamedDelayingQueue(string(RoleNode)),
		store:    informer.Informer().GetStore(),

		instances: instances,
		scraper:   newScraper(),
		keys:      make(map[string]string),
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

	go func() {
		for n.process(ctx) {
		}
	}()

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
		n.terminateProms(key)
		return true
	}

	node, ok := obj.(*corev1.Node)
	if !ok {
		klog.Warnf("converting to Node object failed, %v", err)
		return true
	}

	info, exists := n.keys[key]
	if exists && info == joinNodeInfo(node) {
		return true
	}

	klog.Infof("found node %s", key)

	n.terminateProms(key)
	n.runProm(ctx, key, node)
	return true
}

func (n *Node) runProm(ctx context.Context, key string, item *corev1.Node) {
	if shouldSkipNode(item) {
		return
	}

	for _, ins := range n.instances {
		if !ins.validator.Matches("", item.Labels) {
			continue
		}

		pr := newNodeParser(item)
		if !pr.shouldScrape(ins.Scrape) {
			continue
		}

		// record key
		klog.Infof("added Node %s", key)
		n.keys[key] = joinNodeInfo(item)

		cfg, err := pr.parsePromConfig(ins)
		if err != nil {
			klog.Warnf("node %s has unexpected url target %v", key, ins.Target)
			continue
		}

		interval := ins.Interval
		urlstr := cfg.urlstr

		opts := buildPromOptions(
			RoleNode, key, n.feeder,
			iprom.WithMeasurementName(cfg.measurement),
			iprom.WithTags(cfg.tags))

		if tlsOpts, err := buildPromOptionsWithAuth(&ins.Auth); err != nil {
			klog.Warnf("node %s has unexpected tls config %ss", key, err)
		} else {
			opts = append(opts, tlsOpts...)
		}

		workerInc(RoleNode, key)
		workerGo.Go(func(_ context.Context) error {
			defer func() {
				workerDec(RoleNode, key)
			}()
			n.scraper.runProm(ctx, key, urlstr, interval, opts)
			return nil
		})
	}
}

func (n *Node) terminateProms(key string) {
	n.scraper.terminateProms(key)
	delete(n.keys, key)
}

func joinNodeInfo(item *corev1.Node) string {
	internalIP := ""
	for _, address := range item.Status.Addresses {
		if address.Type == corev1.NodeInternalIP {
			internalIP = address.Address
		}
	}
	return internalIP
}

func shouldSkipNode(item *corev1.Node) bool {
	if maxedOutClients() {
		return true
	}
	for _, address := range item.Status.Addresses {
		if address.Type == corev1.NodeInternalIP && address.Address == "" {
			return true
		}
	}
	return false
}
