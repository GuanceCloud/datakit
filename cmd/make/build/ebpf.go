// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

//nolint:funlen,gocyclo
func PubDatakitEBpf() error {
	var (
		start  = time.Now()
		basics []ossFile
		// upload all build archs
		curTmpArchs = ParseArchs(Archs)
	)

	// tar files and collect OSS upload/backup info
	for _, arch := range curTmpArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		for _, appName := range StandaloneApps {
			buildPath := filepath.Join(DistDir, "standalone")
			switch appName {
			case "datakit-ebpf":
				if parts[0] != runtime.GOOS {
					continue
				}
				if parts[0] != datakit.OSLinux {
					continue
				}
				if parts[1] != runtime.GOARCH {
					continue
				}
			default:
			}

			curEBpfArchs = append(curEBpfArchs, arch)
			gz, gzpath := tarFiles(DistDir, buildPath, appName, parts[0], parts[1], tarWithReleaseVer)
			basics = append(basics, ossFile{gz, gzpath})
		}
	}

	ossfiles := addOSSFiles(ossCli.WorkDir, basics)

	// test if all file ok before uploading
	for _, x := range ossfiles {
		if _, err := os.Stat(x.local); err != nil {
			return err
		}
	}

	for _, x := range ossfiles {
		fi, _ := os.Stat(x.local)
		l.Debugf("%s => %s(%s)...", x.local, x.remote, humanize.Bytes(uint64(fi.Size())))

		if err := ossCli.Upload(x.local, x.remote); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

// PackageEBPF download exist eBPF input from OSS.
//
// we have build and uploaded eBPF input binary in previous steps.
func PackageEBPF() error {
	l.Debug("Start downloading ebpf...")

	uploadAddr := fmt.Sprintf("%s.%s/%s", ossCli.BucketName, ossCli.Host, ossCli.WorkDir)

	curArch := ParseArchs(Archs)
	for _, arch := range curArch {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid arch: %s", arch)
			l.Error(err)
			NotifyFail(err.Error())
		}
		goos, goarch := parts[0], parts[1]
		if goos == datakit.OSLinux {
			url := "https://" + filepath.Join(uploadAddr, fmt.Sprintf(
				"datakit-ebpf-%s-%s-%s.tar.gz", goos, goarch, ReleaseVersion))
			dir := fmt.Sprintf("%s/%s-%s-%s/externals/", DistDir, AppName, goos, goarch)

			switch goarch {
			case "amd64", "arm64":

				l.Infof("Downloading %s => %s\n", url, dir)
				if err := downloader.Download(httpcli.Cli(nil), url, dir, false, false); err != nil {
					l.Errorf("downloader.Download: %s", err)

					NotifyFail(err.Error())
					return err
				}
			}
		}
	}

	return nil
}
