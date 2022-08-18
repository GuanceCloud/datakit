// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubev1guancebeta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
)

const (
	annotationPromExport  = "datakit/prom.instances"
	annotationPromIPIndex = "datakit/prom.instances.ip_index"
)

type discovery struct {
	client        k8sClientX
	extraTags     map[string]string
	localNodeName string
}

var globalCRDLogsConfList = struct {
	list map[string]string
	mu   sync.Mutex
}{
	make(map[string]string),
	sync.Mutex{},
}

func newDiscovery(client k8sClientX, extraTags map[string]string) *discovery {
	return &discovery{
		client:    client,
		extraTags: extraTags,
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

	runners := d.updateRunners()
	l.Infof("autodiscovery: update input list, len %d", len(runners))

	updateTicker := time.NewTicker(time.Minute * 3)
	defer updateTicker.Stop()

	collectTicker := time.NewTicker(time.Second * 1)
	defer collectTicker.Stop()

	l.Infof("start k8s autodiscovery, node_name is %s", localNodeName)

	for {
		for _, r := range runners {
			if time.Since(r.lastTime) < r.runner.GetIntervalDuration() {
				continue
			}
			l.Debugf("autodiscovery: running collect from source %s", r.source)
			if err := r.runner.RunningCollect(); err != nil {
				l.Warnf("autodiscovery, source %s collect err %s", r.source, err)
			}
			r.lastTime = time.Now()
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("stop k8s autodiscovery")
			return

		case <-updateTicker.C:
			runners = d.updateRunners()
			l.Infof("autodiscovery: update input list, len %d", len(runners))

		case <-collectTicker.C:
			// nil
		}
	}
}

func (d *discovery) updateRunners() []*discoveryRunner {
	runners := []*discoveryRunner{}
	runners = append(runners, d.fetchPromInputs()...)
	runners = append(runners, d.fetchDatakitCRDInputs()...)
	return runners
}

func (d *discovery) fetchPromInputs() []*discoveryRunner {
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

		runner, err := newDiscoveryRunner(&podMeta{Pod: &list.Items[idx]}, cfg, d.extraTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner err %s", err)
			continue
		}

		res = append(res, runner...)
	}

	return res
}

func (d *discovery) fetchDatakitCRDInputs() []*discoveryRunner {
	var res []*discoveryRunner

	fn := func(ins kubev1guancebeta1.DatakitInstance, pod *podMeta) {
		l.Debugf("autodiscovery: find CRD inputConf, pod_name: %s, pod_namespace: %s, conf: %s", pod.Name, pod.Namespace, ins.InputConf)
		runner, err := newDiscoveryRunner(pod, ins.InputConf, d.extraTags)
		if err != nil {
			l.Warnf("autodiscovery: new runner from crd, err: %s", err)
			return
		}
		res = append(res, runner...)
	}

	err := d.processCRDWithPod(fn)
	if err != nil {
		// TODO:
		// 对 error 内容进行子串判定，不再打印这个错误
		// 避免因为 k8s 客户端没有 datakits resource 而获取失败，频繁报错
		// “could not find the requested resource” 为 k8s api 实际返回的 error message，可能会因为版本不同而变更
		if !strings.Contains(err.Error(), "could not find the requested resource") {
			return nil
		}
		l.Warnf("autodiscovery: failed to get datakits, err: %s, retry in a minute", err)
		return nil
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

	err := d.processCRDWithPod(fn)
	if err != nil {
		if !strings.Contains(err.Error(), "could not find the requested resource") {
			return
		}
		l.Warnf("autodiscovery: failed to get datakits, err: %s, retry in a minute", err)
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
			lableSelector, err := d.getDeploymentLabelSelector(ins.K8sNamespace, ins.K8sDeployment)
			if err != nil {
				l.Debugf("autodiscovery: not found deployment %s labelSelector, error %s, namespace %s", ins.K8sDeployment, err, ins.K8sNamespace)
				continue
			}
			opt := metav1.ListOptions{
				LabelSelector: joinLabelSelector(lableSelector),
				FieldSelector: "spec.nodeName=" + d.localNodeName,
			}

			pods, err := d.client.getPodsForNamespace(ins.K8sNamespace).List(context.Background(), opt)
			if err != nil {
				l.Warnf("autodiscovery: failed to get pods from node_name %s, namespace %s, app %s, err: %s, retry in a minute",
					d.localNodeName,
					ins.K8sNamespace,
					ins.K8sDeployment,
					err)
				continue
			}

			for idx := range pods.Items {
				fn(ins, &podMeta{Pod: &pods.Items[idx]})
			}
		}
	}

	return nil
}

func (d *discovery) getDeploymentLabelSelector(namespace, deployment string) (map[string]string, error) {
	deploymentObj, err := d.client.getDeploymentsForNamespace(namespace).Get(context.Background(), deployment, metaV1GetOption)
	if err != nil {
		return nil, err
	}
	if deploymentObj.Spec.Selector == nil {
		return nil, fmt.Errorf("invalid lableSelector")
	}
	return deploymentObj.Spec.Selector.MatchLabels, nil
}

type discoveryRunner struct {
	source   string
	runner   inputs.InputOnceRunnable
	lastTime time.Time
}

func newDiscoveryRunner(item *podMeta, inputConfig string, extraTags map[string]string) ([]*discoveryRunner, error) {
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
			inp.SetTags(extraTags)
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

func joinLabelSelector(m map[string]string) string {
	var res []string
	for k, v := range m {
		res = append(res, k+"="+v)
	}
	return strings.Join(res, ",")
}
