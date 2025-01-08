// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
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

type AgentAddr struct {
	Host    string
	Port    string
	UDSAddr string
}

const (
	DatakitConfPath = "/usr/local/datakit/conf.d/datakit.conf"
	DefaultDKHost   = "127.0.0.1"
	DefaultDKPort   = "9529"
	DefaultDKUDS    = "/var/run/datakit/datakit.sock"

	// used for container.
	EnvDKSocketAddr = "ENV_DATAKIT_SOCKET_ADDR"
)

func GetDKAddr() *AgentAddr {
	confPath := DatakitConfPath
	var aAddr AgentAddr
	if v := os.Getenv(EnvDKSocketAddr); v != "" {
		if _, err := os.Stat(v); err == nil {
			aAddr.UDSAddr = v
			return &aAddr
		}
	}

	if _, err := os.Stat(DefaultDKUDS); err == nil {
		aAddr.UDSAddr = DefaultDKUDS
	}

	aAddr.Host, aAddr.Port = DefaultDKHost, DefaultDKPort
	v := map[string]any{}
	if _, err := toml.DecodeFile(confPath, &v); err != nil {
		return &aAddr
	}

	if httpAPI, ok := v["http_api"]; ok {
		if v, ok := httpAPI.(map[string]any); ok {
			if l, ok := v["listen"]; ok {
				if l, ok := l.(string); ok {
					if h, p, e := net.SplitHostPort(l); e == nil {
						aAddr.Host, aAddr.Port = h, p
					}
				}
			}
			if l, ok := v["listen_socket"]; ok {
				if l, ok := l.(string); ok && l != "" {
					if _, err := os.Stat(l); err == nil {
						aAddr.UDSAddr = l
					}
				}
			}
		}
	}

	return &aAddr
}
