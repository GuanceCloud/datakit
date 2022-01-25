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

func (d *dockerInput) gatherMetric() ([]inputs.Measurement, error) {
	cList, err := d.getContaierList()
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
	cList, err := d.getContaierList()
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
	cList, err := d.getContaierList()
	if err != nil {
		return err
	}

	for idx, container := range cList {
		if !d.shouldPullContainerLog(&cList[idx]) {
			continue
		}

		l.Infof("add container log, name: %s image: %s", getContainerName(container.Names), container.Image)
		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		go func(container *types.Container) {
			defer func() {
				d.removeFromContainerList(container.ID)
				l.Debugf("remove container log, name: %s image: %s", getContainerName(container.Names), container.Image)
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

const (
	podAnnotationNil = iota + 1
	podAnnotationEnable
	podAnnotationDisable
)

func (d *dockerInput) shouldPullContainerLog(container *types.Container) bool {
	if d.containerInContainerList(container.ID) {
		return false
	}

	image := container.Image

	// TODO
	// 每次获取到容器列表，都要进行以下所有 resist 考查，特别是获取其 k8s Annotation 的配置，需要进行访问和查找
	// 这消耗很大，且没有意义
	// 可以使用 container ID 进行缓存，维持一份名单，通过名单再决定是否进行考查

	podAnnotationState := podAnnotationNil

	func() {
		if d.k8sClient == nil || container.Labels["pod_name"] == "" {
			return
		}
		meta, err := queryPodMetaData(d.k8sClient, container.Labels["pod_name"], container.Labels["pod_namesapce"])
		if err != nil {
			return
		}
		image = meta.containerImage()

		logconf, err := getContainerLogConfig(meta.Annotations)
		if err != nil || logconf == nil {
			return
		}

		if logconf.Disable {
			podAnnotationState = podAnnotationDisable
		} else {
			podAnnotationState = podAnnotationEnable
		}
	}()

	switch podAnnotationState {
	case podAnnotationDisable:
		l.Debugf("ignore containerlog because of annotation disable, name: %s, shortImage: %s", getContainerName(container.Names), image)
		return false
	case podAnnotationEnable:
		return true
	case podAnnotationNil:
		// nil
	}

	if d.ignoreContainer(container) {
		l.Debugf("ignore containerlog because of pause status, name: %s, shortImage: %s", getContainerName(container.Names), image)
		return false
	}

	if d.ignoreImageForLogging(image) {
		l.Debugf("ignore containerlog because of image filter, name: %s, shortImage: %s", getContainerName(container.Names), image)
		return false
	}

	return true
}

func (d *dockerInput) getContaierList() ([]types.Container, error) {
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
	if container.Labels["pod_name"] == "" {
		return
	}

	meta, err := queryPodMetaData(k8sClient, container.Labels["pod_name"], container.Labels["pod_namesapce"])
	if err != nil {
		return
	}
	image = meta.containerImage()
	return
}
