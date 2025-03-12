// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

var (
	defaultHostInputs = []string{
		"cpu",
		"disk",
		"diskio",
		"mem",
		"swap",
		"system",
		"hostobject",
		"net",
		"host_processes",
		"dk",
	}

	defaultHostInputsForLinux = []string{
		"cpu",
		"disk",
		"diskio",
		"mem",
		"swap",
		"system",
		"hostobject",
		"net",
		"host_processes",
		"container",
		"dk",
	}

	defaultHostInputsForMacOS = []string{
		"cpu",
		"disk",
		"diskio",
		"mem",
		"swap",
		"system",
		"hostobject",
		"net",
		"container",
		"dk",

		// host_processes is costly, maybe we should disable default
		"host_processes",
	}

	DataKitVersion = ""

	OTA = false

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

	EnableElection     = ""
	GlobalElectionTags = ""
	ElectionNamespace  = "default"

	RumOriginIPHeader, RumDisable404Page string

	LogLevel, Log, GinLog string

	PProfListen string

	EnableSinker,
	SinkerGlobalCustomerKeys string

	LimitDisabled int

	LimitCPUMax, LimitCPUCores float64
	LimitCPUMin                float64 // deprecated

	LimitMemMax      int64
	CryptoAESKey     string
	CryptoAESKeyFile string
)

// generate default inputs list.
func mergeDefaultInputs(defaultList, enabledList []string, appendDefault bool) []string {
	if len(enabledList) == 0 {
		return defaultList // no inputs enabled(disabled), enable all default inputs
	}

	l.Infof("enabled input list: %+#v", enabledList)

	res := []string{}
	blackList := map[string]bool{}
	whiteList := map[string]bool{}

	for _, elem := range enabledList {
		if elem == "-" { // disabled all
			for _, x := range defaultList {
				res = append(res, "-"+x) // prefixed '-' to disable the input
			}
			return res
		}

		res = append(res, elem) // may be 'foo' or '-foo'
		if strings.HasPrefix(elem, "-") {
			blackList[elem] = true
			l.Warnf("input %q disabled", elem)
		} else {
			whiteList[elem] = true
		}
	}

	// why people specify both list?
	// we drop white list: only accept black list
	if len(blackList) > 0 && len(whiteList) > 0 {
		whiteList = map[string]bool{}
	}

	//
	// merge default enabled inputs
	//
	if len(blackList) > 0 {
		for _, elem := range defaultList {
			if _, ok := blackList["-"+elem]; !ok { // not disabled, then enable it
				l.Infof("input %q enabled", elem)
				res = append(res, elem)
			}
		}
	}

	if len(whiteList) > 0 {
		for _, elem := range defaultList {
			if appendDefault {
				l.Infof("input %q enabled", elem)
				res = append(res, elem)
			} else {
				// disable them
				if _, ok := whiteList[elem]; !ok { // not enabled, then disable it
					l.Warnf("input %q disabled", elem)
					res = append(res, "-"+elem)
				}
			}
		}
	}

	// compact the list, remove duplicates.
	set := map[string]bool{}
	for _, elem := range res {
		set[elem] = true
	}

	res = []string{}
	for k := range set {
		res = append(res, k)
	}

	// keep sorted
	sort.Strings(res)

	return res
}

func setupDefaultInputs(mc *config.Config, arg string, defaultList []string, upgrade bool) {
	if upgrade {
		if arg == "" {
			if len(mc.DefaultEnabledInputs) == 0 { // all default inputs disabled
				mc.DefaultEnabledInputs = mergeDefaultInputs(defaultList, []string{"-"}, true)
			} else {
				mc.DefaultEnabledInputs = mergeDefaultInputs(defaultList, mc.DefaultEnabledInputs, true)
			}
		} else {
			mc.DefaultEnabledInputs = mergeDefaultInputs(defaultList, strings.Split(arg, ","), false)
		}
	} else {
		if arg == "" {
			mc.DefaultEnabledInputs = mergeDefaultInputs(defaultList, nil, false)
		} else {
			mc.DefaultEnabledInputs = mergeDefaultInputs(defaultList, strings.Split(arg, ","), false)
		}
	}
}

func writeDefInputToMainCfg(mc *config.Config, upgrade bool) {
	hostInputs := defaultHostInputs

	switch runtime.GOOS {
	case datakit.OSLinux:
		hostInputs = defaultHostInputsForLinux
	case datakit.OSDarwin:
		hostInputs = defaultHostInputsForMacOS
	}

	setupDefaultInputs(mc, EnableInputs, hostInputs, upgrade)

	if CloudProvider != "" {
		if err := injectCloudProvider(CloudProvider); err != nil {
			l.Warnf("Failed to inject cloud-provider: %s", err.Error())
		} else {
			l.Infof("Set cloud provider to %s ok", CloudProvider)
		}
	} else {
		l.Infof("Cloud provider not set")
	}
}

func injectCloudProvider(p string) error {
	switch p {
	case "aliyun", "tencent", "aws", "hwcloud", "azure":

		l.Infof("try set cloud provider to %s...", p)

		conf := preEnableHostobjectInput(p)

		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, "host"), datakit.ConfPerm); err != nil {
			l.Fatalf("failed to init hostobject conf: %s", err.Error())
		}

		cfgpath := filepath.Join(datakit.ConfdDir, "host", "hostobject.conf")
		if err := os.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
			return err
		}

	case "": // pass

	default:
		l.Warnf("Unknown cloud provider %q, ignored", p)
	}

	return nil
}

func preEnableHostobjectInput(cloud string) []byte {
	// I don't want to import hostobject input, cause the installer binary bigger
	sample := []byte(`
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

## Ignore mount points by filesystem type. Default ignored following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...`)

	conf := bytes.ReplaceAll(sample,
		[]byte(`# cloud_provider = "aliyun"`),
		[]byte(fmt.Sprintf(`  cloud_provider = "%s"`, cloud)))

	return conf
}

func getDataway() (*dataway.Dataway, error) {
	dw := dataway.NewDefaultDataway()

	if DatawayURLs != "" {
		urls := strings.Split(DatawayURLs, ",")

		if Proxy != "" {
			l.Debugf("set proxy to %s", Proxy)
			dw.HTTPProxy = Proxy
		}

		if err := dw.Init(dataway.WithURLs(urls...)); err != nil {
			return nil, err
		} else {
			tokens := dw.GetTokens()
			if len(tokens) == 0 {
				return nil, fmt.Errorf("dataway token should not be empty")
			}

			if err := dataway.CheckToken(tokens[0]); err != nil {
				return nil, err
			}
			return dw, nil
		}
	} else {
		return nil, fmt.Errorf("dataway is not set")
	}
}

func loadInstallerEnvs(mc *config.Config) *config.Config {
	var err error
	// prepare dataway info and check token format
	if len(DatawayURLs) != 0 {
		mc.Dataway, err = getDataway()
		if err != nil {
			l.Errorf("getDataway failed: %s", err.Error())
			l.Fatal(err)
		}

		l.Infof("Set dataway to %s", DatawayURLs)

		mc.Dataway.GlobalCustomerKeys = dataway.ParseGlobalCustomerKeys(SinkerGlobalCustomerKeys)
		mc.Dataway.EnableSinker = (EnableSinker != "")

		l.Infof("Set dataway global sinker customer keys: %+#v", mc.Dataway.GlobalCustomerKeys)
	}

	if InstrumentationEnabled != "" {
		mc.HTTPAPI.ListenSocket = "/var/run/datakit/datakit.sock"
		mc.APMInject.InstrumentationEnabled = InstrumentationEnabled
	}

	if OTA {
		mc.AutoUpdate = OTA
		l.Info("set auto update(OTA enabled)")
	}

	if HTTPPublicAPIs != "" {
		apis := strings.Split(HTTPPublicAPIs, ",")
		idx := 0
		for _, api := range apis {
			api = strings.TrimSpace(api)
			if api != "" {
				if !strings.HasPrefix(api, "/") {
					api = "/" + api
				}
				apis[idx] = api
				idx++
			}
		}
		mc.HTTPAPI.PublicAPIs = apis[:idx]
		l.Infof("set PublicAPIs to %s", strings.Join(apis[:idx], ","))
	}

	if DCAEnable != "" {
		config.Cfg.DCAConfig.Enable = true
		l.Infof("set dca enabled")
	}

	if DCAWebsocketServer != "" {
		config.Cfg.DCAConfig.WebsocketServer = DCAWebsocketServer
		config.Cfg.DCAConfig.Enable = true // enable dca if websocket server is set
		l.Infof("set dca enabe: %v, websocket server: %s", config.Cfg.DCAConfig.Enable, config.Cfg.DCAConfig.WebsocketServer)
	}

	if PProfListen != "" {
		config.Cfg.PProfListen = PProfListen
		l.Infof("pprof enabled? %v, listen on %s", config.Cfg.EnablePProf, config.Cfg.PProfListen)
	}

	// Only supports linux and windows
	if LimitDisabled != 1 && (runtime.GOOS == datakit.OSLinux || runtime.GOOS == datakit.OSWindows) {
		mc.ResourceLimitOptions.Enable = true

		if LimitCPUMax > 0 {
			mc.ResourceLimitOptions.CPUMax = LimitCPUMax
		}

		if LimitCPUCores > 0 {
			mc.ResourceLimitOptions.CPUCores = LimitCPUCores

			// we passed limit-cpu-cores, so overwrite exist cpu-max config
			mc.ResourceLimitOptions.CPUMax = 0
		}

		if LimitMemMax > 0 {
			l.Infof("resource limit set max memory to %dMB", LimitMemMax)
			mc.ResourceLimitOptions.MemMax = LimitMemMax
		} else {
			l.Infof("resource limit max memory not set")
		}

		l.Infof("resource limit enabled under %s, cpu: %f, cores: %f, mem: %dMB",
			runtime.GOOS,
			mc.ResourceLimitOptions.CPUMax,
			mc.ResourceLimitOptions.CPUCores,
			mc.ResourceLimitOptions.MemMax)
	} else {
		mc.ResourceLimitOptions.Enable = false
		l.Infof("resource limit disabled, OS: %s", runtime.GOOS)
	}

	if HostName != "" {
		mc.Environments["ENV_HOSTNAME"] = HostName
		l.Infof("set ENV_HOSTNAME to %s", HostName)
	}

	if WALWorkers != 0 {
		mc.Dataway.WAL.Workers = WALWorkers
		l.Infof("set WAL workers to %d", mc.Dataway.WAL.Workers)
	}

	if WALCapacity != mc.Dataway.WAL.MaxCapacityGB {
		mc.Dataway.WAL.MaxCapacityGB = WALCapacity
		l.Infof("set WAL cap to %f GB", mc.Dataway.WAL.MaxCapacityGB)
	}

	if GlobalHostTags != "" {
		mc.GlobalHostTags = config.ParseGlobalTags(GlobalHostTags)
		l.Infof("set global host tags to %+#v", mc.GlobalHostTags)
	}

	if GlobalElectionTags != "" {
		mc.Election.Tags = config.ParseGlobalTags(GlobalElectionTags)
		l.Infof("set global election tags to %+#v", mc.Election.Tags)
	}

	if EnableElection != "" {
		mc.Election.Enable = true
		l.Infof("election enabled? %v", true)
	}

	if ElectionNamespace != "" {
		mc.Election.Namespace = ElectionNamespace
		l.Infof("set election namespace to %s", mc.Election.Namespace)
	}

	if HTTPListen != "" || HTTPPort != 0 {
		taddr, err := net.ResolveTCPAddr("tcp", mc.HTTPAPI.Listen)
		if err != nil {
			l.Warnf("invalid lagacy HTTP listen %q", mc.HTTPAPI.Listen)
		} else {
			if HTTPPort == 0 && taddr.Port != 0 { // use lagacy port
				HTTPPort = taddr.Port
			}

			if HTTPListen == "" && taddr.IP.String() != "" {
				HTTPListen = taddr.IP.String() // use lagacy ip
			}
		}

		mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", HTTPListen, HTTPPort)
		l.Infof("set HTTP listen to %s", mc.HTTPAPI.Listen)
	}

	mc.InstallVer = DataKitVersion
	l.Infof("install version %s", mc.InstallVer)

	if DatakitName != "" {
		mc.Name = DatakitName
		l.Infof("set datakit name to %s", mc.Name)
	}

	if CryptoAESKey != "" || CryptoAESKeyFile != "" {
		if mc.Crypto != nil {
			mc.Crypto.AESKey = CryptoAESKey
			mc.Crypto.AESKeyFile = CryptoAESKeyFile
			l.Infof("set datakit crypto key=%s or crypto key file=%s", mc.Crypto.AESKey, mc.Crypto.AESKeyFile)
		}
	}

	// 一个 func 最大 230 行
	addConfdConfig(mc)

	if GitURL != "" {
		mc.GitRepos = &config.GitRepost{
			PullInterval: GitPullInterval,
			Repos: []*config.GitRepository{
				{
					Enable:                true,
					URL:                   GitURL,
					SSHPrivateKeyPath:     GitKeyPath,
					SSHPrivateKeyPassword: GitKeyPW,
					Branch:                GitBranch,
				}, // GitRepository
			}, // Repos
		} // GitRepost
	}

	if RumDisable404Page != "" {
		l.Infof("set disable 404 page: %v", RumDisable404Page)
		mc.HTTPAPI.Disable404Page = true
	}

	if RumOriginIPHeader != "" {
		l.Infof("set rum origin IP header: %s", RumOriginIPHeader)
		mc.HTTPAPI.RUMOriginIPHeader = RumOriginIPHeader
	}

	if LogLevel != "" {
		mc.Logging.Level = LogLevel
		l.Infof("set log level to %s", LogLevel)
	}

	if Log != "" {
		mc.Logging.Log = Log
		l.Infof("set log to %s", Log)
	}

	if GinLog != "" {
		l.Infof("set gin log to %s", GinLog)
		mc.Logging.GinLog = GinLog
	}

	if javaHome := getJavaHome(); javaHome != "" {
		if mc.RemoteJob == nil {
			mc.RemoteJob = &io.RemoteJob{}
		}
		mc.RemoteJob.JavaHome = javaHome
	}
	return mc
}
