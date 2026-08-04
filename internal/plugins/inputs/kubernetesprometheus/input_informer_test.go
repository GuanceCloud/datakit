// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestInformerListOptions(t *testing.T) {
	general := metav1.ListOptions{}
	setInformerListLimit(&general)
	require.Equal(t, informerListLimit, general.Limit)
	require.Empty(t, general.FieldSelector)

	pod := metav1.ListOptions{}
	setPodInformerListOptions(&pod, "node-a")
	require.Equal(t, informerListLimit, pod.Limit)
	require.Equal(t, "spec.nodeName=node-a", pod.FieldSelector)
}

func TestInformerFactoriesSelectPodScope(t *testing.T) {
	clientset := &kubernetes.Clientset{}

	withoutPods := newInformerFactories(clientset, true, "node-a", false)
	require.Nil(t, withoutPods.pod)

	clusterWide := newInformerFactories(clientset, false, "node-a", true)
	require.True(t, clusterWide.pod == clusterWide.general)
	require.False(t, clusterWide.podIsSeparate)

	nodeLocal := newInformerFactories(clientset, true, "node-a", true)
	require.True(t, nodeLocal.pod != nodeLocal.general)
	require.True(t, nodeLocal.podIsSeparate)
}
