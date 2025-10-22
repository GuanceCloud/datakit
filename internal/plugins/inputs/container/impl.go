// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
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
		if k8sCollector, err := newK8sCollectors(ipt); err != nil {
			l.Errorf("failed to init k8s collector: %s", err)
		} else {
			collectors = append(collectors, k8sCollector)
		}
	}

	return collectors
}

func newContainerCollectors(ipt *Input) []Collector {
	var collectors []Collector

	if config.IsECSFargate() {
		collector, err := createECSFargateCollector(ipt)
		if err != nil {
			l.Errorf("failed to create ECS Fargate collector: %s", err)
			return collectors
		}
		collectors = append(collectors, collector)
		return collectors
	}

	logCoordinator := newLogCoordinator(ipt)
	k8sClient := createK8sClientIfNeeded(ipt, logCoordinator)

	for _, endpoint := range ipt.Endpoints {
		collector, err := createContainerCollector(ipt, endpoint, k8sClient, logCoordinator)
		if err != nil {
			l.Warnf("failed to create collector for endpoint %s: %s", endpoint, err)
			continue
		}
		collectors = append(collectors, collector)
	}

	return collectors
}

func newK8sCollectors(ipt *Input) (Collector, error) {
	client, err := k8sclient.NewKubernetesClientInCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	cfg := buildK8sConfig(ipt)
	return kubernetes.NewKubeCollector(client, &cfg, ipt.chPause)
}

func createECSFargateCollector(ipt *Input) (Collector, error) {
	var baseURL string

	if v4 := config.ECSFargateBaseURIV4(); v4 != "" {
		baseURL = v4
		l.Infof("connecting to ECS Fargate v4 at %s", v4)
	} else if v3 := config.ECSFargateBaseURIV3(); v3 != "" {
		baseURL = v3
		l.Infof("connecting to ECS Fargate v3 at %s", v3)
	} else {
		return nil, fmt.Errorf("no valid ECS Fargate URL found, only v3 and v4 are supported")
	}

	collector, err := newECSFargate(ipt, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECS Fargate collector: %w", err)
	}

	return collector, nil
}

func createK8sClientIfNeeded(ipt *Input, logCoordinator *containerLogCoordinator) k8sclient.Client {
	if !datakit.Docker || !config.IsKubernetes() {
		return nil
	}

	client, err := k8sclient.NewKubernetesClientInCluster()
	if err != nil {
		l.Warnf("unable to connect k8s client: %s", err)
		return nil
	}

	crdWatcherG.Go(func(_ context.Context) error {
		startLoggingConfigWatcher(client, logCoordinator)
		return nil
	})

	return client
}

func createContainerCollector(ipt *Input, endpoint string, k8sClient k8sclient.Client, logCoordinator *containerLogCoordinator) (Collector, error) {
	if err := checkEndpoint(endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint %s: %w", endpoint, err)
	}

	collector, err := newContainerCollector(ipt, endpoint, getMountPoint(), k8sClient, logCoordinator)
	if err != nil {
		return nil, fmt.Errorf("failed to create container collector: %w", err)
	}

	l.Infof("connected to runtime at %s", endpoint)
	return collector, nil
}

func buildK8sConfig(ipt *Input) kubernetes.Config {
	tags := inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, "")
	if name := getClusterNameK8s(); name != "" {
		tags["cluster_name_k8s"] = name
	}

	labelOptions := buildLabelOptions(ipt)
	podFilter := createPodFilter(ipt)

	l.Infof("use labels %s for k8s non-metric", labelOptions.nonMetric.keys)
	l.Infof("use labels %s for k8s metric", labelOptions.metric.keys)

	return kubernetes.Config{
		NodeLocal:        ipt.EnableK8sNodeLocal,
		EnableK8sMetric:  ipt.EnableK8sMetric,
		EnableK8sObject:  true,
		EnablePodMetric:  ipt.EnablePodMetric,
		EnableK8sEvent:   ipt.EnableK8sEvent,
		EnableCollectJob: ipt.EnableCollectK8sJob,

		MetricCollecInterval: ipt.MetricCollecInterval,
		ObjectCollecInterval: ipt.ObjectCollecInterval,

		PodFilterForMetric:            podFilter,
		EnableExtractK8sLabelAsTagsV1: ipt.EnableExtractK8sLabelAsTags,
		LabelAsTagsForMetric: kubernetes.LabelsOption{
			All:  labelOptions.metric.all,
			Keys: labelOptions.metric.keys,
		},
		LabelAsTagsForNonMetric: kubernetes.LabelsOption{
			All:  labelOptions.nonMetric.all,
			Keys: labelOptions.nonMetric.keys,
		},
		ExtraTags: tags,
		Feeder:    ipt.Feeder,
	}
}

func createPodFilter(ipt *Input) filter.Filter {
	if len(ipt.PodIncludeMetric) == 0 && len(ipt.PodExcludeMetric) == 0 {
		return nil
	}

	podFilter, err := filter.NewFilter(ipt.PodIncludeMetric, ipt.PodExcludeMetric)
	if err != nil {
		l.Warnf("failed to create pod filter: %s", err)
		return nil
	}
	return podFilter
}

func newLogCoordinator(ipt *Input) *containerLogCoordinator {
	defaults := newLoggingDefaults(ipt)
	return newContainerLogCoordinator(defaults)
}
