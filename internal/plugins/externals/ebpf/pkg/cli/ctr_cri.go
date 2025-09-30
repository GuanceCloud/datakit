package cli

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/goccy/go-json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var CriRuntimeEndpoint = []string{
	"unix:///var/run/dockershim.sock",
	"unix:///var/run/cri-dockerd.sock",
	"unix:///var/run/containerd/containerd.sock",
	"unix:///var/run/crio/crio.sock",
}

type CRIClient struct {
	cli     runtime.RuntimeServiceClient
	timeout time.Duration
}

func NewCRIDefault(endpoint ...string) (criLi []*CRIClient, warns []error) {
	if len(endpoint) == 0 {
		endpoint = append(endpoint, CriRuntimeEndpoint...)
	}
	for _, v := range endpoint {
		if cri, err := NewCRIClient(v); err != nil {
			warns = append(warns, err)
		} else {
			criLi = append(criLi, cri)
		}
	}
	return criLi, warns
}

func NewCRIClient(criEndpoint string, timeout ...time.Duration) (*CRIClient, error) {
	// parse unix socket
	u, err := url.Parse(criEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CRI endpoint: %w", err)
	}
	if u.Scheme != "unix" {
		return nil, fmt.Errorf("invalid CRI endpoint: %s", criEndpoint)
	}

	if _, err := os.Stat(u.Path); err != nil {
		return nil, fmt.Errorf("failed to stat CRI endpoint: %w", err)
	}

	var dur time.Duration
	if len(timeout) > 0 {
		dur = timeout[0]
	} else {
		dur = time.Second * 5
	}

	conn, err := grpc.Dial(criEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial CRI endpoint: %w", err)
	}

	c := &CRIClient{
		cli:     runtime.NewRuntimeServiceClient(conn),
		timeout: dur,
	}

	// check if the CRI endpoint is working
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := c.cli.Version(ctx, &runtime.VersionRequest{}); err != nil {
		return nil, fmt.Errorf("failed to get CRI version: %w", err)
	}

	return c, nil
}

func (cri *CRIClient) ListContainers() ([]*runtime.Container, error) {
	if cri.cli == nil {
		return nil, errors.New("cri client is nil")
	}

	ctx := context.Background()
	if cri.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cri.timeout)
		defer cancel()
	}

	resp, err := cri.cli.ListContainers(ctx, &runtime.ListContainersRequest{
		Filter: &runtime.ContainerFilter{
			State: &runtime.ContainerStateValue{
				State: runtime.ContainerState_CONTAINER_RUNNING,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return resp.GetContainers(), nil
}

type criInfo struct {
	Pid    int `json:"pid"`
	Config struct {
		Envs []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"envs"`
	} `json:"config"`
}

func (cri *CRIClient) GetContainerPID(containerID string) (int, error) {
	if cri.cli == nil {
		return 0, errors.New("cri client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cri.timeout)
	defer cancel()

	response, err := cri.cli.ContainerStatus(ctx, &runtime.ContainerStatusRequest{
		ContainerId: containerID,
		Verbose:     true,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get container status: %w", err)
	}
	infoStr := response.GetInfo()["info"]
	if infoStr == "" {
		return 0, fmt.Errorf("failed to get container info")
	}

	var info criInfo
	if err := json.Unmarshal([]byte(infoStr), &info); err != nil {
		return 0, fmt.Errorf("failed to unmarshal container info: %w", err)
	}

	// 返回容器的 PID
	return info.Pid, nil
}
