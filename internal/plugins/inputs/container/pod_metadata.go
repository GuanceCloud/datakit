// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"errors"
	"fmt"
	"time"

	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type podCacheMissReason string

type podMetadataUsage string

const (
	podCacheMissNotSynced   podCacheMissReason = "not_synced"
	podCacheMissNotFound    podCacheMissReason = "not_found"
	podCacheMissReasonError podCacheMissReason = "error"

	podMetadataUsageLogging podMetadataUsage = "logging"
	podMetadataUsageMetric  podMetadataUsage = "metric"
	podMetadataUsageObject  podMetadataUsage = "object"
)

type podCacheMissError struct {
	reason podCacheMissReason
	err    error
}

func (e *podCacheMissError) Error() string {
	return fmt.Sprintf("pod cache lookup failed (%s): %v", e.reason, e.err)
}

func (e *podCacheMissError) Unwrap() error { return e.err }

type podMetadataProvider interface {
	Get(ctx context.Context, namespace, name string) (*corev1.Pod, error)
}

type apiPodMetadataProvider struct {
	client k8sclient.Client
}

func newAPIPodMetadataProvider(client k8sclient.Client) podMetadataProvider {
	if client == nil {
		return nil
	}
	return &apiPodMetadataProvider{client: client}
}

func (p *apiPodMetadataProvider) Get(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return p.client.GetPods(namespace).Get(ctx, name, metav1.GetOptions{ResourceVersion: "0"})
}

type informerPodMetadataProvider struct {
	lister    corev1listers.PodLister
	hasSynced cache.InformerSynced
}

func (p *informerPodMetadataProvider) Get(_ context.Context, namespace, name string) (*corev1.Pod, error) {
	if !p.hasSynced() {
		return nil, &podCacheMissError{
			reason: podCacheMissNotSynced,
			err:    errors.New("informer cache is not synced"),
		}
	}

	pod, err := p.lister.Pods(namespace).Get(name)
	if err == nil {
		return pod, nil
	}

	reason := podCacheMissReasonError
	if apierrors.IsNotFound(err) {
		reason = podCacheMissNotFound
	}
	return nil, &podCacheMissError{reason: reason, err: err}
}

func newPodCacheMissWarningLimiter() *rate.Limiter {
	return rate.NewLimiter(rate.Every(time.Minute), 1)
}

func (c *containerCollector) getPodMetadata(
	ctx context.Context,
	usage podMetadataUsage,
	namespace, name string,
) (*corev1.Pod, bool) {
	if c.podMetadata == nil || name == "" {
		return nil, false
	}

	pod, err := c.podMetadata.Get(ctx, namespace, name)
	if err == nil {
		return pod, true
	}

	var cacheMiss *podCacheMissError
	if !errors.As(err, &cacheMiss) {
		l.Warnf("query pod failed, err: %s", err)
		return nil, false
	}

	reason := cacheMiss.reason
	switch reason {
	case podCacheMissNotSynced, podCacheMissNotFound, podCacheMissReasonError:
		// Keep the public metric label bounded to the declared reasons.
	default:
		reason = podCacheMissReasonError
	}
	podCacheMissVec.WithLabelValues(string(usage), string(reason)).Inc()

	if c.podCacheMissWarningLimiter == nil || c.podCacheMissWarningLimiter.Allow() {
		l.Warnf("query pod metadata from informer cache failed, usage=%s, err: %s", usage, err)
	}
	return nil, false
}
