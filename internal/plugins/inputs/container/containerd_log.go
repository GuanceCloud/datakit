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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
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

func (c *containerdInput) tailingLogs(info *containerLogInfo) {
	g := goroutine.NewGroup(goroutine.Option{Name: "containerd-logs/" + info.containerName})
	done := make(chan interface{})

	for _, cfg := range info.logConfigs {
		if cfg.Disable {
			continue
		}

		if c.logTable.inTable(info.id, cfg.Path) {
			continue
		}

		opt := &tailer.Option{
			Source:                   cfg.Source,
			Service:                  cfg.Service,
			Pipeline:                 cfg.Pipeline,
			CharacterEncoding:        cfg.CharacterEncoding,
			MultilinePatterns:        cfg.MultilinePatterns,
			GlobalTags:               cfg.Tags,
			BlockingMode:             c.ipt.LoggingBlockingMode,
			MinFlushInterval:         c.ipt.LoggingMinFlushInterval,
			MaxMultilineLifeDuration: c.ipt.LoggingMaxMultilineLifeDuration,
			Done:                     done,
		}

		if cfg.Type == "file" {
			opt.Mode = tailer.FileMode
		} else {
			opt.Mode = tailer.ContainerdMode
		}
		_ = opt.Init()

		path := logsJoinRootfs(cfg.Path)

		tail, err := tailer.NewTailerSingle(path, opt)
		if err != nil {
			l.Errorf("failed to create containerd-log collection %s for %s, err: %s", path, info.containerName, err)
			continue
		}

		c.logTable.addToTable(info.id, cfg.Path, done)

		g.Go(func(ctx context.Context) error {
			defer c.logTable.removePathFromTable(info.id, cfg.Path)
			tail.Run()
			return nil
		})
	}
}

func (c *containerdInput) queryContainerLogInfo(resp *cri.ContainerStatusResponse) *containerLogInfo {
	status := resp.GetStatus()

	var originalName string
	if status.GetMetadata() != nil {
		originalName = status.GetMetadata().GetName()
	}

	labels := status.GetLabels()
	info := &containerLogInfo{
		runtimeType:   "containerd",
		id:            status.GetId(),
		originalName:  originalName,
		containerName: getContainerNameForLabels(labels),
		podName:       getPodNameForLabels(labels),
		podNamespace:  getPodNamespaceForLabels(labels),
		logPath:       status.GetLogPath(),
	}

	if info.containerName == "" {
		info.containerName = originalName
	}

	if status.GetImage() != nil {
		info.image = status.GetImage().GetImage()
	}

	if c.k8sClient != nil && info.podName != "" {
		meta, err := queryPodMetaData(c.k8sClient, info.podName, info.podNamespace)
		if err != nil {
			l.Warnf("failed to query containerd %s info from k8s, err: %s, skip", info.containerName, err)
		} else {
			img := meta.containerImage(info.containerName)
			if img != "" {
				info.image = img
			}

			annotations := meta.annotations()

			// ex: datakit/logs
			if v := annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, "")]; v != "" {
				info.logConfigStr = v
			}

			// ex: datakit/nginx.logs
			if v := annotations[fmt.Sprintf(logConfigAnnotationKeyFormat, info.containerName+".")]; v != "" {
				info.logConfigStr = v
			}
		}
	}

	// ex: DATAKIT_LOGS_CONFIG
	if in := resp.GetInfo(); in != nil {
		criInfo, err := parseCriInfo(in["info"])
		if err != nil {
			l.Warnf("unable to parse containerd %s info, err: %s, skip", info.containerName, err)
		} else {
			if v := criInfo.Config.Envs.Find("DATAKIT_LOGS_CONFIG"); v != "" {
				info.logConfigStr = v
			}
		}
	}

	l.Debugf("containerd %s use logConfig: '%s'", info.containerName, info.logConfigStr)
	return info
}
