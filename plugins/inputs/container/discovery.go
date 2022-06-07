// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"

	"strconv"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	v1 "k8s.io/api/core/v1"
)

const (
	annotationPromExport  = "datakit/prom.instances"
	annotationPromIPIndex = "datakit/prom.instances.ip_index"
)

func tryRunInput(item *v1.Pod) error {
	config, ok := item.Annotations[annotationPromExport]
	if !ok {
		return nil
	}

	l.Debugf("k8s export, find prom export, config: %s", config)

	if _, ok := discoveryInputsMap[string(item.UID)]; ok {
		return nil
	}

	instance := discoveryInput{
		id:     string(item.UID),
		name:   "prom",
		config: complatePromConfig(config, item),
		extraTags: map[string]string{
			"pod_name":  item.Name,
			"node_name": item.Spec.NodeName,
			"namespace": defaultNamespace(item.Namespace),
		},
	}

	return instance.run()
}

var discoveryInputsMap = make(map[string][]inputs.Input)

type discoveryInput struct {
	id        string
	name      string
	config    string
	configMD5 string
	extraTags map[string]string
}

func (d *discoveryInput) run() error {
	inputInstances, err := config.LoadSingleConf(d.config, inputs.Inputs)
	if err != nil {
		return err
	}

	if len(inputInstances) != 1 {
		l.Warnf("discover invalid input conf, only 1 type of input allowed in annotation, but got %d, ignored", len(inputInstances))
		return nil
	}

	var inputList []inputs.Input
	for _, arr := range inputInstances {
		inputList = arr
		break // get the first iterate elem in the map
	}
	// add to inputsMap
	discoveryInputsMap[d.id] = inputList

	l.Infof("discovery: add %s inputs, len %d", d.name, len(inputList))
	// add to inputsMap
	discoveryInputsMap[d.configMD5] = nil

	// input run() 不受全局 election 影响
	// election 模块运行在此之前，且其列表是固定的
	g := datakit.G("kubernetes-autodiscovery")
	for _, ii := range inputList {
		if ii == nil {
			l.Debugf("skip non-datakit-input %s", d.name)
			continue
		}

		if inp, ok := ii.(inputs.OptionalInput); ok {
			inp.SetTags(d.extraTags)
		}

		func(name string, ii inputs.Input) {
			g.Go(func(ctx context.Context) error {
				l.Infof("discovery: starting input %s ...", name)
				// main
				ii.Run()
				l.Infof("discovery: input %s exited", d.name)
				return nil
			})
		}(d.name, ii)
	}

	return nil
}

func complatePromConfig(config string, podObj *v1.Pod) string {
	podIP := podObj.Status.PodIP

	func() {
		indexStr, ok := podObj.Annotations[annotationPromIPIndex]
		if !ok {
			return
		}
		idx, err := strconv.Atoi(indexStr)
		if err != nil {
			l.Warnf("annotation prom.ip_index parse err: %s", err)
			return
		}
		if !(0 <= idx && idx < len(podObj.Status.PodIPs)) {
			l.Warnf("annotation prom.ip_index %d outrange, len(PodIPs) %d", idx, len(podObj.Status.PodIPs))
			return
		}
		podIP = podObj.Status.PodIPs[idx].IP
	}()

	config = strings.ReplaceAll(config, "$IP", podIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", podObj.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", podObj.Name)

	return config
}
