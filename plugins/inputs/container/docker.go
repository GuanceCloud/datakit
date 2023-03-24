// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dockerInput struct {
	ipt       *Input
	client    *dockerClient
	k8sClient k8sClientX // container log 需要添加 pod 信息，所以存一份 k8sclient

	loggingFilter    filter.Filter
	containerLogList map[string]interface{}
	mu               sync.Mutex
}

func newDockerInput(ipt *Input) (*dockerInput, error) {
	d := &dockerInput{
		containerLogList: make(map[string]interface{}),
		ipt:              ipt,
	}

	client, err := newDockerClient(ipt.DockerEndpoint, nil)
	if err != nil {
		return nil, err
	}
	d.client = client

	if !d.pingOK() {
		return nil, fmt.Errorf("cannot connect to the Docker daemon at %s", ipt.DockerEndpoint)
	}

	if err := d.createLoggingFilters(ipt.ContainerIncludeLog, ipt.ContainerExcludeLog); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *dockerInput) stop() {
	// empty interface
}

func (d *dockerInput) pingOK() bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ping, err := d.client.Ping(ctx)
	if err != nil {
		l.Warnf("docker ping error: %s", err)
		return false
	}
	if ping.APIVersion == "" || ping.OSType == "" {
		return false
	}
	return true
}

func (d *dockerInput) gatherMetric() ([]inputs.Measurement, error) {
	cList, err := d.getContainerList()
	if err != nil {
		return nil, err
	}

	var (
		res []inputs.Measurement
		mu  sync.Mutex
	)

	g := goroutine.NewGroup(goroutine.Option{Name: goroutineGroupName})

	for idx := range cList {
		func(c *types.Container) {
			g.Go(func(ctx context.Context) error {
				if d.ignoreContainer(c) {
					return nil
				}

				m, err := gatherDockerContainerMetric(d.client, d.k8sClient, c)
				if err != nil {
					return nil
				}
				m.tags.append(d.ipt.Tags)

				mu.Lock()
				res = append(res, m)
				mu.Unlock()
				return nil
			})
		}(&cList[idx])
	}

	_ = g.Wait()
	return res, nil
}

func (d *dockerInput) gatherObject() ([]inputs.Measurement, error) {
	cList, err := d.getContainerList()
	if err != nil {
		return nil, err
	}

	var (
		res []inputs.Measurement
		mu  sync.Mutex
	)

	g := goroutine.NewGroup(goroutine.Option{Name: goroutineGroupName})
	for idx := range cList {
		func(c *types.Container) {
			g.Go(func(ctx context.Context) error {
				if d.ignoreContainer(c) {
					return nil
				}
				m, err := gatherDockerContainerObject(d.client, d.k8sClient, c)
				if err != nil {
					return nil
				}
				m.tags.append(d.ipt.Tags)

				mu.Lock()
				res = append(res, m)
				mu.Unlock()
				return nil
			})
		}(&cList[idx])
	}

	_ = g.Wait()

	return res, nil
}

func (d *dockerInput) watchNewLogs() error {
	cList, err := d.getRunningContainerList()
	if err != nil {
		return err
	}

	g := goroutine.NewGroup(goroutine.Option{Name: goroutineGroupName})

	for idx, container := range cList {
		if !d.shouldPullContainerLog(&cList[idx]) {
			continue
		}

		l.Infof("add container log, containerName: %s image: %s", getContainerName(container.Names), container.Image)
		// Start a new goroutine for every new container that has logs to collect
		func(container *types.Container) {
			g.Go(func(ctx context.Context) error {
				if err := d.tailingLog(context.Background(), container); err != nil {
					if !errors.Is(err, context.Canceled) {
						l.Warnf("tail containerLog: %s", err)
					}
				}
				return nil
			})
		}(&cList[idx])
	}

	return nil
}

type podAnnotationStateType int

const (
	podAnnotationNil podAnnotationStateType = iota + 1
	podAnnotationEnable
	podAnnotationDisable
)

func (d *dockerInput) shouldPullContainerLog(container *types.Container) bool {
	if d.containerInContainerList(container.ID) {
		return false
	}

	image := container.Image

	// TODO
	// 每次获取到容器列表，都要进行以下审核，特别是获取其 k8s Annotation 的配置，需要进行访问和查找
	// 这消耗很大，且没有意义
	// 可以使用 container ID 进行缓存，维持一份名单，通过名单再决定是否进行考查

	podAnnotationState := podAnnotationNil

	func() {
		podName := getPodNameForLabels(container.Labels)
		if d.k8sClient == nil || podName == "" {
			return
		}
		podNamespace := getPodNamespaceForLabels(container.Labels)

		meta, err := queryPodMetaData(d.k8sClient, podName, podNamespace)
		if err != nil {
			return
		}
		if containerImage := meta.containerImage(getContainerNameForLabels(container.Labels)); containerImage != "" {
			image = containerImage
		}
		podAnnotationState = getPodAnnotationState(container.Labels, meta)
	}()

	switch podAnnotationState {
	case podAnnotationDisable:
		return false
	case podAnnotationEnable:
		return true
	case podAnnotationNil:
		// nil
	}

	if d.ignoreContainer(container) {
		l.Debugf("ignore containerlog because of pause status, containerName:%s, shortImage:%s", getContainerName(container.Names), image)
		return false
	}

	if d.ignoreImageForLogging(image) {
		l.Debugf("ignore containerlog because of image filter, containerName:%s, shortImage:%s", getContainerName(container.Names), image)
		return false
	}

	return true
}

func getPodAnnotationState(labels map[string]string, meta *podMeta) podAnnotationStateType {
	if meta == nil {
		return podAnnotationNil
	}

	// 优先使用 Pod Annotations 的 datakit/logs 配置
	// 其次使用全局 CRD 列表中 Pod UID 对应的 datakit/logs

	var conf string
	if meta.annotations() != nil && meta.annotations()[containerLogConfigKey] != "" {
		conf = meta.annotations()[containerLogConfigKey]
	} else {
		globalCRDLogsConfList.mu.Lock()
		conf = globalCRDLogsConfList.list[string(meta.UID)]
		globalCRDLogsConfList.mu.Unlock()
	}

	logconf, err := parseContainerLogConfig(conf)
	if err != nil || logconf == nil {
		return podAnnotationNil
	}

	if logconf.Disable {
		l.Debugf("ignore containerlog because of annotation disable, podName:%s, containerName:%s",
			getPodNameForLabels(labels), getContainerNameForLabels(labels))
		return podAnnotationDisable
	}

	if len(logconf.OnlyImages) == 0 {
		return podAnnotationEnable
	}

	f, err := filter.NewIncludeExcludeFilter(splitRules(logconf.OnlyImages), nil)
	if err != nil {
		l.Warnf("failed to new filter of only_images, err: %s", err)
		return podAnnotationEnable
	}

	podContainerName := getContainerNameForLabels(labels)
	image := meta.containerImage(podContainerName)
	if image != "" && f.Match(image) {
		l.Debugf("match pod only_images, name:%s, image: %s", podContainerName, image)
		return podAnnotationEnable
	}
	l.Debugf("ignore pod container, name:%s, image: %s", podContainerName, image)
	return podAnnotationDisable
}

func (d *dockerInput) getContainerList() ([]types.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cList, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get container list: %w", err)
	}
	return cList, nil
}

func (d *dockerInput) getRunningContainerList() ([]types.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cList, err := d.client.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("failed to get container list: %w", err)
	}
	return cList, nil
}

func (d *dockerInput) createLoggingFilters(include, exclude []string) error {
	in := splitRules(include)
	ex := splitRules(exclude)

	f, err := filter.NewIncludeExcludeFilter(in, ex)
	if err != nil {
		return err
	}

	d.loggingFilter = f
	return nil
}

func (d *dockerInput) ignoreImageForLogging(image string) (ignore bool) {
	if d.loggingFilter == nil {
		return
	}
	// 注意，match 和 ignore 是相反的逻辑
	// 如果 match 通过，则表示不需要 ignore
	// 所以要取反
	return !d.loggingFilter.Match(image)
}

func (d *dockerInput) ignoreContainer(container *types.Container) bool {
	if !isRunningContainer(container.State) {
		return true
	}
	if d.ipt.ExcludePauseContainer && isPauseContainer(container.Command) {
		return true
	}
	return false
}

// splitRules
//   切分带有 'image:' 前缀的字符串 kv，返回其 values
//   ex: ["image:img_*", "image:img01*", "xx:xx"]
//   return: ["img_*", "img01*"]
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

//nolint:deadcode,unused
func getImageOfPodContainer(container *types.Container, k8sClient k8sClientX) (image string) {
	image = container.Image

	if k8sClient == nil {
		return
	}
	if getPodNameForLabels(container.Labels) == "" {
		return
	}

	meta, err := queryPodMetaData(k8sClient, getPodNameForLabels(container.Labels), getPodNamespaceForLabels(container.Labels))
	if err != nil {
		return
	}
	image = meta.containerImage(getContainerNameForLabels(container.Labels))
	return
}
