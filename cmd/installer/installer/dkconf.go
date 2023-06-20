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
		"self",
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
		"self",
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
		"self",

		// host_processes is costly, maybe we should disable default
		"host_processes",
	}

	DataKitVersion = ""

	OTA = false

	EnableInputs,
	CloudProvider,
	Proxy,
	Dataway string

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

	Sinker string

	CgroupDisabled int
	LimitCPUMax,
	LimitCPUMin float64
	LimitMemMax int64
)

func writeDefInputToMainCfg(mc *config.Config) {
	hostInputs := defaultHostInputs

	switch runtime.GOOS {
	case datakit.OSLinux:
		hostInputs = defaultHostInputsForLinux
	case datakit.OSDarwin:
		hostInputs = defaultHostInputsForMacOS
	}

	// Enable default input, auto remove duplicated input name.
	if EnableInputs == "" {
		x := strings.Join(hostInputs, ",")

		cp.Infof("Use default enabled inputs '%s'\n", x)
		mc.EnableDefaultsInputs(x)
	} else {
		cp.Infof("Set default inputs '%s'...\n", EnableInputs)
		mc.EnableDefaultsInputs(EnableInputs)
	}

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
			config.Cfg.Dataway = dw
			return dw, nil
		}
	} else {
		return nil, fmt.Errorf("dataway is not set")
	}
}
