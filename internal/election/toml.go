// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

// ElectionCfg defined election configure in datakit.conf.
type ElectionCfg struct {
	Enable             bool `toml:"enable"`
	EnableNamespaceTag bool `toml:"enable_namespace_tag"`

	Namespace string            `toml:"namespace"`
	Tags      map[string]string `toml:"tags"`
}
