// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package election

// ElectionCfg defined election configure in datakit.conf.
type ElectionCfg struct {
	Enable             bool     `toml:"enable"`
	NodeWhitelist      []string `toml:"node_whitelist"`
	EnableNamespaceTag bool     `toml:"enable_namespace_tag"`
	OperatorURL        string   `toml:"operator_url"`

	Namespace string            `toml:"namespace"`
	Tags      map[string]string `toml:"tags"`
}

func (cfg *ElectionCfg) Normalize() error {
	normalized, err := NormalizeOperatorURL(cfg.OperatorURL)
	if err != nil {
		return err
	}
	cfg.OperatorURL = normalized
	return nil
}
