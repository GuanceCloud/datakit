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

	"github.com/GuanceCloud/cliutils"
	humanize "github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

//nolint:funlen,gocyclo
func PubDatakitEBpf() error {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch ReleaseType {
	case ReleaseTesting, ReleaseProduction, ReleaseLocal:
		tag := strings.ToUpper(ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		return fmt.Errorf("unknown release type: %s", ReleaseType)
	}

	if ak == "" || sk == "" {
		return fmt.Errorf("OSS %s/%s not set",
			strings.ToUpper(ReleaseType)+"_OSS_ACCESS_KEY",
			strings.ToUpper(ReleaseType)+"_OSS_SECRET_KEY")
	}

	ossSlice := strings.SplitN(UploadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", UploadAddr)
	}
	OSSPath = ossSlice[1]

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    OSSPath,
	}

	if err := oc.Init(); err != nil {
		return err
	}

	// upload all build archs
	curTmpArchs := ParseArchs(Archs)

	basics := map[string]string{}

	// tar files and collect OSS upload/backup info
	for _, arch := range curTmpArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		for _, appName := range StandaloneApps {
			buildPath := filepath.Join(BuildDir, "standalone")
			switch appName {
			case "datakit-ebpf":
				if parts[0] != runtime.GOOS {
					continue
				}
				if parts[0] != "linux" {
					continue
				}
				if parts[1] != runtime.GOARCH {
					continue
				}
			default:
			}
			curEBpfArchs = append(curEBpfArchs, arch)
			gz, gzP := tarFiles(PubDir, buildPath, appName, parts[0], parts[1], TarWithRlsVer)
			basics[gz] = gzP
		}
	}

	ossfiles := addOSSFiles(OSSPath, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			return err
		}
	}

	for k, v := range ossfiles {
		fi, _ := os.Stat(v)
		l.Debugf("%s => %s(%s)...", v, k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(v, k); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

func PackageeBPF() {
	curArch := ParseArchs(Archs)
	for _, arch := range curArch {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid arch: %s", arch)
			l.Error(err)
			NotifyFail(err.Error())
		}
		goos, goarch := parts[0], parts[1]
		if goos == "linux" {
			url := "https://" + filepath.Join(UploadAddr, fmt.Sprintf(
				"datakit-ebpf-%s-%s-%s.tar.gz", goos, goarch, ReleaseVersion))
			dir := fmt.Sprintf("%s/%s-%s-%s/externals/", BuildDir, AppName, goos, goarch)

			switch goarch {
			case "amd64", "arm64":
				if err := downloader.Download(ihttp.Cli(nil), url, dir, false, false); err != nil {
					l.Error(err)
					NotifyFail(err.Error())
				}
			}
		}
	}
}
