// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var ExportTODO,
	ExportDocDir,
	ExportVersion,
	ExportIgnore,
	ExportIntegrationDir string

func BuildExport() error {
	inputs.TODO = ExportTODO

	return doExport()
}

func doExport() error {
	var exporters []export.Exporter
	if ExportDocDir != "" {
		fmt.Printf("export guance docs to %q...\n", ExportDocDir)
		exporters = append(exporters,
			export.NewGuanceDodcs(
				export.WithTopDir(ExportDocDir),
				export.WithVersion(ExportVersion),
				export.WithExclude(ExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	if ExportIntegrationDir != "" {
		fmt.Printf("export integrations to %q...\n", ExportIntegrationDir)
		exporters = append(exporters,
			export.NewIntegration(
				export.WithTopDir(ExportIntegrationDir),
				export.WithVersion(ExportVersion),
				export.WithExclude(ExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	fmt.Printf("run %d exporters...\n", len(exporters))
	for _, exporter := range exporters {
		if err := exporter.Export(); err != nil {
			return err
		}

		if err := exporter.Check(); err != nil {
			return err
		}
	}

	return nil
}
