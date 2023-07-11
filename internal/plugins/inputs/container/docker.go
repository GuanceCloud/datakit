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

	"github.com/docker/docker/api/types"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type dockerInput struct {
	ipt       *Input
	client    *dockerClient
	k8sClient k8sClientX // container log 需要添加 pod 信息，所以存一份 k8sclient

	loggingFilter filter.Filter
	logTable      *logTable
}

func newDockerInput(ipt *Input) (*dockerInput, error) {
	d := &dockerInput{
		ipt:      ipt,
		logTable: newLogTable(),
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

	var newIDs []string
	for _, container := range cList {
		newIDs = append(newIDs, container.ID)
	}
	d.cleanMissingContainerLog(newIDs)

	l.Infof("docker container IDs: %v", newIDs)

	for idx := range cList {
		info := d.queryContainerLogInfo(context.Background(), &cList[idx])
		if info == nil {
			continue
		}

		if !d.shouldPullContainerLog(&cList[idx], info) {
			continue
		}

		if err := info.parseLogConfigs(); err != nil {
			l.Warn(err)
			continue
		}

		info.addStdout()
		info.fillTags()

		d.ipt.setLoggingExtraSourceMapToLogConfigs(info.logConfigs)
		d.ipt.setLoggingSourceMultilineMapToLogConfigs(info.logConfigs)
		d.ipt.setLoggingAutoMultilineToLogConfigs(info.logConfigs)

		d.ipt.setExtractK8sLabelAsTagsToLogConfigs(info.logConfigs, info.podLabels)
		d.ipt.setTagsToLogConfigs(info.logConfigs, info.tags)
		d.ipt.setGlobalTagsToLogConfigs(info.logConfigs)

		l.Debugf("docker container %s info: %#v", info.containerName, info)

		d.tailingLogs(info)
	}

	l.Debugf("current docker logtable: %s", d.logTable.String())

	return nil
}

func (d *dockerInput) cleanMissingContainerLog(newIDs []string) {
	missingIDs := d.logTable.findDifferences(newIDs)
	for _, id := range missingIDs {
		l.Infof("clean log collection for container id %s", id)
		d.logTable.closeFromTable(id)
		d.logTable.removeFromTable(id)
	}
}

func (d *dockerInput) shouldPullContainerLog(container *types.Container, info *containerLogInfo) bool {
	if d.logTable.inTable(info.id, info.logPath) {
		return false
	}

	if info.enabled() {
		return true
	}

	if d.ignoreContainer(container) {
		return false
	}

	if d.ignoreImageForLogging(info.image) {
		return false
	}

	return true
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
