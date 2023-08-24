// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type container struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	loggingFilter filter.Filter
	logTable      *logTable
	extraTags     map[string]string
}

func newContainer(ipt *Input, endpoint string, mountPoint string, k8sClient k8sclient.Client) (Collector, error) {
	filter, err := newFilters(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog)
	if err != nil {
		return nil, err
	}

	var r runtime.ContainerRuntime

	// docker not supported CRI
	if verifyErr := runtime.VerifyDockerRuntime(endpoint); verifyErr == nil {
		r, err = runtime.NewDockerRuntime(endpoint, mountPoint)
	} else {
		r, err = runtime.NewCRIRuntime(endpoint, mountPoint)
	}

	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	return &container{
		ipt:           ipt,
		runtime:       r,
		k8sClient:     k8sClient,
		loggingFilter: filter,
		logTable:      newLogTable(),
		extraTags:     tags,
	}, nil
}

func (c *container) Metric() ([]inputs.Measurement, error) {
	cList, err := c.runtime.ListContainers()
	if err != nil {
		return nil, err
	}

	var res []inputs.Measurement
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "container-metric"})

	for idx := range cList {
		func(info *runtime.Container) {
			g.Go(func(ctx context.Context) error {
				point := c.transToPoint(info)
				mu.Lock()
				res = append(res, &containerMetric{point})
				mu.Unlock()
				return nil
			})
		}(cList[idx])
	}

	_ = g.Wait()

	return res, nil
}

func (c *container) Object() ([]inputs.Measurement, error) {
	cList, err := c.runtime.ListContainers()
	if err != nil {
		return nil, err
	}

	var res []inputs.Measurement
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "container-object"})

	for idx := range cList {
		func(info *runtime.Container) {
			g.Go(func(ctx context.Context) error {
				point := c.transToPoint(info)
				point.SetTag("name", info.ID)
				point.SetField("age", time.Since(time.Unix(0, info.CreatedAt)).Milliseconds()/1e3)

				mu.Lock()
				res = append(res, &containerObject{point})
				mu.Unlock()
				return nil
			})
		}(cList[idx])
	}

	_ = g.Wait()

	return res, nil
}

func (c *container) Logging() error {
	cList, err := c.runtime.ListContainers()
	if len(cList) == 0 && err != nil {
		return err
	}

	var newIDs []string
	for _, info := range cList {
		newIDs = append(newIDs, info.ID)
	}
	c.cleanMissingContainerLog(newIDs)

	for _, info := range cList {
		l.Debugf("find container %s info: %#v", info.Name, info)

		instance := c.queryContainerLogInfo(info)
		if instance == nil {
			continue
		}

		if err := instance.parseLogConfigs(); err != nil {
			l.Warn(err)
			continue
		}

		if !c.shouldPullContainerLog(instance) {
			continue
		}

		instance.addStdout()
		instance.fillLogType(info.RuntimeName)
		instance.setTagsToLogConfigs(instance.tags())
		instance.setTagsToLogConfigs(c.extraTags)
		instance.setCustomerTags(instance.podLabels, config.Cfg.Dataway.GlobalCustomerKeys)

		c.ipt.setLoggingExtraSourceMapToLogConfigs(instance.configs)
		c.ipt.setLoggingSourceMultilineMapToLogConfigs(instance.configs)
		c.ipt.setLoggingAutoMultilineToLogConfigs(instance.configs)

		c.tailingLogs(instance)
	}

	l.Infof("current container logtable: %s", c.logTable.String())

	return nil
}

func (c *container) Name() string {
	return "container"
}

func (c *container) shouldPullContainerLog(ins *logInstance) bool {
	if ins.enabled() {
		return true
	}
	if c.ignoreImageForLogging(ins.image) {
		return false
	}
	return true
}

func (c *container) transToPoint(info *runtime.Container) typed.PointKV {
	p := typed.NewPointKV()
	p.SetTag("container_id", info.ID)
	p.SetTag("container_runtime", info.RuntimeName)
	p.SetTag("state", info.State)

	if info.Name != "" {
		p.SetTag("container_runtime_name", info.Name)
	} else {
		p.SetTag("container_runtime_name", "unknown")
	}

	if name := getContainerNameForLabels(info.Labels); name != "" {
		p.SetTag("container_name", name)
	} else {
		p.SetTag("container_name", info.Name)
	}

	if containerIsFromKubernetes(info.Labels) {
		p.SetTag("container_type", "kubernetes")
	} else {
		p.SetTag("container_type", info.RuntimeName)
	}

	top, err := c.runtime.ContainerTop(info.ID)
	if err != nil {
		l.Warnf("unable to query container stats, err: %s", err)
	} else {
		p.SetField("cpu_usage", top.CPUUsage)
		p.SetField("cpu_numbers", top.CPUCores)
		if top.CPUCores != 0 {
			p.SetField("cpu_usage_base100", top.CPUUsage/float64(top.CPUCores))
		}

		p.SetField("mem_usage", top.MemoryWorkingSet)

		if top.MemoryLimit != 0 {
			p.SetField("mem_limit", top.MemoryLimit)
			p.SetField("mem_used_percent_base_limit", float64(top.MemoryWorkingSet)/float64(top.MemoryLimit)*100)
		}

		if top.MemoryCapacity != 0 {
			p.SetField("mem_capacity", top.MemoryCapacity)
			p.SetField("mem_used_percent", float64(top.MemoryWorkingSet)/float64(top.MemoryCapacity)*100)
		}

		p.SetField("network_bytes_rcvd", top.NetworkRcvd)
		p.SetField("network_bytes_sent", top.NetworkSent)

		// not supported containerd/CRI
		if top.BlockRead != 0 {
			p.SetField("block_read_byte", top.BlockRead)
		}
		if top.BlockWrite != 0 {
			p.SetField("block_write_byte", top.BlockWrite)
		}
	}

	if uid := getPodUIDForLabels(info.Labels); uid != "" {
		p.SetTag("pod_uid", uid)
	}

	podName := getPodNameForLabels(info.Labels)
	if podName != "" {
		p.SetTag("pod_name", podName)
	}
	namespace := getPodNamespaceForLabels(info.Labels)
	if namespace != "" {
		p.SetTag("namespace", namespace)
	}

	image := info.Image

	if c.k8sClient != nil && podName != "" {
		podInfo, err := c.queryPodInfo(context.Background(), podName, namespace)
		if err != nil {
			l.Warnf("container %s, %s", info.Name, err)
		} else {
			ownerKind, ownerName := podInfo.owner()
			if ownerKind != "" && ownerName != "" {
				p.SetTag(ownerKind, ownerName)
			}

			// use Image from Pod Container
			img := podInfo.containerImage(getContainerNameForLabels(info.Labels))
			if img != "" {
				image = img
			}

			p.SetCustomerTags(podInfo.pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)
		}
	}

	imageName, shortName, imageTag := runtime.ParseImage(image)
	p.SetTag("image", image)
	p.SetTag("image_name", imageName)
	p.SetTag("image_short_name", shortName)
	p.SetTag("image_tag", imageTag)

	p.SetTags(c.extraTags)
	return p
}

func newFilters(include, exclude []string) (filter.Filter, error) {
	in := splitRules(include)
	ex := splitRules(exclude)
	return filter.NewIncludeExcludeFilter(in, ex)
}

func (c *container) ignoreImageForLogging(image string) (ignore bool) {
	if c.loggingFilter == nil {
		return
	}
	return !c.loggingFilter.Match(image)
}

func getPodNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.name"]
}

func getPodUIDForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.uid"]
}

func getPodNamespaceForLabels(labels map[string]string) string {
	return labels["io.kubernetes.pod.namespace"]
}

func getContainerNameForLabels(labels map[string]string) string {
	return labels["io.kubernetes.container.name"]
}

// splitRules
//
//	split 'image:' kv，return values
//	ex, in: ["image:img_*", "image:img01*", "xx:xx"] return: ["img_*", "img01*"]
func splitRules(arr []string) (rules []string) {
	for _, str := range arr {
		x := strings.Split(str, ":")
		if len(x) != 2 {
			continue
		}
		if strings.HasPrefix(str, "image:") {
			rules = append(rules, x[1])
		}
	}
	return
}

func containerIsFromKubernetes(labels map[string]string) bool {
	uid, ok := labels["io.kubernetes.pod.uid"]
	return ok && uid != ""
}
