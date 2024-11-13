// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package io

type RemoteJob struct {
	Enable   bool     `toml:"enable"`
	ENVs     []string `toml:"envs"`
	Interval string   `toml:"interval"`
	JavaHome string   `toml:"java_home"`
}
