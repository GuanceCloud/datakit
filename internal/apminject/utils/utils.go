// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package utils contains utils
package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

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

	// disable inject.
	EnvDKAPMINJECT = "ENV_DATAKIT_DISABLE_APM_INS"
)

var (
	ErrPyLibNotFound    = errors.New("ddtrace-run not found")
	ErrParseJavaVersion = errors.New(("failed to parse java version"))
	ErrJavaLibNotFound  = errors.New("dd-java-agent.jar not found")
	ErrUnsupportedJava  = errors.New(("unsupported java version"))
	ErrAlreadyInjected  = errors.New(("already injected"))
	ErrInjectDisabled   = errors.New("inject disabled")
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

func GetJavaVersion(s string) (int, error) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return 0, ErrParseJavaVersion
	}

	idx := strings.Index(lines[0], "\"")
	if idx == -1 {
		return 0, ErrParseJavaVersion
	}
	idxTail := strings.LastIndex(lines[0], "\"")
	if idx == -1 {
		return 0, ErrParseJavaVersion
	}

	versionStr := lines[0][idx+1 : idxTail-1]
	li := strings.Split(versionStr, ".")
	if len(li) < 2 {
		return 0, ErrParseJavaVersion
	}

	v, err := strconv.Atoi(li[0])
	if err != nil {
		return 0, err
	}

	if v == 1 {
		v, err = strconv.Atoi(li[1])
		if err != nil {
			return 0, err
		}
		return v, nil
	} else {
		return v, nil
	}
}

func CheckDisableInjFromEnv(k, v string) bool {
	if k == EnvDKAPMINJECT {
		switch strings.ToLower(v) {
		case "f", "false", "0":
			return false
		default:
			return true
		}
	}

	return false
}
