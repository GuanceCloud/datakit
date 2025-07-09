// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package changes

type Manifests struct {
	K8sManifest  *Manifest
	HostManifest *Manifest
}

type Manifest struct {
	Version string   `toml:"version"`
	Changes []Change `toml:"change"`
}

type Change struct {
	ID      string `toml:"id"`
	Title   I18n   `toml:"title"`
	Message I18n   `toml:"message"`
}

type I18n struct {
	Zh string `toml:"zh"`
	En string `toml:"en"`
}

type Language int

const (
	LangEn Language = iota
	LangZh
)
