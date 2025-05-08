// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func buildDatakitHelm() error {
	files := [][2]string{
		{filepath.Join(DistDir, "helm-values.yaml"), filepath.Join(HelmChartDir, "values.yaml")},
		{filepath.Join(DistDir, "helm-Chart.yaml"), filepath.Join(HelmChartDir, "Chart.yaml")},
		{filepath.Join(DistDir, "helm-README.md"), filepath.Join(HelmChartDir, "README.md")},
		{filepath.Join(DistDir, "helm-questions.md"), filepath.Join(HelmChartDir, "questions.yml")},
	}

	// copy generated files to helm chart path
	for _, x := range files {
		if data, err := os.ReadFile(x[0]); err != nil {
			return err
		} else {
			if err := os.WriteFile(x[1], data, os.ModePerm); err != nil {
				return err
			}
		}
	}

	// run helm package command
	cmdArgs := []string{
		"helm", "package", HelmChartDir,
		"--version", strings.Split(ReleaseVersion, "-")[0],
		"--app-version", ReleaseVersion,
		"--destination", DistDir,
	}

	msg, err := runEnv(cmdArgs, nil)
	if err != nil {
		return fmt.Errorf("failed to run %v: %w, msg: %s", cmdArgs, err, string(msg))
	}
	return nil
}
