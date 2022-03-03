package container

import (
	"context"

	//nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

	if !shouldForkInput(item.Spec.NodeName) {
		l.Debugf("should not fork input, pod-nodeName:%s", item.Spec.NodeName)
		return nil
	}

	md5str := md5sum(config)
	if _, ok := discoveryInputsMap[md5str]; ok {
		return nil
	}

	instance := discoveryInput{
		name:      "prom",
		config:    complatePromConfig(config, item),
		configMD5: md5str,
		extraTags: map[string]string{
			"pod_name":  item.Name,
			"node_name": item.Spec.NodeName,
			"namespace": defaultNamespace(item.Namespace),
		},
	}

	return instance.run()
}

// map[md5sum(cfg)] = nil.
var discoveryInputsMap = make(map[string]interface{})

type discoveryInput struct {
	name      string
	config    string
	configMD5 string
	extraTags map[string]string
}

func (d *discoveryInput) run() error {
	creator, ok := inputs.Inputs["prom"]
	if !ok {
		return fmt.Errorf("unreachable, invalid inputName")
	}

	inputList, err := config.LoadInputConfig(d.config, creator)
	if err != nil {
		return err
	}

	// add to inputsMap
	discoveryInputsMap[d.configMD5] = nil

	l.Infof("discovery: add %s inputs, len %d", d.name, len(inputList))

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

func shouldForkInput(nodeName string) bool {
	if !datakit.Docker {
		return true
	}
	// ENV NODE_NAME 在 daemonset.yaml 配置，是当前程序所在的 Node 名称
	// Node 名称匹配，表示运行在同一个 Node，此时才需要 fork

	// Node 名称为空属于 unreachable
	return datakit.GetEnv("NODE_NAME") == nodeName
}

func md5sum(str string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
