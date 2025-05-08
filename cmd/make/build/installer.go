// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func buildInstaller(outdir, goos, goarch string) error {
	l.Debugf("building %s-%s/installer...", goos, goarch)

	installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
	if goos == datakit.OSWindows {
		installerExe += winBinSuffix
	}

	var cmdArgs []string
	if RaceDetection != "off" && runtime.GOOS == goos && runtime.GOARCH == goarch {
		l.Infof("race deteciton enabled")
		cmdArgs = []string{
			"go", "build", "-race",
		}
	} else {
		cmdArgs = []string{
			"go", "build",
		}
	}

	ldflags := fmt.Sprintf("-w -s -X main.DataKitBaseURL=%s -X main.DataKitVersion=%s", DownloadCDN, ReleaseVersion)

	l.Infof("set ldflags on installer: %q", ldflags)
	cmdArgs = append(cmdArgs, []string{
		"-o", filepath.Join(outdir, installerExe),
		"-ldflags", ldflags,
		"cmd/installer/main.go",
	}...)

	var envs []string
	if RaceDetection != "off" && runtime.GOOS == goos && runtime.GOARCH == goarch {
		envs = []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
			"CGO_ENABLED=1",
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
