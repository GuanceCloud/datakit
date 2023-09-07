// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
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
	Dataway string

	HTTPPublicAPIs        string
	HTTPDisabledAPIs      string
	InstallRUMSymbolTools int

	DCAWhiteList,
	DCAEnable,
	DCAListen string

	HTTPPort int
	HTTPListen,
	DatakitName,
	GlobalHostTags,
	HostName,
	IPDBType string

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
	LimitCPUMax,
	LimitCPUMin float64
	LimitMemMax int64
)

// generate default inputs list.
func mergeDefaultInputs(defaultList, enabledList []string, appendDefault bool) []string {
	if len(enabledList) == 0 {
		return defaultList // no inputs enabled(disabled), enable all default inputs
	}

	cp.Infof("enabledList: %+#v\n", enabledList)

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
			cp.Warnf("input %q disabled\n", elem)
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
				cp.Infof("input %q enabled\n", elem)
				res = append(res, elem)
			}
		}
	}

	if len(whiteList) > 0 {
		for _, elem := range defaultList {
			if appendDefault {
				cp.Infof("input %q enabled\n", elem)
				res = append(res, elem)
			} else {
				// disable them
				if _, ok := whiteList[elem]; !ok { // not enabled, then disable it
					cp.Warnf("input %q disabled\n", elem)
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

func setupDefaultInputs(mc *config.Config, arg string, list []string, upgrade bool) {
	if upgrade {
		if len(mc.DefaultEnabledInputs) == 0 { // all default inputs disabled
			mc.DefaultEnabledInputs = mergeDefaultInputs(list, []string{"-"}, true)
		} else {
			mc.DefaultEnabledInputs = mergeDefaultInputs(list, mc.DefaultEnabledInputs, true)
		}
	} else {
		if arg == "" {
			mc.DefaultEnabledInputs = mergeDefaultInputs(list, nil, false)
		} else {
			mc.DefaultEnabledInputs = mergeDefaultInputs(list, strings.Split(arg, ","), false)
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
			cp.Warnf("Failed to inject cloud-provider: %s\n", err.Error())
		} else {
			cp.Infof("Set cloud provider to %s ok\n", CloudProvider)
		}
	} else {
		cp.Infof("Cloud provider not set\n")
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
		if err := ioutil.WriteFile(cfgpath, conf, datakit.ConfPerm); err != nil {
			return err
		}

	case "": // pass

	default:
		cp.Warnf("Unknown cloud provider %s, ignored\n", p)
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
	dw := &dataway.Dataway{}

	if Dataway != "" {
		dw.URLs = strings.Split(Dataway, ",")

		if Proxy != "" {
			l.Debugf("set proxy to %s", Proxy)
			dw.HTTPProxy = Proxy
		}

		if err := dw.Init(); err != nil {
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
