// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1beta1 wraps Datakit resource by kubernetes client-gen.
package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Datakit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DatakitSpec `json:"spec,omitempty"`
}

type DatakitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Datakit `json:"items,omitempty"`
}

type DatakitSpec struct {
	Selector  *metav1.LabelSelector `json:"selector"`
	Instances []DatakitInstance     `json:"instances,omitempty"`
}

type DatakitInstance struct {
	K8sNamespace  string `json:"k8sNamespace"`
	K8sDeployment string `json:"k8sDeployment"`
	LogsConf      string `json:"datakit/logs"`
	InputConf     string `json:"inputConf"`
}
