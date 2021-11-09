package build

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func tarFiles(goos, goarch string) {
	gz := filepath.Join(PubDir, Release, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		AppName, goos, goarch, ReleaseVersion))
	args := []string{
		`czf`,
		gz,
		`-C`,
		// the whole buildDir/datakit-<goos>-<goarch> dir
		filepath.Join(BuildDir, fmt.Sprintf("%s-%s-%s", AppName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...) //nolint:gosec

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	l.Debugf("tar %s...", gz)
	if err := cmd.Run(); err != nil {
		l.Fatal(err)
	}
}

//nolint:funlen,gocyclo
func PubDatakit() {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch Release {
	case `test`, `local`, `release`, `preprod`:
		tag := strings.ToUpper(Release)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		l.Fatalf("unknown release type: %s", Release)
	}

	if ak == "" || sk == "" {
		l.Fatalf("oss access key or secret key missing, tag=%s", strings.ToUpper(Release))
	}

	ossSlice := strings.SplitN(DownloadAddr, "/", 2)
	if len(ossSlice) != 2 {
		l.Fatalf("downloadAddr:%s err", DownloadAddr)
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
		l.Fatal(err)
	}

	// upload all build archs
	archs := parseArchs(Archs)

	ossfiles := map[string]string{
		path.Join(OSSPath, "version"):                                     path.Join(PubDir, Release, "version"),
		path.Join(OSSPath, "install.sh"):                                  "install.sh",
		path.Join(OSSPath, "install.ps1"):                                 "install.ps1",
		path.Join(OSSPath, fmt.Sprintf("install-%s.sh", ReleaseVersion)):  "install.sh",
		path.Join(OSSPath, fmt.Sprintf("install-%s.ps1", ReleaseVersion)): "install.ps1",
	}

	if Archs == datakit.OSArchDarwinAmd64 {
		delete(ossfiles, path.Join(OSSPath, "version"))
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range archs {
		if arch == datakit.OSArchDarwinAmd64 && runtime.GOOS != datakit.OSDarwin {
			l.Warn("Not a darwin system, skip the upload of related files.")
			continue
		}

		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppName, goos+"-"+goarch, ReleaseVersion)

		installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
		installerExeWithVer := fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
		if parts[0] == datakit.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
			installerExeWithVer = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
		}

		ossfiles[path.Join(OSSPath, gzName)] = path.Join(PubDir, Release, gzName)
		ossfiles[path.Join(OSSPath, installerExe)] = path.Join(PubDir, Release, installerExe)
		ossfiles[path.Join(OSSPath, installerExeWithVer)] = path.Join(PubDir, Release, installerExe)
	}

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			l.Fatal(err)
		}
	}

	for k, v := range ossfiles {
		fi, _ := os.Stat(v)
		l.Debugf("%s => %s(%s)...", v, k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(v, k); err != nil {
			l.Fatal(err)
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}
