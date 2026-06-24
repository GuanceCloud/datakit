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

	"gopkg.in/yaml.v2"
)

func buildDatakitHelm() error {
	if SkipHelm != 0 {
		l.Warnf("build helm packages skipped")
		return nil
	}

	files := [][2]string{
		{filepath.Join(DistDir, "helm-values.yaml"), filepath.Join(HelmChartDir, "values.yaml")},
		{filepath.Join(DistDir, "helm-values-gke-autopilot.yaml"), filepath.Join(HelmChartDir, "values-gke-autopilot.yaml")},
		{filepath.Join(DistDir, "helm-Chart.yaml"), filepath.Join(HelmChartDir, "Chart.yaml")},
		{filepath.Join(DistDir, "helm-README.md"), filepath.Join(HelmChartDir, "README.md")},
		{filepath.Join(DistDir, "helm-questions.md"), filepath.Join(HelmChartDir, "questions.yml")},
	}

	if err := copyFiles(files); err != nil {
		return err
	}

	if err := packageHelmChart(HelmChartDir); err != nil {
		return err
	}

	return buildGKEAutopilotHelm()
}

func buildGKEAutopilotHelm() error {
	chartDir := filepath.Join(DistDir, "datakit-gke-autopilot-chart")
	if err := os.RemoveAll(chartDir); err != nil {
		return err
	}
	if err := copyDir(HelmChartDir, chartDir); err != nil {
		return err
	}

	if err := mergeYAMLFiles(
		filepath.Join(DistDir, "helm-values.yaml"),
		filepath.Join(DistDir, "helm-values-gke-autopilot.yaml"),
		filepath.Join(chartDir, "values.yaml"),
	); err != nil {
		return err
	}

	files := [][2]string{
		{filepath.Join(DistDir, "helm-Chart-gke-autopilot.yaml"), filepath.Join(chartDir, "Chart.yaml")},
		{filepath.Join(DistDir, "helm-README.md"), filepath.Join(chartDir, "README.md")},
		{filepath.Join(DistDir, "helm-questions.md"), filepath.Join(chartDir, "questions.yml")},
	}
	if err := copyFiles(files); err != nil {
		return err
	}

	return packageHelmChart(chartDir)
}

func copyFiles(files [][2]string) error {
	for _, x := range files {
		data, err := os.ReadFile(x[0])
		if err != nil {
			return err
		}
		if err := os.WriteFile(x[1], data, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func mergeYAMLFiles(basePath, overridePath, dstPath string) error {
	baseData, err := os.ReadFile(basePath)
	if err != nil {
		return err
	}
	overrideData, err := os.ReadFile(overridePath)
	if err != nil {
		return err
	}

	base := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(baseData, &base); err != nil {
		return err
	}
	override := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(overrideData, &override); err != nil {
		return err
	}

	mergeYAMLMap(base, override)

	data, err := yaml.Marshal(base)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, os.ModePerm)
}

func mergeYAMLMap(dst, src map[interface{}]interface{}) {
	for key, srcVal := range src {
		srcMap, srcOK := srcVal.(map[interface{}]interface{})
		dstMap, dstOK := dst[key].(map[interface{}]interface{})
		if srcOK && dstOK {
			mergeYAMLMap(dstMap, srcMap)
			continue
		}
		dst[key] = srcVal
	}
}

func packageHelmChart(chartDir string) error {
	cmdArgs := []string{
		"helm", "package", chartDir,
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

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		switch {
		case d.Type()&os.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		case d.IsDir():
			return os.MkdirAll(target, info.Mode())
		case d.Type().IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode())
		default:
			return nil
		}
	})
}
