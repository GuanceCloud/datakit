// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/option"
)

type container struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	nodeName                      string
	enableCollectLogging          bool
	enableExtractK8sLabelAsTagsV1 bool
	podLabelAsTagsForNonMetric    labelsOption
	podLabelAsTagsForMetric       labelsOption

	maxConcurrent int
	logFilter     filter.Filter
	logTable      *logTable
	extraTags     map[string]string
}

func newECSFargate(ipt *Input, agentURL string) (Collector, error) {
	r, err := runtime.NewECSFargateRuntime(agentURL)
	if err != nil {
		return nil, err
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	return &container{
		ipt:                  ipt,
		runtime:              r,
		enableCollectLogging: false,
		extraTags:            tags,
	}, nil
}

var containerExistList sync.Map

func newContainer(ipt *Input, endpoint string, mountPoint string, k8sClient k8sclient.Client) (Collector, error) {
	logFilter, err := filter.NewFilter(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog)
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

	version, err := r.Version()
	if err != nil {
		return nil, fmt.Errorf("get runtime version err: %w", err)
	}
	l.Infof("runtime platform %s, api-version %s", version.PlatformName, version.APIVersion)

	key := fmt.Sprintf("%s:%s", version.PlatformName, version.APIVersion)
	if _, exist := containerExistList.Load(key); exist {
		return nil, fmt.Errorf("runtime %s already exists", key)
	} else {
		containerExistList.Store(key, nil)
	}

	tags := inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")

	optForNonMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2, config.Cfg.Dataway.GlobalCustomerKeys)
	optForMetric := buildLabelsOption(ipt.ExtractK8sLabelAsTagsV2ForMetric, config.Cfg.Dataway.GlobalCustomerKeys)

	maxConcurrent := datakit.AvailableCPUs + 1
	if ipt.ContainerMaxConcurrent > 0 {
		maxConcurrent = ipt.ContainerMaxConcurrent
	}

	return &container{
		ipt:                           ipt,
		runtime:                       r,
		k8sClient:                     k8sClient,
		nodeName:                      config.Cfg.Hostname,
		enableCollectLogging:          true,
		enableExtractK8sLabelAsTagsV1: ipt.DeprecatedEnableExtractK8sLabelAsTags,
		podLabelAsTagsForNonMetric:    optForNonMetric,
		podLabelAsTagsForMetric:       optForMetric,
		maxConcurrent:                 maxConcurrent,
		logFilter:                     logFilter,
		logTable:                      newLogTable(),
		extraTags:                     tags,
	}, nil
}

func (c *container) Metric(feed func(pts []*point.Point) error, opts ...option.CollectOption) {
	if !c.ipt.EnableContainerMetric {
		l.Info("collect container metric: off")
		return
	}

	opt := option.DefaultOption()
	for _, fn := range opts {
		fn(opt)
	}
	if opt.OnlyElection {
		return
	}

	c.gather("metric", feed)
}

func (c *container) Object(feed func(pts []*point.Point) error, opts ...option.CollectOption) {
	opt := option.DefaultOption()
	for _, fn := range opts {
		fn(opt)
	}
	if opt.OnlyElection {
		return
	}

	c.gather("object", feed)
}

// Container collection (Docker/CRI/Fargate) not uses election.
const containerElection = false

func (c *container) Election() bool { return containerElection }

func (c *container) gather(category string, feed func(pts []*point.Point) error) {
	var opts []point.Option

	switch category {
	case "metric":
		opts = point.DefaultMetricOptions()
	case "object":
		opts = point.DefaultObjectOptions()
	default:
		// unreachable
		return
	}

	wrapFeed := func(pts []*point.Point) error {
		err := feed(pts)
		if err == nil {
			collectPtsVec.WithLabelValues(category).Add(float64(len(pts)))
		}
		return err
	}

	start := time.Now()
	if err := c.gatherResource(category, opts, wrapFeed); err != nil {
		l.Errorf("feed container-%s error: %s", category, err.Error())
		c.ipt.Feeder.FeedLastError(err.Error(), metrics.WithLastErrorInput(inputName))
	}
	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

func (c *container) gatherResource(category string, opts []point.Option, feed func(pts []*point.Point) error) error {
	cList, err := c.runtime.ListContainers()
	if err != nil {
		l.Warn(err)

		return nil
	}

	var res []*typed.PointKV
	var mu sync.Mutex

	setPodLabelAsTagsForMetric := func(p *typed.PointKV, labels map[string]string) {
		p.SetLabelAsTags(labels, c.podLabelAsTagsForMetric.all, c.podLabelAsTagsForMetric.keys)
	}
	setPodLabelAsTagsForObject := func(p *typed.PointKV, labels map[string]string) {
		p.SetLabelAsTags(labels, c.podLabelAsTagsForNonMetric.all, c.podLabelAsTagsForNonMetric.keys)
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "container-collect-" + category})
	for idx := range cList {
		if isPauseContainer(cList[idx]) {
			continue
		}

		func(info *runtime.Container) {
			g.Go(func(ctx context.Context) error {
				var pt *typed.PointKV
				switch category {
				case "metric":
					pt = c.transformPoint(info, setPodLabelAsTagsForMetric)
					pt.SetTags(c.extraTags)
				case "object":
					pt = c.transformPoint(info, setPodLabelAsTagsForObject)
					pt.SetTags(c.extraTags)
					pt.SetFields(transLabels(info.Labels))
					pt.SetTag("name", info.ID)
					pt.SetField("age", time.Since(time.Unix(0, info.CreatedAt)).Milliseconds()/1e3)
					pt.SetField("message", typed.TrimString(pt.String(), maxMessageLength))

				default:
					// nil
				}

				if pt != nil {
					mu.Lock()
					res = append(res, pt)
					mu.Unlock()
				}
				return nil
			})
		}(cList[idx])

		if (idx+1)%c.maxConcurrent == 0 {
			if err = g.Wait(); err != nil {
				l.Warn("waiting err: %s", err)
			}
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if category == "metric" && c.nodeName != "" {
		res = append(res, buildCountPoint(c.nodeName, res)...)
	}

	return feed(transToPoint(res, opts))
}

func (c *container) Logging(_ func([]*point.Point) error) {
	if !c.enableCollectLogging {
		return
	}

	cList, err := c.runtime.ListContainers()
	if len(cList) == 0 && err != nil {
		l.Warn("not found containers, err: %s", err)
		return
	}

	var newIDs []string
	for _, info := range cList {
		newIDs = append(newIDs, info.ID)
	}
	c.cleanMissingContainerLog(newIDs)

	for _, info := range cList {
		if isPauseContainer(info) {
			continue
		}

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
		instance.fillSource()
		instance.checkTagsKey()

		instance.setTagsToLogConfigs(instance.tags())
		instance.setTagsToLogConfigs(c.extraTags)
		if c.enableExtractK8sLabelAsTagsV1 {
			instance.setLabelAsTags(instance.podLabels, true /*all labels*/, nil)
		} else {
			instance.setLabelAsTags(instance.podLabels, c.podLabelAsTagsForNonMetric.all, c.podLabelAsTagsForNonMetric.keys)
		}

		c.ipt.setLoggingExtraSourceMapToLogConfigs(instance.configs)
		c.ipt.setLoggingSourceMultilineMapToLogConfigs(instance.configs)
		c.ipt.setLoggingAutoMultilineToLogConfigs(instance.configs)

		c.tailingLogs(instance)
	}

	l.Debugf("current container logtable: %s", c.logTable.String())
}

func (c *container) Name() string {
	return "container"
}

func (c *container) shouldPullContainerLog(ins *logInstance) bool {
	if ins.enabled() {
		return true
	}

	if ins.ownerKind == "job" || ins.ownerKind == "cronjob" {
		return false
	}

	pass := c.logFilter.Match(filter.FilterImage, ins.image) ||
		c.logFilter.Match(filter.FilterImageName, ins.imageName) ||
		c.logFilter.Match(filter.FilterImageShortName, ins.imageShortName) ||
		c.logFilter.Match(filter.FilterNamespace, ins.podNamespace)

	return pass
}

func (c *container) transformPoint(info *runtime.Container, setPodLabelAsTags func(p *typed.PointKV, labels map[string]string)) *typed.PointKV {
	p := typed.NewPointKV(containerMeasurement)

	p.SetTag("container_id", info.ID)
	p.SetTag("container_runtime", info.RuntimeName)
	p.SetTag("container_runtime_version", info.RuntimeVersion)
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
		p.SetField("cpu_usage", top.CPUPercent)
		p.SetField("cpu_usage_millicores", top.CPUUsageMillicores)
		p.SetField("cpu_numbers", top.CPUCores)
		if top.CPUCores != 0 {
			p.SetField("cpu_usage_base100", top.CPUPercent/float64(top.CPUCores))
		}

		p.SetField("mem_usage", top.MemoryWorkingSet)
		if top.MemoryCapacity != 0 && top.MemoryCapacity != math.MaxInt64 {
			p.SetField("mem_capacity", top.MemoryCapacity)
			p.SetField("mem_used_percent", float64(top.MemoryWorkingSet)/float64(top.MemoryCapacity)*100)
		}

		p.SetField("network_bytes_rcvd", top.NetworkRcvd)
		p.SetField("network_bytes_sent", top.NetworkSent)

		// only supported docker
		if top.BlockRead != 0 {
			p.SetField("block_read_byte", top.BlockRead)
		}
		if top.BlockWrite != 0 {
			p.SetField("block_write_byte", top.BlockWrite)
		}
	}

	p.SetTagIfNotEmpty("pod_uid", getPodUIDForLabels(info.Labels))
	podName := getPodNameForLabels(info.Labels)
	p.SetTagIfNotEmpty("pod_name", podName)
	namespace := getPodNamespaceForLabels(info.Labels)
	p.SetTagIfNotEmpty("namespace", namespace)

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

			cpuLimit, memLimit := podInfo.limit(getContainerNameForLabels(info.Labels))

			if cpuLimit != 0 {
				p.SetField("cpu_limit_millicores", cpuLimit)
				if top != nil {
					p.SetField("cpu_usage_base_limit", float64(top.CPUUsageMillicores)/float64(cpuLimit)*100)
				}
			}
			if memLimit != 0 {
				p.SetField("mem_limit", memLimit)
				if top != nil {
					p.SetField("mem_used_percent_base_limit", float64(top.MemoryWorkingSet)/float64(memLimit)*100)
				}
			}

			cpuRequest, memRequest := podInfo.request(getContainerNameForLabels(info.Labels))
			if cpuRequest != 0 {
				p.SetField("cpu_request_millicores", cpuRequest)
				if top != nil {
					p.SetField("cpu_usage_base_request", float64(top.CPUUsageMillicores)/float64(cpuRequest)*100)
				}
			}
			if memRequest != 0 {
				p.SetField("mem_request", memRequest)
				if top != nil {
					p.SetField("mem_used_percent_base_request", float64(top.MemoryWorkingSet)/float64(memRequest)*100)
				}
			}

			setPodLabelAsTags(p, podInfo.pod.Labels)
		}
	}

	p.SetTag("image", image)

	// only ecs fargate
	p.SetTagIfNotEmpty("aws_ecs_cluster_name", getAWSClusterNameForLabels(info.Labels))
	p.SetTagIfNotEmpty("task_family", getTaskFamilyForLabels(info.Labels))
	p.SetTagIfNotEmpty("task_version", getTaskVersionForLabels(info.Labels))
	p.SetTagIfNotEmpty("task_arn", getTaskARNForLabels(info.Labels))

	return p
}

/// get container info for k8s

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

func getDockerTypeForLabels(labels map[string]string) string {
	return labels["io.kubernetes.docker.type"]
}

func isPauseContainer(info *runtime.Container) bool {
	typ := getDockerTypeForLabels(info.Labels)
	return typ == "podsandbox"
}

/// get task info for ecs fargate

func getAWSClusterNameForLabels(labels map[string]string) string {
	return trimClusterName(labels["com.amazonaws.ecs.cluster"])
}

func getTaskFamilyForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-definition-family"]
}

func getTaskVersionForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-definition-version"]
}

func getTaskARNForLabels(labels map[string]string) string {
	return labels["com.amazonaws.ecs.task-arn"]
}

func trimClusterName(s string) string {
	// e.g. arn:aws-cn:ecs:cn-north-1:123123123:cluster/datakit-dev-cluster
	flag := "cluster/"
	index := strings.LastIndex(s, flag)
	if index == -1 {
		return ""
	}
	return s[index+len(flag):]
}

func containerIsFromKubernetes(labels map[string]string) bool {
	uid, ok := labels["io.kubernetes.pod.uid"]
	return ok && uid != ""
}

func transToPoint(pts []*typed.PointKV, opts []point.Option) []*point.Point {
	if len(pts) == 0 {
		return nil
	}

	var res []*point.Point
	t := time.Now()

	for _, pt := range pts {
		r := point.NewPointV2(
			pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(pt.Fields())...),
			append(opts, point.WithTime(t))...,
		)
		res = append(res, r)
	}

	return res
}

func buildCountPoint(nodeName string, pts []*typed.PointKV) (res []*typed.PointKV) {
	nsCount := calculateCount(pts)
	for ns, count := range nsCount {
		pt := typed.NewPointKV("kubernetes")
		pt.SetTag("node_name", nodeName)
		pt.SetTag("namespace", ns)
		pt.SetField("container", count)
		res = append(res, pt)
	}
	return
}

func calculateCount(pts []*typed.PointKV) map[string]int {
	count := make(map[string]int)
	for _, pt := range pts {
		if ns := pt.GetTag("namespace"); ns != "" {
			count[ns]++
		}
	}
	return count
}

func transLabels(labels map[string]string) map[string]interface{} {
	// empty array
	labelsString := "[]"
	if len(labels) != 0 {
		var lb []string
		for k, v := range labels {
			lb = append(lb, k+":"+v)
		}

		b, err := json.Marshal(lb)
		if err == nil {
			labelsString = string(b)
		}
	}
	// http://gitlab.jiagouyun.com/cloudcare-tools/kodo/-/issues/61#note_11580
	return map[string]interface{}{
		"df_label":            labelsString,
		"df_label_permission": "read_only",
		"df_label_source":     "datakit",
	}
}
