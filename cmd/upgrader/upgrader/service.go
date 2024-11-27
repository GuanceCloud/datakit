// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"gopkg.in/natefinch/lumberjack.v2"
	net2 "k8s.io/apimachinery/pkg/util/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

const (
	ExitStatusUnableToRun     = 101
	HTTPServerExitNotExpected = 107
	ExitStatusAlreadyRunning  = 120
)

var (
	l = logger.DefaultSLogger(ServiceName)

	PidFile = filepath.Join(InstallDir, ServiceName+".pid")
)

const (
	DatakitCmd         = "datakit"
	BuildEntranceFile  = "cmd/upgrader/main.go"
	BuildBinName       = "dk_upgrader"
	ServiceName        = BuildBinName
	DarwinServiceName  = "com.guance." + ServiceName
	ServiceDescription = "datakit upgrade service"
)

var optionalInstallDir = map[string]string{
	datakit.OSArchWinAmd64: `C:\Program Files\` + ServiceName,
	datakit.OSArchWin386:   `C:\Program Files (x86)\` + ServiceName,

	datakit.OSArchLinuxArm:    `/usr/local/` + ServiceName,
	datakit.OSArchLinuxArm64:  `/usr/local/` + ServiceName,
	datakit.OSArchLinuxAmd64:  `/usr/local/` + ServiceName,
	datakit.OSArchLinux386:    `/usr/local/` + ServiceName,
	datakit.OSArchDarwinAmd64: `/usr/local/` + ServiceName,
	datakit.OSArchDarwinArm64: `/usr/local/` + ServiceName,
}

var (
	InstallDir     = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	DefaultLogDir  = filepath.Join("/var/log", ServiceName)
	MainConfigFile = filepath.Join(InstallDir, "main.conf")

	defaultServiceOpts = map[string]interface{}{
		"RestartSec":         10, // 重启间隔.
		"StartLimitInterval": 60, // 60秒内5次重启之后便不再启动.
		"StartLimitBurst":    5,
		"OnFailure":          "restart", // windows
	}

	defaultServiceImpl = newProgram()
)

type serviceImpl struct {
	entry func(*serviceImpl)
	stop  chan struct{}
	done  chan struct{}
}

func newProgram() *serviceImpl {
	return &serviceImpl{
		entry: entryFunc,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

func createDirs() error {
	for _, dir := range []string{
		InstallDir,
		DefaultLogDir,
	} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create %s failed: %w", dir, err)
		}
	}
	return nil
}

func (p *serviceImpl) Start(s service.Service) error {
	if p.entry == nil {
		return fmt.Errorf("entry not set")
	}

	p.entry(p)
	return nil
}

func (p *serviceImpl) Stop(s service.Service) error {
	close(p.stop)

	// We must wait here:
	// On Windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-p.done
	return nil
}

func NewDefaultService(username string, args []string) (service.Service, error) {
	l = logger.SLogger(ServiceName)
	return NewService(defaultServiceImpl, username, args)
}

func NewService(program service.Interface, username string, args []string) (service.Service, error) {
	if program == nil {
		program = defaultServiceImpl
	}

	executable := filepath.Join(InstallDir, BuildBinName)
	if runtime.GOOS == datakit.OSWindows {
		executable += ".exe"
	}

	scfg := &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: ServiceDescription,
		Executable:  executable,
		Arguments:   args,
		Option:      defaultServiceOpts,
		UserName:    username,
	}

	if runtime.GOOS == "darwin" {
		scfg.Name = "com.guance." + ServiceName
	}

	svc, err := service.New(program, scfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create service: %w", err)
	}

	return svc, nil
}

func entryFunc(p *serviceImpl) {
	serv := startHTTPServer()

	go func() {
		if err := startDCA(p); err != nil {
			l.Errorf("startDCA failed: %s", err.Error())
		}
	}()

	go func() {
		select {
		case <-p.stop:
			if err := shutdownWithTimeout(serv, time.Second*5); err != nil {
				l.Warnf("datakit upgrade service shutdown err: %s", err)
			}
			close(p.done)

		case err := <-httpServClosed:
			l.Errorf("http server exit abnormal: %s", err)
			os.Exit(HTTPServerExitNotExpected)
		}
	}()
}

func nopMiddleware(c *gin.Context) {
	c.Next()
}

func getIPVerifyMiddleware() gin.HandlerFunc {
	if len(Cfg.IPWhiteList) == 0 {
		return nopMiddleware
	}

	ipWhiteListMap := make(map[string]struct{}, len(Cfg.IPWhiteList))

	for _, ip := range Cfg.IPWhiteList {
		// We use net.LookupHost func to check the ip validity
		addrs, err := net.LookupHost(ip)
		if err != nil {
			l.Warnf("the IP [%s] in ip_whitelist setting is illegal: %s", ip, err)
			continue
		}

		for _, addr := range addrs {
			ipWhiteListMap[addr] = struct{}{}
		}
	}

	if len(ipWhiteListMap) == 0 {
		return nopMiddleware
	}

	return func(c *gin.Context) {
		clientIP := net2.GetClientIP(c.Request)
		if _, ok := ipWhiteListMap[clientIP.String()]; !ok && !clientIP.IsLoopback() {
			c.Abort()
			c.String(http.StatusForbidden, "request now allowed")
			return
		}
		c.Next()
	}
}

func shutdownWithTimeout(serv *http.Server, timeout time.Duration) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	if err := serv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown the http server err: %w", err)
	}
	return nil
}

type pingInfo struct {
	Content httpapi.Ping `json:"content"`
}

func getGinLog() io.Writer {
	if Cfg.Logging.GinLog == "stdout" {
		return os.Stdout
	}
	rotate := logger.MaxSize
	if Cfg.Logging.Rotate > 0 {
		rotate = Cfg.Logging.Rotate
	}
	rotateBackups := logger.MaxBackups
	if Cfg.Logging.RotateBackups > 0 {
		rotateBackups = Cfg.Logging.RotateBackups
	}
	return &lumberjack.Logger{
		Filename:   Cfg.Logging.GinLog,
		MaxSize:    rotate, // MB
		MaxBackups: rotateBackups,
		MaxAge:     30, // day
	}
}

func getGinErrLogger() io.Writer {
	if Cfg.Logging.GinErrLog == "stderr" {
		return os.Stderr
	}
	rotate := logger.MaxSize
	if Cfg.Logging.Rotate > 0 {
		rotate = Cfg.Logging.Rotate
	}
	rotateBackups := logger.MaxBackups
	if Cfg.Logging.RotateBackups > 0 {
		rotateBackups = Cfg.Logging.RotateBackups
	}

	return &lumberjack.Logger{
		Filename:   Cfg.Logging.GinErrLog,
		MaxSize:    rotate, // MB
		MaxBackups: rotateBackups,
		MaxAge:     30, // day
	}
}
