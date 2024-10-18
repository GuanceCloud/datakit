// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cfg is used to control the behavior of this package.
package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	DefaultWorkDir = "/usr/local/pprofparser"
	DefaultCfgName = "conf.yml"
	DefaultCfgPath = DefaultWorkDir + "/" + DefaultCfgName
)

const (
	EnvLocal      = "local"
	EnvDev        = "dev"
	EnvTest       = "test"
	EnvPre        = "pre"
	EnvProduction = "prod"
)

var (
	Cfg *Config
)

func Load(file string) error {
	reader, err := os.Open(file)
	if err != nil {
		var wd string
		if wd, err = os.Getwd(); err == nil {
			cfgFile := filepath.Join(wd, "cfg", DefaultCfgName)
			reader, err = os.Open(cfgFile)
		}
		if err != nil {
			return fmt.Errorf("read config file fail: %w", err)
		}
	}
	defer func() {
		_ = reader.Close()
	}()

	decoder := yaml.NewDecoder(reader)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return fmt.Errorf("decode yaml config file fail: %w", err)
	}
	Cfg = &cfg
	return nil
}

type Config struct {
	Serv    Server  `yaml:"server"`
	Log     Log     `yaml:"log"`
	Gin     Gin     `yaml:"gin"`
	Oss     Oss     `yaml:"oss"`
	Storage Storage `yaml:"storage"`
}

// Server configuration
type Server struct {
	Addr string `yaml:"addr"`
	Port string `yaml:"port"`
}

// Log log configuration
type Log struct {
	Path  string `yaml:"path"`
	File  string `yaml:"file"`
	Level string `yaml:"level"`
}

// Gin gin configuration
type Gin struct {
	RunMode  string `yaml:"run_mode"`
	Log      string `yaml:"log"`
	ErrorLog string `yaml:"error_log"`
}

// Oss aliyun oss configuration
type Oss struct {
	Host          string `yaml:"host"`
	AccessKey     string `yaml:"access_key"`
	SecretKey     string `yaml:"secret_key"`
	ProfileBucket string `yaml:"profile_bucket"`
	ProfileDir    string `yaml:"profile_dir"`
}

type Disk struct {
	ProfileDir string `yaml:"profile_dir"`
}

type Storage struct {
	Disk Disk `yaml:"disk"`
	Oss  Oss  `yaml:"oss"`
}
