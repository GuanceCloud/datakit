// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps DataKit resource by kubernetes client-gen.
package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DataKit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DataKitSpec `json:"spec"`
}

type DataKitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DataKit `json:"items"`
}

type DataKitSpec struct {
	Selector      *metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`
	InputConf     string                `json:"input-conf"`
	K8sDeployment string                `json:"k8s-deployment"`
	K8sNamespace  string                `json:"k8s-namespace"`
	Tags          string                `json:"tags"`
}
