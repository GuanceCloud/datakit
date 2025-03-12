// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"fmt"
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

	RuntimeRunc   = "runc"
	RuntimeDkRunc = "dk-runc"
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
	DkHost string
	DkPort string
	DkUds  string

	SDHost string
	SDPort string
	SDUds  string
}

const (
	DatakitConfPath = "/usr/local/datakit/conf.d/datakit.conf"
	DefaultDKHost   = "127.0.0.1"
	DefaultDKPort   = "9529"
	DefaultDKUDS    = "/var/run/datakit/datakit.sock"

	StatsDConfPath    = "/usr/local/datakit/conf.d/statsd/statsd.conf"
	DefaultStatsDHost = "127.0.0.1"
	DefaultStatsDPort = "8125"
	DefaultStatsDUDS  = "/var/run/datakit/statsd.sock"

	// used for container.
	EnvDKSocketAddr     = "ENV_DATAKIT_SOCKET_ADDR"
	EnvStatsdSocketAddr = "ENV_DATAKIT_STATSD_SOCKET_ADDR"
)

type iptsCfg struct {
	Inputs map[string][]map[string]any `toml:"inputs"`
}

func GetDKAddr() *AgentAddr {
	var aAddr AgentAddr

	// unix addr first
	if v := getEnv(EnvDKSocketAddr, DefaultDKUDS); v != "" {
		if v := getEnv(EnvStatsdSocketAddr, DefaultStatsDUDS); v != "" {
			if _, err := os.Stat(v); err == nil {
				aAddr.SDUds = v
			}
		}
		if _, err := os.Stat(v); err == nil {
			aAddr.DkUds = v
			return &aAddr
		}
	}

	aAddr.SDHost = DefaultStatsDHost
	aAddr.SDPort = DefaultStatsDPort

	var sdCfg iptsCfg
	if err := loadConfig(StatsDConfPath, &sdCfg); err == nil {
		if sd, ok := sdCfg.Inputs["statsd"]; ok && len(sd) > 0 {
			if c, ok := sd[0]["service_address"]; ok {
				if c, ok := c.(string); ok {
					if host, port, err := net.SplitHostPort(c); err == nil {
						if host == "" {
							host = "127.0.0.1"
						}
						aAddr.SDHost = host
						aAddr.SDPort = port
					}
				}
			}
		}
	}

	aAddr.DkHost, aAddr.DkPort = DefaultDKHost, DefaultDKPort

	var dkconf map[string]any
	if err := loadConfig(DatakitConfPath, &dkconf); err == nil {
		if httpAPI, ok := dkconf["http_api"]; ok {
			if v, ok := httpAPI.(map[string]any); ok {
				if l, ok := v["listen"]; ok {
					if l, ok := l.(string); ok {
						if h, p, e := net.SplitHostPort(l); e == nil {
							aAddr.DkHost, aAddr.DkPort = h, p
						}
					}
				}
				if l, ok := v["listen_socket"]; ok {
					if l, ok := l.(string); ok && l != "" {
						if _, err := os.Stat(l); err == nil {
							aAddr.DkUds = l
						}
					}
				}
			}
		}
	}

	return &aAddr
}

func loadConfig(filePath string, v any) error {
	if _, err := toml.DecodeFile(filePath, v); err != nil {
		return fmt.Errorf("failed to load config %w", err)
	}
	return nil
}

func getEnv(name, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
