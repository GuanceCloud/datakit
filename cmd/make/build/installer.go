// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func generateInstallScript() error {
	x := struct {
		InstallBaseURL string
		Version        string
	}{
		InstallBaseURL: DownloadCDN,
		Version:        ReleaseVersion,
	}

	l.Infof("generating install scripts for version %q with base download URL %q",
		ReleaseVersion, DownloadCDN)

	for k, v := range map[string]string{
		"install.sh.template":   "install.sh",
		"install.ps1.template":  "install.ps1",
		"datakit.yaml.template": "datakit.yaml",
	} {
		txt, err := ioutil.ReadFile(filepath.Clean(k))
		if err != nil {
			return err
		}

		t := template.New("")
		t, err = t.Parse(string(txt))
		if err != nil {
			return err
		}

		fd, err := os.OpenFile(filepath.Clean(v),
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}

		l.Infof("creating install script %s", v)
		if err := t.Execute(fd, x); err != nil {
			return err
		}

		fd.Close() //nolint:errcheck,gosec
	}

	return nil
}

func buildInstaller(outdir, goos, goarch string) error {
	l.Debugf("building %s-%s/installer...", goos, goarch)

	installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
	if goos == datakit.OSWindows {
		installerExe += winBinSuffix
	}

	var cmdArgs []string
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		l.Infof("race deteciton enabled")
		cmdArgs = []string{
			"go", "build", "-race",
		}
	} else {
		cmdArgs = []string{
			"go", "build",
		}
	}

	cmdArgs = append(cmdArgs, []string{
		"-o", filepath.Join(outdir, installerExe),
		"-ldflags",
		fmt.Sprintf("-w -s -X main.DataKitBaseURL=%s -X main.DataKitVersion=%s",
			DownloadCDN,
			ReleaseVersion),
		"cmd/installer/main.go",
	}...)

	var envs []string
	if RaceDetection && runtime.GOOS == goos && runtime.GOARCH == goarch {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			"CGO_ENABLED=on",
		}
	} else {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			"CGO_ENABLED=0",
		}
	}

	msg, err := runEnv(cmdArgs, envs)
	if err != nil {
		return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s", cmdArgs, envs, err, string(msg))
	}
	return nil
}
