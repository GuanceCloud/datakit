// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterLoggingConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterLoggingConfigSpec `json:"spec,omitempty"`
}

type ClusterLoggingConfigSpec struct {
	Selector        Selector `json:"selector"`
	PodTargetLabels []string `json:"podTargetLabels,omitempty"`
	Configs         []Config `json:"configs,omitempty"`
}

type Selector struct {
	NamespaceRegex   string `json:"namespaceRegex,omitempty"`
	PodRegex         string `json:"podRegex,omitempty"`
	PodLabelSelector string `json:"podLabelSelector,omitempty"`
	ContainerRegex   string `json:"containerRegex,omitempty"`
}

type Config struct {
	Type                  string            `json:"type"`
	Source                string            `json:"source"`
	Disable               bool              `json:"disable,omitempty"`
	Path                  string            `json:"path,omitempty"`
	StorageIndex          string            `json:"storage_index,omitempty"`
	Service               string            `json:"service,omitempty"`
	CharacterEncoding     string            `json:"character_encoding,omitempty"`
	Pipeline              string            `json:"pipeline,omitempty"`
	Multiline             string            `json:"multiline_match,omitempty"`
	RemoveAnsiEscapeCodes bool              `json:"remove_ansi_escape_codes,omitempty"`
	FromBeginning         bool              `json:"from_beginning,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterLoggingConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ClusterLoggingConfig `json:"items,omitempty"`
}
