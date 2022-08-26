// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps Datakit resource by kubernetes client-gen.
package v1beta1

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// GuanceV1Client is used to interact with features provided by the guance.com group.
type GuanceV1Client struct {
	restClient rest.Interface
}

func (c *GuanceV1Client) Datakits(namespace string) DatakitInterface {
	return newDatakits(c, namespace)
}

// NewForConfig creates a new GuanceV1Client for the given config.
func NewForConfig(c *rest.Config) (*GuanceV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &GuanceV1Client{client}, nil
}

// NewForConfigOrDie creates a new GuanceV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *GuanceV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new GuanceV1Client for the given RESTClient.
func New(c rest.Interface) *GuanceV1Client {
	return &GuanceV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *GuanceV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
