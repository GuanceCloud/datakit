// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"strings"
)

type Option struct {
	// DataKit connection options
	DatakitHTTPEndpoint string `mapstructure:"datakit-http-endpoint"`

	// Common options
	LogPath  string `mapstructure:"log-path"`
	LogLevel string `mapstructure:"log-level"`
	Tags     string `mapstructure:"tags"`

	// Journald specific options
	Paths         string `mapstructure:"paths"`
	Units         string `mapstructure:"units"`
	Priorities    string `mapstructure:"priorities"`
	ExcludeFields string `mapstructure:"exclude-fields"`
	TailOnly      bool   `mapstructure:"tail-only"`
	MaxEntries    int    `mapstructure:"max-entries"`
	SaveCursor    bool   `mapstructure:"save-cursor"`
	CursorFile    string `mapstructure:"cursor-file"`
}

func (opt *Option) parseTags() map[string]string {
	tags := make(map[string]string)
	if opt.Tags == "" {
		return tags
	}

	pairs := strings.Split(opt.Tags, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return tags
}

type IInput interface {
	Run()
	Terminate()
}
