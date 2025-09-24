// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"regexp"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	corev1 "k8s.io/api/core/v1"
)

// PodName                = "__kubernetes_pod_name"
// PodNamespace           = "__kubernetes_pod_namespace"
// PodNodeName            = "__kubernetes_pod_node_name"
// PodIP                  = "__kubernetes_pod_ip"
// PodHostIP              = "__kubernetes_pod_host_ip"
// PodLabel               = "__kubernetes_pod_label_%s"
// PodAnnotation          = "__kubernetes_pod_annotation_%s"
// PodContainerPortNumber = "__kubernetes_pod_container_%s_port_%s_number"

var PodValueFroms = []struct {
	key keyMatcher
	fn  func(*corev1.Pod, []string) string
}{
	{
		key: newKeyMatcher("__kubernetes_pod_name"),
		fn:  func(item *corev1.Pod, _ []string) string { return item.Name },
	},
	{
		key: newKeyMatcher("__kubernetes_pod_namespace"),
		fn:  func(item *corev1.Pod, _ []string) string { return item.Namespace },
	},
	{
		key: newKeyMatcher("__kubernetes_pod_node_name"),
		fn:  func(item *corev1.Pod, _ []string) string { return item.Spec.NodeName },
	},
	{
		key: newKeyMatcher("__kubernetes_pod_ip"),
		fn:  func(item *corev1.Pod, _ []string) string { return item.Status.PodIP },
	},
	{
		key: newKeyMatcher("__kubernetes_pod_host_ip"),
		fn:  func(item *corev1.Pod, _ []string) string { return item.Status.HostIP },
	},
	{
		// e.g. __kubernetes_pod_label_app
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_pod_label_(.+)`)),
		fn: func(item *corev1.Pod, args []string) string {
			if len(args) != 1 {
				return ""
			}
			return item.Labels[args[0]]
		},
	},
	{
		// e.g. __kubernetes_pod_annotation_prometheus.io/scheme
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_pod_annotation_(.+)`)),
		fn: func(item *corev1.Pod, args []string) string {
			if len(args) != 1 {
				return ""
			}
			if args[0] == annotationPrometheusioPath && item.Annotations[args[0]] == "" {
				return "/metrics"
			}
			return item.Annotations[args[0]]
		},
	},
	{
		// e.g. __kubernetes_pod_container_nginx_port_metrics_number
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`^__kubernetes_pod_container_(.*?)_port_(.*?)_number$`)),
		fn: func(item *corev1.Pod, args []string) string {
			if len(args) != 2 {
				return ""
			}
			for _, container := range item.Spec.Containers {
				if container.Name != args[0] {
					continue
				}
				for _, port := range container.Ports {
					if port.Name == args[1] {
						return strconv.Itoa(int(port.ContainerPort))
					}
				}
			}
			return ""
		},
	},
	{
		// e.g. __kubernetes_pod_container_port_metrics_number
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`^__kubernetes_pod_container_port_(.*?)_number$`)),
		fn: func(item *corev1.Pod, args []string) string {
			if len(args) != 1 {
				return ""
			}
			for _, container := range item.Spec.Containers {
				for _, port := range container.Ports {
					if port.Name == args[0] {
						return strconv.Itoa(int(port.ContainerPort))
					}
				}
			}
			return ""
		},
	},
}

type podParser struct{ item *corev1.Pod }

func newPodParser(item *corev1.Pod) *podParser { return &podParser{item} }

func (p *podParser) shouldScrape(scrape string) bool {
	if scrape == matchedScrape {
		return true
	}
	_, res := p.matches(scrape)
	return res == matchedScrape
}

func (p *podParser) parsePromConfig(ins *Instance) (*basePromConfig, error) {
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
		if matched, res := matchInstanceOrHost(v, u.Host); matched {
			if res != "" {
				tags[k] = res
			}
			continue
		}
		if matched, res := p.matches(v); matched && res != "" {
			if v == "__kubernetes_pod_node_name" {
				res = config.RenameNode(res)
			}
			if v == podAnnotationParamTags {
				m := config.ParseGlobalTags(res)
				for key, val := range m {
					tags[key] = val
				}
				continue
			}
			tags[k] = res
			continue
		}
		if !isKeywords(v) {
			tags[k] = v
		}
	}

	measurement := ins.Measurement
	if matched, res := p.matches(ins.Measurement); matched && res != "" {
		measurement = res
	}
	if measurement == podAnnotationParamMeasurement {
		measurement = ""
	}

	return &basePromConfig{
		urlstr:              u.String(),
		measurement:         measurement,
		keepExistMetricName: ins.keepExistMetricName,
		honorTimestamps:     ins.honorTimestamps,
		tags:                tags,
	}, nil
}

func (p *podParser) matches(key string) (matched bool, res string) {
	for _, v := range PodValueFroms {
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
