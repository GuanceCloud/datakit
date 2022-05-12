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
		return err
	}

	for _, resp := range list.GetContainers() {
		status, err := c.criClient.ContainerStatus(context.Background(), &cri.ContainerStatusRequest{ContainerId: resp.Id})
		if err != nil {
			l.Warn(err)
			continue
		}

		logpath := status.GetStatus().GetLogPath()

		if c.inLogList(logpath) {
			continue
		}

		tags := map[string]string{
			// "container_type": "containerd",
			// "state":          "running",
			"container_name": getContainerNameForLabels(status.GetStatus().GetLabels()),
			"container_id":   status.GetStatus().GetId(),
			"pod_name":       getPodNameForLabels(status.GetStatus().GetLabels()),
			"namespace":      getPodNamespaceForLabels(status.GetStatus().GetLabels()),
		}
		// add extra tags
		for k, v := range c.cfg.extraTags {
			if _, ok := tags[k]; !ok {
				tags[k] = v
			}
		}

		if image := status.GetStatus().GetImage(); image != nil {
			// 如果能找到 pod image，则使用它
			imageName, imageShortName, imageTag := ParseImage(image.Image)
			tags["image"] = image.Image
			tags["image_name"] = imageName
			tags["image_short_name"] = imageShortName
			tags["image_tag"] = imageTag
		}

		source := getContainerNameForLabels(status.GetStatus().GetLabels())
		if n := status.GetStatus().Metadata; n != nil {
			source = n.Name
		}

		opt := &tailer.Option{
			Source:     source,
			GlobalTags: tags,
		}

		logconf, err := getContainerLogConfig(status.GetStatus().GetAnnotations())
		if err != nil {

			l.Warn(err)
			continue
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

			l.Debugf("use container logconfig:%#v, containerName:%s", logconf, tags["container_name"])
		}

		_ = opt.Init()

		t, err := tailer.NewTailerSingle(logpath, opt)
		if err != nil {
			l.Warn(err)
			continue
		}

		c.addToLogList(logpath)
		go func(logpath string) {
			defer c.removeFromLogList(logpath)
			t.Run()
		}(logpath)
	}

	return nil
}
