// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
)

const (
	annotationPromExport  = "datakit/prom.instances"
	annotationPromIPIndex = "datakit/prom.instances.ip_index"

	annotationPrometheusioScrape = "prometheus.io/scrape"
	annotationPrometheusioPort   = "prometheus.io/port"
	annotationPrometheusioPath   = "prometheus.io/path"
	annotationPrometheusioScheme = "prometheus.io/scheme"
)

var (
	defaultPromScheme = "http"
	defaultPromPath   = "/metrics"
)

type discovery struct {
	client                k8sClientX
	extraTags             map[string]string
	localNodeName         string
	extractK8sLabelAsTags bool

	enablePrometheusPodAnnotations     bool
	enablePrometheusServiceAnnotations bool
	enablePrometheusPodMonitors        bool
	enablePrometheusServiceMonitors    bool

	prometheusMonitoringExtraConfig *prometheusMonitoringExtraConfig

	pause   bool
	chPause chan bool
	done    <-chan interface{}
}

func newDiscovery(client k8sClientX, done <-chan interface{}) *discovery {
	return &discovery{
		client:  client,
		chPause: make(chan bool, maxPauseCh),
		done:    done,
	}
}

func (d *discovery) start() {
	if d.client == nil {
		l.Warn("invalid k8s client, input autodiscovery start failed")
		return
	}

	localNodeName, err := getLocalNodeName()
	if err != nil {
		l.Warnf("autodiscovery: %s", err)
		return
	}
	d.localNodeName = localNodeName

	var (
		runners         []*promRunner
		electionRunners []*promRunner
	)

	runners = d.updateRunners()
	l.Infof("autodiscovery: update input list, len %d", len(runners))

	if d.election() {
		electionRunners = d.updateElectionRunners()
		l.Infof("autodiscovery: update electionInput list, len %d", len(electionRunners))
	}

	updateTicker := time.NewTicker(time.Minute * 3)
	defer updateTicker.Stop()

	collectTicker := time.NewTicker(time.Second * 1)
	defer collectTicker.Stop()

	l.Infof("start k8s autodiscovery, node_name is %s", localNodeName)

	for {
		for _, r := range runners {
			r.runOnce()
		}

		if d.pause {
			l.Debug("not leader, skipped")
		} else if d.election() {
			for _, r := range electionRunners {
				r.runOnce()
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("autodiscovery: exit")
			return

		case <-d.done:
			l.Info("autodiscovery: terminate")
			return

		case <-updateTicker.C:
			runners = d.updateRunners()
			l.Infof("autodiscovery: update input list, len %d", len(runners))

			if d.election() {
				electionRunners = d.updateElectionRunners()
				l.Infof("autodiscovery: update electionInput list, len %d", len(electionRunners))
			}

		case <-collectTicker.C:
			// nil

		case d.pause = <-d.chPause:
		}
	}
}

func (d *discovery) election() bool {
	return d.enablePrometheusServiceAnnotations || d.enablePrometheusServiceMonitors
}

func (d *discovery) updateRunners() []*promRunner {
	runners := []*promRunner{}
	runners = append(runners, d.newPromFromPodAnnotationExport()...)
	runners = append(runners, d.newPromFromDatakitCRD()...)

	if d.enablePrometheusPodAnnotations {
		runners = append(runners, d.newPromFromPodAnnotationKeys()...)
	}
	if d.enablePrometheusPodMonitors {
		runners = append(runners, d.newPromForPodMonitors()...)
	}
	return runners
}

func (d *discovery) updateElectionRunners() []*promRunner {
	runners := []*promRunner{}

	if d.enablePrometheusServiceAnnotations {
		runners = append(runners, d.newPromFromServiceAnnotations()...)
	}
	if d.enablePrometheusServiceMonitors {
		runners = append(runners, d.newPromForServiceMonitors()...)
	}

	return runners
}

func (d *discovery) newPromFromPodAnnotationKeys() []*promRunner {
	var res []*promRunner

	opt := metav1.ListOptions{FieldSelector: "spec.nodeName=" + d.localNodeName}
	list, err := d.client.getPods().List(context.Background(), opt)
	if err != nil {
		l.Warnf("autodiscovery: failed to get pods, err: %s, retry in a minute", d.localNodeName, err)
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
			d.extraTags,
		)
		if err != nil {
			l.Warnf("autodiscovery: failed to new PromRunner of Pod %s, err: %s, retry in a minute", item.Name, err)
			continue
		}

		if d.extractK8sLabelAsTags {
			runner.addTags(item.Labels)
		}
		runner.addSingleTag("namespace", item.Namespace)
		runner.addSingleTag("pod_name", item.Name)

		l.Infof("autodiscovery: created PromRunner of Pod %s, urls %s", item.Name, runner.conf.URLs)
		res = append(res, runner)
	}

	return res
}

func (d *discovery) newPromFromServiceAnnotations() []*promRunner {
	var res []*promRunner

	list, err := d.client.getServices().List(context.Background(), metaV1ListOption)
	if err != nil {
		l.Warnf("autodiscovery: failed to get services, err: %s, retry in a minute", d.localNodeName, err)
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
			d.extraTags,
		)
		if err != nil {
			l.Warnf("autodiscovery: failed to new PromRunner of Service %s, err: %s, retry in a minute", item.Name, err)
			continue
		}

		runner.addSingleTag("namespace", item.Namespace)
		runner.addSingleTag("service_name", item.Name)

		l.Infof("autodiscovery: created PromRunner of Pod %s, urls %s", item.Name, runner.conf.URLs)
		res = append(res, runner)
	}

	return res
}

func (d *discovery) newPromFromPodAnnotationExport() []*promRunner {
	var res []*promRunner

	opt := metav1.ListOptions{FieldSelector: "spec.nodeName=" + d.localNodeName}
	list, err := d.client.getPods().List(context.Background(), opt)
	if err != nil {
		l.Warnf("autodiscovery: failed to get pods from node_name %s, err: %s, retry in a minute", d.localNodeName, err)
		return nil
	}

	for idx := range list.Items {
		cfg, ok := list.Items[idx].Annotations[annotationPromExport]
		if !ok {
			continue
		}

		runner, err := newPromRunnersForPod(&podMeta{Pod: &list.Items[idx]}, cfg, d.extraTags, d.extractK8sLabelAsTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner err %s", err)
			continue
		}

		res = append(res, runner...)
	}

	return res
}

func (d *discovery) newPromFromDatakitCRD() []*promRunner {
	var res []*promRunner

	fn := func(ins kubev1guancebeta1.DatakitInstance, pod *podMeta) {
		l.Debugf("autodiscovery: find CRD inputConf, pod_name: %s, pod_namespace: %s, conf: %s", pod.Name, pod.Namespace, ins.InputConf)

		runner, err := newPromRunnersForPod(pod, ins.InputConf, d.extraTags, d.extractK8sLabelAsTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner from crd, err: %s", err)
			return
		}

		res = append(res, runner...)
	}

	if err := d.processCRDWithPod(fn); err != nil {
		l.Debugf("autodiscovery: failed to get datakits, err: %s, retry in a minute", err)
		return nil
	}

	return res
}

func (d *discovery) newPromForPodMonitors() []*promRunner {
	var res []*promRunner

	list, err := d.client.getPrmetheusPodMonitors().List(context.Background(), metaV1ListOption)
	if err != nil {
		l.Warnf("autodiscovery: failed to get podMonitors, err: %s, retry in a minute", d.localNodeName, err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil {
			continue
		}

		l.Infof("autodiscovery: find podMonitor %s spec: %#v", item.Name, item)

		var pods []*podMeta

		if item.Spec.NamespaceSelector.Any {
			pods = d.getPodsFromLabelSelector("", "podMonitors", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					pods = append(pods, d.getPodsFromLabelSelector(namespace, "podMonitors", &list.Items[idx].Spec.Selector)...)
				}
			} else {
				pods = d.getPodsFromLabelSelector(item.Namespace, "podMonitors", &list.Items[idx].Spec.Selector)
			}
		}

		l.Infof("autodiscovery: find %d pods from podMonitor %s", len(pods), item.Name)

		for _, pod := range pods {
			for _, metricsEndpoints := range item.Spec.PodMetricsEndpoints {
				var port int
				if metricsEndpoints.Port != "" {
					port = pod.containerPort(metricsEndpoints.Port)
					if port == -1 {
						l.Warnf("autodiscovery: not found port %s for podMonitor %s podName %s, ignored", metricsEndpoints.Port, item.Name, pod.Name)
						continue
					}
				}

				u, err := getPromURL(
					pod.Status.PodIP,
					strconv.Itoa(port),
					metricsEndpoints.Scheme,
					metricsEndpoints.Path,
				)
				if err != nil {
					// unreachable
					continue
				}
				u.RawQuery = url.Values(metricsEndpoints.Params).Encode()

				l.Infof("autodiscovery: new PromRunner for podMonitor %s podName %s, url %s", item.Name, pod.Name, u.String())

				conf := &promConfig{
					Source: fmt.Sprintf("k8s.podMonitor/%s::%s", item.Name, pod.Name),
					URLs:   []string{u.String()},
				}
				if val, err := time.ParseDuration(metricsEndpoints.Interval); err != nil {
					conf.Interval = defaultPrometheusioInterval
				} else {
					conf.Interval = val
				}

				if d.prometheusMonitoringExtraConfig != nil {
					l.Debugf("autodiscovery: matching promConfig %#v", d.prometheusMonitoringExtraConfig)
					newconf := d.prometheusMonitoringExtraConfig.matchPromConfig(pod.Labels, pod.Namespace)
					if newconf != nil {
						l.Debugf("autodiscovery: use promConfig %#v for podMonitor %s podName %s", newconf, item.Name, pod.Name)
						conf = mergePromConfig(conf, newconf)
					}
				}

				runner, err := newPromRunnerWithConfig(conf)
				if err != nil {
					l.Warnf("autodiscovery: failed to new PromRunner of podMonitor %s podName %s, err: %s", item.Name, pod.Name, err)
					continue
				}

				runner.addSingleTag("namespace", pod.Namespace)
				runner.addSingleTag("pod", pod.Name)

				l.Infof("autodiscovery: new PromRunner for podMonitor %s podName %s, urls: %#v", item.Name, pod.Name, runner.conf.URLs)
				res = append(res, runner)
			}
		}
	}

	return res
}

func (d *discovery) newPromForServiceMonitors() []*promRunner {
	var res []*promRunner

	list, err := d.client.getPrmetheusServiceMonitors().List(context.Background(), metaV1ListOption)
	if err != nil {
		l.Warnf("autodiscovery: failed to get serviceMonitors, err: %s, retry in a minute", d.localNodeName, err)
		return nil
	}

	for idx, item := range list.Items {
		if item == nil {
			continue
		}

		l.Infof("autodiscovery: find serviceMonitor %s spec: %#v", item.Name, item)

		var services []*serviceMeta

		if item.Spec.NamespaceSelector.Any {
			services = d.getServicesFromLabelSelector("", "serviceMonitors", &list.Items[idx].Spec.Selector)
		} else {
			if len(item.Spec.NamespaceSelector.MatchNames) != 0 {
				for _, namespace := range item.Spec.NamespaceSelector.MatchNames {
					services = append(services, d.getServicesFromLabelSelector(namespace, "serviceMonitors", &list.Items[idx].Spec.Selector)...)
				}
			} else {
				services = d.getServicesFromLabelSelector(item.Namespace, "serviceMonitors", &list.Items[idx].Spec.Selector)
			}
		}

		l.Infof("autodiscovery: find %d services from serviceMonitor %s", len(services), item.Name)

		for _, service := range services {
			for _, endpoint := range item.Spec.Endpoints {
				var port int
				if endpoint.Port != "" {
					port = service.servicePort(endpoint.Port)
					if port == -1 {
						l.Warnf("autodiscovery: not found port %s for serviceMonitor %s serviceName %s, ignored", endpoint.Port, item.Name, service.Name)
						continue
					}
				}

				u, err := getPromURL(
					fmt.Sprintf("%s.%s", service.Name, service.Namespace),
					strconv.Itoa(port),
					endpoint.Scheme,
					endpoint.Path,
				)
				if err != nil {
					// unreachable
					continue
				}
				u.RawQuery = url.Values(endpoint.Params).Encode()

				l.Infof("autodiscovery: new PromRunner for serviceMonitor %s serviceName %s, url %s", item.Name, service.Name, u.String())

				conf := &promConfig{
					Source: fmt.Sprintf("k8s.serviceMonitor/%s::%s", item.Name, service.Name),
					URLs:   []string{u.String()},
				}
				if val, err := time.ParseDuration(endpoint.Interval); err != nil {
					conf.Interval = defaultPrometheusioInterval
				} else {
					conf.Interval = val
				}

				if d.prometheusMonitoringExtraConfig != nil {
					l.Debugf("autodiscovery: matching promConfig %#v", d.prometheusMonitoringExtraConfig)
					newConf := d.prometheusMonitoringExtraConfig.matchPromConfig(service.Labels, service.Namespace)
					if newConf != nil {
						l.Debugf("autodiscovery: use promConfig %#v for serviceMonitor %s serviceName %s", newConf, item.Name, service.Name)
						conf = mergePromConfig(conf, newConf)
					}
				}

				runner, err := newPromRunnerWithConfig(conf)
				if err != nil {
					l.Warnf("autodiscovery: failed to new PromRunner of serviceMonitor %s serviceName %s, err: %s", item.Name, service.Name, err)
					continue
				}

				runner.addSingleTag("namespace", service.Namespace)
				runner.addSingleTag("service", service.Name)

				l.Infof("autodiscovery: new promInput for serviceMonitor %s serviceName %s, urls: %s", item.Name, service.Name, runner.conf.URLs)
				res = append(res, runner)
			}
		}
	}

	return res
}

type datakitCRDHandler func(kubev1guancebeta1.DatakitInstance, *podMeta)

func (d *discovery) processCRDWithPod(fn datakitCRDHandler) error {
	list, err := d.client.getDatakits().List(context.Background(), metaV1ListOption)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		for _, ins := range item.Spec.Instances {
			if ins.K8sNamespace == "" {
				l.Warn("autodiscovery: invalid empty K8sNamespace")
				continue
			}

			pods := []*podMeta{}

			if ins.K8sDaemonSet != "" {
				selector, err := d.getDaemonSetLabelSelector(ins.K8sNamespace, ins.K8sDaemonSet)
				if err != nil {
					l.Debugf("autodiscovery: not found DaemonSet %s LabelSelector, error %s, namespace %s", ins.K8sDaemonSet, err, ins.K8sNamespace)
				} else {
					t := d.getPodsFromLabelSelector(ins.K8sNamespace, ins.K8sDaemonSet, selector)
					pods = append(pods, t...)
				}
			}

			if ins.K8sDeployment != "" {
				selector, err := d.getDeploymentLabelSelector(ins.K8sNamespace, ins.K8sDeployment)
				if err != nil {
					l.Debugf("autodiscovery: not found Deployment %s LabelSelector, error %s, namespace %s", ins.K8sDeployment, err, ins.K8sNamespace)
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

func (d *discovery) getServicesFromLabelSelector(namespace, appName string, selector *metav1.LabelSelector) (res []*serviceMeta) {
	if selector == nil {
		return nil
	}
	opt := metav1.ListOptions{
		LabelSelector: newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String(),
	}

	services, err := d.client.getServicesForNamespace(namespace).List(context.Background(), opt)
	if err != nil {
		l.Warnf("autodiscovery: failed to get services from node_name %s, namespace %s, app %s, err: %s, retry in 3 minute",
			d.localNodeName, namespace, appName, err)
		return
	}
	for idx := range services.Items {
		res = append(res, &serviceMeta{Service: &services.Items[idx]})
	}
	return
}

func (d *discovery) getPodsFromLabelSelector(namespace, appName string, selector *metav1.LabelSelector) (res []*podMeta) {
	if selector == nil {
		return nil
	}
	opt := metav1.ListOptions{
		LabelSelector: newLabelSelector(selector.MatchLabels, selector.MatchExpressions).String(),
		FieldSelector: "spec.nodeName=" + d.localNodeName,
	}
	pods, err := d.client.getPodsForNamespace(namespace).List(context.Background(), opt)
	if err != nil {
		l.Warnf("autodiscovery: failed to get pods from node_name %s, namespace %s, app %s, err: %s, retry in 3 minute",
			d.localNodeName, namespace, appName, err)
		return
	}
	for idx := range pods.Items {
		res = append(res, &podMeta{Pod: &pods.Items[idx]})
	}
	return
}

func (d *discovery) getDaemonSetLabelSelector(namespace, daemonset string) (*metav1.LabelSelector, error) {
	daemonsetObj, err := d.client.getDaemonSetsForNamespace(namespace).Get(context.Background(), daemonset, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if daemonsetObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid DaemonSet LabelSelector")
	}
	return daemonsetObj.Spec.Selector, nil
}

func (d *discovery) getDeploymentLabelSelector(namespace, deployment string) (*metav1.LabelSelector, error) {
	deploymentObj, err := d.client.getDeploymentsForNamespace(namespace).Get(context.Background(), deployment, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if deploymentObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid Deployment LabelSelector")
	}
	return deploymentObj.Spec.Selector, nil
}

func newPromRunnersForPod(item *podMeta, inputConfig string, extraTags map[string]string, extractK8sLabelAsTags bool) ([]*promRunner, error) {
	l.Debugf("autodiscovery: new runner, source: %s, config: %s", item.Name, inputConfig)

	runners, err := newPromRunnerWithTomlConfig(completePromConfig(inputConfig, item))
	if err != nil {
		return nil, err
	}

	for _, runner := range runners {
		runner.addTags(extraTags)
		// extract pod labels as tags
		if extractK8sLabelAsTags {
			runner.addTags(item.Labels)
		}
	}

	return runners, nil
}

func completePromConfig(config string, item *podMeta) string {
	podIP := item.Status.PodIP

	// 从 ip 列表中使用 index 获取 ip
	func() {
		indexStr, ok := item.Annotations[annotationPromIPIndex]
		if !ok {
			return
		}
		idx, err := strconv.Atoi(indexStr)
		if err != nil {
			l.Warnf("autodiscovery: source %s annotation prom.ip_index parse err: %s", item.Name, err)
			return
		}
		if !(0 <= idx && idx < len(item.Status.PodIPs)) {
			l.Warnf("autodiscovery: source %s annotation prom.ip_index %d outrange, len(PodIPs) %d", item.Name, idx, len(item.Status.PodIPs))
			return
		}
		podIP = item.Status.PodIPs[idx].IP
	}()

	config = strings.ReplaceAll(config, "$IP", podIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", item.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", item.Name)
	config = strings.ReplaceAll(config, "$NODENAME", item.Spec.NodeName)

	return config
}

func getLocalNodeName() (string, error) {
	var e string
	if os.Getenv("NODE_NAME") != "" {
		e = os.Getenv("NODE_NAME")
	}
	if os.Getenv("ENV_K8S_NODE_NAME") != "" {
		e = os.Getenv("ENV_K8S_NODE_NAME")
	}
	if e != "" {
		return e, nil
	}
	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}

func parseScrapeFromProm(scrape string) bool {
	if scrape == "" {
		return false
	}
	b, _ := strconv.ParseBool(scrape)
	return b
}
