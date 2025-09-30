// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"fmt"
	"regexp"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	corev1 "k8s.io/api/core/v1"
)

// ServiceName               = "__kubernetes_service_name"
// ServiceNamespace          = "__kubernetes_service_namespace"
// ServiceLabel              = "__kubernetes_service_label_%s"
// ServiceAnnotation         = "__kubernetes_service_annotation_%s"
// ServicePort               = "__kubernetes_service_port_%s_port"
// ServicePortTarget         = "__kubernetes_service_port_%s_targetport"
// ServiceTargetPodName      = "__kubernetes_service_target_pod_name"
// ServiceTargetPodNamespace = "__kubernetes_service_target_pod_namespace"

var ServiceValueFroms = []struct {
	key keyMatcher
	fn  func(*corev1.Service, []string) string
}{
	{
		key: newKeyMatcher("__kubernetes_service_name"),
		fn:  func(item *corev1.Service, _ []string) string { return item.Name },
	},
	{
		key: newKeyMatcher("__kubernetes_service_namespace"),
		fn:  func(item *corev1.Service, _ []string) string { return item.Namespace },
	},
	{
		// e.g. __kubernetes_service_label_kubernetes.io/cluster-service
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_service_label_(.+)`)),
		fn: func(item *corev1.Service, args []string) string {
			if len(args) != 1 {
				return ""
			}
			return item.Labels[args[0]]
		},
	},
	{
		// e.g. __kubernetes_service_annotation_prometheus.io/port
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`__kubernetes_service_annotation_(.+)`)),
		fn: func(item *corev1.Service, args []string) string {
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
		// e.g. __kubernetes_service_port_metrics_port
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`^__kubernetes_service_port_(.*?)_port$`)),
		fn: func(item *corev1.Service, args []string) string {
			if len(args) != 1 {
				return ""
			}
			for _, port := range item.Spec.Ports {
				if port.Name != args[0] {
					continue
				}
				return strconv.Itoa(int(port.Port))
			}
			return ""
		},
	},
	{
		// e.g. __kubernetes_service_port_metrics_targetport
		key: newKeyMatcherWithRegexp(regexp.MustCompile(`^__kubernetes_service_port_(.*?)_targetport$`)),
		fn: func(item *corev1.Service, args []string) string {
			if len(args) != 1 {
				return ""
			}
			for _, port := range item.Spec.Ports {
				if port.Name != args[0] {
					continue
				}
				return port.TargetPort.String()
			}
			return ""
		},
	},

	// convert to endpoints
	{
		key: newKeyMatcher("__kubernetes_service_target_kind"),
		fn: func(_ *corev1.Service, _ []string) string {
			return "__kubernetes_endpoints_address_target_kind"
		},
	},
	{
		key: newKeyMatcher("__kubernetes_service_target_name"),
		fn: func(_ *corev1.Service, _ []string) string {
			return "__kubernetes_endpoints_address_target_name"
		},
	},
	{
		key: newKeyMatcher("__kubernetes_service_target_namespace"),
		fn: func(_ *corev1.Service, _ []string) string {
			return "__kubernetes_endpoints_address_target_namespace"
		},
	},
	// deprecated, use __kubernetes_service_target_name
	{
		key: newKeyMatcher("__kubernetes_service_target_pod_name"),
		fn: func(_ *corev1.Service, _ []string) string {
			return "__kubernetes_endpoints_address_target_pod_name"
		},
	},
	// deprecated, use __kubernetes_service_target_namespace
	{
		key: newKeyMatcher("__kubernetes_service_target_pod_namespace"),
		fn: func(_ *corev1.Service, _ []string) string {
			return "__kubernetes_endpoints_address_target_pod_namespace"
		},
	},
}

type serviceParser struct{ item *corev1.Service }

func newServiceParser(item *corev1.Service) *serviceParser { return &serviceParser{item} }

func (s *serviceParser) shouldScrape(scrape string) bool {
	if scrape == matchedScrape {
		return true
	}
	_, res := s.matches(scrape)
	return res == matchedScrape
}

func (s *serviceParser) transToEndpointsInstance(ins *Instance) *Instance {
	newIns := &Instance{
		Role:       "endpoints",
		Namespaces: deepCopySlice(ins.Namespaces),
		Selector:   selectorToString(s.item.Labels),
		Target: Target{
			Scheme:  ins.Target.Scheme,
			Address: ins.Target.Address,
			Port:    ins.Target.Port,
			Path:    ins.Target.Path,
			Params:  ins.Target.Params,
		},
		Headers: ins.Headers,
		Custom: Custom{
			Measurement:         ins.Measurement,
			JobAsMeasurement:    ins.Custom.JobAsMeasurement,
			keepExistMetricName: ins.keepExistMetricName,
			honorTimestamps:     ins.honorTimestamps,
			Tags:                map[string]string{},
		},
		Auth: Auth{
			BearerTokenFile: ins.Auth.BearerTokenFile,
			TLSConfig:       deepCopyTLSConfig(ins.Auth.TLSConfig),
		},
	}

	if _, res := s.matches(ins.Scrape); res != "" {
		newIns.Scrape = res
	}

	// Target
	if _, res := s.matches(ins.Target.Scheme); res != "" {
		newIns.Target.Scheme = res
	}
	if _, res := s.matches(ins.Target.Port); res != "" {
		// The port is integer
		if _, err := strconv.Atoi(res); err == nil {
			newIns.Target.Port = res
		} else {
			// The port is string, convert to endpoints port
			newIns.Target.Port = fmt.Sprintf("__kubernetes_endpoints_port_%s_number", res)
		}
	}
	if _, res := s.matches(ins.Target.Path); res != "" {
		newIns.Target.Path = res
	}

	// Custom
	if _, res := s.matches(ins.Custom.Measurement); res != "" {
		newIns.Custom.Measurement = res
	}
	if newIns.Custom.Measurement == serviceAnnotationParamMeasurement {
		newIns.Custom.Measurement = ""
	}
	for k, v := range ins.Tags {
		value := v
		if matched, res := s.matches(v); matched && res != "" {
			if v == serviceAnnotationParamTags {
				m := config.ParseGlobalTags(res)
				for key, val := range m {
					newIns.Custom.Tags[key] = val
				}
				continue
			}
			value = res
		}
		newIns.Custom.Tags[k] = value
	}

	return newIns
}

func (s *serviceParser) matches(key string) (matched bool, res string) {
	for _, v := range ServiceValueFroms {
		matched, args := v.key.matches(key)
		if !matched {
			continue
		}
		if res := v.fn(s.item, args); res != "" {
			return true, res
		}
	}
	return false, ""
}

func deepCopySlice(arr []string) []string {
	return append(make([]string, 0, len(arr)), arr...)
}

func deepCopyTLSConfig(cfg *dknet.TLSClientConfig) *dknet.TLSClientConfig {
	if cfg == nil {
		return nil
	}
	// not copy Base64
	return &dknet.TLSClientConfig{
		CaCerts:            deepCopySlice(cfg.CaCerts),
		Cert:               cfg.Cert,
		CertKey:            cfg.CertKey,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.ServerName,
	}
}
