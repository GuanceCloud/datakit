// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultNamespace = ""

func (d *Discovery) newPromFromPodAnnotations() []*promRunner {
	var res []*promRunner

	pods := d.getLocalPodsFromLabelSelector("pod-export", defaultNamespace, nil)

	for _, pod := range pods {
		if !parseScrapeFromProm(pod.Annotations[annotationPrometheusioScrape]) {
			continue
		}

		measurementName := ""
		if meas := pod.Annotations[annotationPrometheusioParamMeasurement]; meas != "" {
			measurementName = meas
		}

		// scheme://podIP:port/path?params
		urlstr, err := joinPromURL(pod.Status.PodIP,
			pod.Annotations[annotationPrometheusioPort],
			pod.Annotations[annotationPrometheusioScheme],
			pod.Annotations[annotationPrometheusioPath],
			"")
		if err != nil {
			klog.Warnf("failed to parse config of pod %s, err: %s, skip", pod.Name, err)
			continue
		}

		config := newPromConfig(withSource(fmt.Sprintf("k8s/pod-annotations(%s)", pod.Name)),
			withMeasurementName(measurementName),
			withURLs([]string{urlstr}),
			withTag("namespace", pod.Namespace),
			withTag("pod_name", pod.Name),
			withTagIfNotEmpty(queryPodOwner(pod)),
			withTags(d.cfg.ExtraTags),
			withCustomerTags(pod.Labels, d.cfg.CustomerKeys))

		runner, err := newPromRunnerWithConfig(d, config)
		if err != nil {
			klog.Warnf("failed to create runner of pod %s, err: %s, skip", pod.Name, err)
			continue
		}
		klog.Infof("create prom runner of pod %s, urls %s", pod.Name, runner.conf.URLs)
		res = append(res, runner)
	}

	return res
}

func (d *Discovery) newPromFromServiceAnnotations() []*promRunner {
	var res []*promRunner

	svcs := d.getServicesFromLabelSelector("service-export", defaultNamespace, nil)

	for _, svc := range svcs {
		if !parseScrapeFromProm(svc.Annotations[annotationPrometheusioScrape]) {
			continue
		}

		selector := &metav1.LabelSelector{MatchLabels: svc.Spec.Selector}
		pods := d.getLocalPodsFromLabelSelector("service-export-to-pod", defaultNamespace, selector)

		for _, pod := range pods {
			measurementName := ""
			if meas := svc.Annotations[annotationPrometheusioParamMeasurement]; meas != "" {
				measurementName = meas
			}

			// This port is an annotation(prometheus.io/port) for the Service.
			// If this port does not match the Pod port, the collect will fail.

			// scheme://podIP:svc-port/path?params
			urlstr, err := joinPromURL(pod.Status.PodIP,
				svc.Annotations[annotationPrometheusioPort],
				svc.Annotations[annotationPrometheusioScheme],
				svc.Annotations[annotationPrometheusioPath],
				"")
			if err != nil {
				klog.Warnf("failed to parse config of svc %s to pod %s, err: %s, skip", svc.Name, pod.Name, err)
				continue
			}

			config := newPromConfig(withSource(fmt.Sprintf("k8s/service-annotations(%s)/pod(%s)", svc.Name, pod.Name)),
				withMeasurementName(measurementName),
				withURLs([]string{urlstr}),
				withTag("namespace", pod.Namespace),
				withTag("service_name", svc.Name),
				withTag("pod_name", pod.Name),
				withTagIfNotEmpty(queryPodOwner(pod)),
				withTags(d.cfg.ExtraTags),
				withCustomerTags(pod.Labels, d.cfg.CustomerKeys))

			runner, err := newPromRunnerWithConfig(d, config)
			if err != nil {
				klog.Warnf("failed to create runner of svc %s pod %s, err: %s, skip", svc.Name, pod.Name, err)
				continue
			}
			klog.Infof("created prom runner of svc %s to pod %s, urls %s", svc.Name, pod.Name, runner.conf.URLs)
			res = append(res, runner)
		}
	}

	return res
}

func (d *Discovery) newPromFromPodAnnotationExport() []*promRunner {
	var res []*promRunner

	pods := d.getLocalPodsFromLabelSelector("pod-export-config", defaultNamespace, nil)

	for _, pod := range pods {
		cfgStr, ok := pod.Annotations[annotationPromExport]
		if !ok {
			continue
		}

		runners, err := newPromRunnersForPod(d, pod, cfgStr)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s export-config, err: %s, skip", pod.Name, err)
			continue
		}

		for _, runner := range runners {
			if runner.conf == nil {
				continue
			}
			withTag("namespace", pod.Namespace)(runner.conf)
			withTag("pod_name", pod.Name)(runner.conf)
			withTagIfNotEmpty(queryPodOwner(pod))(runner.conf)
			withTags(d.cfg.ExtraTags)(runner.conf)
			withCustomerTags(pod.Labels, d.cfg.CustomerKeys)(runner.conf)

			klog.Infof("created prom runner of pod-export-config %s, urls %s", pod.Name, runner.conf.URLs)
			res = append(res, runner)
		}
	}

	return res
}

func (d *Discovery) newPromFromDatakitCRD() []*promRunner {
	var res []*promRunner

	fn := func(ins kubev1guancebeta1.DatakitInstance, pod *apicorev1.Pod) {
		klog.Debugf("find CRD inputConf, pod %s, namespace %s, conf: %s", pod.Name, pod.Namespace, ins.InputConf)

		runners, err := newPromRunnersForPod(d, pod, ins.InputConf)
		if err != nil {
			klog.Warnf("failed to new prom runner of crd, err: %s, skip", err)
			return
		}

		for _, runner := range runners {
			if runner.conf == nil {
				continue
			}
			withTag("namespace", pod.Namespace)(runner.conf)
			withTag("pod_name", pod.Name)(runner.conf)
			withTagIfNotEmpty(queryPodOwner(pod))(runner.conf)
			withTags(d.cfg.ExtraTags)(runner.conf)
			withCustomerTags(pod.Labels, d.cfg.CustomerKeys)(runner.conf)
			res = append(res, runner)
		}
	}

	if err := d.processCRDWithPod(fn); err != nil {
		klog.Debugf("failed to get datakits, err: %s", err)
		return nil
	}

	return res
}

func (d *Discovery) newPromForPodMonitors() []*promRunner {
	var res []*promRunner

	list, err := d.client.GetPrmetheusPodMonitors(defaultNamespace).List(context.Background(), metaV1ListOption)
	if err != nil {
		klog.Warnf("failed to get PodMonitor, err: %s", err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil || len(item.Spec.PodMetricsEndpoints) == 0 {
			continue
		}

		pods := []*apicorev1.Pod{}

		if item.Spec.NamespaceSelector.Any {
			pods = d.getLocalPodsFromLabelSelector("PodMonitor", defaultNamespace, &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					pods = append(pods, d.getLocalPodsFromLabelSelector("PodMonitor", namespace, &list.Items[idx].Spec.Selector)...)
				}
			} else {
				pods = d.getLocalPodsFromLabelSelector("PodMonitor", item.Namespace, &list.Items[idx].Spec.Selector)
			}
		}

		klog.Infof("find %d pods from PodMonitor %s", len(pods), item.Name)

		for _, metricsEndpoints := range item.Spec.PodMetricsEndpoints {
			for _, pod := range pods {
				port := findContainerPortForPod(pod, metricsEndpoints.Port)
				if port == -1 {
					klog.Warnf("not found port %s for PodMonitor %s pod %s, skip", metricsEndpoints.Port, item.Name, pod.Name)
					continue
				}

				// scheme://podIP:port/path?params
				urlstr, err := joinPromURL(pod.Status.PodIP,
					strconv.Itoa(port),
					metricsEndpoints.Scheme,
					metricsEndpoints.Path,
					url.Values(metricsEndpoints.Params).Encode())
				if err != nil {
					klog.Warnf("failed to parse config of pod %s, err: %s, skip", pod.Name, err)
					continue
				}

				measurementName := ""
				if meas := getParamMeasurement(metricsEndpoints.Params); meas != "" {
					measurementName = meas
				}

				config := newPromConfig(withSource(fmt.Sprintf("k8s/pod-monitor(%s)/pod(%s)", item.Name, pod.Name)),
					withMeasurementName(measurementName),
					withURLs([]string{urlstr}),
					withTag("namespace", pod.Namespace),
					withTag("pod_name", pod.Name),
					withTagIfNotEmpty(queryPodOwner(pod)),
					withTags(d.cfg.ExtraTags),
					withTags(getTargetLabels(pod.Labels, item.Spec.PodTargetLabels)),
					withCustomerTags(pod.Labels, d.cfg.CustomerKeys),
					withInterval(metricsEndpoints.Interval))

				runner, err := newPromRunnerWithConfig(d, config)
				if err != nil {
					klog.Warnf("failed to new prom runner of PodMonitor %s pod %s, err: %s, skip", item.Name, pod.Name, err)
					continue
				}
				klog.Infof("create prom runner for PodMonitor %s pod %s, urls: %#v", item.Name, pod.Name, runner.conf.URLs)
				res = append(res, runner)
			}
		}
	}

	return res
}

func (d *Discovery) newPromForServiceMonitors() []*promRunner {
	var res []*promRunner

	list, err := d.client.GetPrmetheusServiceMonitors(defaultNamespace).List(context.Background(), metaV1ListOption)
	if err != nil {
		klog.Warnf("failed to get ServiceMonitor, err: %s", err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil {
			continue
		}
		if len(item.Spec.Endpoints) == 0 {
			continue
		}

		var svcs []*apicorev1.Service

		if item.Spec.NamespaceSelector.Any {
			svcs = d.getServicesFromLabelSelector("ServiceMonitor", defaultNamespace, &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					svcs = append(svcs, d.getServicesFromLabelSelector("ServiceMonitor", namespace, &list.Items[idx].Spec.Selector)...)
				}
			} else {
				svcs = d.getServicesFromLabelSelector("ServiceMonitor", item.Namespace, &list.Items[idx].Spec.Selector)
			}
		}

		klog.Infof("find %d services from ServiceMonitor %s", len(svcs), item.Name)

		for _, svc := range svcs {
			selector := &metav1.LabelSelector{MatchLabels: svc.Spec.Selector}
			pods := d.getLocalPodsFromLabelSelector("ServiceMonitor", defaultNamespace, selector)

			for _, endpoint := range item.Spec.Endpoints {
				for _, pod := range pods {
					port := findContainerPortForService(svc, pod, endpoint.Port)
					if port == -1 {
						klog.Warnf("not found port %s for ServiceMonitor %s pod %s, skip", endpoint.Port, item.Name, pod.Name)
						continue
					}

					// scheme://podIP:port/path?params
					urlstr, err := joinPromURL(pod.Status.PodIP,
						strconv.Itoa(port),
						endpoint.Scheme,
						endpoint.Path,
						url.Values(endpoint.Params).Encode())
					if err != nil {
						klog.Warnf("failed to parse config of svc %s to pod %s, err: %s, skip", svc.Name, pod.Name, err)
						continue
					}

					measurementName := ""
					if meas := getParamMeasurement(endpoint.Params); meas != "" {
						measurementName = meas
					}

					config := newPromConfig(withSource(fmt.Sprintf("k8s/service-monitor(%s)/service(%s)/pod(%s)", item.Name, svc.Name, pod.Name)),
						withMeasurementName(measurementName),
						withURLs([]string{urlstr}),
						withTag("namespace", pod.Namespace),
						withTag("pod_name", pod.Name),
						withTag("service_name", svc.Name),
						withTagIfNotEmpty(queryPodOwner(pod)),
						withTags(d.cfg.ExtraTags),
						withTags(getTargetLabels(pod.Labels, item.Spec.PodTargetLabels)),
						withTags(getTargetLabels(svc.Labels, item.Spec.TargetLabels)),
						withCustomerTags(pod.Labels, d.cfg.CustomerKeys),
						withInterval(endpoint.Interval))

					runner, err := newPromRunnerWithConfig(d, config)
					if err != nil {
						klog.Warnf("failed to new PromRunner of serviceMonitor %s service %s, err: %s", item.Name, service.Name, err)
						continue
					}
					klog.Infof("create prom runner for ServiceMonitor %s service %s, urls: %s", item.Name, service.Name, runner.conf.URLs)
					res = append(res, runner)
				}
			}
		}
	}

	return res
}

type datakitCRDHandler func(kubev1guancebeta1.DatakitInstance, *apicorev1.Pod)

func (d *Discovery) processCRDWithPod(fn datakitCRDHandler) error {
	list, err := d.client.GetDatakits(defaultNamespace).List(context.Background(), metaV1ListOption)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		for _, ins := range item.Spec.Instances {
			if ins.K8sNamespace == "" {
				klog.Warn("invalid empty namespace")
				continue
			}

			pods := []*apicorev1.Pod{}

			if ins.K8sDaemonSet != "" {
				selector, err := d.getDaemonSetLabelSelector(ins.K8sNamespace, ins.K8sDaemonSet)
				if err != nil {
					klog.Debugf("not found DaemonSet %s LabelSelector, error %s, namespace %s", ins.K8sDaemonSet, err, ins.K8sNamespace)
				} else {
					t := d.getLocalPodsFromLabelSelector("datakit-crd/"+ins.K8sDaemonSet, ins.K8sNamespace, selector)
					pods = append(pods, t...)
				}
			}

			if ins.K8sDeployment != "" {
				selector, err := d.getDeploymentLabelSelector(ins.K8sNamespace, ins.K8sDeployment)
				if err != nil {
					klog.Debugf("not found Deployment %s LabelSelector, error %s, namespace %s", ins.K8sDeployment, err, ins.K8sNamespace)
				} else {
					t := d.getLocalPodsFromLabelSelector("datakit-crd/"+ins.K8sDeployment, ins.K8sNamespace, selector)
					pods = append(pods, t...)
				}
			}

			for _, pod := range pods {
				fn(ins, pod)
			}
		}
	}

	return nil
}

func (d *Discovery) getLocalPodsFromLabelSelector(source, namespace string, selector *metav1.LabelSelector) (res []*apicorev1.Pod) {
	opt := metav1.ListOptions{
		ResourceVersion: "0",
		FieldSelector:   "spec.nodeName=" + d.localNodeName,
	}
	if selector != nil {
		opt.LabelSelector = newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String()
	}

	list, err := d.client.GetPods(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods from namespace '%s' by %s, err: %s, skip", namespace, source, err)
		return
	}
	for idx := range list.Items {
		if list.Items[idx].Status.Phase == apicorev1.PodRunning {
			res = append(res, &list.Items[idx])
		}
	}
	return
}

func (d *Discovery) getServicesFromLabelSelector(source, namespace string, selector *metav1.LabelSelector) (res []*apicorev1.Service) {
	opt := metav1.ListOptions{
		ResourceVersion: "0",
	}
	if selector != nil {
		opt.LabelSelector = newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String()
	}

	list, err := d.client.GetServices(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get services from namespace '%s' by %s, err: %s, skip", namespace, source, err)
		return
	}
	for idx := range list.Items {
		res = append(res, &list.Items[idx])
	}
	return
}

func (d *Discovery) getDaemonSetLabelSelector(namespace, daemonset string) (*metav1.LabelSelector, error) {
	daemonsetObj, err := d.client.GetDaemonSets(namespace).Get(context.Background(), daemonset, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if daemonsetObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid DaemonSet LabelSelector")
	}
	return daemonsetObj.Spec.Selector, nil
}

func (d *Discovery) getDeploymentLabelSelector(namespace, deployment string) (*metav1.LabelSelector, error) {
	deploymentObj, err := d.client.GetDeployments(namespace).Get(context.Background(), deployment, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if deploymentObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid Deployment LabelSelector")
	}
	return deploymentObj.Spec.Selector, nil
}

func newPromRunnersForPod(discovery *Discovery, pod *apicorev1.Pod, inputConfig string) ([]*promRunner, error) {
	cfg := completePromConfig(inputConfig, pod)
	return newPromRunnerWithTomlConfig(discovery, cfg)
}

func getTargetLabels(labels map[string]string, target []string) map[string]string {
	if len(labels) == 0 || len(target) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, key := range target {
		value, ok := labels[key]
		if !ok {
			continue
		}
		m[replaceLabelKey(key)] = value
	}
	return m
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
