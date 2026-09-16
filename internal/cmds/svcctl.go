// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"time"

	"github.com/kardianos/service"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

type ServiceAction string

const (
	ServiceActionStart     ServiceAction = "start"
	ServiceActionStop      ServiceAction = "stop"
	ServiceActionRestart   ServiceAction = "restart"
	ServiceActionUninstall ServiceAction = "uninstall"
	ServiceActionReinstall ServiceAction = "reinstall"
)

type ServiceOptions struct {
	LogPath string
	Action  ServiceAction
}

func RunService(opts ServiceOptions) error {
	ConfigureCommandLog(opts.LogPath)

	switch opts.Action {
	case ServiceActionRestart:
		if err := RestartDatakit(); err != nil {
			return fmt.Errorf("restart DataKit failed: %w; using command to restart: %s", err, serviceCommandHint(runtime.GOOS, opts.Action))
		}
		cp.Infof("Restart DataKit OK\n")
		return nil
	case ServiceActionStop:
		if err := stopDatakit(); err != nil {
			return fmt.Errorf("stop DataKit failed: %w", err)
		}
		cp.Infof("Stop DataKit OK\n")
		return nil
	case ServiceActionStart:
		if err := startDatakit(); err != nil {
			return fmt.Errorf("start DataKit failed: %w; using command to start: %s", err, serviceCommandHint(runtime.GOOS, opts.Action))
		}
		cp.Infof("Start DataKit OK\n")
		return nil
	case ServiceActionUninstall:
		if err := uninstallDatakit(); err != nil {
			return fmt.Errorf("uninstall DataKit failed: %w", err)
		}
		cp.Infof("Uninstall DataKit OK\n")
		return nil
	case ServiceActionReinstall:
		tryLoadMainCfg()
		if err := reinstallDatakit(config.Cfg); err != nil {
			return fmt.Errorf("reinstall DataKit failed: %w", err)
		}
		cp.Infof("Reinstall DataKit OK\n")
		return nil
	default:
		return fmt.Errorf("unknown service action: %s", opts.Action)
	}
}

func serviceCommandHint(goos string, action ServiceAction) string {
	switch goos {
	case datakit.OSWindows:
		if action == ServiceActionRestart {
			return "Restart-Service -Name datakit"
		}
		return "Start-Service -Name datakit"
	case datakit.OSDarwin:
		const start = "sudo launchctl load -w /Library/LaunchDaemons/com.datakit.plist"
		if action == ServiceActionRestart {
			return "sudo launchctl unload -w /Library/LaunchDaemons/com.datakit.plist && " + start
		}
		return start
	default:
		return "systemctl " + string(action) + " datakit"
	}
}

func isRoot() error {
	if runtime.GOOS == datakit.OSWindows {
		return nil // under windows, there is no root user
	}

	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("get user failed: %w", err)
	}

	if !datakit.IsAdminUser(u.Username) {
		return fmt.Errorf("not admin user, current is %s", u.Username)
	}

	return nil
}

func stopDatakit() error {
	if err := isRoot(); err != nil {
		return err
	}

	// BUG: current service package can't Control service under windows, we use powershell's command instead
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Stop-Service", "datakit"}...)
		return cmd.Run()
	}

	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	return stopService(svc)
}

func stopService(svc service.Service) error {
	if svc.Platform() != "linux-systemd" {
		status, err := svc.Status()
		if err != nil {
			return err
		}
		if status == service.StatusStopped {
			return nil
		}
	}

	l.Info("stoping datakit...")
	// 不能一直等待阻塞的 chan 或者 waitgroup到超时时间被强制 kill 时才退出
	errChan := make(chan error, 1)

	g.Go(func(ctx context.Context) error {
		errChan <- svc.Stop()
		return nil
	})

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-time.After(time.Second * 30):
		return errors.New("datakit.service stop-sigterm timed out")
	}
	return nil
}

func startDatakit() error {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Start-Service", "datakit"}...)
		return cmd.Run()
	}

	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	return startService(svc)
}

func startService(svc service.Service) error {
	// systemd can start failed units. The service library reports that state as
	// an error, so a Status preflight would prevent recovery before Start runs.
	if svc.Platform() == "linux-systemd" {
		return svc.Start()
	}

	status, err := svc.Status()
	if err != nil {
		return err
	}

	if status == service.StatusRunning {
		l.Info("datakit service is already running")
		return nil
	}

	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service failed: %s, ignored", err)
	}

	if err := svc.Start(); err != nil {
		return err
	}

	return nil
}

func RestartDatakit() error {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Restart-Service", "datakit"}...)
		return cmd.Run()
	}

	if err := isRoot(); err != nil {
		return err
	}
	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	return restartService(svc)
}

func restartService(svc service.Service) error {
	// Let systemd manage the restart job and its configured stop timeout,
	// including recovery from a unit that is already in the failed state.
	if svc.Platform() == "linux-systemd" {
		return svc.Restart()
	}

	if err := stopService(svc); err != nil {
		return err
	}

	return startService(svc)
}

func uninstallDatakit() error {
	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	if err := service.Control(svc, "stop"); err != nil {
		return err
	}

	l.Info("uninstall datakit...")
	return service.Control(svc, "uninstall")
}

func reinstallDatakit(mc *config.Config) error {
	var opts []dkservice.ServiceOption

	if mc.ResourceLimitOptions.Enable {
		opts = append(opts,
			dkservice.WithMemLimit(fmt.Sprintf("%dM", mc.ResourceLimitOptions.MemMax)),
			dkservice.WithCPULimit(mc.ResourceLimitOptions.CPUCores),
		)
	}

	if runtime.GOOS == datakit.OSLinux { // only linux add user to daemon service
		l.Infof("reinstallDatakit with user: %s", mc.DatakitUser)
		opts = append(opts, dkservice.WithUser(mc.DatakitUser))
	}

	svc, err := dkservice.NewService(opts...)
	if err != nil {
		return err
	}

	l.Info("re-install datakit...")
	if err := service.Control(svc, "install"); err != nil {
		return err
	}

	return service.Control(svc, "start")
}
