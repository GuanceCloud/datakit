//go:build linux
// +build linux

package l4log

import (
	"fmt"
	"net/url"
	"os"

	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

type containerRuntime interface {
	ListContainers() ([]containerSnapshot, error)
	Close() error
}

type containerSnapshot struct {
	ID     string
	Pid    int
	Labels map[string]string
}

type containerRuntimeAdapter struct {
	rt cruntime.ContainerRuntime
}

func newContainerRuntime(endpoint string) (containerRuntime, error) {
	var (
		rt  cruntime.ContainerRuntime
		err error
	)
	if verifyErr := cruntime.VerifyDockerRuntime(endpoint); verifyErr == nil {
		rt, err = cruntime.NewDockerRuntime(endpoint, "")
	} else {
		rt, err = cruntime.NewCRIRuntime(endpoint, "")
	}
	if err != nil {
		return nil, err
	}
	return &containerRuntimeAdapter{rt: rt}, nil
}

func newContainerRuntimes(endpoints []string) []containerRuntime {
	var runtimes []containerRuntime
	for _, ep := range endpoints {
		if err := checkEndpoint(ep); err != nil {
			log.Warnf("skip connect to %s: %s", ep, err.Error())
			continue
		}

		rt, err := newContainerRuntime(ep)
		if err != nil {
			log.Warnf("skip connect to %s: %s", ep, err.Error())
			continue
		}

		log.Infof("connect to %s success", ep)
		runtimes = append(runtimes, rt)
	}
	return runtimes
}

func (a *containerRuntimeAdapter) ListContainers() ([]containerSnapshot, error) {
	if a == nil || a.rt == nil {
		return nil, fmt.Errorf("container runtime is nil")
	}

	containers, err := a.rt.ListContainers()
	if err != nil {
		return nil, err
	}

	snapshots := make([]containerSnapshot, 0, len(containers))
	for _, ctr := range containers {
		if ctr == nil {
			continue
		}
		snapshots = append(snapshots, containerSnapshot{
			ID:     ctr.ID,
			Pid:    ctr.Pid,
			Labels: ctr.Labels,
		})
	}
	return snapshots, nil
}

func (a *containerRuntimeAdapter) Close() error {
	if a == nil || a.rt == nil {
		return nil
	}
	closer, ok := a.rt.(interface{ Close() error })
	if !ok || closer == nil {
		return nil
	}
	return closer.Close()
}

func closeContainerRuntimes(runtimes []containerRuntime) {
	for _, rt := range runtimes {
		if rt == nil {
			continue
		}
		if err := rt.Close(); err != nil {
			log.Warnf("close container runtime failed: %v", err)
		}
	}
}

// checkEndpoint check if endpoint is valid, copy from internal/plugins/inputs/container/impl.go
func checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s, err: %w", endpoint, err)
	}

	switch u.Scheme {
	case "unix":
		// nil
	default:
		return fmt.Errorf("using %s as endpoint is not supported protocol", endpoint)
	}

	info, err := os.Stat(u.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("endpoint %s does not exist, maybe it is not running", endpoint)
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("endpoint %s cannot be a directory", u.Path)
	}

	return nil
}
