// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func tarFiles(pubPath, buildPath, appName, goos, goarch string) (string, string) {
	gz := fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		appName, goos, goarch, ReleaseVersion)
	gzPath := filepath.Join(pubPath, ReleaseType, gz)

	args := []string{
		`czf`,
		gzPath,
		`-C`,
		// the whole basePath/appName-<goos>-<goarch> dir
		filepath.Join(buildPath, fmt.Sprintf("%s-%s-%s", appName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...) //nolint:gosec

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	l.Debugf("tar %s...", gzPath)
	if err := cmd.Run(); err != nil {
		l.Fatal(err)
	}
	return gz, gzPath
}

func generateInstallScript() error {
	x := struct {
		InstallBaseURL string
		Version        string
	}{
		InstallBaseURL: DownloadAddr,
		Version:        ReleaseVersion,
	}

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

func generateMetaInfo() error {
	return cmds.ExportMetaInfo("measurements-meta.json")
}

func generatePipelineDoc() error {
	encoding := base64.StdEncoding
	protoPrefix, descPrefix := "函数原型：", "函数说明："
	// Write function description & prototype.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		lines := strings.Split(plDoc.Doc, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, protoPrefix) {
				proto := strings.TrimPrefix(line, protoPrefix)
				// Prototype line contains starting and trailing ` only.
				if len(proto) >= 2 && strings.Index(proto, "`") == 0 && strings.Index(proto[1:], "`") == len(proto[1:])-1 {
					proto = proto[1 : len(proto)-1]
				}
				plDoc.Prototype = proto
			} else if strings.HasPrefix(line, descPrefix) {
				plDoc.Description = strings.TrimPrefix(line, descPrefix)
			}
		}
	}
	// Encode Markdown docs with base64.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		plDoc.Doc = encoding.EncodeToString([]byte(plDoc.Doc))
		plDoc.Prototype = encoding.EncodeToString([]byte(plDoc.Prototype))
		plDoc.Description = encoding.EncodeToString([]byte(plDoc.Description))
	}
	exportPLDocs := struct {
		Version   string                  `json:"version"`
		Docs      string                  `json:"docs"`
		Functions map[string]*funcs.PLDoc `json:"functions"`
	}{
		Version:   git.Version,
		Docs:      "经过 base64 编码的 pipeline 函数文档，包括各函数原型、函数说明、使用示例",
		Functions: funcs.PipelineFunctionDocs,
	}
	data, err := json.Marshal(exportPLDocs)
	if err != nil {
		return err
	}
	f, err := os.Create("pipeline-docs.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func generatePipelineScripts() error {
	encoding := base64.StdEncoding
	demoMap, err := config.GetPipelineDemoMap()
	if err != nil {
		return err
	}

	// Encode script and log examples with base64.
	for scriptName, demo := range demoMap {
		demo.Pipeline = encoding.EncodeToString([]byte(demo.Pipeline))
		for n, e := range demo.Examples {
			demo.Examples[n] = encoding.EncodeToString([]byte(e))
		}
		demoMap[scriptName] = demo
	}

	data, err := json.Marshal(demoMap)
	if err != nil {
		return err
	}
	f, err := os.Create("internal-pipelines.json")
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func addOSSFiles(ossPath string, files map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range files {
		res[path.Join(ossPath, k)] = v
	}
	return res
}

//nolint:funlen,gocyclo
func PubDatakit() error {
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

	ossSlice := strings.SplitN(DownloadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", DownloadAddr)
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
	curArchs = parseArchs(Archs)

	if err := generateInstallScript(); err != nil {
		return err
	}

	if err := generateMetaInfo(); err != nil {
		return err
	}

	if err := generatePipelineDoc(); err != nil {
		return err
	}

	if err := generatePipelineScripts(); err != nil {
		return err
	}

	basics := map[string]string{
		"version":                 path.Join(PubDir, ReleaseType, "version"),
		"datakit.yaml":            "datakit.yaml",
		"install.sh":              "install.sh",
		"install.ps1":             "install.ps1",
		"measurements-meta.json":  "measurements-meta.json",
		"pipeline-docs.json":      "pipeline-docs.json",
		"internal-pipelines.json": "internal-pipelines.json",
		fmt.Sprintf("datakit-%s.yaml", ReleaseVersion): "datakit.yaml",
		fmt.Sprintf("install-%s.sh", ReleaseVersion):   "install.sh",
		fmt.Sprintf("install-%s.ps1", ReleaseVersion):  "install.ps1",
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}
		goos, goarch := parts[0], parts[1]

		gzName, gzPath := tarFiles(PubDir, BuildDir, AppName, parts[0], parts[1])
		// gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppName, goos+"-"+goarch, ReleaseVersion)
		basics[gzName] = gzPath

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
			gz, gzP := tarFiles(PubDir, buildPath, appName, parts[0], parts[1])
			basics[gz] = gzP
		}

		installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
		installerExeWithVer := fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
		if parts[0] == datakit.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
			installerExeWithVer = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
		}

		basics[gzName] = gzPath
		basics[installerExe] = path.Join(PubDir, ReleaseType, installerExe)
		basics[installerExeWithVer] = path.Join(PubDir, ReleaseType, installerExe)
	}

	// Darwin release not under CI, so disable upload `version' file under darwin,
	// only upload darwin related files.
	if Archs == datakit.OSArchDarwinAmd64 && runtime.GOOS == datakit.OSDarwin {
		delete(basics, "version")
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

	ossSlice := strings.SplitN(DownloadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", DownloadAddr)
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
	curTmpArchs := parseArchs(Archs)

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
			gz, gzP := tarFiles(PubDir, buildPath, appName, parts[0], parts[1])
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
