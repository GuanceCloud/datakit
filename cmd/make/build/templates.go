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
	"text/template"
)

func buidlDCATemplates() error {
	x := struct {
		Version,
		BrandDomain,
		DockerImageRepo string
	}{
		Version:         DCAVersion,
		DockerImageRepo: brand(Brand).dcaDockerImageRepo(),
		BrandDomain:     brand(Brand).domain(),
	}

	l.Infof("generating install scripts on %+#v", x)

	for k, v := range map[string]string{
		"templates/dca.template.yaml": filepath.Join(DistDir, "dca.yaml"),
	} {
		txt, err := os.ReadFile(filepath.Clean(k))
		if err != nil {
			return err
		}

		t := template.New("")
		t, err = t.Parse(string(txt))
		if err != nil {
			return err
		}

		fd, err := os.OpenFile(filepath.Clean(v), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}

		l.Infof("creating install script %s from %s", v, k)
		if err := t.Execute(fd, x); err != nil {
			return fmt.Errorf("generate from %q failed: %w", k, err)
		}

		fd.Close() //nolint:errcheck,gosec
	}

	return nil
}

func buildTemplates() error {
	x := struct {
		InstallBaseURL,
		Version,
		DockerImageRepo,
		BrandDomain,
		HelmVersion, // NOTE: helm required the version to be 1.2.3, not 1.2.3-iss-xxxx
		BrandChartRepo string
	}{
		InstallBaseURL:  DownloadCDN,
		Version:         ReleaseVersion,
		DockerImageRepo: brand(Brand).dockerImageRepo(),
		BrandDomain:     brand(Brand).domain(),
		HelmVersion:     strings.Split(ReleaseVersion, "-")[0],
		BrandChartRepo:  brand(Brand).chartRepo(),
	}

	l.Infof("generating install scripts on %+#v", x)

	for k, v := range map[string]string{
		"templates/install.template.sh":           filepath.Join(DistDir, "install.sh"),
		"templates/install.template.ps1":          filepath.Join(DistDir, "install.ps1"),
		"templates/datakit.template.yaml":         filepath.Join(DistDir, "datakit.yaml"),
		"templates/datakit-elinker.template.yaml": filepath.Join(DistDir, "datakit-elinker.yaml"),
		"templates/charts-values.template.yaml":   filepath.Join(DistDir, "helm-values.yaml"),
		"templates/charts-Chart.template.yaml":    filepath.Join(DistDir, "helm-Chart.yaml"),
		"templates/charts-readme.template.md":     filepath.Join(DistDir, "helm-README.md"),
		"templates/charts-questions.template.yml": filepath.Join(DistDir, "helm-questions.md"),
	} {
		txt, err := os.ReadFile(filepath.Clean(k))
		if err != nil {
			return err
		}

		t := template.New("")
		t, err = t.Parse(string(txt))
		if err != nil {
			return err
		}

		fd, err := os.OpenFile(filepath.Clean(v), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}

		l.Infof("creating install script %s from %s", v, k)
		if err := t.Execute(fd, x); err != nil {
			return fmt.Errorf("generate from %q failed: %w", k, err)
		}

		fd.Close() //nolint:errcheck,gosec
	}

	return nil
}
