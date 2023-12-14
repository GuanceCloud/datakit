// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

const (
	ECSFargateRuntime = "ecsfargate"

	taskPath  = "/task"
	statsPath = "/stats"
)

type ecsfargateClient struct {
	agentURL string
	baseURL  string
	client   http.RoundTripper
}

func NewECSFargateRuntime(agentURL string) (ContainerRuntime, error) {
	u, err := url.Parse(agentURL)
	if err != nil {
		return nil, fmt.Errorf("parse ecsfargate url error: %w", err)
	}

	c := &ecsfargateClient{
		agentURL: u.String(),
		client:   dkhttp.DefTransport(),
	}

	baseURL, err := c.parseBaseURL(u.String())
	if err != nil {
		return nil, err
	}

	c.baseURL = baseURL
	return c, nil
}

func (c *ecsfargateClient) ListContainers() ([]*Container, error) {
	urlstr := makeURL(c.agentURL, taskPath)

	var v ECSFargateTask
	if err := c.get(context.Background(), urlstr, &v); err != nil {
		return nil, err
	}

	var res []*Container

	for _, container := range v.Containers {
		r := &Container{
			ID:          container.DockerID,
			Name:        container.Name,
			Image:       container.Image,
			Labels:      copyMap(container.Labels),
			RuntimeName: ECSFargateRuntime,
			State:       container.DesiredStatus,
		}

		tim, err := time.Parse(time.RFC3339Nano, container.CreatedAt)
		if err == nil {
			r.CreatedAt = tim.UnixNano()
		}

		res = append(res, r)
	}

	return res, nil
}

func (c *ecsfargateClient) ContainerStatus(id string) (*ContainerStatus, error) {
	return nil, fmt.Errorf("ecsfargate not available status")
}

func (c *ecsfargateClient) ContainerTop(id string) (*ContainerTop, error) {
	urlstr := makeURL(c.baseURL, id, statsPath)

	var stats dockertypes.StatsJSON
	if err := c.get(context.Background(), urlstr, &stats); err != nil {
		return nil, err
	}

	top := ContainerTop{ID: stats.ID}

	// cpu usage
	top.CPUUsage = calculateCPUPercentUnix(&stats)

	// cpu cores
	top.CPUCores = int(stats.CPUStats.OnlineCPUs)

	// memory usage and menory limit
	top.MemoryWorkingSet = calculateMemUsageUnixNoCache(stats.MemoryStats)
	top.MemoryLimit = int64(stats.MemoryStats.Stats["hierarchical_memory_limit"])
	top.MemoryCapacity = int64(math.MaxInt64)

	// block io
	top.BlockRead, top.BlockWrite = calculateBlockIO(stats.BlkioStats)

	// network
	var rx, tx uint64
	if len(stats.Networks) != 0 {
		for name, network := range stats.Networks {
			if name == "lo" {
				continue
			}
			rx += network.RxBytes
			tx += network.TxBytes
		}
	}
	top.NetworkRcvd = int64(rx)
	top.NetworkSent = int64(tx)

	return &top, nil
}

func (c *ecsfargateClient) parseBaseURL(agentURL string) (string, error) {
	var v ECSFargateContainer

	if err := c.get(context.Background(), agentURL, &v); err != nil {
		return "", fmt.Errorf("check agent err: %w", err)
	}

	id := v.DockerID
	if id == "" {
		return "", fmt.Errorf("unexpected DockerID, agent url %s", agentURL)
	}

	baseURL := strings.TrimSuffix(agentURL, id)
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url %s, err: %w", baseURL, err)
	}

	return u.String(), nil
}

func (c *ecsfargateClient) get(ctx context.Context, urlstr string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", urlstr, nil)
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}

	resp, err := c.client.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		var msg string
		if buf, err := io.ReadAll(resp.Body); err == nil {
			msg = string(buf)
		}
		return fmt.Errorf("unexpected HTTP response, url %s, status code %d, message: %s", urlstr, resp.StatusCode, msg)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode json, err: %w", err)
	}

	return nil
}

func makeURL(urlstr string, requestPaths ...string) string {
	u, err := url.Parse(urlstr)
	if err != nil {
		// unreachable
		return ""
	}

	// e.g. "/v4/<id>/stats"
	paths := []string{u.Path}
	paths = append(paths, requestPaths...)

	u.Path = path.Join(paths...)
	return u.String()
}

// ECSFargateTask represents a task as returned by the ECS metadata API v3 or v4.
type ECSFargateTask struct {
	ClusterName           string                `json:"Cluster"`
	Containers            []ECSFargateContainer `json:"Containers"`
	KnownStatus           string                `json:"KnownStatus"`
	TaskARN               string                `json:"TaskARN"`
	Family                string                `json:"Family"`
	Version               string                `json:"Revision"`
	Limits                map[string]float64    `json:"Limits,omitempty"`
	DesiredStatus         string                `json:"DesiredStatus"`
	LaunchType            string                `json:"LaunchType,omitempty"` // present only in v4
	ContainerInstanceTags map[string]string     `json:"ContainerInstanceTags,omitempty"`
	TaskTags              map[string]string     `json:"TaskTags,omitempty"`
}

// ECSFargateContainer represents a container within a task.
type ECSFargateContainer struct {
	DockerID      string            `json:"DockerID"`
	Name          string            `json:"Name"`
	DockerName    string            `json:"DockerName"`
	Image         string            `json:"Image"`
	ImageID       string            `json:"ImageID,omitempty"`
	Labels        map[string]string `json:"Labels,omitempty"`
	DesiredStatus string            `json:"DesiredStatus"`
	// See https://github.com/aws/amazon-ecs-agent/blob/master/agent/api/container/status/containerstatus.go
	KnownStatus  string              `json:"KnownStatus"`
	Limits       map[string]uint64   `json:"Limits,omitempty"`
	CreatedAt    string              `json:"CreatedAt,omitempty"`
	StartedAt    string              `json:"StartedAt,omitempty"` // 2017-11-17T17:14:07.781711848Z
	Type         string              `json:"Type"`
	ContainerARN string              `json:"ContainerARN,omitempty"` // present only in v4
	Networks     []ECSFargateNetwork `json:"Networks,omitempty"`
	Ports        []ECSFargatePort    `json:"Ports,omitempty"`
	LogDriver    string              `json:"LogDriver,omitempty"`   // present only in v4
	LogOptions   map[string]string   `json:"LogOptions,omitempty"`  // present only in v4
	Snapshotter  string              `json:"Snapshotter,omitempty"` // present only in v4
	ExitCode     string              `json:"ExitCode,omitempty"`
}

// ECSFargateNetwork represents the network of a container.
type ECSFargateNetwork struct {
	NetworkMode   string   `json:"NetworkMode"`   // supports awsvpc and bridge
	IPv4Addresses []string `json:"IPv4Addresses"` // one-element list
	MACAddress    string   `json:"MACAddress"`
}

// ECSFargatePort represents the ports of a container.
type ECSFargatePort struct {
	ContainerPort uint16 `json:"ContainerPort,omitempty"`
	Protocol      string `json:"Protocol,omitempty"`
	HostPort      uint16 `json:"HostPort,omitempty"`
}
