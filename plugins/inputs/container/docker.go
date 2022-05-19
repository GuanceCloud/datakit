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
	"time"

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type dockerInput struct {
	client dockerClientX
	// container log 需要添加 pod 信息，所以存一份 k8sclient
	k8sClient k8sClientX

	containerLogList map[string]context.CancelFunc

	metricFilter  filter.Filter
	loggingFilter filter.Filter

	cfg *dockerInputConfig
}

type dockerInputConfig struct {
	endpoint string

	excludePauseContainer  bool
	removeLoggingAnsiCodes bool

	containerIncludeMetric []string
	containerExcludeMetric []string

	containerIncludeLog []string
	containerExcludeLog []string

	extraTags map[string]string
}

func newDockerInput(cfg *dockerInputConfig) (*dockerInput, error) {
	d := &dockerInput{
		containerLogList: make(map[string]context.CancelFunc),
		cfg:              cfg,
	}

	client, err := newDockerClient(cfg.endpoint, nil)
	if err != nil {
		return nil, err
	}
	d.client = client

	if !d.pingOK() {
		return nil, fmt.Errorf("cannot connect to the Docker daemon at unix:///var/run/docker.sock")
	}

	if err := d.createMetricFilters(cfg.containerIncludeMetric, cfg.containerExcludeMetric); err != nil {
		return nil, err
	}
	if err := d.createLoggingFilters(cfg.containerIncludeLog, cfg.containerExcludeLog); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *dockerInput) stop() {
	d.cancelTails()
}

func (d *dockerInput) pingOK() bool {
	ping, err := d.client.Ping(context.TODO())
	if err != nil {
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
		wg  sync.WaitGroup

		now = time.Now()
	)

	for idx := range cList {
		wg.Add(1)

		go func(c *types.Container) {
			defer wg.Done()

			image := getImageOfPodContainer(c, d.k8sClient)
			if d.ignoreContainer(c) || d.ignoreImageForMetric(image) {
				return
			}

			m, err := gatherDockerContainerMetric(d.client, d.k8sClient, c)
			if err != nil {
				return
			}
			m.tags.append(d.cfg.extraTags)
			m.time = now

			mu.Lock()
			res = append(res, m)
			mu.Unlock()
		}(&cList[idx])
	}

	wg.Wait()
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
		wg  sync.WaitGroup

		now = time.Now()
	)

	for idx := range cList {
		wg.Add(1)

		go func(c *types.Container) {
			defer wg.Done()

			if d.ignoreContainer(c) {
				return
			}
			m, err := gatherDockerContainerObject(d.client, d.k8sClient, c)
			if err != nil {
				return
			}
			m.tags.append(d.cfg.extraTags)
			m.time = now

			mu.Lock()
			res = append(res, m)
			mu.Unlock()
		}(&cList[idx])
	}

	wg.Wait()
	return res, nil
}

func (d *dockerInput) watchNewContainerLogs() error {
	cList, err := d.getContainerList()
	if err != nil {
		return err
	}

	for idx, container := range cList {
		if !d.shouldPullContainerLog(&cList[idx]) {
			continue
		}

		l.Infof("add container log, containerName: %s image: %s", getContainerName(container.Names), container.Image)
		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		go func(container *types.Container) {
			defer func() {
				d.removeFromContainerList(container.ID)
				l.Debugf("remove container log, containerName: %s image: %s", getContainerName(container.Names), container.Image)
			}()

			if err := d.watchingContainerLog(ctx, container); err != nil {
				if !errors.Is(err, context.Canceled) {
					l.Errorf("tailContainerLog: %s", err)
				}
			}
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
		podAnnotationState = getPodAnnotationState(container, meta)
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

func getPodAnnotationState(container *types.Container, meta *podMeta) podAnnotationStateType {
	if meta == nil {
		return podAnnotationNil
	}

	logconf, err := getContainerLogConfig(meta.Annotations)
	if err != nil || logconf == nil {
		return podAnnotationNil
	}

	if logconf.Disable {
		l.Debugf("ignore containerlog because of annotation disable, podName:%s, containerName:%s",
			getPodNameForLabels(container.Labels), getContainerName(container.Names))
		return podAnnotationDisable
	}

	if len(logconf.OnlyImages) == 0 {
		return podAnnotationEnable
	}

	f, err := filter.NewIncludeExcludeFilter(splitRules(logconf.OnlyImages), nil)
	if err != nil {
		l.Warnf("failed to new filter of only_images, err:%w", err)
		return podAnnotationEnable
	}

	podContainerName := getContainerNameForLabels(container.Labels)
	image := meta.containerImage(podContainerName)
	if image != "" && f.Match(image) {
		l.Debugf("match pod only_images, name:%s, image: %s", podContainerName, image)
		return podAnnotationEnable
	}
	l.Debugf("ignore pod container, name:%s, image: %s", podContainerName, image)
	return podAnnotationDisable
}

func (d *dockerInput) getContainerList() ([]types.Container, error) {
	cList, err := d.client.ContainerList(context.Background(), dockerContainerListOption)
	if err != nil {
		return nil, fmt.Errorf("failed to get container list: %w", err)
	}

	return cList, nil
}

func (d *dockerInput) createMetricFilters(include, exclude []string) error {
	in := splitRules(include)
	ex := splitRules(exclude)

	f, err := filter.NewIncludeExcludeFilter(in, ex)
	if err != nil {
		return err
	}

	d.metricFilter = f
	return nil
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

func (d *dockerInput) ignoreImageForMetric(image string) (ignore bool) {
	if d.metricFilter == nil {
		return
	}
	// 注意，match 和 ignore 是相反的逻辑
	// 如果 match 通过，则表示不需要 ignore
	// 所以要取反
	return !d.metricFilter.Match(image)
}

func (d *dockerInput) ignoreImageForLogging(image string) (ignore bool) {
	if d.loggingFilter == nil {
		return
	}
	return !d.loggingFilter.Match(image)
}

func (d *dockerInput) ignoreContainer(container *types.Container) bool {
	if !isRunningContainer(container.State) {
		return true
	}
	if d.cfg.excludePauseContainer && isPauseContainer(container.Command) {
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
