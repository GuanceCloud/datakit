// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"time"

	"github.com/kardianos/service"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

func runServiceFlags() error {
	if *flagServiceRestart {
		if err := RestartDatakit(); err != nil {
			cp.Errorf("[E] restart DataKit failed:%s\n using command to restart: %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		cp.Infof("Restart DataKit OK\n")
		os.Exit(0)
	}

	if *flagServiceStop {
		if err := stopDatakit(); err != nil {
			cp.Errorf("[E] stop DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Stop DataKit OK\n")
		os.Exit(0)
	}

	if *flagServiceStart {
		if err := startDatakit(); err != nil {
			cp.Errorf("[E] start DataKit failed: %s\n using command to stop : %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		cp.Infof("Start DataKit OK\n") // TODO: 需说明 PID 是多少
		os.Exit(0)
	}

	if *flagServiceUninstall {
		if err := uninstallDatakit(); err != nil {
			cp.Errorf("[E] uninstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Uninstall DataKit OK\n")
		os.Exit(0)
	}

	if *flagServiceReinstall {
		tryLoadMainCfg()

		if err := reinstallDatakit(config.Cfg); err != nil {
			cp.Errorf("[E] reinstall DataKit failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Reinstall DataKit OK\n")
		os.Exit(0)
	}

	return fmt.Errorf("no action specified")
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

	status, err := svc.Status()
	if err != nil {
		return err
	}

	if status == service.StatusStopped {
		return nil
	}

	l.Info("stoping datakit...")
	// 不能一直等待阻塞的 chan 或者 waitgroup到超时时间被强制 kill 时才退出
	errChan := make(chan error, 1)

	g.Go(func(ctx context.Context) error {
		errChan <- service.Control(svc, "stop")
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

	if err := service.Control(svc, "start"); err != nil {
		return err
	}

	return nil
}

func RestartDatakit() error {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Restart-Service", "datakit"}...)
		return cmd.Run()
	}

	if err := stopDatakit(); err != nil {
		return err
	}

	if err := startDatakit(); err != nil {
		return err
	}

	return nil
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
			dkservice.WithCPULimit(fmt.Sprintf("%f%%",
				resourcelimit.CPUCoresToCPUMax(mc.ResourceLimitOptions.CPUCores))),
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
