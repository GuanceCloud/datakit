// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrometheusMonitoringExtraConfig struct {
	Matches []struct {
		NamespaceSelector struct {
			Any             bool     `json:"any,omitempty"`
			MatchNamespaces []string `json:"matchNamespaces,omitempty"`
		} `json:"namespaceSelector,omitempty"`

		Selector metav1.LabelSelector `json:"selector"`

		PromConfig *promConfig `json:"promConfig"`
	} `json:"matches"`
}

func (p *PrometheusMonitoringExtraConfig) matchPromConfig(targetLabels map[string]string, namespace string) *promConfig {
	if len(p.Matches) == 0 {
		return nil
	}

	for _, match := range p.Matches {
		if !match.NamespaceSelector.Any {
			if len(match.NamespaceSelector.MatchNamespaces) != 0 &&
				slices.Index(match.NamespaceSelector.MatchNamespaces, namespace) == -1 {
				continue
			}
		}
		if !newLabelSelector(match.Selector.MatchLabels, match.Selector.MatchExpressions).Matches(targetLabels) {
			continue
		}
		return match.PromConfig
	}

	return nil
}

func mergePromConfig(c1 *promConfig, c2 *promConfig) *promConfig {
	c3 := &promConfig{
		Source:   c1.Source,
		Interval: c1.Interval,
		URLs:     c1.URLs,
		Tags:     c1.Tags,
	}

	c3.IgnoreReqErr = c2.IgnoreReqErr
	c3.MetricTypes = c2.MetricTypes
	c3.MetricNameFilter = c2.MetricNameFilter
	c3.MetricNameFilterIgnore = c2.MetricNameFilterIgnore
	c3.MeasurementPrefix = c2.MeasurementPrefix
	c3.MeasurementName = c2.MeasurementName
	c3.Measurements = c2.Measurements

	c3.TLSOpen = c2.TLSOpen
	c3.UDSPath = c2.UDSPath
	c3.CacertFile = c2.CacertFile
	c3.CertFile = c2.CertFile
	c3.KeyFile = c2.KeyFile

	c3.TagsIgnore = c2.TagsIgnore
	c3.TagsRename = c2.TagsRename
	c3.AsLogging = c2.AsLogging
	c3.IgnoreTagKV = c2.IgnoreTagKV

	c3.HTTPHeaders = c2.HTTPHeaders
	c3.DisableInfoTag = c2.DisableInfoTag

	for k, v := range c2.Tags {
		if _, ok := c3.Tags[k]; !ok {
			c3.Tags[k] = v
		}
	}

	return c3
}
