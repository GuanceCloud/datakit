package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"

	"github.com/dustin/go-humanize"
)

var (
	installerExe string
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func (vd *versionDesc) withoutGitCommit() string {
	parts := strings.Split(vd.Version, "-")
	if len(parts) != 3 {
		l.Fatalf("version info not in v<x.x>-<n>-g<commit-id> format: %s", vd.Version)
	}

	return strings.Join(parts[:2], "-")
}

func tarFiles(goos, goarch string) {

	gz := filepath.Join(PubDir, Release, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		AppName, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
		`-C`,
		// the whole buildDir/datakit-<goos>-<goarch> dir
		filepath.Join(BuildDir, fmt.Sprintf("%s-%s-%s", AppName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		l.Fatal(err)
	}
}

func getCurrentVersionInfo(url string) *versionDesc {

	l.Infof("get current online version: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		l.Fatal(err)
	}

	if resp.StatusCode != 200 {
		l.Warn("get current online version failed, ignored")
		return nil
	}

	defer resp.Body.Close()
	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Fatal(err)
	}

	l.Infof("current online version: %s", string(info))
	var vd versionDesc
	if err := json.Unmarshal(info, &vd); err != nil {
		l.Fatal(err)
	}
	return &vd
}

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

	// 请求线上版本信息
	url := fmt.Sprintf("http://%s.%s/%s/%s", bucket, ossHost, OSSPath, "version")
	curVd := getCurrentVersionInfo(url)

	// upload all build archs
	archs := []string{}

	switch Archs {
	case "all":
		archs = OSArches
	case "local":
		archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
	default:
		archs = strings.Split(Archs, "|")
	}

	ossfiles := map[string]string{
		path.Join(PubDir, Release, "version"): path.Join(OSSPath, "version"),
	}

	renameOssFiles := map[string]string{}
	var verId string

	if curVd != nil && curVd.Version == git.Version {
		l.Warnf("Current verison is the newest (%s <=> %s). Exit now.", curVd.Version, git.Version)
		os.Exit(0)
	}

	// rename installer
	if curVd != nil {
		verId = curVd.withoutGitCommit()
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppName, goos+"-"+goarch, git.Version)

		ossfiles[path.Join(PubDir, Release, gzName)] = path.Join(OSSPath, gzName)

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)

			if curVd != nil {
				renameOssFiles[path.Join(OSSPath, installerExe)] =
					path.Join(OSSPath, fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, verId))
			}

		} else {
			installerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)

			if curVd != nil {
				renameOssFiles[path.Join(OSSPath, installerExe)] =
					path.Join(OSSPath, fmt.Sprintf("installer-%s-%s-%s", goos, goarch, verId))
			}
		}

		ossfiles[path.Join(PubDir, Release, installerExe)] = path.Join(OSSPath, installerExe)
	}

	// backup old installer script online, make it possible to install old version if required
	for k, v := range renameOssFiles {
		if err := oc.Move(k, v); err != nil {
			l.Debugf("backup %s -> %s failed: %s, ignored", k, v, err.Error())
			continue
		}

		l.Debugf("backup %s -> %s ok", k, v)
	}

	// test if all file ok before uploading
	for k, _ := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			l.Fatal(err)
		}
	}

	for k, v := range ossfiles {

		fi, _ := os.Stat(k)
		l.Debugf("upload %s(%s)...", k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(k, v); err != nil {
			l.Fatal(err)
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}

func PubTelegraf() {

	var archs []string
	switch Archs {
	case "all":
		archs = OSArches
	case "local":
		archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
	default:
		archs = strings.Split(Archs, "|")
	}

	start := time.Now()
	var ak, sk, bucket, ossHost string
	objPath := "datakit/telegraf"

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

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   128 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    objPath,
	}

	if err := oc.Init(); err != nil {
		l.Fatal(err)
	}

	ossfiles := map[string]string{}
	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		relPath := fmt.Sprintf("%s-%s/agent", parts[0], parts[1])

		switch parts[0] {
		case "windows":
			relPath += ".exe"
		default: // pass
		}

		gz := fmt.Sprintf("agent-%s-%s.tar.gz", parts[0], parts[1])

		cmd := exec.Command("tar", []string{"czf",
			path.Join(PubDir, gz),
			path.Join(PubDir, relPath)}...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			l.Fatal(err)
		}

		ossfiles[path.Join(PubDir, gz)] = path.Join(objPath, gz)
	}

	for k, _ := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			l.Fatal(err)
		}
	}

	for k, v := range ossfiles {
		l.Debugf("upload %s -> %s.%s...", k, bucket, path.Join(ossHost, v))
		if err := oc.Upload(k, v); err != nil {
			l.Fatal(err)
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}
