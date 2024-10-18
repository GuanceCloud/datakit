// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"net/http"
	"strings"
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

	cli *http.Client

	offline             bool
	offlineLauncherPath string
	offlineAPMLibPath   string

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

func WithOnline(cli *http.Client, launcherURL string) Opt {
	return func(c *config) {
		c.launcherURL = launcherURL
		c.cli = cli
		c.offline = false
	}
}

func WithOffline(launcher, apmLib string) Opt {
	return func(c *config) {
		c.offline = true
		c.offlineLauncherPath = launcher
		c.offlineAPMLibPath = apmLib
	}
}

func WithForceUpgradeLib(y bool) Opt {
	return func(c *config) {
		c.forceUpgradeAPMLib = y
	}
}
