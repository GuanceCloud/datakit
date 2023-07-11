// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package man

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

//go:embed doc/*
var AllDocs embed.FS

//go:embed dashboard/*
var AllDashboards embed.FS

//go:embed monitor/*
var AllMonitors embed.FS

func dashboardTryLoad(name string, lang inputs.I18n) ([]byte, error) {
	arr := []string{
		// load under dashboard/{zh,en}/name dir
		filepath.Join("dashboard", lang.String(), name, "meta.json"),

		// or under dashboard/name.json
		filepath.Join("dashboard", name+".json"),
	}

	for _, f := range arr {
		j, err := AllDashboards.ReadFile(f)
		if err == nil {
			return j, nil
		}
	}

	return nil, fmt.Errorf("dashboard not found in %s", strings.Join(arr, ","))
}

func monitorTryLoad(name string, lang inputs.I18n) ([]byte, error) {
	arr := []string{
		// load under monitor/{zh,en}/name dir
		filepath.Join("monitor", lang.String(), name, "meta.json"),

		// or under monitor/name.json dir
		filepath.Join("monitor", name+".json"),
	}

	for _, f := range arr {
		j, err := AllDashboards.ReadFile(f)
		if err == nil {
			return j, nil
		}
	}

	return nil, fmt.Errorf("monitor not found in %s", strings.Join(arr, ","))
}
