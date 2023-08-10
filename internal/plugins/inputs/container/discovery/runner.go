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

	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *Discovery) newPromFromPodAnnotationKeys() []*promRunner {
	var res []*promRunner

	opt := metav1.ListOptions{FieldSelector: "spec.nodeName=" + d.localNodeName}
	list, err := d.client.GetPods().List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods, err: %s", err)
		return nil
	}

	for _, item := range list.Items {
		if !parseScrapeFromProm(item.Annotations[annotationPrometheusioScrape]) {
			continue
		}

		runner, err := newPromRunnerWithURLParams(
			"k8s.pod/"+item.Name,
			item.Status.PodIP, // podIP:port
			item.Annotations[annotationPrometheusioPort],
			item.Annotations[annotationPrometheusioScheme],
			item.Annotations[annotationPrometheusioPath],
			d.cfg.ExtraTags,
		)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s, err: %s, skip", item.Name, err)
			continue
		}

		runner.addSingleTag("namespace", item.Namespace)
		runner.addSingleTag("pod", item.Name)
		runner.addTags(d.cfg.ExtraTags)

		klog.Infof("create prom runner of pod %s, urls %s", item.Name, runner.conf.URLs)
		res = append(res, runner)
	}

	return res
}

func (d *Discovery) newPromFromServiceAnnotations() []*promRunner {
	var res []*promRunner

	list, err := d.client.GetServices().List(context.Background(), metaV1ListOption)
	if err != nil {
		klog.Warnf("failed to get services, err: %s", err)
		return nil
	}

	for _, item := range list.Items {
		if !parseScrapeFromProm(item.Annotations[annotationPrometheusioScrape]) {
			continue
		}

		runner, err := newPromRunnerWithURLParams(
			"k8s.service/"+item.Name,
			fmt.Sprintf("%s.%s", item.Name, item.Namespace), // service_name.service_namespace:port
			item.Annotations[annotationPrometheusioPort],
			item.Annotations[annotationPrometheusioScheme],
			item.Annotations[annotationPrometheusioPath],
			d.cfg.ExtraTags,
		)
		if err != nil {
			klog.Warnf("failed to new prom runner of service %s, err: %s, skip", item.Name, err)
			continue
		}

		runner.addSingleTag("namespace", item.Namespace)
		runner.addSingleTag("service", item.Name)
		runner.addTags(d.cfg.ExtraTags)

		klog.Infof("created prom runner of service %s, urls %s", item.Name, runner.conf.URLs)
		res = append(res, runner)
	}

	return res
}

func (d *Discovery) newPromFromPodAnnotationExport() []*promRunner {
	var res []*promRunner

	opt := metav1.ListOptions{FieldSelector: "spec.nodeName=" + d.localNodeName}
	list, err := d.client.GetPods().List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods from node %s, err: %s", d.localNodeName, err)
		return nil
	}

	for idx, item := range list.Items {
		cfg, ok := item.GetAnnotations()[annotationPromExport]
		if !ok {
			continue
		}

		runners, err := newPromRunnersForPod(&list.Items[idx], cfg)
		if err != nil {
			klog.Warnf("failed to new prom runner of pod %s export, err: %s, skip", item.Name, err)
			continue
		}

		for _, runner := range runners {
			runner.addSingleTag("namespace", item.Namespace)
			runner.addSingleTag("pod", item.Name)
			runner.addTags(d.cfg.ExtraTags)

			klog.Infof("created prom runner of PodExport %s, urls %s", item.Name, runner.conf.URLs)
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
			runner.addSingleTag("namespace", pod.Namespace)
			runner.addSingleTag("pod", pod.Name)
			runner.addTags(d.cfg.ExtraTags)

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

	list, err := d.client.GetPrmetheusPodMonitors().List(context.Background(), metaV1ListOption)
	if err != nil {
		klog.Warnf("failed to get PodMonitor, err: %s", err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil {
			continue
		}

		pods := []*apicorev1.Pod{}

		if item.Spec.NamespaceSelector.Any {
			pods = d.getPodsFromLabelSelector("", "PodMonitor", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					pods = append(pods, d.getPodsFromLabelSelector(namespace, "PodMonitor", &list.Items[idx].Spec.Selector)...)
				}
			} else {
				pods = d.getPodsFromLabelSelector(item.Namespace, "PodMonitor", &list.Items[idx].Spec.Selector)
			}
		}

		klog.Infof("find %d pods from PodMonitor %s", len(pods), item.Name)

		for _, pod := range pods {
			for _, metricsEndpoints := range item.Spec.PodMetricsEndpoints {
				var port int
				if metricsEndpoints.Port != "" {
					port = findContainerPort(pod, metricsEndpoints.Port)
					if port == -1 {
						klog.Warnf("not found port %s for PodMonitor %s pod %s, skip", metricsEndpoints.Port, item.Name, pod.Name)
						continue
					}
				}

				u, err := getPromURL(pod.Status.PodIP, strconv.Itoa(port), metricsEndpoints.Scheme, metricsEndpoints.Path)
				if err != nil {
					// unreachable
					continue
				}
				u.RawQuery = url.Values(metricsEndpoints.Params).Encode()

				conf := &promConfig{
					Source: fmt.Sprintf("k8s.podMonitor/%s::%s", item.Name, pod.Name),
					URLs:   []string{u.String()},
				}
				if val, err := time.ParseDuration(metricsEndpoints.Interval); err != nil {
					conf.Interval = defaultPrometheusioInterval
				} else {
					conf.Interval = val
				}

				if d.cfg.PrometheusMonitoringExtraConfig != nil {
					klog.Debugf("matching promConfig %#v", d.cfg.PrometheusMonitoringExtraConfig)
					newconf := d.cfg.PrometheusMonitoringExtraConfig.matchPromConfig(pod.Labels, pod.Namespace)
					if newconf != nil {
						klog.Debugf("use promConfig %#v for PodMonitor %s podName %s", newconf, item.Name, pod.Name)
						conf = mergePromConfig(conf, newconf)
					}
				}

				runner, err := newPromRunnerWithConfig(conf)
				if err != nil {
					klog.Warnf("failed to new prom runner of PodMonitor %s pod %s, err: %s", item.Name, pod.Name, err)
					continue
				}

				runner.addSingleTag("namespace", pod.Namespace)
				runner.addSingleTag("pod", pod.Name)
				runner.addTags(d.cfg.ExtraTags)

				klog.Infof("create prom runner for PodMonitor %s pod %s, urls: %#v", item.Name, pod.Name, runner.conf.URLs)
				res = append(res, runner)
			}
		}
	}

	return res
}

func (d *Discovery) newPromForServiceMonitors() []*promRunner {
	var res []*promRunner

	list, err := d.client.GetPrmetheusServiceMonitors().List(context.Background(), metaV1ListOption)
	if err != nil {
		klog.Warnf("failed to get ServiceMonitor, err: %s", d.localNodeName, err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil {
			continue
		}

		var services []*apicorev1.Service

		if item.Spec.NamespaceSelector.Any {
			services = d.getServicesFromLabelSelector("", "ServiceMonitor", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					services = append(services, d.getServicesFromLabelSelector(namespace, "ServiceMonitor", &list.Items[idx].Spec.Selector)...)
				}
			} else {
				services = d.getServicesFromLabelSelector(item.Namespace, "ServiceMonitor", &list.Items[idx].Spec.Selector)
			}
		}

		klog.Infof("find %d services from ServiceMonitor %s", len(services), item.Name)

		for _, service := range services {
			for _, endpoint := range item.Spec.Endpoints {
				var port int
				if endpoint.Port != "" {
					port = findServicePort(service, endpoint.Port)
					if port == -1 {
						klog.Warnf("not found port %s for ServiceMonitor %s service %s, skip", endpoint.Port, item.Name, service.Name)
						continue
					}
				}

				u, err := getPromURL(fmt.Sprintf("%s.%s", service.Name, service.Namespace), strconv.Itoa(port), endpoint.Scheme, endpoint.Path)
				if err != nil {
					// unreachable
					continue
				}
				u.RawQuery = url.Values(endpoint.Params).Encode()

				conf := &promConfig{
					Source: fmt.Sprintf("k8s.serviceMonitor/%s::%s", item.Name, service.Name),
					URLs:   []string{u.String()},
				}
				if val, err := time.ParseDuration(endpoint.Interval); err != nil {
					conf.Interval = defaultPrometheusioInterval
				} else {
					conf.Interval = val
				}

				if d.cfg.PrometheusMonitoringExtraConfig != nil {
					klog.Debugf("matching promConfig %#v", d.cfg.PrometheusMonitoringExtraConfig)
					newConf := d.cfg.PrometheusMonitoringExtraConfig.matchPromConfig(service.Labels, service.Namespace)
					if newConf != nil {
						klog.Debugf("use promConfig %#v for serviceMonitor %s service %s", newConf, item.Name, service.Name)
						conf = mergePromConfig(conf, newConf)
					}
				}

				runner, err := newPromRunnerWithConfig(conf)
				if err != nil {
					klog.Warnf("failed to new PromRunner of serviceMonitor %s service %s, err: %s", item.Name, service.Name, err)
					continue
				}

				runner.addSingleTag("namespace", service.Namespace)
				runner.addSingleTag("service", service.Name)
				runner.addTags(d.cfg.ExtraTags)

				klog.Infof("create prom runner for ServiceMonitor %s service %s, urls: %s", item.Name, service.Name, runner.conf.URLs)
				res = append(res, runner)
			}
		}
	}

	return res
}

type datakitCRDHandler func(kubev1guancebeta1.DatakitInstance, *apicorev1.Pod)

func (d *Discovery) processCRDWithPod(fn datakitCRDHandler) error {
	list, err := d.client.GetDatakits().List(context.Background(), metaV1ListOption)
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
					t := d.getPodsFromLabelSelector(ins.K8sNamespace, ins.K8sDaemonSet, selector)
					pods = append(pods, t...)
				}
			}

			if ins.K8sDeployment != "" {
				selector, err := d.getDeploymentLabelSelector(ins.K8sNamespace, ins.K8sDeployment)
				if err != nil {
					klog.Debugf("not found Deployment %s LabelSelector, error %s, namespace %s", ins.K8sDeployment, err, ins.K8sNamespace)
				} else {
					t := d.getPodsFromLabelSelector(ins.K8sNamespace, ins.K8sDeployment, selector)
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

func (d *Discovery) getServicesFromLabelSelector(namespace, appName string, selector *metav1.LabelSelector) (res []*apicorev1.Service) {
	if selector == nil {
		return nil
	}
	opt := metav1.ListOptions{
		LabelSelector: newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String(),
	}

	services, err := d.client.GetServicesForNamespace(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get services from node %s, namespace %s, app %s, err: %s",
			d.localNodeName, namespace, appName, err)
		return
	}
	for idx := range services.Items {
		res = append(res, &services.Items[idx])
	}
	return
}

func (d *Discovery) getPodsFromLabelSelector(namespace, appName string, selector *metav1.LabelSelector) (res []*apicorev1.Pod) {
	if selector == nil {
		return nil
	}
	opt := metav1.ListOptions{
		LabelSelector: newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String(),
		FieldSelector: "spec.nodeName=" + d.localNodeName,
	}
	pods, err := d.client.GetPodsForNamespace(namespace).List(context.Background(), opt)
	if err != nil {
		klog.Warnf("failed to get pods from node %s, namespace %s, app %s, err: %s",
			d.localNodeName, namespace, appName, err)
		return
	}
	for idx := range pods.Items {
		res = append(res, &pods.Items[idx])
	}
	return
}

func (d *Discovery) getDaemonSetLabelSelector(namespace, daemonset string) (*metav1.LabelSelector, error) {
	daemonsetObj, err := d.client.GetDaemonSetsForNamespace(namespace).Get(context.Background(), daemonset, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if daemonsetObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid DaemonSet LabelSelector")
	}
	return daemonsetObj.Spec.Selector, nil
}

func (d *Discovery) getDeploymentLabelSelector(namespace, deployment string) (*metav1.LabelSelector, error) {
	deploymentObj, err := d.client.GetDeploymentsForNamespace(namespace).Get(context.Background(), deployment, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if deploymentObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid Deployment LabelSelector")
	}
	return deploymentObj.Spec.Selector, nil
}

func newPromRunnersForPod(item *apicorev1.Pod, inputConfig string) ([]*promRunner, error) {
	cfg := completePromConfig(inputConfig, item)
	return newPromRunnerWithTomlConfig(cfg)
}
