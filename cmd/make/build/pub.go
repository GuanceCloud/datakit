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

func tarFiles(goos, goarch string) {
	gz := filepath.Join(PubDir, ReleaseType, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
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
	// Encode Markdown docs with base64.
	for _, plDoc := range funcs.PipelineFunctionDocs {
		plDoc.Doc = encoding.EncodeToString([]byte(plDoc.Doc))
	}
	exportPLDocs := struct {
		Version   string                  `json:"version"`
		Docs      string                  `json:"docs"`
		Functions map[string]*funcs.PLDoc `json:"functions"`
	}{
		Version:   git.Version,
		Docs:      "经过 base64 编码的 pipeline 函数文档，包括各函数原型、说明、使用示例",
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
	scriptMap, err := config.GetScriptMap()
	if err != nil {
		return err
	}
	encoding := base64.StdEncoding
	// Encode Markdown docs with base64.
	for name, script := range scriptMap {
		scriptMap[name] = encoding.EncodeToString([]byte(script))
	}
	data, err := json.Marshal(scriptMap)
	if err != nil {
		return err
	}
	f, err := os.Create("internal-pipelines.json")
	if err != nil {
		return err
	}
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
		if arch == datakit.OSArchDarwinAmd64 && runtime.GOOS != datakit.OSDarwin {
			l.Warn("Not a darwin system, skip the upload of related files.")
			continue
		}

		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
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

		basics[gzName] = path.Join(PubDir, ReleaseType, gzName)
		basics[installerExe] = path.Join(PubDir, ReleaseType, installerExe)
		basics[installerExeWithVer] = path.Join(PubDir, ReleaseType, installerExe)
	}

	// Darwin release not under CI, so disable upload `version' file under darwin,
	// only upload darwin related files.
	if Archs == datakit.OSArchDarwinAmd64 {
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
