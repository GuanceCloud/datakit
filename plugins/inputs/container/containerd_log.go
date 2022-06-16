// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
	"fmt"
	"net"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"google.golang.org/grpc"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	kubeRuntimeAPIVersion = "0.1.0"
	maxMsgSize            = 1024 * 1024 * 16
)

func newCRIClient(endpoint string) (cri.RuntimeServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	//nolint:lll
	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithInsecure(), grpc.WithDialer(dial), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize))) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	return cri.NewRuntimeServiceClient(conn), nil
}

func getCRIRuntimeVersion(client cri.RuntimeServiceClient) (*cri.VersionResponse, error) {
	ctx, cancel := getContextWithTimeout(time.Second * 10)
	defer cancel()
	return client.Version(ctx, &cri.VersionRequest{Version: kubeRuntimeAPIVersion})
}

func getContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}

func (c *containerdInput) addToLogList(logpath string) {
	c.logpathList[logpath] = nil
}

func (c *containerdInput) removeFromLogList(logpath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.logpathList, logpath)
}

func (c *containerdInput) inLogList(logpath string) bool {
	_, ok := c.logpathList[logpath]
	return ok
}

func (c *containerdInput) watchNewLogs() error {
	list, err := c.criClient.ListContainers(context.Background(), &cri.ListContainersRequest{Filter: nil})
	if err != nil {
		return fmt.Errorf("failed to get cri-ListContainers err: %w", err)
	}

	for _, container := range list.GetContainers() {
		resp, err := c.criClient.ContainerStatus(context.Background(), &cri.ContainerStatusRequest{ContainerId: container.Id})
		if err != nil {
			l.Warnf("failed to get cri-container respone, id: %s, err: %s", container.Id, err)
			continue
		}

		status := resp.GetStatus()
		if status == nil {
			continue
		}

		logpath := status.GetLogPath()

		if c.inLogList(logpath) {
			continue
		}

		tags := map[string]string{
			// "state":          "running",
			"container_name": getContainerNameForLabels(status.GetLabels()),
			"container_id":   status.GetId(),
			"pod_name":       getPodNameForLabels(status.GetLabels()),
			"namespace":      getPodNamespaceForLabels(status.GetLabels()),
		}
		if c.criRuntimeVersion != nil {
			tags["container_type"] = c.criRuntimeVersion.RuntimeName
		}
		// add extra tags
		for k, v := range c.cfg.extraTags {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}

		if image := status.GetImage(); image != nil {
			// 如果能找到 pod image，则使用它
			imageName, imageShortName, imageTag := ParseImage(image.Image)
			tags["image"] = image.Image
			tags["image_name"] = imageName
			tags["image_short_name"] = imageShortName
			tags["image_tag"] = imageTag
		}

		source := getContainerNameForLabels(status.GetLabels())
		if n := status.Metadata; n != nil {
			source = n.Name
		}

		opt := &tailer.Option{
			Source:     source,
			GlobalTags: tags,
		}

		logconf, err := getContainerLogConfig(status.GetAnnotations())
		if err != nil {
			l.Warnf("invalid logconfig from annotation, err: %s, skip", err)
		}

		if logconf != nil {
			if logconf.Source != "" {
				opt.Source = logconf.Source
			}
			if logconf.Service != "" {
				opt.Service = logconf.Service
			}
			opt.Pipeline = logconf.Pipeline
			opt.MultilineMatch = logconf.Multiline

			l.Debugf("use container logconfig:%#v, containerId: %s, source: %s, logpath: %s", logconf, container.Id, opt.Source, logpath)
		}

		_ = opt.Init()

		t, err := tailer.NewTailerSingle(logpath, opt)
		if err != nil {
			l.Warnf("failed to new containerd log, containerId: %s, source: %s, logpath: %s, err: %s", container.Id, opt.Source, logpath, err)
			continue
		}

		c.addToLogList(logpath)
		l.Infof("add containerd log, containerId: %s, source: %s, logpath: %s", container.Id, opt.Source, logpath)

		go func(logpath string) {
			defer c.removeFromLogList(logpath)
			t.Run()
		}(logpath)
	}

	return nil
}
