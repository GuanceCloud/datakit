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
		"statsd",
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
		"statsd",
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
		"statsd",

		// host_processes is costly, maybe we should disable default
		"host_processes",
	}
)

// generate default inputs list.
func (args *InstallerArgs) mergeDefaultInputs(defaultList, enabledList []string, appendDefault bool) []string {
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

func (args *InstallerArgs) setupDefaultInputs(mc *config.Config, defaultList []string) {
	if args.FlagDKUpgrade {
		if args.EnableInputs == "" {
			if len(mc.DefaultEnabledInputs) == 0 { // all default inputs disabled
				mc.DefaultEnabledInputs = args.mergeDefaultInputs(defaultList, []string{"-"}, true)
			} else {
				mc.DefaultEnabledInputs = args.mergeDefaultInputs(defaultList, mc.DefaultEnabledInputs, true)
			}
		} else {
			mc.DefaultEnabledInputs = args.mergeDefaultInputs(defaultList, strings.Split(args.EnableInputs, ","), false)
		}
	} else {
		if args.EnableInputs == "" {
			mc.DefaultEnabledInputs = args.mergeDefaultInputs(defaultList, nil, false)
		} else {
			mc.DefaultEnabledInputs = args.mergeDefaultInputs(defaultList, strings.Split(args.EnableInputs, ","), false)
		}
	}
}

// WriteDefInputs inject default inputs into datakit.con.
func (args *InstallerArgs) WriteDefInputs(mc *config.Config) error {
	hostInputs := defaultHostInputs

	switch runtime.GOOS {
	case datakit.OSLinux:
		hostInputs = defaultHostInputsForLinux
	case datakit.OSDarwin:
		hostInputs = defaultHostInputsForMacOS
	}

	args.setupDefaultInputs(mc, hostInputs)

	if args.CloudProvider != "" {
		if err := args.injectCloudProvider(); err != nil {
			l.Warnf("Failed to inject cloud-provider: %s", err.Error())
			return err
		} else {
			l.Infof("Set cloud provider to %s ok", args.CloudProvider)
		}
	} else {
		l.Infof("Cloud provider not set")
	}

	return nil
}

func (args *InstallerArgs) injectCloudProvider() error {
	switch args.CloudProvider {
	case "aliyun", "tencent", "aws", "hwcloud", "azure":

		l.Infof("try set cloud provider to %s...", args.CloudProvider)

		conf := args.preEnableHostobjectInput()

		if err := os.MkdirAll(filepath.Join(datakit.ConfdDir, "host"), datakit.ConfPerm); err != nil {
			l.Errorf("failed to init hostobject conf: %s", err.Error())
			return err
		}

		cfgpath := filepath.Join(datakit.ConfdDir, "host", "hostobject.conf")
		if err := os.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
			l.Errorf("WriteFile: %s", err.Error())
			return err
		}

	case "": // pass

	default:
		l.Warnf("Unknown cloud provider %q, ignored", args.CloudProvider)
	}

	return nil
}

func (args *InstallerArgs) preEnableHostobjectInput() []byte {
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
		[]byte(fmt.Sprintf(`  cloud_provider = "%s"`, args.CloudProvider)))

	return conf
}

func (args *InstallerArgs) getDataway() (*dataway.Dataway, error) {
	dw := dataway.NewDefaultDataway()

	if args.DatawayURLs != "" {
		urls := strings.Split(args.DatawayURLs, ",")

		if args.Proxy != "" {
			l.Debugf("set proxy to %s", args.Proxy)
			dw.HTTPProxy = args.Proxy
		}

		if err := dw.Init(dataway.WithURLs(urls...)); err != nil {
			return nil, err
		} else {
			tokens := dw.GetTokens()
			if len(tokens) == 0 {
				return nil, dataway.ErrEmptyToken
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

// LoadInstallerArgs apply args settings to mc.
func (args *InstallerArgs) LoadInstallerArgs(mc *config.Config) (*config.Config, error) {
	var err error

	if args.FlagUserName != "" {
		mc.DatakitUser = args.FlagUserName
	}

	// setup dataway and check token format
	if len(args.DatawayURLs) != 0 {
		mc.Dataway, err = args.getDataway()
		if err != nil {
			return mc, fmt.Errorf("getDataway: %w", err)
		}

		l.Infof("Set dataway to %s", args.DatawayURLs)

		mc.Dataway.GlobalCustomerKeys = dataway.ParseGlobalCustomerKeys(args.SinkerGlobalCustomerKeys)
		mc.Dataway.EnableSinker = (args.EnableSinker != "")

		l.Infof("Set dataway global sinker customer keys: %+#v", mc.Dataway.GlobalCustomerKeys)
	}

	if args.InstrumentationEnabled != "" {
		mc.HTTPAPI.ListenSocket = "/var/run/datakit/datakit.sock"
		mc.APMInject.InstrumentationEnabled = args.InstrumentationEnabled
	}

	if args.OTA {
		mc.AutoUpdate = args.OTA
		l.Info("set auto update(OTA enabled)")
	}

	if args.HTTPPublicAPIs != "" {
		apis := strings.Split(args.HTTPPublicAPIs, ",")
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

	if args.DCAEnable != "" {
		config.Cfg.DCAConfig.Enable = true
		l.Infof("set dca enabled")
	}

	if args.DCAWebsocketServer != "" {
		config.Cfg.DCAConfig.WebsocketServer = args.DCAWebsocketServer
		config.Cfg.DCAConfig.Enable = true // enable dca if websocket server is set
		l.Infof("set dca enabe: %v, websocket server: %s", config.Cfg.DCAConfig.Enable, config.Cfg.DCAConfig.WebsocketServer)
	}

	if args.PProfListen != "" {
		config.Cfg.PProfListen = args.PProfListen
		l.Infof("pprof enabled? %v, listen on %s", config.Cfg.EnablePProf, config.Cfg.PProfListen)
	}

	if args.LimitDisabled != 1 {
		if mc.ResourceLimitOptions.Enable { // resource-limit not disabled before upgrade/install
			if args.LimitCPUMax > 0 {
				mc.ResourceLimitOptions.CPUMax = args.LimitCPUMax
			}

			// apply args to datakit.conf or from datakit.conf to args
			if args.LimitCPUCores > 0 {
				mc.ResourceLimitOptions.CPUCores = args.LimitCPUCores

				// we passed limit-cpu-cores, so reset cpu-max config and deprecated it
				mc.ResourceLimitOptions.CPUMax = 0
			}

			if args.LimitMemMax > 0 {
				mc.ResourceLimitOptions.MemMax = args.LimitMemMax
			}

			mc.ResourceLimitOptions.Setup()

			l.Infof("resource limit enabled under %s, cpu: %f, cores: %f, mem: %dMB",
				runtime.GOOS,
				mc.ResourceLimitOptions.CPUMax,
				mc.ResourceLimitOptions.CPUCores,
				mc.ResourceLimitOptions.MemMax)
		}
	} else {
		mc.ResourceLimitOptions.Enable = false
		l.Infof("resource limit disabled, OS: %s", runtime.GOOS)
	}

	if args.HostName != "" {
		mc.Environments["ENV_HOSTNAME"] = args.HostName
		l.Infof("set ENV_HOSTNAME to %s", args.HostName)
	}

	if args.WALWorkers != 0 {
		mc.Dataway.WAL.Workers = args.WALWorkers
		l.Infof("set WAL workers to %d", mc.Dataway.WAL.Workers)
	}

	if args.WALCapacity > 0 && args.WALCapacity != mc.Dataway.WAL.MaxCapacityGB {
		mc.Dataway.WAL.MaxCapacityGB = args.WALCapacity
		l.Infof("set WAL cap to %f GB", mc.Dataway.WAL.MaxCapacityGB)
	}

	if args.GlobalHostTags != "" {
		mc.GlobalHostTags = config.ParseGlobalTags(args.GlobalHostTags)
		l.Infof("set global host tags to %+#v", mc.GlobalHostTags)
	}

	if args.GlobalElectionTags != "" {
		mc.Election.Tags = config.ParseGlobalTags(args.GlobalElectionTags)
		l.Infof("set global election tags to %+#v", mc.Election.Tags)
	}

	if args.EnableElection != "" {
		mc.Election.Enable = true
		l.Infof("election enabled? %v", true)
	}

	if args.ElectionNamespace != "" {
		mc.Election.Namespace = args.ElectionNamespace
		l.Infof("set election namespace to %s", mc.Election.Namespace)
	}

	if args.HTTPListen != "" || args.HTTPPort != 0 {
		taddr, err := net.ResolveTCPAddr("tcp", mc.HTTPAPI.Listen)
		if err != nil {
			l.Warnf("invalid lagacy HTTP listen %q", mc.HTTPAPI.Listen)
		} else {
			if args.HTTPPort == 0 && taddr.Port != 0 { // use lagacy port
				args.HTTPPort = taddr.Port
			}

			if args.HTTPListen == "" && taddr.IP.String() != "" {
				args.HTTPListen = taddr.IP.String() // use lagacy ip
			}
		}

		mc.HTTPAPI.Listen = fmt.Sprintf("%s:%d", args.HTTPListen, args.HTTPPort)
		l.Infof("set HTTP listen to %s", mc.HTTPAPI.Listen)
	}

	if args.HTTPSocket != "" {
		mc.HTTPAPI.ListenSocket = args.HTTPSocket
		l.Infof("set HTTP socket to %q", mc.HTTPAPI.ListenSocket)
	}

	mc.InstallVer = args.DataKitVersion
	l.Infof("install version %s", mc.InstallVer)

	if args.DatakitName != "" {
		mc.Name = args.DatakitName
		l.Infof("set datakit name to %s", mc.Name)
	}

	if args.CryptoAESKey != "" || args.CryptoAESKeyFile != "" {
		if mc.Crypto != nil {
			mc.Crypto.AESKey = args.CryptoAESKey
			mc.Crypto.AESKeyFile = args.CryptoAESKeyFile
			l.Infof("set datakit crypto key=%s or crypto key file=%s", mc.Crypto.AESKey, mc.Crypto.AESKeyFile)
		}
	}

	args.addConfdConfig(mc)

	if args.GitURL != "" {
		mc.GitRepos = &config.GitRepost{
			PullInterval: args.GitPullInterval,
			Repos: []*config.GitRepository{
				{
					Enable:                true,
					URL:                   args.GitURL,
					SSHPrivateKeyPath:     args.GitKeyPath,
					SSHPrivateKeyPassword: args.GitKeyPW,
					Branch:                args.GitBranch,
				}, // GitRepository
			}, // Repos
		} // GitRepost
	}

	if args.RumDisable404Page != "" {
		l.Infof("set disable 404 page: %v", args.RumDisable404Page)
		mc.HTTPAPI.Disable404Page = true
	}

	if args.RumOriginIPHeader != "" {
		l.Infof("set rum origin IP header: %s", args.RumOriginIPHeader)
		mc.HTTPAPI.RUMOriginIPHeader = args.RumOriginIPHeader
	}

	if args.LogLevel != "" {
		mc.Logging.Level = args.LogLevel
		l.Infof("set log level to %s", args.LogLevel)
	}

	if args.Log != "" {
		mc.Logging.Log = args.Log
		l.Infof("set log to %s", args.Log)
	}

	if args.GinLog != "" {
		l.Infof("set gin log to %s", args.GinLog)
		mc.Logging.GinLog = args.GinLog
	}

	if javaHome := getJavaHome(); javaHome != "" {
		if mc.RemoteJob == nil {
			mc.RemoteJob = &io.RemoteJob{}
		}
		mc.RemoteJob.JavaHome = javaHome
	}

	return mc, nil
}
