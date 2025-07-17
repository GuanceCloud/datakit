// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package changes

import (
	"fmt"
	"strings"
)

type Manifests struct {
	K8sManifest  *Manifest
	HostManifest *Manifest
}

type Manifest struct {
	Version string   `toml:"version"`
	Changes []Change `toml:"change"`
}

func (m *Manifest) MDTable(lang string) string {
	if len(m.Changes) == 0 {
		return ""
	}

	arr := []string{
		`| ID | Title & Message |`,
		`|--- |--- |`,
	}

	for _, c := range m.Changes {
		switch lang {
		// change \n as line break in markdown table column
		case "en":
			arr = append(arr, fmt.Sprintf("|**%s**|**Title**: %s<br>**Message**:<br>%s|", c.ID, c.Title.En, strings.ReplaceAll(c.Message.En, "\n", "<br>")))
		case "zh":
			arr = append(arr, fmt.Sprintf("|**%s**|**Title**: %s<br>**Message**:<br>%s|", c.ID, c.Title.Zh, strings.ReplaceAll(c.Message.Zh, "\n", "<br>")))
		}
	}

	return strings.Join(arr, "\n")
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
