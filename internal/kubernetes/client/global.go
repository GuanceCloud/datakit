// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type APIClient struct {
	Clientset       *kubernetes.Clientset
	InformerFactory informers.SharedInformerFactory
}

func GetAPIClient() (*APIClient, error) {
	restConfig, err := DefaultConfigInCluster()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0,
		informers.WithTweakListOptions(
			func(v *v1.ListOptions) {
				v.Limit = 50
			},
		),
	)
	return &APIClient{
		Clientset:       clientset,
		InformerFactory: informerFactory,
	}, nil
}
