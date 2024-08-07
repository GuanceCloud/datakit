// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"regexp"
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// NodeName                = "__kubernetes_node_name"
// NodeAnnotation          = "__kubernetes_node_annotation_%s"
// NodeLabel               = "__kubernetes_node_label_%s"
// NodeAddressHostname     = "__kubernetes_node_address_Hostname"
// NodeAddressInternalIP   = "__kubernetes_node_address_InternalIP"
// NodeAddressExternalIP   = "__kubernetes_node_address_ExternalIP"
// NodeKubeletEndpointPort = "__kubernetes_node_kubelet_endpoint_port"

var NodeValueFroms = []struct {
	key keyMatcher
	fn  func(*corev1.Node, []string) string
}{
	{
		key: newKeyMatcher("__kubernetes_node_name"),
		fn:  func(item *corev1.Node, _ []string) string { return item.Name },
	},
	{
		// e.g. __kubernetes_node_label_kubernetes.io/os
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_node_label_(.+)`)),
		fn: func(item *corev1.Node, args []string) string {
			if len(args) != 1 {
				return ""
			}
			return item.Labels[args[0]]
		},
	},
	{
		// e.g. __kubernetes_node_annotation_kubernetes.io/ttl
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_node_annotation_(.+)`)),
		fn: func(item *corev1.Node, args []string) string {
			if len(args) != 1 {
				return ""
			}
			return item.Annotations[args[0]]
		},
	},
	{
		key: newKeyMatcher("__kubernetes_node_address_Hostname"),
		fn: func(item *corev1.Node, _ []string) string {
			for _, address := range item.Status.Addresses {
				if address.Type == corev1.NodeHostName {
					return address.Address
				}
			}
			return ""
		},
	},
	{
		key: newKeyMatcher("__kubernetes_node_address_InternalIP"),
		fn: func(item *corev1.Node, _ []string) string {
			for _, address := range item.Status.Addresses {
				if address.Type == corev1.NodeInternalIP {
					return address.Address
				}
			}
			return ""
		},
	},
	{
		key: newKeyMatcher("__kubernetes_node_address_ExternalIP"),
		fn: func(item *corev1.Node, _ []string) string {
			for _, address := range item.Status.Addresses {
				if address.Type == corev1.NodeExternalIP {
					return address.Address
				}
			}
			return ""
		},
	},
	{
		key: newKeyMatcher("__kubernetes_node_kubelet_endpoint_port"),
		fn: func(item *corev1.Node, _ []string) string {
			return strconv.Itoa(int(item.Status.DaemonEndpoints.KubeletEndpoint.Port))
		},
	},
}

type nodeParser struct{ item *corev1.Node }

func newNodeParser(item *corev1.Node) *nodeParser { return &nodeParser{item} }

func (p *nodeParser) shouldScrape(scrape string) bool {
	if scrape == matchedScrape {
		return true
	}
	_, res := p.matches(scrape)
	return res == matchedScrape
}

func (p *nodeParser) parsePromConfig(ins *Instance) (*basePromConfig, error) {
	elems := []string{ins.Scheme, ins.Address, ins.Port, ins.Path}

	for idx, elem := range elems {
		if matched, res := p.matches(elem); matched && res != "" {
			elems[idx] = res
			continue
		}
	}

	u, err := buildURLWithParams(elems[0], elems[1], elems[2], elems[3], ins.Params)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{}

	for k, v := range ins.Tags {
		switch v {
		case MateInstanceTag:
			tags[k] = u.Host
		case MateHostTag:
			if host := splitHost(u.Host); host != "" {
				tags[k] = host
			}
		default:
			if matched, res := p.matches(v); matched && res != "" {
				tags[k] = res
			} else {
				tags[k] = v
			}
		}
	}

	measurement := ins.Measurement
	if matched, res := p.matches(ins.Measurement); matched && res != "" {
		measurement = res
	}

	return &basePromConfig{
		urlstr:      u.String(),
		measurement: measurement,
		tags:        tags,
	}, nil
}

func (p *nodeParser) matches(key string) (matched bool, res string) {
	for _, v := range NodeValueFroms {
		matched, args := v.key.matches(key)
		if !matched {
			continue
		}
		if res := v.fn(p.item, args); res != "" {
			return true, res
		}
	}
	return false, ""
}
