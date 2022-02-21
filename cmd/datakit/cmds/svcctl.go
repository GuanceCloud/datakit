package cmds

import (
	"errors"
	"fmt"
	"os/exec"
	"os/user"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

func runServiceFlags() error {
	if *flagServiceRestart {
		return restartDatakit()
	}

	if *flagServiceStop {
		return stopDatakit()
	}

	if *flagServiceStart {
		return startDatakit()
	}

	if *flagServiceUninstall {
		return uninstallDatakit()
	}

	if *flagServiceReinstall {
		return reinstallDatakit()
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

	if u.Username != "root" {
		return fmt.Errorf("not root user, current is %s", u.Username)
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
	go func() {
		errChan <- service.Control(svc, "stop")
	}()
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-time.After(time.Second * 10):
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

func restartDatakit() error {
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

func reinstallDatakit() error {
	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	l.Info("re-install datakit...")
	if err := service.Control(svc, "install"); err != nil {
		return err
	}

	return service.Control(svc, "start")
}

func datakitStatus() (string, error) {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Get-Service", "datakit"}...)
		res, err := cmd.CombinedOutput()
		return string(res), err
	}

	svc, err := dkservice.NewService()
	if err != nil {
		return "", err
	}

	status, err := svc.Status()
	if err != nil {
		return "", err
	}
	switch status {
	case service.StatusUnknown:
		return "unknown", nil
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	default:
		return "", fmt.Errorf("should not been here")
	}
}
