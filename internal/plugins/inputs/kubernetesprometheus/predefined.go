// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ipt *Input) applyPredefinedInstances() {
	keys := getExtraLabelAsTags()

	if ipt.EnableDiscoveryOfPrometheusPodAnnotations {
		ins := &Instance{
			Role:   "pod",
			Scrape: "__kubernetes_pod_annotation_" + annotationPrometheusioScrape,
			Target: Target{
				// Scheme: "__kubernetes_pod_annotation_" + annotationPrometheusioScheme,
				Scheme: "http",
				Port:   "__kubernetes_pod_annotation_" + annotationPrometheusioPort,
				Path:   "__kubernetes_pod_annotation_" + annotationPrometheusioPath,
			},
			Custom: Custom{
				Measurement:      "__kubernetes_pod_annotation_" + annotationPrometheusioParamMeasurement,
				JobAsMeasurement: false,
				Tags: map[string]string{
					"instance":  "__kubernetes_mate_instance",
					"host":      "__kubernetes_mate_host",
					"namespace": "__kubernetes_pod_namespace",
					"pod_name":  "__kubernetes_pod_name",
				},
			},
		}

		for _, key := range keys {
			ins.Custom.Tags[key] = "__kubernetes_pod_label_" + key
		}

		ins.setDefault(ipt)
		ins.Custom.keepExistMetricName = !disableKeepExistMetricName()
		ipt.InstanceManager.Instances = append(ipt.InstanceManager.Instances, ins)

		klog.Info("apply PodAnnotations from predefined instance")
	}

	if ipt.EnableDiscoveryOfPrometheusServiceAnnotations {
		ins := &Instance{
			Role:   "service",
			Scrape: "__kubernetes_service_annotation_" + annotationPrometheusioScrape,
			Target: Target{
				// Scheme: "__kubernetes_service_annotation_" + annotationPrometheusioScheme,
				Scheme: "http",
				Port:   "__kubernetes_service_annotation_" + annotationPrometheusioPort,
				Path:   "__kubernetes_service_annotation_" + annotationPrometheusioPath,
			},
			Custom: Custom{
				Measurement:      "__kubernetes_service__annotation_" + annotationPrometheusioParamMeasurement,
				JobAsMeasurement: false,
				Tags: map[string]string{
					"instance":     "__kubernetes_mate_instance",
					"host":         "__kubernetes_mate_host",
					"namespace":    "__kubernetes_service_namespace",
					"service_name": "__kubernetes_service_name",
					"pod_name":     "__kubernetes_service_target_name",
				},
			},
		}

		for _, key := range keys {
			ins.Custom.Tags[key] = "__kubernetes_service_label_" + key
		}

		ins.setDefault(ipt)
		ins.Custom.keepExistMetricName = !disableKeepExistMetricName()
		ipt.InstanceManager.Instances = append(ipt.InstanceManager.Instances, ins)

		klog.Info("apply ServiceAnnotations from predefined instance")
	}
}

func (ipt *Input) applyCRDs(ctx context.Context, client client.Client, scrapeManager scrapeManagerInterface) {
	if ipt.EnableDiscoveryOfPrometheusPodMonitors || ipt.EnableDiscoveryOfPrometheusServiceMonitors {
		klog.Info("apply PodMonitors and ServiceMonitors from predefined instance")
		asTags := getExtraLabelAsTags()

		managerGo.Go(func(_ context.Context) error {
			tick := time.NewTicker(time.Second * 20)
			defer tick.Stop()

			for {
				select {
				case <-ctx.Done():
					klog.Info("podmonitor/servicemonitor fetcher exit")
					return nil

				case <-tick.C:
					if ipt.EnableDiscoveryOfPrometheusPodMonitors {
						if err := fetchPodMonitor(ctx, ipt, client, scrapeManager, asTags); err != nil {
							klog.Warn(err)
						}
					}

					if ipt.EnableDiscoveryOfPrometheusServiceMonitors {
						if err := fetchServiceMonitor(ctx, ipt, client, scrapeManager, asTags); err != nil {
							klog.Warn(err)
						}
					}
				}
			}
		})
	}
}

func fetchPodMonitor(
	ctx context.Context,
	ipt *Input,
	client client.Client,
	scrapeManager scrapeManagerInterface,
	asTags []string,
) error {
	list, err := client.GetPrmetheusPodMonitors("").List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return err
	}

	for idx, item := range list.Items {
		if item == nil || len(item.Spec.PodMetricsEndpoints) == 0 {
			continue
		}

		var instances []*Instance

		for _, endpoints := range item.Spec.PodMetricsEndpoints {
			ins := &Instance{
				Role:   "pod",
				Scrape: "true",
				Target: Target{
					Scheme: endpoints.Scheme,
					Port:   fmt.Sprintf("__kubernetes_pod_container_port_%s_number", endpoints.Port),
					Path:   endpoints.Path,
					Params: url.Values(endpoints.Params).Encode(),
				},
				Custom: Custom{
					Measurement:      getParamMeasurement(endpoints.Params),
					JobAsMeasurement: false,
					Tags: map[string]string{
						"instance":  "__kubernetes_mate_instance",
						"host":      "__kubernetes_mate_host",
						"namespace": "__kubernetes_pod_namespace",
						"pod_name":  "__kubernetes_pod_name",
					},
				},
			}

			for _, key := range asTags {
				ins.Custom.Tags[key] = "__kubernetes_pod_label_" + key
			}

			if endpoints.TLSConfig != nil {
				ins.Auth = Auth{
					TLSConfig: &dknet.TLSClientConfig{
						InsecureSkipVerify: endpoints.TLSConfig.SafeTLSConfig.InsecureSkipVerify,
					},
				}
				if ins.Target.Scheme == "" {
					ins.Target.Scheme = "https"
				}
			}

			for _, labelName := range item.Spec.PodTargetLabels {
				ins.Custom.Tags[labelName] = "__kubernetes_pod_label_" + labelName
			}

			ins.setDefault(ipt)
			ins.Custom.keepExistMetricName = !disableKeepExistMetricName()
			instances = append(instances, ins)
		}

		pods := []*apicorev1.Pod{}

		if item.Spec.NamespaceSelector.Any {
			pods = getLocalPodsFromLabelSelector(client, ipt.nodeName, "", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					pods = append(pods, getLocalPodsFromLabelSelector(client, ipt.nodeName, namespace, &list.Items[idx].Spec.Selector)...)
				}
			} else {
				pods = getLocalPodsFromLabelSelector(client, ipt.nodeName, item.Namespace, &list.Items[idx].Spec.Selector)
			}
		}

		if len(pods) == 0 {
			continue
		}

		p := &Pod{
			role:      RolePodMonitor,
			instances: instances,
			scrape:    scrapeManager,
			feeder:    ipt.feeder,
		}

		for _, podItem := range pods {
			if shouldSkipPod(podItem) {
				continue
			}

			key := fmt.Sprintf("podMonitor:%s/pod:%s", item.Name, podItem.Name)
			traits := podTraits(podItem)

			if p.scrape.isTraitsExists(p.role, key, traits) {
				continue
			}

			p.startScrape(ctx, key, traits, podItem)
		}
	}

	return nil
}

func fetchServiceMonitor(
	ctx context.Context,
	ipt *Input,
	client client.Client,
	scrapeManager scrapeManagerInterface,
	asTags []string,
) error {
	list, err := client.GetPrmetheusServiceMonitors("").List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return err
	}

	for idx, item := range list.Items {
		if item == nil || len(item.Spec.Endpoints) == 0 {
			continue
		}

		var instances []*Instance

		for _, endpoints := range item.Spec.Endpoints {
			ins := &Instance{
				Role:   "endpoints",
				Scrape: "true",
				Target: Target{
					Scheme: endpoints.Scheme,
					Port:   fmt.Sprintf("__kubernetes_endpoints_port_%s_number", endpoints.Port),
					Path:   endpoints.Path,
					Params: url.Values(endpoints.Params).Encode(),
				},
				Custom: Custom{
					Measurement:      getParamMeasurement(endpoints.Params),
					JobAsMeasurement: false,
					Tags: map[string]string{
						"instance":  "__kubernetes_mate_instance",
						"host":      "__kubernetes_mate_host",
						"namespace": "__kubernetes_endpoints_namespace",
						"pod_name":  "__kubernetes_endpoints_address_target_name",
					},
				},
			}

			for _, key := range asTags {
				ins.Custom.Tags[key] = "__kubernetes_endpoints_label_" + key
			}

			if endpoints.TLSConfig != nil {
				ins.Auth = Auth{
					TLSConfig: &dknet.TLSClientConfig{
						InsecureSkipVerify: endpoints.TLSConfig.SafeTLSConfig.InsecureSkipVerify,
					},
				}
				if ins.Target.Scheme == "" {
					ins.Target.Scheme = "https"
				}
			}

			for _, labelName := range item.Spec.TargetLabels {
				ins.Custom.Tags[labelName] = "__kubernetes_endpoints_label_" + labelName
			}

			ins.setDefault(ipt)
			ins.Custom.keepExistMetricName = !disableKeepExistMetricName()
			instances = append(instances, ins)
		}

		endpointsList := []*apicorev1.Endpoints{}

		if item.Spec.NamespaceSelector.Any {
			endpointsList = getLocalEndpointsFromLabelSelector(client, ipt.nodeName, "", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					endpointsList = append(endpointsList,
						getLocalEndpointsFromLabelSelector(client, ipt.nodeName, namespace, &list.Items[idx].Spec.Selector)...,
					)
				}
			} else {
				endpointsList = getLocalEndpointsFromLabelSelector(client, ipt.nodeName, item.Namespace, &list.Items[idx].Spec.Selector)
			}
		}

		if len(endpointsList) == 0 {
			continue
		}

		for _, ep := range endpointsList {
			for insIdx, ins := range instances {
				key := fmt.Sprintf("serviceMonitor:%s/endpoints:%s/ins[%d]", item.Name, ep.Name, insIdx)
				tryCreateScrapeForEndpoints(ctx, RoleServiceMonitor, key, ep, ins, scrapeManager, ipt.feeder)
			}
		}
	}

	return nil
}

func getLocalPodsFromLabelSelector(
	client client.Client,
	nodeName, namespace string,
	selector *metav1.LabelSelector,
) (res []*apicorev1.Pod) {
	opt := metav1.ListOptions{
		ResourceVersion: "0",
		FieldSelector:   "spec.nodeName=" + nodeName,
	}
	if selector != nil {
		opt.LabelSelector = newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String()
	}

	list, err := client.GetPods(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods from namespace '%s', err: %s", namespace, err)
		return
	}
	for idx := range list.Items {
		if list.Items[idx].Status.Phase == apicorev1.PodRunning {
			res = append(res, &list.Items[idx])
		}
	}
	return
}

func getLocalEndpointsFromLabelSelector(
	client client.Client,
	nodeName, namespace string,
	selector *metav1.LabelSelector,
) (res []*apicorev1.Endpoints) {
	opt := metav1.ListOptions{ResourceVersion: "0"}
	if selector != nil {
		opt.LabelSelector = newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String()
	}

	list, err := client.GetEndpoints(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get endpoints from namespace '%s', err: %s", namespace, err)
		return
	}
	for idx := range list.Items {
		if len(list.Items[idx].Subsets) != 0 && len(list.Items[idx].Subsets[0].Addresses) != 0 {
			res = append(res, &list.Items[idx])
		}
	}
	return
}

func getParamMeasurement(params map[string][]string) string {
	meas, ok := params["measurement"]
	if ok {
		if len(meas) > 0 {
			return meas[0]
		}
	}
	return ""
}

func getExtraLabelAsTags() []string {
	str := os.Getenv("ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC")
	var keys []string

	if err := json.Unmarshal([]byte(str), &keys); err != nil {
		return nil
	}

	res := unique(append(keys, config.Cfg.Dataway.GlobalCustomerKeys...))
	sort.Strings(res)
	return res
}

func disableKeepExistMetricName() bool {
	str := os.Getenv("ENV_INPUT_CONTAINER_KEEP_EXIST_PROMETHEUS_METRIC_NAME")
	return str == "false"
}
