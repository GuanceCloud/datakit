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
	"google.golang.org/grpc/credentials/insecure"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	kubeRuntimeAPIVersion = "0.1.0"
	maxMsgSize            = 1024 * 1024 * 16
)

func newCRIClient(endpoint string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return grpc.DialContext(
		ctx,
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dial),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)),
	)
}

func getCRIRuntimeVersion(client cri.RuntimeServiceClient) (*cri.VersionResponse, error) {
	ctx, cancel := getContextWithTimeout(time.Second * 10)
	defer cancel()
	return client.Version(ctx, &cri.VersionRequest{Version: kubeRuntimeAPIVersion})
}

func getContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func dial(ctx context.Context, addr string) (net.Conn, error) {
	var timeout time.Duration = 0
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
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
	var name string
	if status.GetMetadata() != nil && status.GetMetadata().Name != "" {
		name = status.GetMetadata().Name
	} else {
		name = "unknown"
	}

	logpath := logsJoinRootfs(status.GetLogPath())

	if !tailer.FileIsActive(logpath, ignoreDeadLogDuration) {
		l.Debugf("container %s file %s is not active, larger than %s, ignored", name, logpath, ignoreDeadLogDuration)
		return nil
	}

	info := &containerLogBasisInfo{
		name:                  name,
		id:                    status.GetId(),
		logPath:               logpath,
		labels:                status.GetLabels(),
		tags:                  make(map[string]string),
		extraSourceMap:        c.ipt.LoggingExtraSourceMap,
		sourceMultilineMap:    c.ipt.LoggingSourceMultilineMap,
		autoMultilinePatterns: c.ipt.getAutoMultilinePatterns(),
		extractK8sLabelAsTags: c.ipt.ExtractK8sLabelAsTags,
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

	opt, _ := composeTailerOption(c.k8sClient, info)
	opt.Mode = tailer.ContainerdMode
	opt.BlockingMode = c.ipt.LoggingBlockingMode
	opt.MinFlushInterval = c.ipt.LoggingMinFlushInterval
	opt.MaxMultilineLifeDuration = c.ipt.LoggingMaxMultilineLifeDuration
	opt.Done = c.ipt.semStop.Wait()
	_ = opt.Init()

	l.Debugf("use container-log opt:%#v, containerId: %s", opt, status.GetId())

	t, err := tailer.NewTailerSingle(info.logPath, opt)
	if err != nil {
		l.Warnf("failed to new containerd log, containerId: %s, source: %s, logpath: %s, err: %s", status.Id, opt.Source, info.logPath, err)
		return err
	}

	// 这里添加原始 logpath，而不是修改过的
	c.addToLogList(status.GetLogPath())
	l.Infof("add containerd log, containerId: %s, source: %s, logpath: %s", status.Id, opt.Source, info.logPath)
	defer func() {
		c.removeFromLogList(info.logPath)
		l.Infof("remove containerd log, containerId: %s, source: %s, logpath: %s", status.Id, opt.Source, info.logPath)
	}()

	t.Run()
	return nil
}
