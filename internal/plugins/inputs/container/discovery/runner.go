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
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
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

		runner, err := newPromRunnerWithURLParams(
			"k8s.pod/"+pod.Name,
			queryPodOwner(pod),
			pod.Status.PodIP, // podIP:port
			pod.Annotations[annotationPrometheusioPort],
			pod.Annotations[annotationPrometheusioScheme],
			pod.Annotations[annotationPrometheusioPath],
		)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s, err: %s, skip", pod.Name, err)
			continue
		}

		runner.setTag("namespace", pod.Namespace)
		runner.setTag("pod_name", pod.Name)
		runner.setTags(d.cfg.ExtraTags)
		runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

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
			runner, err := newPromRunnerWithURLParams(
				fmt.Sprintf("k8s.service(%s)-pod/%s", svc.Name, pod.Name),
				queryPodOwner(pod),
				pod.Status.PodIP, // podIP:port
				svc.Annotations[annotationPrometheusioPort],
				svc.Annotations[annotationPrometheusioScheme],
				svc.Annotations[annotationPrometheusioPath],
			)
			if err != nil {
				klog.Warnf("failed to new prom runner of service %s to pod %s, err: %s, skip", svc.Name, pod.Name, err)
				continue
			}

			runner.setTag("namespace", pod.Namespace)
			runner.setTag("service_name", svc.Name)
			runner.setTag("pod_name", pod.Name)
			runner.setTags(d.cfg.ExtraTags)
			runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

			klog.Infof("created prom runner of service %s to pod %s, urls %s", svc.Name, pod.Name, runner.conf.URLs)
			res = append(res, runner)
		}
	}

	return res
}

func (d *Discovery) newPromFromPodAnnotationExport() []*promRunner {
	var res []*promRunner

	pods := d.getLocalPodsFromLabelSelector("pod-export-config", defaultNamespace, nil)

	for _, pod := range pods {
		cfg, ok := pod.Annotations[annotationPromExport]
		if !ok {
			continue
		}

		runners, err := newPromRunnersForPod(pod, cfg)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s export-config, err: %s, skip", pod.Name, err)
			continue
		}

		for _, runner := range runners {
			runner.setTag("namespace", pod.Namespace)
			runner.setTag("pod_name", pod.Name)
			runner.setTags(d.cfg.ExtraTags)
			runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

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

		runners, err := newPromRunnersForPod(pod, ins.InputConf)
		if err != nil {
			klog.Warnf("failed to new prom runner of crd, err: %s, skip", err)
			return
		}

		for _, runner := range runners {
			runner.setTag("namespace", pod.Namespace)
			runner.setTag("pod_name", pod.Name)
			runner.setTags(d.cfg.ExtraTags)
			runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

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
		if item == nil {
			continue
		}
		if len(item.Spec.PodMetricsEndpoints) == 0 {
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

				u, err := getPromURL(pod.Status.PodIP, strconv.Itoa(port), metricsEndpoints.Scheme, metricsEndpoints.Path)
				if err != nil {
					// unreachable
					continue
				}
				u.RawQuery = url.Values(metricsEndpoints.Params).Encode()

				conf := &promConfig{
					Source:          fmt.Sprintf("k8s.podMonitor(%s)-pod/%s", item.Name, pod.Name),
					URLs:            []string{u.String()},
					MeasurementName: queryPodOwner(pod),
				}
				if val, err := time.ParseDuration(metricsEndpoints.Interval); err != nil {
					conf.Interval = defaultPrometheusioInterval
				} else {
					conf.Interval = val
				}

				runner, err := newPromRunnerWithConfig(conf)
				if err != nil {
					klog.Warnf("failed to new prom runner of PodMonitor %s pod %s, err: %s", item.Name, pod.Name, err)
					continue
				}

				runner.setTag("namespace", pod.Namespace)
				runner.setTag("pod_name", pod.Name)
				runner.setTags(d.cfg.ExtraTags)
				runner.setTags(getTargetLabels(pod.Labels, item.Spec.PodTargetLabels))
				runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

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

					u, err := getPromURL(pod.Status.PodIP, strconv.Itoa(port), endpoint.Scheme, endpoint.Path)
					if err != nil {
						// unreachable
						continue
					}
					u.RawQuery = url.Values(endpoint.Params).Encode()

					conf := &promConfig{
						Source:          fmt.Sprintf("k8s.serviceMonitor(%s)-service(%s)-pod/%s", item.Name, svc.Name, pod.Name),
						URLs:            []string{u.String()},
						MeasurementName: queryPodOwner(pod),
					}
					if val, err := time.ParseDuration(endpoint.Interval); err != nil {
						conf.Interval = defaultPrometheusioInterval
					} else {
						conf.Interval = val
					}

					runner, err := newPromRunnerWithConfig(conf)
					if err != nil {
						klog.Warnf("failed to new PromRunner of serviceMonitor %s service %s, err: %s", item.Name, service.Name, err)
						continue
					}

					runner.setTag("namespace", pod.Namespace)
					runner.setTag("pod_name", pod.Name)
					runner.setTag("service_name", svc.Name)
					runner.setTags(getTargetLabels(pod.Labels, item.Spec.PodTargetLabels))
					runner.setTags(getTargetLabels(svc.Labels, item.Spec.TargetLabels))
					runner.setTags(d.cfg.ExtraTags)
					runner.setCustomerTags(pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)

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
		res = append(res, &list.Items[idx])
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

func newPromRunnersForPod(pod *apicorev1.Pod, inputConfig string) ([]*promRunner, error) {
	cfg := completePromConfig(inputConfig, pod)
	return newPromRunnerWithTomlConfig(cfg)
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
