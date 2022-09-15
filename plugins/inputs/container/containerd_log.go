// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"context"
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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logpathList[logpath] = nil
}

func (c *containerdInput) removeFromLogList(logpath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.logpathList, logpath)
}

func (c *containerdInput) inLogList(logpath string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.logpathList[logpath]
	return ok
}

func (c *containerdInput) tailingLog(status *cri.ContainerStatus) error {
	info := &containerLogBasisInfo{
		id:                    status.GetId(),
		logPath:               status.GetLogPath(),
		labels:                status.GetLabels(),
		tags:                  make(map[string]string),
		extraSourceMap:        c.ipt.LoggingExtraSourceMap,
		sourceMultilineMap:    c.ipt.LoggingSourceMultilineMap,
		autoMultilinePatterns: c.ipt.getAutoMultilinePatterns(),
		extractK8sLabelAsTags: c.ipt.ExtractK8sLabelAsTags,
	}

	if status.GetMetadata() != nil && status.GetMetadata().Name != "" {
		info.name = status.GetMetadata().Name
	} else {
		info.name = "unknown"
	}

	if n := status.GetImage(); n != nil {
		info.image = n.Image
	}

	if c.criRuntimeVersion != nil {
		l.Debugf("containedlog runtime: '%s'", c.criRuntimeVersion.RuntimeName)
		info.tags["container_type"] = c.criRuntimeVersion.RuntimeName
	} else {
		l.Debug("containedlog runtime: default 'containerd'")
		info.tags["container_type"] = "containerd"
	}
	// add extra tags
	for k, v := range c.ipt.Tags {
		if _, ok := info.tags[k]; !ok {
			info.tags[k] = v
		}
	}

	opt := composeTailerOption(c.k8sClient, info)
	opt.Mode = tailer.ContainerdMode
	opt.BlockingMode = c.ipt.LoggingBlockingMode
	opt.Done = c.ipt.semStop.Wait()

	t, err := tailer.NewTailerSingle(info.logPath, opt)
	if err != nil {
		l.Warnf("failed to new containerd log, containerId: %s, source: %s, logpath: %s, err: %s", status.Id, opt.Source, info.logPath, err)
		return err
	}

	c.addToLogList(info.logPath)
	l.Infof("add containerd log, containerId: %s, source: %s, logpath: %s", status.Id, opt.Source, info.logPath)
	defer func() {
		c.removeFromLogList(info.logPath)
		l.Infof("remove containerd log, containerId: %s, source: %s, logpath: %s", status.Id, opt.Source, info.logPath)
	}()

	t.Run()
	return nil
}
