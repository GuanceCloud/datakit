// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
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
		l.Debugf("export guance docs to %q...", ExportDocDir)
		exporters = append(exporters,
			export.NewGuanceDodcs(
				export.WithTopDir(ExportDocDir),
				export.WithVersion(ExportVersion),
				export.WithDCAVersion(DCAVersion),
				export.WithExclude(ExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	if ExportIntegrationDir != "" {
		l.Debugf("export integrations to %q...", ExportIntegrationDir)
		exporters = append(exporters,
			export.NewIntegration(
				export.WithTopDir(ExportIntegrationDir),
				export.WithVersion(ExportVersion),
				export.WithDCAVersion(DCAVersion),
				export.WithExclude(ExportIgnore),
				export.WithIgnoreMissing(true),
			))
	}

	l.Debugf("run %d exporters...", len(exporters))
	for _, exporter := range exporters {
		l.Debugf("run exporter %+#v...", exporter)

		if err := exporter.Export(); err != nil {
			return err
		}

		if err := exporter.Check(); err != nil {
			return err
		}
	}

	return nil
}
