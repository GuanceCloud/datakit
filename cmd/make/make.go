//+build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	flagBinary       = flag.String("binary", "", "binary name to build")
	flagName         = flag.String("name", *flagBinary, "same as -binary")
	flagBuildDir     = flag.String("build-dir", "build", "output of build files")
	flagMain         = flag.String(`main`, `main.go`, `binary build entry`)
	flagDownloadAddr = flag.String("download-addr", "", "")
	flagPubDir       = flag.String("pub-dir", "pub", "")
	flagArchs        = flag.String("archs", "linux/amd64", "os archs")
	flagRelease      = flag.String(`release`, ``, `build for local/test/alpha/preprod/release`)
	flagPub          = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/alpha/release/preprod`)

	installerExe string

	/* Use:
			go tool dist list
		to get your current os/arch list

	aix/ppc64

	android/386
	android/amd64
	android/arm
	android/arm64

	darwin/386
	darwin/amd64
	darwin/arm
	darwin/arm64

	dragonfly/amd64

	freebsd/386
	freebsd/amd64
	freebsd/arm

	illumos/amd64

	js/wasm

	linux/386
	linux/amd64
	linux/arm
	linux/arm64
	linux/mips
	linux/mips64
	linux/mips64le
	linux/mipsle
	linux/ppc64
	linux/ppc64le
	linux/s390x

	nacl/386
	nacl/amd64p32
	nacl/arm

	netbsd/386
	netbsd/amd64
	netbsd/arm
	netbsd/arm64

	openbsd/386
	openbsd/amd64
	openbsd/arm
	openbsd/arm64

	plan9/386
	plan9/amd64
	plan9/arm

	solaris/amd64

	windows/386
	windows/amd64
	windows/arm */

	osarches = []string{
		//`freebsd/386`,
		//`freebsd/amd64`,

		`linux/386`,
		`linux/amd64`,
		`linux/arm`,
		`linux/arm64`,

		`darwin/amd64`,

		`windows/amd64`,
		`windows/386`,
	}
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date"`
	Uploader string `json:"uploader"`
}

func (vd *versionDesc) withoutGitCommit() string {
	parts := strings.Split(vd.Version, "-")
	if len(parts) != 3 {
		log.Fatalf("version info not in v<x.x>-<n>-g<commit-id> format: %s", vd.Version)
	}

	return strings.Join(parts[:2], "-")
}

func runEnv(args, env []string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	if err := cmd.Run(); err != nil {
		log.Printf("[error] failed to run %v, envs: %v: %v", args, env, err)
	}
}

func run(args ...string) {
	runEnv(args, nil)
}

func compileArch(bin, goos, goarch, dir string) {

	output := path.Join(dir, bin)
	if goos == "windows" {
		output += ".exe"
	}

	args := []string{
		"go", "build",
		"-o", output,
		"-ldflags", "-w -s",
		*flagMain,
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		`GO111MODULE=off`,
		"CGO_ENABLED=0",
	}

	log.Printf("[debug] building % 13s, envs: %v.", fmt.Sprintf("%s-%s", goos, goarch), env)
	runEnv(args, env)
}

func compile() {
	start := time.Now()

	compileTask := func(bin, goos, goarch, dir string) {
		compileArch(bin, goos, goarch, dir)
	}

	var archs []string

	if *flagArchs == "all" {
		archs = osarches
	} else {
		archs = strings.Split(*flagArchs, "|")
	}

	for _, arch := range archs {

		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			log.Fatalf("invalid arch %q", parts)
		}

		goos, goarch := parts[0], parts[1]

		dir := fmt.Sprintf("build/%s-%s-%s", *flagName, goos, goarch)

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("failed to mkdir: %v", err)
		}

		dir, err = filepath.Abs(dir)
		if err != nil {
			log.Fatal("[fatal] %v", err)
		}

		compileTask(*flagBinary, goos, goarch, dir)

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
		} else {
			installerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		buildInstaller(path.Join(*flagPubDir, *flagRelease), goos, goarch)
	}

	log.Printf("build elapsed %v", time.Since(start))
}

type installInfo struct {
	Name         string
	DownloadAddr string
	Version      string
}

func getCurrentVersionInfo(url string) *versionDesc {

	log.Printf("get current online version: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("[fatal] %s", err.Error())
	}

	if resp.StatusCode != 200 {
		log.Printf("[warn] get current online version failed, ignored")
		return nil
	}

	defer resp.Body.Close()
	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("current online version: %s", string(info))
	var vd versionDesc
	if err := json.Unmarshal(info, &vd); err != nil {
		log.Fatal(err)
	}
	return &vd
}

func releaseAgent() {
	var ak, sk, bucket, ossHost string
	objPath := *flagName

	// 在你本地设置好这些 oss-key 环境变量
	switch *flagRelease {
	case `test`, `local`, `release`:
		tag := strings.ToUpper(*flagRelease)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		log.Fatalf("unknown release type: %s", *flagRelease)
	}

	if ak == "" || sk == "" {
		log.Fatalf("[fatal] oss access key or secret key missing, tag=%s", strings.ToUpper(*flagRelease))
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
		log.Fatalf("[fatal] %s", err)
	}

	versionFile := `version`

	// 请求线上版本信息
	url := fmt.Sprintf("http://%s.%s/%s/%s", bucket, ossHost, *flagName, versionFile)
	curVd := getCurrentVersionInfo(url)

	// upload all build archs
	archs := []string{}
	switch *flagArchs {
	case "all":
		archs = osarches
	default:
		archs = strings.Split(*flagArchs, "|")
	}

	ossfiles := map[string]string{
		path.Join(*flagPubDir, *flagRelease, "version"): path.Join(objPath, "version"),
	}

	renameOssFiles := map[string]string{}
	var verId string

	if curVd != nil && curVd.Version == git.Version {
		log.Printf("[warn] Current verison is the newest (%s <=> %s). Exit now.", curVd.Version, git.Version)
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
			log.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, goos+"-"+goarch, git.Version)

		ossfiles[path.Join(*flagPubDir, *flagRelease, gzName)] = path.Join(objPath, gzName)

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)

			if curVd != nil {
				renameOssFiles[path.Join(objPath, installerExe)] =
					path.Join(objPath, fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, verId))
			}

		} else {
			installerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)

			if curVd != nil {
				renameOssFiles[path.Join(objPath, installerExe)] =
					path.Join(objPath, fmt.Sprintf("installer-%s-%s-%s", goos, goarch, verId))
			}
		}

		ossfiles[path.Join(*flagPubDir, *flagRelease, installerExe)] = path.Join(objPath, installerExe)
	}

	// backup old installer script online, make it possible to install old version if required
	for k, v := range renameOssFiles {
		if err := oc.Move(k, v); err != nil {
			log.Printf("[debug] backup %s -> %s failed: %s, ignored", k, v, err.Error())
			continue
		}

		log.Printf("[debug] backup %s -> %s ok", k, v)
	}

	// test if all file ok before uploading
	for k, _ := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			log.Fatalf("[error] %s", err.Error())
		}
	}

	for k, v := range ossfiles {
		log.Printf("[debug] upload %s -> %s ...", k, v)
		if err := oc.Upload(k, v); err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Done :)")
}

func main() {

	var err error

	flag.Parse()
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if *flagPub {
		releaseAgent()
		return
	}

	// create version info
	vd := &versionDesc{
		Version:  strings.TrimSpace(git.Version),
		Date:     git.BuildAt,
		Uploader: git.Uploader,
	}

	versionInfo, err := json.Marshal(vd)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(path.Join(*flagPubDir, *flagRelease, "version"), versionInfo, 0666); err != nil {
		log.Fatal(err)
	}

	os.RemoveAll(*flagBuildDir)
	_ = os.MkdirAll(*flagBuildDir, os.ModePerm)
	compile()
}

func tarFiles(goos, goarch string) {

	var telegrafPath string

	suffix := goos + "-" + goarch

	switch suffix {
	case `freebsd-386`, `freebsd-amd64`,
		`linux-386`, `linux-amd64`,
		`linux-arm`, `linux-arm64`,
		`darwin-amd64`:
		telegrafPath = path.Join("embed", suffix, "agent")

	case `windows-amd64`, `windows-386`:
		telegrafPath = path.Join("embed", suffix, "agent.exe")
	}

	gz := path.Join(*flagPubDir, *flagRelease, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		*flagName, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
		telegrafPath,
		`-C`,
		// the whole build/datakit-<goos>-<goarch> dir
		path.Join(*flagBuildDir, fmt.Sprintf("%s-%s-%s", *flagName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[debug] tar %s ok", gz)
}

func buildInstaller(outdir, goos, goarch string) {

	log.Printf("[debug] build %s/%s installer to %s...", goos, goarch, outdir)

	gzName := fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, goos+"-"+goarch, git.Version)

	args := []string{
		"go", "build",
		"-ldflags", fmt.Sprintf("-w -s -X main.DataKitGzipUrl=https://%s/%s", *flagDownloadAddr, gzName),
		"-o", path.Join(outdir, installerExe),
		"cmd/installer/installer.go",
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
	}

	runEnv(args, env)
}
