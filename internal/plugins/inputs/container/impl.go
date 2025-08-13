// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/kubernetes"
)

var l = logger.DefaultSLogger(inputName)

type Collector interface {
	StartCollect()
}

func newCollectors(ipt *Input) []Collector {
	collectors := newContainerCollectors(ipt)

	if datakit.Docker && config.IsKubernetes() {
		k8sCollectors, err := newK8sCollectors(ipt)
		if err != nil {
			l.Errorf("init the k8s fail, err: %s", err)
		} else {
			collectors = append(collectors, k8sCollectors)
		}
	}

	return collectors
}

func newContainerCollectors(ipt *Input) []Collector {
	var collectors []Collector

	if config.IsECSFargate() {
		var baseURL string

		v4 := config.ECSFargateBaseURIV4()
		if v4 != "" {
			baseURL = v4
			l.Infof("connect ecsfargate v4 with url %s", v4)
		} else {
			v3 := config.ECSFargateBaseURIV3()
			if v3 != "" {
				baseURL = v3
				l.Infof("connect ecsfargate v3 with url %s", v4)
			}
		}

		if baseURL != "" {
			collector, err := newECSFargate(ipt, baseURL)
			if err != nil {
				l.Errorf("unable to connect ecsfargate url %s, err: %s", baseURL, err)
			}
			collectors = append(collectors, collector)
		} else {
			l.Errorf("unexpected ecsfargate url, version only be v3 or v4")
		}

		return collectors
	}

	for _, endpoint := range ipt.Endpoints {
		if err := checkEndpoint(endpoint); err != nil {
			l.Warnf("%s, skip", err)
			continue
		}

		var client k8sclient.Client
		var err error
		if datakit.Docker && config.IsKubernetes() {
			client, err = k8sclient.NewKubernetesClientInCluster()
			if err != nil {
				l.Warnf("unable to connect k8s client, err: %s, skip", err)
			}
		}

		collector, err := newContainerCollector(ipt, endpoint, getMountPoint(), client)
		if err != nil {
			l.Warnf("cannot connect endpoint(%s), err: %s", endpoint, err)
			continue
		}

		l.Infof("connect runtime with %s", endpoint)
		collectors = append(collectors, collector)
	}

	return collectors
}

func newK8sCollectors(ipt *Input) (Collector, error) {
	client, err := k8sclient.NewKubernetesClientInCluster()
	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, "")
	if name := getClusterNameK8s(); name != "" {
		tags["cluster_name_k8s"] = name
	}

	optForNonMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys)
	optForMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2ForMetric, config.Cfg.Dataway.GlobalCustomerKeys)

	var podFilterForMetric filter.Filter
	if len(ipt.PodIncludeMetric) != 0 || len(ipt.PodExcludeMetric) != 0 {
		podFilter, err := filter.NewFilter(ipt.PodIncludeMetric, ipt.PodExcludeMetric)
		if err != nil {
			return nil, fmt.Errorf("new k8s collector failed, err: %w", err)
		}
		podFilterForMetric = podFilter
	}

	l.Infof("Use labels %s for k8s non-metric", optForNonMetric.keys)
	l.Infof("Use labels %s for k8s metric", optForMetric.keys)

	cfg := kubernetes.Config{
		NodeLocal:        ipt.EnableK8sNodeLocal,
		EnableK8sMetric:  ipt.EnableK8sMetric,
		EnableK8sObject:  true,
		EnablePodMetric:  ipt.EnablePodMetric,
		EnableK8sEvent:   ipt.EnableK8sEvent,
		EnableCollectJob: ipt.EnableCollectK8sJob,

		MetricCollecInterval: ipt.MetricCollecInterval,
		ObjectCollecInterval: ipt.ObjectCollecInterval,

		PodFilterForMetric:            podFilterForMetric,
		EnableExtractK8sLabelAsTagsV1: ipt.EnableExtractK8sLabelAsTags,
		LabelAsTagsForMetric: kubernetes.LabelsOption{
			All:  optForMetric.all,
			Keys: optForMetric.keys,
		},
		LabelAsTagsForNonMetric: kubernetes.LabelsOption{
			All:  optForNonMetric.all,
			Keys: optForNonMetric.keys,
		},
		ExtraTags: tags,
		Feeder:    ipt.Feeder,
	}

	return kubernetes.NewKubeCollector(client, &cfg, ipt.chPause)
}
