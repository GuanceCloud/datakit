// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

// DefaultInstallArgs get empty installer instance.
func DefaultInstallArgs() *InstallerArgs {
	return &InstallerArgs{}
}

type InstallerArgs struct {
	DataKitBaseURL,
	DataKitVersion string

	BrandURL string

	OTA bool

	InstallExternals string

	EnableInputs,
	CloudProvider,
	Proxy,
	DatawayURLs string

	HTTPPublicAPIs string

	// Deprecated.
	HTTPDisabledAPIs string

	InstallRUMSymbolTools int

	DCAEnable,
	DCAWebsocketServer string

	HTTPPort int
	HTTPSocket,
	HTTPListen,
	DatakitName,
	GlobalHostTags,
	HostName,
	IPDBType string

	InstrumentationEnabled string

	// WAL options.
	WALWorkers  int
	WALCapacity float64

	ConfdBackend,
	ConfdBasicAuth,
	ConfdClientCaKeys,
	ConfdClientCert,
	ConfdClientKey,
	ConfdBackendNodes,
	ConfdPassword,
	ConfdScheme,
	ConfdSeparator,
	ConfdUsername,
	ConfdAccessKey,
	ConfdSecretKey,
	ConfdConfdNamespace,
	ConfdPipelineNamespace,
	ConfdRegion string
	ConfdCircleInterval int

	GitURL,
	GitKeyPath,
	GitKeyPW,
	GitBranch,
	GitPullInterval string

	EnableElection,
	GlobalElectionTags string
	ElectionNamespace string // = "default"

	RumOriginIPHeader, RumDisable404Page string

	LogLevel, Log, GinLog string

	PProfListen string

	EnableSinker,
	SinkerGlobalCustomerKeys string

	LimitDisabled int

	LimitCPUMax, LimitCPUCores float64
	LimitCPUMin                float64 // deprecated
	shouldReinstallService     bool

	LimitMemMax      int64
	CryptoAESKey     string
	CryptoAESKeyFile string

	FlagDKUpgrade,
	FlagOffline,
	FlagDownloadOnly,
	FlagInfo bool

	FlagUserName,
	FlagInstallLog,
	FlagLite,
	FlagELinker,
	FlagSrc string

	IsELinker,
	IsLite bool

	FlagUpgraderIPWhiteList,
	FlagUpgraderListen string
	FlagUpgraderEnabled,
	FlagInstallOnly int

	DistBaseURL,
	DistDataURL,
	DistDatakitLiteURL,
	DistDkUpgraderURL,
	DistDatakitELinkerURL,
	DistDatakitAPMInjectURL,
	DistDatakitAPMInjJavaLibURL,
	DistDatakitURL string
}

func (args *InstallerArgs) UpdateDownloadURLs() error {
	var (
		prefix   = "https://"
		baseURL  = args.DataKitBaseURL
		brandURL = args.BrandURL
		err      error
	)

	if args.DistBaseURL != "" {
		baseURL = args.DistBaseURL
		if strings.HasPrefix(args.DistBaseURL, "http") {
			prefix = "" // clear prefix
		}
	}

	if _, err = url.Parse(baseURL); err != nil {
		return fmt.Errorf("%q can not parse to URL: %w", baseURL, err)
	}

	if args.DistDataURL, err = url.JoinPath(prefix+baseURL, "data.tar.gz"); err != nil {
		return err
	}

	if args.DistDkUpgraderURL, err = url.JoinPath(prefix+baseURL,
		fmt.Sprintf("%s-%s-%s.tar.gz",
			upgrader.BuildBinName,
			runtime.GOOS,
			runtime.GOARCH)); err != nil {
		return err
	}

	if args.DistDatakitURL, err = url.JoinPath(prefix+baseURL,
		fmt.Sprintf("datakit-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			args.DataKitVersion)); err != nil {
		return err
	}

	if args.DistDatakitLiteURL, err = url.JoinPath(prefix+baseURL,
		fmt.Sprintf("datakit_lite-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			args.DataKitVersion)); err != nil {
		return err
	}

	if args.DistDatakitELinkerURL, err = url.JoinPath(prefix+baseURL,
		fmt.Sprintf("datakit_elinker-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			args.DataKitVersion)); err != nil {
		return err
	}

	if args.DistDatakitAPMInjectURL, err = url.JoinPath(prefix+baseURL,
		fmt.Sprintf("datakit-apm-inject-%s-%s-%s.tar.gz",
			runtime.GOOS,
			runtime.GOARCH,
			args.DataKitVersion)); err != nil {
		return err
	}

	if args.DistDatakitAPMInjJavaLibURL, err = url.JoinPath(prefix+brandURL,
		"dd-image/dd-java-agent.jar"); err != nil {
		return err
	}

	// 如果命令行传入 -installer_base_url，则表明是离线安装，此刻离线的 java ddtrace 库地址
	// 是用户自己本地下载的，根据已有[文档](https://docs.guance.com/datakit/datakit-offline-install/#offline-advanced)，
	// 这个目录是 *apm_lib*，故此处调整一下 lib 的路径。
	if args.DistBaseURL != "" {
		if args.DistDatakitAPMInjJavaLibURL, err = url.JoinPath(prefix+args.DistBaseURL, "apm_lib/dd-java-agent.jar"); err != nil {
			return err
		}
	}

	return nil
}

var (
	liteReg    = regexp.MustCompile(`Build Tag:\s*lite`)
	eLinkerReg = regexp.MustCompile(`Build Tag:\s*elinker`)
)

func (args *InstallerArgs) SetDatakitLiteAndELinker() {
	switch {
	case len(args.FlagLite) > 0:
		v, err := strconv.ParseBool(args.FlagLite)
		if err != nil {
			l.Warnf("parse flag 'lite' error: %s", err.Error())
		} else {
			args.IsLite = v
		}
	case len(args.FlagELinker) > 0:
		if v, err := strconv.ParseBool(args.FlagELinker); err != nil {
			l.Warnf("parse flag 'elinker' error: %s", err.Error())
		} else {
			args.IsELinker = v
		}
	case args.FlagDKUpgrade: // only for upgrading datakit
		cmd := exec.Command(datakit.DatakitBinaryPath(), "version") //nolint:gosec
		res, err := cmd.CombinedOutput()
		if err != nil {
			l.Warnf("check version failed: %s", err.Error())
		} else {
			args.IsLite = liteReg.Match(res)
			args.IsELinker = eLinkerReg.Match(res)
		}
	}
}

func (args *InstallerArgs) setupServiceOptions() *service.Config {
	var (
		def          = config.DefaultConfig()
		rl           = def.ResourceLimitOptions // use cpu/mem limit from default configure
		limitUpdated = false
	)

	// setup CPU limit
	if args.LimitCPUMax > 0.0 {
		rl.CPUCores = resourcelimit.CPUMaxToCores(args.LimitCPUMax)
		limitUpdated = true
	}

	if args.LimitCPUCores > 0.0 { // cpu-cores override above cpu-max
		rl.CPUCores = args.LimitCPUCores
		limitUpdated = true
	}

	// setup mem limit
	if args.LimitMemMax > 0 {
		rl.MemMax = args.LimitMemMax
		limitUpdated = true
	}

	rl.Setup() // apply

	svcopts := []dkservice.ServiceOption{
		dkservice.WithMemLimit(fmt.Sprintf("%dM", rl.MemMax)),
		dkservice.WithCPULimit(fmt.Sprintf("%f%%", rl.CPUMax())),
	}

	if runtime.GOOS == datakit.OSLinux && args.FlagUserName != "" {
		svcopts = append(svcopts, dkservice.WithUser(args.FlagUserName))

		if !datakit.IsAdminUser(args.FlagUserName) && limitUpdated && args.FlagDKUpgrade {
			// for non-admin user, during upgrade, if user specified new cpu/mem limit, we should
			// apply them to datakit.service.
			args.shouldReinstallService = true
		}
	}

	return dkservice.ApplyOptions(dkservice.DefaultServiceConfig(), svcopts...)
}

// SetupService detect if datakit is installed, and try to stop service before install/upgrade.
func (args *InstallerArgs) SetupService() (service.Service, error) {
	svc, err := dkservice.NewServiceOnConfigure(args.setupServiceOptions())
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return nil, err
	}

	svcStatus, err := svc.Status()
	if err != nil {
		if errors.Is(err, service.ErrNotInstalled) {
			l.Infof("datakit service not installed before")
		} else {
			l.Warnf("svc.Status: %s, ignored", err.Error())
		}
	} else {
		switch svcStatus {
		case service.StatusUnknown: // not installed
			cp.Infof("DataKit service maybe not installed\n")
		case service.StatusStopped: // pass
			cp.Infof("DataKit service stopped\n")
		case service.StatusRunning:
			cp.Infof("Stopping running DataKit...\n")
			if err = service.Control(svc, "stop"); err != nil {
				cp.Warnf("stop service failed %s, ignored\n", err.Error())
			}
		}
	}

	return svc, nil
}

func (args *InstallerArgs) SetupUpgraderService() {
	var wlist []string
	if len(args.FlagUpgraderIPWhiteList) > 0 {
		wlist = strings.Split(strings.TrimSpace(args.FlagUpgraderIPWhiteList), ",")
	}

	// Apply options from exist datakit.conf.
	// During upgrade, we still able to re-install dk-upgrader service, at the
	// time, we should reuse datakit's exist configures(such as datakit HTTP API host),
	// not read these configures from installer args.
	mc := config.Cfg

	opts := []upgrader.UpgraderOpt{
		upgrader.WithDKUpgrade(args.FlagDKUpgrade),
		upgrader.WithUpgradeService(args.FlagUpgraderEnabled != 0),
		upgrader.WithInstallOnly(args.FlagInstallOnly != 0),
		upgrader.WithListen(args.FlagUpgraderListen),
		upgrader.WithIPWhiteList(wlist),
		upgrader.WithInstallBaseURL(args.DistBaseURL),
		upgrader.WithDatakitAPIListen(mc.HTTPAPI.Listen),
		upgrader.WithProxy(mc.Dataway.HTTPProxy),
		upgrader.WithDatakitAPIHTTPS(mc.HTTPAPI.HTTPSEnabled()),
	}

	if runtime.GOOS == datakit.OSLinux {
		opts = append(opts, upgrader.WithInstallUserName(mc.DatakitUser))
	}

	if err := upgrader.InstallUpgradeService(opts...); err != nil {
		cp.Warnf("upgrader service install/start failed: %s, ignored", err.Error())
	}
}

func (args *InstallerArgs) SetupUserGroup(mc *config.Config) (err error) {
	// NOTE: make group name same as user name
	username, groupname := args.FlagUserName, args.FlagUserName

	if len(username) == 0 || username == "root" || runtime.GOOS != datakit.OSLinux {
		l.Infof("skip user/group settings")
		return nil
	}

	// check if group and username exist.
	if _, err = user.LookupGroup(groupname); err != nil {
		l.Errorf("Group %s not existed! Please create it first.", groupname)
		return fmt.Errorf("user.LookupGroup(%q): %w", groupname, err)
	}

	if _, err = user.Lookup(args.FlagUserName); err != nil {
		l.Errorf("User %s not existed! Please create it first.", args.FlagUserName)
		return fmt.Errorf("user.Lookup(%q): %w", args.FlagUserName, err)
	}

	l.Infof("datakit service run as user: %q", args.FlagUserName)

	// NOTE: following works only running under linux...
	logDir := filepath.Dir(mc.Logging.Log)
	l.Infof("logDir = %s", logDir)

	// make dirs.
	if err = mkdir(datakit.InstallDir, os.ModePerm); err != nil {
		l.Errorf("make installDir failed: %v", err)
		return err
	}

	if err = mkdir(logDir, os.ModePerm); err != nil {
		l.Errorf("make defaultLogDir failed: %v", err)
		return err
	}

	// chown.
	if err = executeCmd("chown", "-R",
		fmt.Sprintf("%s:%s", username, groupname),
		datakit.InstallDir, logDir); err != nil {
		l.Errorf("chown failed: %v", err)
		return err
	}

	// chmod.
	if err = executeCmd("chmod", "-R", "755", datakit.InstallDir, logDir); err != nil {
		l.Errorf("chmod failed: %v", err)
		return err
	}

	// chown.
	if err = executeCmd("chown", "-R",
		fmt.Sprintf("%s:%s", username, groupname),
		upgrader.InstallDir, upgrader.DefaultLogDir); err != nil {
		l.Errorf("chown failed: %v", err)
		return err
	}

	// chmod.
	if err = executeCmd("chmod", "-R", "755", upgrader.InstallDir, upgrader.DefaultLogDir); err != nil {
		l.Errorf("chmod failed: %v", err)
		return err
	}

	return nil
}

func mkdir(path string, perm os.FileMode) error {
	l.Infof("MkdirAll %s => %s", path, perm.String())
	return os.MkdirAll(path, perm)
}

func executeCmd(name string, arg ...string) error {
	l.Infof("executing %s %v", name, arg)
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}
