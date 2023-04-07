// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

const (
	ENVDatakitHome    = "DATAKIT_HOME"
	ENVDKUpgraderHome = "DKUPGRADER_HOME"
)

var (
	windowsCmdErrMsg = "Stop-Service -Name " + ServiceName
	darwinCmdErrMsg  = "sudo launchctl unload /Library/LaunchDaemons/" + DarwinServiceName + ".plist"
	linuxCmdErrMsg   = "systemctl stop " + ServiceName

	errMsg = map[string]string{
		datakit.OSWindows: windowsCmdErrMsg,
		datakit.OSLinux:   linuxCmdErrMsg,
		datakit.OSDarwin:  darwinCmdErrMsg,
	}

	fsHelpName = "help"

	fsENVName          = "env"
	fsENV              = pflag.NewFlagSet(fsENVName, pflag.ContinueOnError)
	flagENVPath        = fsENV.String("PATH", "", "set service runtime PATH environment variable, append to current PATH")
	flagENVDatakitHome = fsENV.String(ENVDatakitHome, "", "set datakit home env variable")
	flagENVHomeDir     = fsENV.String(ENVDKUpgraderHome, "", "set dk_upgrader home env variable")
	fsENVUsage         = func() {
		fmt.Printf("usage: dk_upgrader env [options]\n\n")
		fmt.Printf("set dk_upgrader service runtime environmant virables\n\n")
		fmt.Println(fsENV.FlagUsagesWrapped(0))
	}

	//
	// service management related flags.
	//
	fsServiceName        = "service"
	fsService            = pflag.NewFlagSet(fsServiceName, pflag.ContinueOnError)
	flagServiceStatus    = fsService.Bool("status", false, "Show dk_upgrader service runtime status")
	flagServiceRestart   = fsService.BoolP("restart", "R", false, "restart dk_upgrader service")
	flagServiceStop      = fsService.BoolP("stop", "T", false, "stop dk_upgrader service")
	flagServiceStart     = fsService.BoolP("start", "S", false, "start dk_upgrader service")
	flagServiceUninstall = fsService.BoolP("uninstall", "U", false, "uninstall dk_upgrader service")
	flagServiceReinstall = fsService.BoolP("reinstall", "I", false, "reinstall dk_upgrader service")
	fsServiceUsage       = func() {
		fmt.Printf("usage: dk_upgrader service [options]\n\n")
		fmt.Printf("Service used to manage dk_upgrader service\n\n")
		fmt.Println(fsService.FlagUsagesWrapped(0))
	}

	g = datakit.G("control")
)

func setENVVariables() {
	if *flagENVPath != "" {
		pathListSeparator := string(os.PathListSeparator)
		envPATH := strings.Trim(*flagENVPath, pathListSeparator)
		if envPATH != "" {
			curPath := strings.TrimRight(os.Getenv("PATH"), pathListSeparator) + pathListSeparator + envPATH
			curPath = strings.Trim(curPath, pathListSeparator)
			if err := os.Setenv("PATH", curPath); err != nil {
				L().Errorf("unable to set PATH env variable: %s", err)
			}
		}
	}

	if *flagENVDatakitHome != "" {
		if err := os.Setenv(ENVDatakitHome, strings.TrimSpace(*flagENVDatakitHome)); err != nil {
			L().Warnf("unable to set %s envronment variable: %s", ENVDatakitHome, err)
		}
	}

	if *flagENVHomeDir != "" {
		if err := os.Setenv(ENVDKUpgraderHome, strings.TrimSpace(*flagENVHomeDir)); err != nil {
			L().Warnf("unable to set %s environment variable: %s", ENVDKUpgraderHome, err)
		}
	}
}

func runServiceFlags() error {
	if *flagServiceStatus {
		status, err := dKUpgraderStatus()
		if err != nil {
			cp.Errorf("unable to get dk_upgrader running status: %s", err)
			os.Exit(1)
		}
		cp.Infof("%s service is %s\n", ServiceName, status)
		os.Exit(0)
	}

	if *flagServiceRestart {
		if err := restartDKUpgrader(); err != nil {
			cp.Errorf("[E] restart dk_upgrader failed:%s\n using command to restart: %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		cp.Infof("Restart dk_upgrader OK\n")
		os.Exit(0)
	}

	if *flagServiceStop {
		if err := stopDKUpgrader(); err != nil {
			cp.Errorf("[E] stop dk_upgrader failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Stop dk_upgrader OK\n")
		os.Exit(0)
	}

	if *flagServiceStart {
		if err := startDKUpgrader(); err != nil {
			cp.Errorf("[E] start dk_upgrader failed: %s\n using command to stop : %s\n", err.Error(), errMsg[runtime.GOOS])
			os.Exit(-1)
		}

		cp.Infof("Start dk_upgrader OK\n") // TODO: 需说明 PID 是多少
		os.Exit(0)
	}

	if *flagServiceUninstall {
		if err := uninstallDKUpgrader(); err != nil {
			cp.Errorf("[E] uninstall dk_upgrader failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Uninstall dk_upgrader OK\n")
		os.Exit(0)
	}

	if *flagServiceReinstall {
		if err := reinstallDKUpgrader(); err != nil {
			cp.Errorf("[E] reinstall dk_upgrader failed: %s\n", err.Error())
			os.Exit(-1)
		}

		cp.Infof("Reinstall dk_upgrader OK\n")
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

	if u.Username != "root" {
		return fmt.Errorf("not root user, current is %s", u.Username)
	}

	return nil
}

func stopDKUpgrader() error {
	if err := isRoot(); err != nil {
		return err
	}

	// BUG: current service package can't Control service under windows, we use powershell's command instead
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Stop-Service", ServiceName}...)
		return cmd.Run()
	}

	svc, err := NewDefaultService("", nil)
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

	L().Info("stoping dk_upgrader...")
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
		return errors.New("dk_upgrader.service stop-sigterm timed out")
	}
	return nil
}

func startDKUpgrader() error {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Start-Service", ServiceName}...)
		return cmd.Run()
	}

	svc, err := NewDefaultService("", nil)
	if err != nil {
		return err
	}

	status, err := svc.Status()
	if err != nil {
		return err
	}

	if status == service.StatusRunning {
		L().Info("dk_upgrader service is already running")
		return nil
	}

	if err := service.Control(svc, "start"); err != nil {
		return err
	}

	return nil
}

func restartDKUpgrader() error {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Restart-Service", ServiceName}...)
		return cmd.Run()
	}

	if err := stopDKUpgrader(); err != nil {
		return err
	}

	if err := startDKUpgrader(); err != nil {
		return err
	}

	return nil
}

func uninstallDKUpgrader() error {
	svc, err := NewDefaultService("", nil)
	if err != nil {
		return err
	}

	if err := service.Control(svc, "stop"); err != nil {
		L().Warnf("unable to stop service: %s", err)
	}

	L().Info("uninstall dk_upgrader...")
	return service.Control(svc, "uninstall")
}

func reinstallDKUpgrader() error {
	svc, err := NewDefaultService("", nil)
	if err != nil {
		return err
	}

	L().Info("re-install dk_upgrader...")
	if err := service.Control(svc, "install"); err != nil {
		return err
	}

	return service.Control(svc, "start")
}

func dKUpgraderStatus() (string, error) {
	if runtime.GOOS == datakit.OSWindows {
		cmd := exec.Command("powershell", []string{"Get-Service", ServiceName}...)
		res, err := cmd.CombinedOutput()
		return string(res), err
	}

	svc, err := NewDefaultService("", nil)
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

func printHelp(w io.Writer) {
	fmt.Fprintf(w, "\nUsage:\n\n")

	fmt.Fprintf(w, "\tdk_upgrader <command> [arguments]\n\n")

	fmt.Fprintf(w, "The commands are:\n\n")

	fmt.Fprintf(w, "\tservice    manage dk_upgrader service\n")

	fmt.Fprintf(w, "\tenv    set service runtime environment variables\n")

	// TODO: add more commands...

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Use 'dk_upgrader help <command>' for more information about a command.\n\n")
}

func runHelpFlags() {
	switch len(os.Args) {
	case 2: // only 'datakit help'
		printHelp(os.Stdout)
	case 3: // need help for various commands
		switch os.Args[2] {
		case fsServiceName:
			fsServiceUsage()

		case fsENVName:
			fsENVUsage()

		default: // add more
			cp.Errorf("[E] flag provided but not defined: `%s'\n\n", os.Args[2])
			printHelp(os.Stderr)
			os.Exit(1)
		}
	}
}

func ParseAndRunSubCommand() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case fsHelpName:
			runHelpFlags()
			os.Exit(0)

		case fsServiceName:
			if err := fsService.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsServiceUsage()
				os.Exit(-1)
			}

			if err := runServiceFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsENVName:
			if err := fsENV.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Unable to parse env parameters: %s", err)
				fsENVUsage()
				os.Exit(1)
			}
			setENVVariables()
			// We should not call os.Exit() here

		default:
			cp.Errorf("unknown subcommand: %s", os.Args[1])
			printHelp(os.Stderr)
			os.Exit(1)
		}
	}
}
