// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"go.uber.org/atomic"
	"gopkg.in/natefinch/lumberjack.v2"
	net2 "k8s.io/apimachinery/pkg/util/net"
)

const (
	ExitStatusUnableToRun     = 101
	HTTPServerExitNotExpected = 107
	ExitStatusAlreadyRunning  = 120
)

const (
	statusNoUpgrade = 0
	statusUpgrading = 1
)

var (
	GL = func() **logger.Logger {
		l := logger.DefaultSLogger(ServiceName)
		return &l
	}()
	PidFile       = filepath.Join(InstallDir, ServiceName+".pid")
	UpgradeStatus = atomic.NewInt32(0)
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
	InstallDir         = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	DefaultLogDir      = filepath.Join("/var/log", ServiceName)
	MainConfigFile     = filepath.Join(InstallDir, "main.conf")
	defaultServiceOpts = map[string]interface{}{
		"RestartSec":         10, // 重启间隔.
		"StartLimitInterval": 60, // 60秒内5次重启之后便不再启动.
		"StartLimitBurst":    5,
	}
	DefaultProgram = NewProgram()
)

func L() *logger.Logger {
	return *GL
}

type Runnable func(*Program)

type Program struct {
	Run          Runnable
	NotifyToStop chan struct{}
	StopWellDone chan struct{}
}

func NewProgram() *Program {
	return &Program{
		Run:          DoProgramRun,
		NotifyToStop: make(chan struct{}),
		StopWellDone: make(chan struct{}),
	}
}

func CreateDirs() error {
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

func (p *Program) Start(s service.Service) error {
	if p.Run == nil {
		return fmt.Errorf("entry not set")
	}

	p.Run(p)
	return nil
}

func (p *Program) Stop(s service.Service) error {
	close(p.NotifyToStop)

	// We must wait here:
	// On Windows, we stop datakit in services.msc, if datakit process do not
	// echo to here, services.msc will complain the datakit process has been
	// exit unexpected
	<-p.StopWellDone
	return nil
}

func NewDefaultService(username string, args []string) (service.Service, error) {
	return NewService(DefaultProgram, username, args)
}

func NewService(program service.Interface, username string, args []string) (service.Service, error) {
	if program == nil {
		program = DefaultProgram
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

func RunService(serv service.Service) error {
	errch := make(chan error, 32) //nolint:gomnd
	sLogger, err := serv.Logger(errch)
	if err != nil {
		return fmt.Errorf("unable to get service logger: %w", err)
	}

	if err := sLogger.Infof("%s set service logger ok, starting...", ServiceName); err != nil {
		return err
	}

	if err := serv.Run(); err != nil {
		if serr := sLogger.Errorf("start service failed: %s", err.Error()); serr != nil {
			return serr
		}
		return err
	}

	if err := sLogger.Infof("%s service exited", ServiceName); err != nil {
		return err
	}

	return nil
}

func DoProgramRun(p *Program) {
	gin.DefaultErrorWriter = getGinErrLogger()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: uhttp.GinLogFormatter,
		Output:    getGinLog(),
	}))
	router.Use(gin.Recovery())

	if len(Cfg.IPWhiteList) > 0 {
		router.Use(getIPVerifyMiddleware())
	}

	v1 := router.Group("/v1")
	{
		v1.GET("/datakit/version", dkVersion)
		v1.POST("/datakit/upgrade", upgrade)
	}

	serv := &http.Server{
		Addr:    Cfg.Listen,
		Handler: router,
	}

	httpServClosed := make(chan error, 4)
	go func() {
		if err := serv.ListenAndServe(); err != nil {
			L().Infof("datakit manager server return: %s", err)
			httpServClosed <- err
		}
	}()

	go func() {
		select {
		case <-p.NotifyToStop:
			if err := shutdownWithTimeout(serv, time.Second*5); err != nil {
				L().Warnf("datakit upgrade service shutdown err: %s", err)
			}
			close(p.StopWellDone)

		case err := <-httpServClosed:
			L().Errorf("http server exit abnormal: %s", err)
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
			L().Warnf("the IP [%s] in ip_whitelist setting is illegal: %s", ip, err)
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

func fetchCurrentDKVersion() ([]byte, error) {
	resp, err := http.Get("http://127.0.0.1:9529/v1/ping")
	if err != nil {
		return nil, fmt.Errorf("unable to query current Datakit version: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read Datakit ping result")
	}

	return body, nil
}

func dkVersion(ctx *gin.Context) {
	output, err := fetchCurrentDKVersion()
	if err != nil {
		errorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Data(http.StatusOK, "application/json", output)
}

func errorResponse(ctx *gin.Context, code int, err error) {
	ctx.JSON(code, map[string]interface{}{
		"error": err.Error(),
	})
}

func successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

type PingInfo struct {
	Content httpapi.Ping `json:"content"`
}

func upgrade(ctx *gin.Context) {
	L().Infof("receive request: %s", ctx.Request.URL.String())

	if !UpgradeStatus.CompareAndSwap(statusNoUpgrade, statusUpgrading) {
		errorResponse(ctx, http.StatusNotAcceptable, fmt.Errorf("upgrade is on going, please try later"))
		return
	}
	defer UpgradeStatus.Store(statusNoUpgrade)

	output, err := fetchCurrentDKVersion()
	downloadURL := ""

	var DKPingVer PingInfo
	if err == nil {
		if err := json.Unmarshal(output, &DKPingVer); err == nil {
			L().Infof("VersionString: %s, Commit: %s", DKPingVer.Content.Version, DKPingVer.Content.Commit)
		} else {
			L().Warnf("unable to unmarshal json: %s", err)
		}
	} else {
		L().Warnf("unable to check version info from command line: %s", err)
	}

	if Cfg.InstallerBaseURL != "" {
		cmds.OnlineBaseURL = Cfg.InstallerBaseURL
	}

	versions, err := cmds.GetOnlineVersions()
	if err != nil {
		errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("unable to find newer Datakit version: %w", err))
		return
	}

	upToDate := false
	for _, v := range versions {
		L().Infof("VersionString: %s, Commit: %s, ReleaseDate: %s", v.VersionString, v.Commit, v.ReleaseDate)
		if v.DownloadURL != "" {
			// only compare release commit hash
			if v.Commit != DKPingVer.Content.Commit {
				downloadURL = v.DownloadURL
				break
			} else {
				upToDate = true
			}
		}
	}
	if downloadURL == "" {
		if upToDate {
			errorResponse(ctx, http.StatusNotModified, fmt.Errorf("already up to date"))
		} else {
			errorResponse(ctx, http.StatusNotModified, fmt.Errorf("unable to find newer Datakit version"))
		}
		return
	}

	scriptFile, err := saveUpgradeScript(downloadURL)
	if err != nil {
		errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("unable to download upgrade script: %w", err))
		return
	}

	if err := execUpdateCmd(scriptFile); err != nil {
		errorResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	successResponse(ctx, map[string]string{"msg": "success"})
}

func saveUpgradeScript(downloadURL string) (string, error) {
	downloadURL = strings.TrimRight(downloadURL, "/ ")
	scriptName := "datakit-upgrade.sh"
	if runtime.GOOS == datakit.OSWindows {
		downloadURL = fmt.Sprintf("%s/install.ps1", downloadURL)
		scriptName = "datakit-upgrade.ps1"
	} else {
		downloadURL = fmt.Sprintf("%s/install.sh", downloadURL)
	}

	scriptFile := filepath.Join(datakit.InstallDir, scriptName)

	f, err := os.OpenFile(scriptFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("unable to open script file [%s]: %w", scriptFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f) //nolint:errcheck

	code := ""
	if code != "" {
		if _, err := f.WriteString(code); err != nil {
			return "", fmt.Errorf("unable to write script file[%s]: %w", scriptFile, err)
		}
	}

	resp, err := http.Get(downloadURL) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("unable to download script file [%s]: %w", downloadURL, err)
	}

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("request datakit upgrade script [%s] response error status [%s]", downloadURL, resp.Status)
	}
	defer resp.Body.Close() // nolint:errcheck

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("unable to save datakit upgrade script[%s] to local file[%s]: %w", downloadURL, scriptFile, err)
	}
	return scriptFile, nil
}

func execUpdateCmd(scriptFile string) error {
	shell := "bash"
	args := []string{scriptFile}
	if runtime.GOOS == datakit.OSWindows {
		shell = "powershell"
		// Powershell can not invoke a script at a path with blanks
		// see: https://stackoverflow.com/questions/18537098/spaces-cause-split-in-path-with-powershell
		args = []string{
			"-c",
			fmt.Sprintf(`Set-ExecutionPolicy Bypass -scope Process -Force;& "%s"`, scriptFile),
		}
	}

	shellBin, err := exec.LookPath(shell)
	if err != nil {
		return fmt.Errorf("%s command not found: %w", shell, err)
	}

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	cmd := exec.Command(shellBin, args...) // nolint:gosec
	cmd.Stderr = stderr
	cmd.Stdout = stdout

	envs := os.Environ()
	envs = append(envs, "DK_UPGRADE=1")

	proxy := config.Cfg.Dataway.HTTPProxy
	if proxy != "" {
		envs = append(envs, "HTTPS_PROXY="+proxy)
	}

	if Cfg.InstallerBaseURL != "" {
		envs = append(envs, fmt.Sprintf("DK_INSTALLER_BASE_URL=%s", cmds.CanonicalInstallBaseURL(Cfg.InstallerBaseURL)))
	}

	cmd.Env = envs

	L().Infof("run upgrade script envs: %s", strings.Join(cmd.Env, " \t "))

	L().Infof("datakit manager will start execute upgrade cmd: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to execute upgrade cmd[%s]: %w", cmd.String(), err)
	}

	err = cmd.Wait()
	L().Infof("upgrade process stdout: %s", stdout.String())
	L().Infof("upgrade process stderr: %s", stderr.String())
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return fmt.Errorf("upgrade process exit abnormal: %s, err: %w, stdout:%s, stderr: %s",
				ee.ProcessState.String(), ee, stdout.String(), stderr.String())
		}
		return fmt.Errorf("upgrade process execute fail: %w, stdout:%s, stderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
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
