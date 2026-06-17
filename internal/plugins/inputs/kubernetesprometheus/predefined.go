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

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	voidTagKey                        = "__void_tag_key"
	podAnnotationParamMeasurement     = "__kubernetes_pod_annotation_" + annotationPrometheusioParamMeasurement
	podAnnotationParamTags            = "__kubernetes_pod_annotation_" + annotationPrometheusioParamTags
	serviceAnnotationParamMeasurement = "__kubernetes_service_annotation_" + annotationPrometheusioParamMeasurement
	serviceAnnotationParamTags        = "__kubernetes_service_annotation_" + annotationPrometheusioParamTags
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
				Measurement:      podAnnotationParamMeasurement,
				JobAsMeasurement: false,
				Tags: map[string]string{
					"instance":  "__kubernetes_mate_instance",
					"host":      "__kubernetes_mate_host",
					"namespace": "__kubernetes_pod_namespace",
					"pod_name":  "__kubernetes_pod_name",
					voidTagKey:  podAnnotationParamTags,
				},
			},
		}

		for _, key := range keys {
			ins.Custom.Tags[key] = "__kubernetes_pod_label_" + key
		}

		ins.setDefault(ipt)
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
				Measurement:      serviceAnnotationParamMeasurement,
				JobAsMeasurement: false,
				Tags: map[string]string{
					"instance":     "__kubernetes_mate_instance",
					"host":         "__kubernetes_mate_host",
					"namespace":    "__kubernetes_service_namespace",
					"service_name": "__kubernetes_service_name",
					"pod_name":     "__kubernetes_service_target_name",
					voidTagKey:     serviceAnnotationParamTags,
				},
			},
		}

		for _, key := range keys {
			ins.Custom.Tags[key] = "__kubernetes_service_label_" + key
		}

		ins.setDefault(ipt)
		ipt.InstanceManager.Instances = append(ipt.InstanceManager.Instances, ins)

		klog.Info("apply ServiceAnnotations from predefined instance")
	}
}

func (ipt *Input) applyCRDs(ctx context.Context, client k8sclient.Client, scrapeManager scrapeManagerInterface) {
	asTags := getExtraLabelAsTags()
	if ipt.EnableDiscoveryOfPrometheusPodMonitors {
		klog.Info("apply PodMonitors from predefined instance")
		informer, list, watchUnavailable := newPodMonitorInformer(client)
		watcher := newMonitorWatcher(
			RolePodMonitor,
			informer,
			scrapeManager,
			func(ctx context.Context, key string, obj interface{}) ([]string, error) {
				item, ok := obj.(*monitoringv1.PodMonitor)
				if !ok {
					return nil, monitorObjectError(RolePodMonitor, obj)
				}
				return ipt.reconcilePodMonitor(ctx, key, item, client, scrapeManager, asTags)
			},
		)
		watcher.list = list
		watcher.watchUnavailable = watchUnavailable
		managerGo.Go(func(_ context.Context) error {
			watcher.Run(ctx)
			return nil
		})
	}

	if ipt.EnableDiscoveryOfPrometheusServiceMonitors {
		klog.Info("apply ServiceMonitors from predefined instance")
		informer, list, watchUnavailable := newServiceMonitorInformer(client)
		watcher := newMonitorWatcher(
			RoleServiceMonitor,
			informer,
			scrapeManager,
			func(ctx context.Context, key string, obj interface{}) ([]string, error) {
				item, ok := obj.(*monitoringv1.ServiceMonitor)
				if !ok {
					return nil, monitorObjectError(RoleServiceMonitor, obj)
				}
				return ipt.reconcileServiceMonitor(ctx, key, item, client, scrapeManager, asTags)
			},
		)
		watcher.list = list
		watcher.watchUnavailable = watchUnavailable
		managerGo.Go(func(_ context.Context) error {
			watcher.Run(ctx)
			return nil
		})
	}
}

func (ipt *Input) reconcilePodMonitor(
	ctx context.Context,
	monitorKey string,
	item *monitoringv1.PodMonitor,
	client k8sclient.Client,
	scrapeManager scrapeManagerInterface,
	asTags []string,
) ([]string, error) {
	if len(item.Spec.PodMetricsEndpoints) == 0 {
		return nil, nil
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
		instances = append(instances, ins)
	}

	var pods []*apicorev1.Pod
	if item.Spec.NamespaceSelector.Any {
		items, err := getLocalPodsFromLabelSelector(ctx, client, ipt.nodeName, "", &item.Spec.Selector)
		if err != nil {
			return nil, err
		}
		pods = items
	} else if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
		for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
			items, err := getLocalPodsFromLabelSelector(ctx, client, ipt.nodeName, namespace, &item.Spec.Selector)
			if err != nil {
				return nil, err
			}
			pods = append(pods, items...)
		}
	} else {
		items, err := getLocalPodsFromLabelSelector(ctx, client, ipt.nodeName, item.Namespace, &item.Spec.Selector)
		if err != nil {
			return nil, err
		}
		pods = items
	}

	p := &Pod{
		role:      RolePodMonitor,
		instances: instances,
		scrape:    scrapeManager,
		feeder:    ipt.feeder,
	}

	var taskKeys []string
	for _, podItem := range pods {
		if shouldSkipPod(podItem) {
			continue
		}

		key := fmt.Sprintf("podMonitor:%s/pod:%s/%s", monitorKey, podItem.Namespace, podItem.Name)
		taskKeys = append(taskKeys, key)
		traits := podTraits(podItem)

		if p.scrape.isTraitsExists(p.role, key, traits) {
			continue
		}

		p.scrape.removeScrape(p.role, key)
		p.startScrape(ctx, key, traits, podItem)
	}

	return taskKeys, nil
}

func (ipt *Input) reconcileServiceMonitor(
	ctx context.Context,
	monitorKey string,
	item *monitoringv1.ServiceMonitor,
	client k8sclient.Client,
	scrapeManager scrapeManagerInterface,
	asTags []string,
) ([]string, error) {
	if len(item.Spec.Endpoints) == 0 {
		return nil, nil
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
		instances = append(instances, ins)
	}

	var endpointsList []*apicorev1.Endpoints
	if item.Spec.NamespaceSelector.Any {
		items, err := getLocalEndpointsFromLabelSelector(ctx, client, "", &item.Spec.Selector)
		if err != nil {
			return nil, err
		}
		endpointsList = items
	} else if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
		for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
			items, err := getLocalEndpointsFromLabelSelector(ctx, client, namespace, &item.Spec.Selector)
			if err != nil {
				return nil, err
			}
			endpointsList = append(endpointsList, items...)
		}
	} else {
		items, err := getLocalEndpointsFromLabelSelector(ctx, client, item.Namespace, &item.Spec.Selector)
		if err != nil {
			return nil, err
		}
		endpointsList = items
	}

	var taskKeys []string
	for _, ep := range endpointsList {
		for insIdx, ins := range instances {
			key := fmt.Sprintf(
				"serviceMonitor:%s/endpoints:%s/%s/ins[%d]",
				monitorKey, ep.Namespace, ep.Name, insIdx,
			)
			taskKeys = append(taskKeys, key)
			if !scrapeManager.isTraitsExists(RoleServiceMonitor, key, endpointsTraits(ep)) {
				scrapeManager.removeScrape(RoleServiceMonitor, key)
			}
			tryCreateScrapeForEndpoints(ctx, RoleServiceMonitor, key, ep, ins, scrapeManager, ipt.feeder)
		}
	}

	return taskKeys, nil
}

func getLocalPodsFromLabelSelector(
	ctx context.Context,
	client k8sclient.Client,
	nodeName, namespace string,
	selector *metav1.LabelSelector,
) ([]*apicorev1.Pod, error) {
	opt := metav1.ListOptions{
		ResourceVersion: "0",
		FieldSelector:   "spec.nodeName=" + nodeName,
	}
	if selector != nil {
		opt.LabelSelector = labelSelectorToString(selector)
	}

	list, err := client.GetPods(namespace).List(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods from namespace %q: %w", namespace, err)
	}
	var res []*apicorev1.Pod
	for idx := range list.Items {
		if list.Items[idx].Status.Phase == apicorev1.PodRunning {
			res = append(res, &list.Items[idx])
		}
	}
	return res, nil
}

func getLocalEndpointsFromLabelSelector(
	ctx context.Context,
	client k8sclient.Client,
	namespace string,
	selector *metav1.LabelSelector,
) ([]*apicorev1.Endpoints, error) {
	opt := metav1.ListOptions{ResourceVersion: "0"}
	if selector != nil {
		opt.LabelSelector = labelSelectorToString(selector)
	}

	list, err := client.GetEndpoints(namespace).List(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints from namespace %q: %w", namespace, err)
	}
	var res []*apicorev1.Endpoints
	for idx := range list.Items {
		if len(list.Items[idx].Subsets) != 0 && len(list.Items[idx].Subsets[0].Addresses) != 0 {
			res = append(res, &list.Items[idx])
		}
	}
	return res, nil
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
