// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package installer contains installer functions
package installer

import (
	"net/http"
	"strings"
)

const (
	preloadConfigFilePath = "/etc/ld.so.preload"
	dkruncBinName         = "dkrunc"
	launcherName          = "apm_launcher"

	DirInject          = "apm_inject"
	DirInjectSubInject = "inject"
	DirInjectSubLib    = "lib"

	dockerDaemonJSONPath      = "/etc/docker/daemon.json"
	dockerFieldDefaultRuntime = "default-runtime"
	dockerFieldRuntimes       = "runtimes"

	RuntimeRunc   = "runc"
	RuntimeDkRunc = "dk-runc"
)

const (
	KindHOST    = "host"
	KindDocker  = "docker"
	KindDisable = "disable"
)

type config struct {
	enableHostInject   bool
	enableDockerInject bool
	installDir         string
	launcherURL        string
	ddJavaLibURL       string
	pyLib              bool

	cli *http.Client

	forceUpgradeAPMLib bool
}

type Opt func(c *config)

func WithInstallDir(dir string) Opt {
	return func(c *config) {
		c.installDir = dir
	}
}

func WithInstrumentationEnabled(s string) Opt {
	return func(c *config) {
		disable := false
		for _, val := range strings.Split(s, ",") {
			switch strings.TrimSpace(val) {
			case KindDisable:
				disable = true
			case KindDocker:
				c.enableDockerInject = true
			case KindHOST:
				c.enableHostInject = true
			}
		}
		if disable || s == "" {
			c.enableDockerInject = false
			c.enableHostInject = false
		}
	}
}

func WithLauncherURL(cli *http.Client, launcherURL string) Opt {
	return func(c *config) {
		c.cli = cli
		c.launcherURL = launcherURL
	}
}

func WithJavaLibURL(url string) Opt {
	return func(c *config) {
		c.ddJavaLibURL = url
	}
}

func WithPythonLib(ok bool) Opt {
	return func(c *config) {
		c.pyLib = ok
	}
}

func WithForceUpgradeLib(y bool) Opt {
	return func(c *config) {
		c.forceUpgradeAPMLib = y
	}
}
