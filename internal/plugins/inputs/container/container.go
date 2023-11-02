// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/container/option"
)

type container struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	loggingFilters []filter.Filter
	logTable       *logTable
	extraTags      map[string]string
}

func newContainer(ipt *Input, endpoint string, mountPoint string, k8sClient k8sclient.Client) (Collector, error) {
	filters, err := newFilters(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog)
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
		ipt:            ipt,
		runtime:        r,
		k8sClient:      k8sClient,
		loggingFilters: filters,
		logTable:       newLogTable(),
		extraTags:      tags,
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
		c.ipt.Feeder.FeedLastError(err.Error(), dkio.WithLastErrorInput(inputName))
	}
	collectCostVec.WithLabelValues(category).Observe(time.Since(start).Seconds())
}

const goroutineNum = 4

func (c *container) gatherResource(category string, opts []point.Option, feed func(pts []*point.Point) error) error {
	cList, err := c.runtime.ListContainers()
	if err != nil {
		l.Warn(err)

		return nil
	}

	var res []*typed.PointKV
	var mu sync.Mutex

	g := goroutine.NewGroup(goroutine.Option{Name: "container-" + category})
	for idx := range cList {
		if isPauseContainer(cList[idx]) {
			continue
		}

		func(info *runtime.Container) {
			g.Go(func(ctx context.Context) error {
				pt := c.transformPoint(info)
				pt.SetTags(c.extraTags)

				if category == "object" {
					pt.SetTag("name", info.ID)
					pt.SetField("age", time.Since(time.Unix(0, info.CreatedAt)).Milliseconds()/1e3)
				}

				mu.Lock()
				res = append(res, pt)
				mu.Unlock()
				return nil
			})
		}(cList[idx])

		if (idx+1)%goroutineNum == 0 {
			_ = g.Wait()
		}
	}

	if err := g.Wait(); err != nil {
		return err
	} else {
		return feed(transToPoint(res, opts))
	}
}

func (c *container) Logging(_ func([]*point.Point) error) {
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
		instance.setFromBeginning(info.CreatedAt)

		instance.setTagsToLogConfigs(instance.tags())
		instance.setTagsToLogConfigs(c.extraTags)
		instance.setCustomerTags(instance.podLabels, getGlobalCustomerKeys())

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

	ignored := c.ignoreContainerLogging(filterImage, ins.image) ||
		c.ignoreContainerLogging(filterImageName, ins.imageName) ||
		c.ignoreContainerLogging(filterImageShortName, ins.imageShortName) ||
		c.ignoreContainerLogging(filterNamespace, ins.podNamespace)

	return !ignored
}

func (c *container) ignoreContainerLogging(typ filterType, field string) (ignore bool) {
	if field == "" {
		return
	}
	if len(c.loggingFilters) != len(supportedFilterTypes) {
		return
	}

	if c.loggingFilters[typ] != nil {
		ignore = !c.loggingFilters[typ].Match(field)
		l.Debugf("container_log filtered status: %v, type %s, field %s", ignore, typ.String(), field)
	}

	return
}

func (c *container) transformPoint(info *runtime.Container) *typed.PointKV {
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

			p.SetField("cpu_limit", podInfo.cpuLimit(getContainerNameForLabels(info.Labels)))
			p.SetCustomerTags(podInfo.pod.Labels, config.Cfg.Dataway.GlobalCustomerKeys)
		}
	}

	imageName, shortName, imageTag := runtime.ParseImage(image)
	p.SetTag("image", image)
	p.SetTag("image_name", imageName)
	p.SetTag("image_short_name", shortName)
	p.SetTag("image_tag", imageTag)

	return p
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

func getDockerTypeForLabels(labels map[string]string) string {
	return labels["io.kubernetes.docker.type"]
}

func isPauseContainer(info *runtime.Container) bool {
	typ := getDockerTypeForLabels(info.Labels)
	return typ == "podsandbox"
}

type filterType int

const (
	filterImage filterType = iota
	filterImageName
	filterImageShortName
	filterNamespace
)

func (f filterType) String() string {
	switch f {
	case filterImage:
		return "image"
	case filterImageName:
		return "image_name"
	case filterImageShortName:
		return "image_short_name"
	case filterNamespace:
		return "namespace"
	default:
		return "unknown"
	}
}

var supportedFilterTypes = []filterType{filterImage, filterImageName, filterImageShortName, filterNamespace}

func newFilters(include, exclude []string) ([]filter.Filter, error) {
	in := splitRules(include)
	ex := splitRules(exclude)

	supportedFilterTypesLength := len(supportedFilterTypes)
	if len(in) != supportedFilterTypesLength || len(ex) != supportedFilterTypesLength {
		return nil, fmt.Errorf("unreachable, invalid filter type, expect len(%d), actual include: %d exclude: %d",
			supportedFilterTypesLength, len(in), len(ex))
	}

	filters := make([]filter.Filter, supportedFilterTypesLength)

	for _, typ := range supportedFilterTypes {
		filter, err := filter.NewIncludeExcludeFilter(in[typ], ex[typ])
		if err != nil {
			l.Warnf("invalid container_log filter, err: %s, ignored", err)
			continue
		}
		filters[typ] = filter
	}

	return filters, nil
}

func splitRules(arr []string) [][]string {
	rules := make([][]string, len(supportedFilterTypes))

	split := func(str, prefix string) string {
		if !strings.HasPrefix(str, prefix) {
			return ""
		}
		content := strings.TrimPrefix(str, prefix)
		rule := strings.TrimSpace(content)
		if rule == "*" {
			// trans to double star
			return "**"
		}
		return rule
	}

	for _, str := range arr {
		for _, typ := range supportedFilterTypes {
			rule := split(strings.TrimSpace(str), typ.String()+":")
			if rule != "" {
				rules[typ] = append(rules[typ], rule)
			}
		}
	}

	return rules
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
	for _, pt := range pts {
		r := point.NewPointV2(
			pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(pt.Fields())...),
			opts...,
		)
		res = append(res, r)
	}

	return res
}
