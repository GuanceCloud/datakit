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
	client *dockerClient
	// container log 需要添加 pod 信息，所以存一份 k8s client
	k8sClient *k8sClient

	containerLogList map[string]context.CancelFunc

	metricFilter  filter.Filter
	loggingFilter filter.Filter

	cfg *dockerInputConfig

	wg sync.WaitGroup
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
	d.wg.Wait()
}

func (d *dockerInput) gatherMetric() ([]inputs.Measurement, error) {
	cList, err := d.getContaierList()
	if err != nil {
		return nil, err
	}

	var res []inputs.Measurement
	now := time.Now()

	for idx, container := range cList {
		if d.ignoreContainer(&cList[idx]) || d.ignoreImageForMetric(container.Image) {
			continue
		}

		m, err := gatherDockerContainerMetric(d.client, d.k8sClient, &cList[idx])
		if err != nil {
			continue
		}
		m.tags.append(d.cfg.extraTags)
		m.time = now

		res = append(res, m)
	}
	return res, nil
}

func (d *dockerInput) gatherObject() ([]inputs.Measurement, error) {
	cList, err := d.getContaierList()
	if err != nil {
		return nil, err
	}

	var res []inputs.Measurement
	now := time.Now()

	for idx := range cList {
		if d.ignoreContainer(&cList[idx]) {
			continue
		}
		m, err := gatherDockerContainerObject(d.client, d.k8sClient, &cList[idx])
		if err != nil {
			continue
		}
		m.tags.append(d.cfg.extraTags)
		m.time = now

		res = append(res, m)
	}
	return res, nil
}

func (d *dockerInput) watchingNewContainerLog() error {
	cList, err := d.getContaierList()
	if err != nil {
		return err
	}

	for idx, container := range cList {
		if d.containerInContainerList(container.ID) {
			continue
		}
		if d.ignoreContainer(&cList[idx]) || d.ignoreImageForLogging(container.Image) {
			l.Debugf("ignore container log, name: %s", getContainerName(container.Names))
			continue
		}

		l.Infof("add container log, name: %s image: %s", getContainerName(container.Names), container.Image)
		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container *types.Container) {
			defer func() {
				d.wg.Done()
				d.removeFromContainerList(container.ID)
				l.Debugf("remove container log, name: %s image: %s", getContainerName(container.Names), container.Image)
			}()

			if err := d.tailContainerLogs(ctx, container); err != nil {
				if !errors.Is(err, context.Canceled) {
					l.Errorf("tailContainerLogs: %s", err)
				}
			}
		}(&cList[idx])
	}

	return nil
}

func (d *dockerInput) getContaierList() ([]types.Container, error) {
	cList, err := d.client.ContainerList(context.Background(), dockerContaienrListOption)
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
	_, imageShortName, _ := ParseImage(image)

	// 注意，match 和 ignore 是相反的逻辑
	// 如果 match 通过，则表示不需要 ignore
	// 所以要取反
	return !d.metricFilter.Match(imageShortName)
}

func (d *dockerInput) ignoreImageForLogging(image string) (ignore bool) {
	if d.loggingFilter == nil {
		return
	}
	_, imageShortName, _ := ParseImage(image)
	return !d.loggingFilter.Match(imageShortName)
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
