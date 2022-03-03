package container

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
)

var (
	version        = "1.21" // 1.24 is when server first started returning its version
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type dockerClientX interface {
	Info(ctx context.Context) (types.Info, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerTop(ctx context.Context, containerID string, arguments []string) (container.ContainerTopOKBody, error)
	ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

type dockerClient struct {
	client *docker.Client
}

// nolint (is unused)
func newEnvClient() (*dockerClient, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &dockerClient{client}, nil
}

func newDockerClient(host string, tlsConfig *tls.Config) (*dockerClient, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}

	client, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithHTTPClient(httpClient),
		docker.WithVersion(version),
		docker.WithHost(host))
	if err != nil {
		return nil, err
	}

	return &dockerClient{client}, nil
}

func (c *dockerClient) Info(ctx context.Context) (types.Info, error) {
	return c.client.Info(ctx)
}

func (c *dockerClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return c.client.ContainerList(ctx, options)
}

func (c *dockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return c.client.ContainerStats(ctx, containerID, stream)
}

func (c *dockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.client.ContainerInspect(ctx, containerID)
}

func (c *dockerClient) ContainerTop(ctx context.Context, containerID string, arguments []string) (container.ContainerTopOKBody, error) {
	return c.client.ContainerTop(ctx, containerID, arguments)
}

func (c *dockerClient) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}
