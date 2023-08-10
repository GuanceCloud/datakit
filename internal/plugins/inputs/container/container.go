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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type container struct {
	ipt       *Input
	runtime   runtime.ContainerRuntime
	k8sClient k8sclient.Client

	loggingFilter filter.Filter
	logTable      *logTable
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

	return &container{
		ipt:           ipt,
		runtime:       r,
		k8sClient:     k8sClient,
		loggingFilter: filter,
		logTable:      newLogTable(),
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

		if !c.shouldPullContainerLog(info, instance) {
			continue
		}

		instance.addStdout()
		instance.fillLogType(info.RuntimeName)
		instance.setTagsToLogConfigs(instance.tags())
		instance.setTagsToLogConfigs(c.ipt.Tags)
		instance.setTagsToLogConfigs(imageToTags(info.Image))

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

func (c *container) shouldPullContainerLog(info *runtime.Container, instance *logInstance) bool {
	if instance.enabled() {
		return true
	}
	if c.ignoreImageForLogging(info.Image.Image) {
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

	p.SetTag("image", info.Image.Image)
	p.SetTag("image_name", info.Image.ImageName)
	p.SetTag("image_short_name", info.Image.ShortName)
	p.SetTag("image_tag", info.Image.Tag)

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

	if c.k8sClient != nil && podName != "" {
		owner, err := c.queryOwnerFromK8s(context.Background(), podName, namespace)
		if err != nil {
			l.Warnf("container %s, %s", info.Name, err)
		} else {
			switch owner.ownerKind {
			case deploymentKind:
				p.SetTag("deployment", owner.ownerName)
			case daemonsetKind:
				p.SetTag("daemonset", owner.ownerName)
			case statefulsetKind:
				p.SetTag("statefulset", owner.ownerName)
			default:
				// skip
			}
		}
	}

	p.SetTags(c.ipt.Tags)
	return p
}

type ownerKind string

const (
	deploymentKind  ownerKind = "Deployment"
	daemonsetKind   ownerKind = "DaemonSet"
	statefulsetKind ownerKind = "StatefulSet"
)

type ownerInfo struct {
	podName        string
	namespace      string
	podAnnotations map[string]string

	ownerKind ownerKind
	ownerName string
}

func (c *container) queryOwnerFromK8s(ctx context.Context, podName, podNamespace string) (*ownerInfo, error) {
	pod, err := c.k8sClient.GetPodsForNamespace(podNamespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable query pod %s, err: %w", podName, err)
	}

	owner := ownerInfo{
		podName:        podName,
		namespace:      podNamespace,
		podAnnotations: pod.GetAnnotations(),
	}

	if len(pod.OwnerReferences) != 0 {
		switch pod.OwnerReferences[0].Kind {
		case "ReplicaSet":
			replica, repErr := c.k8sClient.GetReplicaSetsForNamespace(podNamespace).Get(ctx, pod.OwnerReferences[0].Name, metav1.GetOptions{})
			if repErr == nil && len(replica.OwnerReferences) != 0 {
				owner.ownerKind = deploymentKind
				owner.ownerName = replica.OwnerReferences[0].Name
			}
		case "DaemonSet":
			owner.ownerKind = daemonsetKind
			owner.ownerName = pod.OwnerReferences[0].Name
		case "StatefulSet":
			owner.ownerKind = statefulsetKind
			owner.ownerName = pod.OwnerReferences[0].Name
		default:
			// skip
		}
	}

	return &owner, nil
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
//   split 'image:' kv，return values
//   ex, in: ["image:img_*", "image:img01*", "xx:xx"] return: ["img_*", "img01*"]
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

func imageToTags(image runtime.Image) map[string]string {
	return map[string]string{
		"image":            image.Image,
		"image_name":       image.ImageName,
		"image_short_name": image.ShortName,
		"image_tag":        image.Tag,
	}
}
