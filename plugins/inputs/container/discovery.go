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
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
	"golang.org/x/exp/slices"
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

	defaultPrometheusioInterval = "60s"
	defaultPromScheme           = "http"
	defaultPromPath             = "/metrics"
)

type discovery struct {
	client                k8sClientX
	extraTags             map[string]string
	localNodeName         string
	extractK8sLabelAsTags bool

	enablePrometheusServiceAnnotations bool
	enablePrometheusPodMonitors        bool
	enablePrometheusServiceMonitors    bool

	prometheusMonitoringExtraConfig *prometheusMonitoringExtraConfig

	pause   bool
	chPause chan bool
	done    <-chan interface{}
}

var globalCRDLogsConfList = struct {
	list map[string]string
	mu   sync.Mutex
}{
	make(map[string]string),
	sync.Mutex{},
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
		runners         []*discoveryRunner
		electionRunners []*discoveryRunner
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
			r.collect()
		}

		if d.pause {
			l.Debug("not leader, skipped")
		} else if d.election() {
			for _, r := range electionRunners {
				r.collect()
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

func (d *discovery) updateRunners() []*discoveryRunner {
	runners := []*discoveryRunner{}
	runners = append(runners, d.fetchPromInputsFromPodAnnotations()...)
	runners = append(runners, d.fetchInputsFromDatakitCRD()...)
	if d.enablePrometheusPodMonitors {
		runners = append(runners, d.fetchPromInputsForPodMonitors()...)
	}
	return runners
}

func (d *discovery) updateElectionRunners() []*discoveryRunner {
	runners := []*discoveryRunner{}
	if d.enablePrometheusServiceAnnotations {
		runners = append(runners, d.fetchPromInputsFromService()...)
	}
	if d.enablePrometheusServiceMonitors {
		runners = append(runners, d.fetchPromInputsForServiceMonitors()...)
	}
	return runners
}

func (d *discovery) fetchPromInputsFromService() []*discoveryRunner {
	var res []*discoveryRunner

	list, err := d.client.getServices().List(context.Background(), metaV1ListOption)
	if err != nil {
		l.Warnf("autodiscovery: failed to get service, err: %s, retry in a minute", d.localNodeName, err)
		return nil
	}

	for _, item := range list.Items {
		scrape, ok := item.Annotations[annotationPrometheusioScrape]
		if !ok {
			continue
		}
		b, err := strconv.ParseBool(scrape)
		if err != nil {
			l.Warnf("autodiscovery: service %s parse %s err: %s, skip", item.Name, scrape, err)
			continue
		}
		if !b {
			continue
		}

		port, ok := item.Annotations[annotationPrometheusioPort]
		if !ok {
			l.Warnf("autodiscovery: not found port from service %s, skip", item.Name)
			continue
		}
		if _, err := strconv.Atoi(port); err != nil {
			l.Warnf("autodiscovery: invalid port %s from service %s, skip", port, item.Name)
			continue
		}

		u := url.URL{
			Scheme: defaultPromScheme,
			Path:   defaultPromPath,
			Host:   fmt.Sprintf("%s.%s:%s", item.Name, item.Namespace, port), // service_name.service_namespace:port
		}

		if s, ok := item.Annotations[annotationPrometheusioScheme]; ok {
			if s == "https" {
				u.Scheme = s
			} else {
				l.Warnf("autodiscovery: invalid scheme %s from service %s, use default 'http'", s, item.Name)
			}
		}

		if p, ok := item.Annotations[annotationPrometheusioPath]; ok {
			u.Path = p
		}

		l.Infof("autodiscovery: new promInput for service %s, url %s", item.Name, u.String())

		promInput := prom.NewProm()
		promInput.Source = "k8s.service/" + item.Name
		promInput.Interval = defaultPrometheusioInterval
		promInput.Election = false
		promInput.MetricTypes = []string{"counter", "gauge"}
		promInput.URLs = []string{u.String()}
		for k, v := range d.extraTags {
			promInput.Tags[k] = v
		}
		promInput.Tags["namespace"] = item.Namespace
		promInput.Tags["service"] = item.Name

		res = append(res, &discoveryRunner{
			runner:   promInput,
			source:   item.Name,
			lastTime: time.Now(),
		})
	}

	return res
}

func (d *discovery) fetchPromInputsFromPodAnnotations() []*discoveryRunner {
	var res []*discoveryRunner

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

		runner, err := newDiscoveryRunnersForPod(&podMeta{Pod: &list.Items[idx]}, cfg, d.extraTags, d.extractK8sLabelAsTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner err %s", err)
			continue
		}

		res = append(res, runner...)
	}

	return res
}

func (d *discovery) fetchInputsFromDatakitCRD() []*discoveryRunner {
	var res []*discoveryRunner

	fn := func(ins kubev1guancebeta1.DatakitInstance, pod *podMeta) {
		l.Debugf("autodiscovery: find CRD inputConf, pod_name: %s, pod_namespace: %s, conf: %s", pod.Name, pod.Namespace, ins.InputConf)
		runner, err := newDiscoveryRunnersForPod(pod, ins.InputConf, d.extraTags, d.extractK8sLabelAsTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner from crd, err: %s", err)
			return
		}
		res = append(res, runner...)
	}

	err := d.processCRDWithPod(fn)
	if err != nil {
		l.Debugf("autodiscovery: failed to get datakits, err: %s, retry in a minute", err)
		return nil
	}

	return res
}

func (d *discovery) fetchPromInputsForPodMonitors() []*discoveryRunner {
	var res []*discoveryRunner

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

		for _, pod := range pods {
			for _, metricsEndpoints := range item.Spec.PodMetricsEndpoints {
				u := url.URL{
					Scheme:   defaultPromScheme,
					Path:     defaultPromPath,
					RawQuery: url.Values(metricsEndpoints.Params).Encode(),
				}
				if metricsEndpoints.Scheme != "" {
					u.Scheme = metricsEndpoints.Scheme
				}
				if metricsEndpoints.Path != "" {
					u.Path = metricsEndpoints.Path
				}

				if metricsEndpoints.Port != "" {
					port := pod.containerPort(metricsEndpoints.Port)
					if port == -1 {
						l.Warnf("autodiscovery: not found port %s for podMonitor %s podName %s, ignored", metricsEndpoints.Port, item.Name, pod.Name)
						continue
					}
					u.Host = fmt.Sprintf("%s:%d", pod.Status.PodIP, port)
				} else {
					//nolint
					// deprecated
					if metricsEndpoints.TargetPort != nil {
						u.Host = fmt.Sprintf("%s:%d", pod.Status.PodIP, metricsEndpoints.TargetPort.IntVal)
					} else {
						l.Warnf("autodiscovery: not found port for podMonitor %s podName %s, ignored", item.Name, pod.Name)
						continue
					}
				}

				l.Infof("autodiscovery: new promInput for podMonitor %s podName %s, url %s", item.Name, pod.Name, u.String())

				promInput := prom.NewProm()
				promInput.Source = fmt.Sprintf("k8s.podMonitor/%s::%s", item.Name, pod.Name)
				promInput.URLs = []string{u.String()}
				for k, v := range d.extraTags {
					promInput.Tags[k] = v
				}
				if metricsEndpoints.Interval != "" {
					promInput.Interval = metricsEndpoints.Interval
				}
				promInput.Tags["namespace"] = pod.Namespace
				promInput.Tags["service"] = pod.Name

				if d.prometheusMonitoringExtraConfig != nil {
					l.Debugf("autodiscovery: matching promConfig %#v", d.prometheusMonitoringExtraConfig)
					promConfig := d.prometheusMonitoringExtraConfig.matchPromConfig(pod.Labels, pod.Namespace)
					if promConfig != nil {
						l.Debugf("autodiscovery: use promConfig %#v for podMonitor %s podName %s", promConfig, item.Name, pod.Name)
						promInput = mergePromConfig(promInput, promConfig)
					}
				}

				l.Infof("autodiscovery: new promInput for podMonitor %s podName %s, config: %#v", item.Name, pod.Name, promInput)

				res = append(res, &discoveryRunner{
					runner:   promInput,
					source:   item.Name + "::" + pod.Name,
					lastTime: time.Now(),
				})
			}
		}
	}

	return res
}

func (d *discovery) fetchPromInputsForServiceMonitors() []*discoveryRunner {
	var res []*discoveryRunner

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

		for _, service := range services {
			for _, endpoint := range item.Spec.Endpoints {
				u := url.URL{
					Scheme:   defaultPromScheme,
					Path:     defaultPromPath,
					RawQuery: url.Values(endpoint.Params).Encode(),
				}
				if endpoint.Scheme != "" {
					u.Scheme = endpoint.Scheme
				}
				if endpoint.Path != "" {
					u.Path = endpoint.Path
				}

				if endpoint.Port != "" {
					port := service.servicePort(endpoint.Port)
					if port == -1 {
						l.Warnf("autodiscovery: not found port %s for serviceMonitor %s serviceName %s, ignored", endpoint.Port, item.Name, service.Name)
						continue
					}
					u.Host = fmt.Sprintf("%s.%s:%d", service.Name, service.Namespace, port)
				} else {
					//nolint
					// deprecated
					if endpoint.TargetPort != nil {
						u.Host = fmt.Sprintf("%s.%s:%d", service.Name, service.Namespace, endpoint.TargetPort.IntVal)
					} else {
						l.Warnf("autodiscovery: not found port for serviceMonitor %s serviceName %s, ignored", item.Name, service.Name)
						continue
					}
				}

				l.Infof("autodiscovery: new promInput for serviceMonitor %s serviceName %s, url %s", item.Name, service.Name, u.String())

				promInput := prom.NewProm()
				promInput.Source = fmt.Sprintf("k8s.serviceMonitor/%s::%s", item.Name, service.Name)
				promInput.URLs = []string{u.String()}
				for k, v := range d.extraTags {
					promInput.Tags[k] = v
				}
				if endpoint.Interval != "" {
					promInput.Interval = endpoint.Interval
				}
				promInput.Tags["namespace"] = service.Namespace
				promInput.Tags["service"] = service.Name

				if d.prometheusMonitoringExtraConfig != nil {
					l.Debugf("autodiscovery: matching promConfig %#v", d.prometheusMonitoringExtraConfig)
					promConfig := d.prometheusMonitoringExtraConfig.matchPromConfig(service.Labels, service.Namespace)
					if promConfig != nil {
						l.Debugf("autodiscovery: use promConfig %#v for serviceMonitor %s serviceName %s", promConfig, item.Name, service.Name)
						promInput = mergePromConfig(promInput, promConfig)
					}
				}

				l.Infof("autodiscovery: new promInput for serviceMonitor %s serviceName %s, config: %#v", item.Name, service.Name, promInput)

				res = append(res, &discoveryRunner{
					runner:   promInput,
					source:   item.Name + "::" + service.Name,
					lastTime: time.Now(),
				})
			}
		}
	}

	return res
}

func (d *discovery) updateGlobalCRDLogsConfList() {
	fn := func(ins kubev1guancebeta1.DatakitInstance, pod *podMeta) {
		// 添加到全局 list
		if ins.LogsConf != "" {
			id := string(pod.UID)
			globalCRDLogsConfList.list[id] = ins.LogsConf
		}
	}

	globalCRDLogsConfList.mu.Lock()
	// reset list
	globalCRDLogsConfList.list = make(map[string]string)
	defer globalCRDLogsConfList.mu.Unlock()

	if err := d.processCRDWithPod(fn); err != nil {
		l.Debugf("autodiscovery: failed to get datakits, err: %s, retry in a minute", err)
	}

	l.Debugf("autodiscovery: find CRD datakit/logs len %d, map<uid:conf>: %v", len(globalCRDLogsConfList.list), globalCRDLogsConfList.list)
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
		l.Warnf("autodiscovery: failed to get pods from node_name %s, namespace %s, app %s, err: %s, retry in 3 minute",
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

type discoveryRunner struct {
	source   string
	runner   inputs.InputOnceRunnable
	lastTime time.Time
}

func (r *discoveryRunner) collect() {
	if time.Since(r.lastTime) < r.runner.GetIntervalDuration() {
		return
	}
	l.Debugf("autodiscovery: running collect from source %s", r.source)
	if err := r.runner.RunningCollect(); err != nil {
		l.Warnf("autodiscovery, source %s collect err %s", r.source, err)
	}
	r.lastTime = time.Now()
}

func newDiscoveryRunnersForPod(item *podMeta,
	inputConfig string,
	extraTags map[string]string,
	extractK8sLabelAsTags bool,
) ([]*discoveryRunner, error) {
	l.Debugf("autodiscovery: new runner, source: %s, config: %s", item.Name, inputConfig)

	inputInstances, err := config.LoadSingleConf(completePromConfig(inputConfig, item), inputs.Inputs)
	if err != nil {
		return nil, err
	}

	if len(inputInstances) != 1 {
		l.Warnf("autodiscovery: discover invalid input conf, only 1 type of input allowed in annotation, but got %d, ignored", len(inputInstances))
		return nil, nil
	}

	var inputList []inputs.Input
	for _, arr := range inputInstances {
		inputList = arr
		break // get the first iterate elem in the map
	}

	var res []*discoveryRunner

	for _, ii := range inputList {
		if ii == nil {
			l.Debugf("skip non-datakit-input %s", item.Name)
			continue
		}

		if _, ok := ii.(inputs.InputOnceRunnable); !ok {
			l.Debugf("unknown input type, unreachable")
			continue
		}

		if inp, ok := ii.(inputs.OptionalInput); ok {
			tags := map[string]string{}

			// extract pod labels as tags
			if extractK8sLabelAsTags {
				for k, v := range item.Labels {
					tags[k] = v
				}

				for k, v := range extraTags {
					tags[k] = v
				}
			} else {
				tags = extraTags
			}

			inp.SetTags(tags)
		}

		res = append(res, &discoveryRunner{
			runner:   ii.(inputs.InputOnceRunnable), // 前面有断言判断
			source:   item.Name,
			lastTime: time.Now(),
		})
	}

	return res, nil
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

type promConfig prom.Input

type prometheusMonitoringExtraConfig struct {
	Matches []struct {
		NamespaceSelector struct {
			Any             bool     `json:"any,omitempty"`
			MatchNamespaces []string `json:"matchNamespaces,omitempty"`
		} `json:"namespaceSelector,omitempty"`

		Selector metav1.LabelSelector `json:"selector"`

		PromConfig *promConfig `json:"promConfig"`
	} `json:"matches"`
}

func (p *prometheusMonitoringExtraConfig) matchPromConfig(targetLabels map[string]string, namespace string) *promConfig {
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

func mergePromConfig(c1 *prom.Input, c2 *promConfig) *prom.Input {
	c3 := &prom.Input{}

	c3.URLs = c1.URLs
	c3.Tags = c1.Tags
	c3.MetricTypes = []string{"counter", "gauge"}

	if c1.Interval != "" {
		c3.Interval = c1.Interval
	} else {
		c3.Interval = defaultPrometheusioInterval
	}

	c3.MetricNameFilter = c2.MetricNameFilter
	c3.MeasurementPrefix = c2.MeasurementPrefix
	c3.MeasurementName = c2.MeasurementName
	c3.Measurements = c2.Measurements
	c3.TagsIgnore = c2.TagsIgnore
	c3.TagsRename = c2.TagsRename
	c3.IgnoreReqErr = c2.IgnoreReqErr
	c3.AsLogging = c2.AsLogging
	c3.IgnoreTagKV = c2.IgnoreTagKV
	c3.HTTPHeaders = c2.HTTPHeaders
	c3.Auth = c2.Auth

	if len(c2.MetricTypes) != 0 {
		c3.MetricTypes = c2.MetricTypes
	}

	for k, v := range c2.Tags {
		if _, ok := c3.Tags[k]; !ok {
			c3.Tags[k] = v
		}
	}

	return c3
}
