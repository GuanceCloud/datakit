// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func runExportFlags() error {
	inputs.TODO = *flagExportTODO

	return doExport()
}

func doExport() error {
	var exporters []export.Exporter
	if *flagExportDocDir != "" {
		cp.Infof("export guance docs to %q...\n", *flagExportDocDir)
		exporters = append(exporters,
			export.NewGuanceDodcs(
				export.WithTopDir(*flagExportDocDir),
				export.WithVersion(*flagExportVersion),
				export.WithExclude(*flagExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	if *flagExportIntegrationDir != "" {
		cp.Infof("export integrations to %q...\n", *flagExportIntegrationDir)
		exporters = append(exporters,
			export.NewIntegration(
				export.WithTopDir(*flagExportIntegrationDir),
				export.WithVersion(*flagExportVersion),
				export.WithExclude(*flagExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	cp.Infof("run %d exporters...\n", len(exporters))
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
