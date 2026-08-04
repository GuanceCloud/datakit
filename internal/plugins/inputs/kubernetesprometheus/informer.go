// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

const informerListLimit int64 = 50

type informerFactories struct {
	general informers.SharedInformerFactory
	pod     informers.SharedInformerFactory

	podIsSeparate bool
}

func newInformerFactories(
	clientset *kubernetes.Clientset,
	nodeLocal bool,
	nodeName string,
	hasPodRole bool,
) *informerFactories {
	general := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0,
		informers.WithTweakListOptions(setInformerListLimit),
	)
	factories := &informerFactories{general: general}
	if !hasPodRole {
		return factories
	}

	if !nodeLocal {
		factories.pod = general
		return factories
	}

	factories.pod = informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			setPodInformerListOptions(options, nodeName)
		}),
	)
	factories.podIsSeparate = true
	return factories
}

func setInformerListLimit(options *metav1.ListOptions) {
	options.Limit = informerListLimit
}

func setPodInformerListOptions(options *metav1.ListOptions, nodeName string) {
	setInformerListLimit(options)
	options.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", nodeName).String()
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	f.general.Start(stopCh)
	if f.podIsSeparate {
		f.pod.Start(stopCh)
	}
}

func (f *informerFactories) WaitForCacheSync(stopCh <-chan struct{}) {
	f.general.WaitForCacheSync(stopCh)
	if f.podIsSeparate {
		f.pod.WaitForCacheSync(stopCh)
	}
}
