// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps Datakit resource by kubernetes client-gen.
package v1beta1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// DatakitsGetter has a method to return a DatakitInterface.
// A group's client should implement this interface.
type DatakitsGetter interface {
	Datakits(namespace string) DatakitInterface
}

// DatakitInterface has methods to work with Datakit resources.
type DatakitInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*Datakit, error)
	List(ctx context.Context, opts metav1.ListOptions) (*DatakitList, error)
	// ...
}

// datakits implements DatakitInterface.
type datakits struct {
	client rest.Interface
	ns     string
}

// newDatakits return a Datakits.
func newDatakits(c *GuanceV1Client, namespace string) *datakits {
	return &datakits{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the datakit, and returns the corresponding datakit object, and an error if there is any.
func (c *datakits) Get(ctx context.Context, name string, opts metav1.GetOptions) (*Datakit, error) {
	result := Datakit{}
	err := c.client.Get().
		Namespace(c.ns).
		Resource("datakits").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

// List takes label and field selectors, and returns the list of Datakits that match those selectors.
func (c *datakits) List(ctx context.Context, opts metav1.ListOptions) (*DatakitList, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result := DatakitList{}
	err := c.client.Get().
		Namespace(c.ns).
		Resource("datakits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(&result)
	return &result, err
}
