//+build ignore

package main

import (
	"encoding/json"
	"flag"
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
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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
	flagRelease      = flag.String(`release`, ``, `build for local/test/preprod/release`)
	flagPub          = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/release/preprod`)
	flagPubAgent     = flag.Bool("pub-agent", false, `publish telegraf`)

	installerExe string

	l *logger.Logger

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

	OSArches = []string{
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

func runEnv(args, env []string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	return cmd.CombinedOutput()
}

func compileArch(bin, goos, goarch, dir string) {

	output := path.Join(dir, bin)
	if goos == "windows" {
		output += ".exe"
	}

	args := []string{
		"go", "build",
		"-o", output,
		"-ldflags",
		"-w -s",
		*flagMain,
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		`GO111MODULE=off`,
		"CGO_ENABLED=0",
	}

	l.Debugf("building %s, envs: %v", fmt.Sprintf("%s-%s/%s", goos, goarch, bin), env)
	msg, err := runEnv(args, env)
	if err != nil {
		l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
	}
}

func compile() {
	start := time.Now()

	compileTask := func(bin, goos, goarch, dir string) {
		compileArch(bin, goos, goarch, dir)
	}

	var archs []string

	if *flagArchs == "all" {
		archs = OSArches
	} else {
		archs = strings.Split(*flagArchs, "|")
	}

	for _, arch := range archs {

		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}

		goos, goarch := parts[0], parts[1]

		dir := fmt.Sprintf("%s/%s-%s-%s", *flagBuildDir, *flagName, goos, goarch)

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			l.Fatalf("failed to mkdir: %v", err)
		}

		dir, err = filepath.Abs(dir)
		if err != nil {
			l.Fatal(err)
		}

		compileTask(*flagBinary, goos, goarch, dir)
		buildExternals(dir, goos, goarch)

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
		} else {
			installerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		buildInstaller(filepath.Join(*flagPubDir, *flagRelease), goos, goarch)
	}

	l.Infof("Done!(elapsed %v)", time.Since(start))
}

type installInfo struct {
	Name         string
	DownloadAddr string
	Version      string
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

func releaseAgent() {
	start := time.Now()
	var ak, sk, bucket, ossHost string
	objPath := *flagName

	// 在你本地设置好这些 oss-key 环境变量
	switch *flagRelease {
	case `test`, `local`, `release`, `preprod`:
		tag := strings.ToUpper(*flagRelease)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		l.Fatalf("unknown release type: %s", *flagRelease)
	}

	if ak == "" || sk == "" {
		l.Fatalf("oss access key or secret key missing, tag=%s", strings.ToUpper(*flagRelease))
	}

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    objPath,
	}

	if err := oc.Init(); err != nil {
		l.Fatal(err)
	}

	versionFile := `version`

	// 请求线上版本信息
	url := fmt.Sprintf("http://%s.%s/%s/%s", bucket, ossHost, *flagName, versionFile)
	curVd := getCurrentVersionInfo(url)

	// upload all build archs
	archs := []string{}
	switch *flagArchs {
	case "all":
		archs = OSArches
	default:
		archs = strings.Split(*flagArchs, "|")
	}

	ossfiles := map[string]string{
		path.Join(*flagPubDir, *flagRelease, "version"): path.Join(objPath, "version"),
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
		l.Debugf("upload %s -> %s ...", k, v)
		if err := oc.Upload(k, v); err != nil {
			l.Fatal(err)
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}

func tarFiles(goos, goarch string) {

	gz := path.Join(*flagPubDir, *flagRelease, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		*flagName, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
		`-C`,
		// the whole *flagBuildDir/datakit-<goos>-<goarch> dir
		path.Join(*flagBuildDir, fmt.Sprintf("%s-%s-%s", *flagName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		l.Fatal(err)
	}
	l.Debugf("tar %s ok", gz)
}

type dkexternal struct {
	name string

	lang string // go/others

	entry     string
	buildArgs []string

	osarchs map[string]bool
	envs    []string

	buildCmd string
}

var (
	externals = []*dkexternal{
		&dkexternal{
			// requirement: apt-get install gcc-multilib
			name: "oraclemonitor",
			lang: "go",

			entry: "main.go",
			osarchs: map[string]bool{
				"linux/amd64": true,
				// `darwin/amd64`: true,
			},

			buildArgs: nil,
			envs: []string{
				"CGO_ENABLED=1",
			},
		},

		&dkexternal{
			name: "tcpdump",
			lang: "go",

			entry: "main.go",
			osarchs: map[string]bool{
				"linux/amd64": true,
				// `darwin/amd64`: true,
			},

			buildArgs: nil,
			envs: []string{
				"CGO_ENABLED=1",
			},
		},

		&dkexternal{
			name: "csv",
			osarchs: map[string]bool{
				`linux/386`:     true,
				`linux/amd64`:   true,
				`linux/arm`:     true,
				`linux/arm64`:   true,
				`darwin/amd64`:  true,
				`windows/amd64`: true,
				`windows/386`:   true,
			},
			buildArgs: []string{"plugins/externals/csv/build.sh"},
			buildCmd:  "bash",
		},
		&dkexternal{
			name: "ansible",
			osarchs: map[string]bool{
				`linux/386`:     true,
				`linux/amd64`:   true,
				`linux/arm`:     true,
				`linux/arm64`:   true,
				`darwin/amd64`:  true,
				`windows/amd64`: true,
				`windows/386`:   true,
			},
			buildArgs: []string{"plugins/externals/ansible/build.sh"},
			buildCmd:  "bash",
		},

		// &dkexternal{
		// 	// requirement: apt-get install gcc-multilib
		// 	name: "skywalkingGrpcV3",
		// 	lang: "go",

		// 	entry: "main.go",
		// 	osarchs: map[string]bool{
		// 		// `linux/386`:   true,
		// 		`linux/amd64`: true,
		// 		//`linux/arm`:     true,
		// 		//`linux/arm64`:   true,
		// 		//`darwin/amd64`:  true,
		// 		`windows/amd64`: true,
		// 		// `windows/386`:   true,
		// 	},

		// 	buildArgs: nil,
		// 	envs: []string{
		// 		"CGO_ENABLED=1",
		// 	},
		// },

		// others...
	}
)

func buildExternals(outdir, goos, goarch string) {
	curOSArch := runtime.GOOS + "/" + runtime.GOARCH

	for _, ex := range externals {
		l.Debugf("building %s-%s/%s", goos, goarch, ex.name)

		if _, ok := ex.osarchs[curOSArch]; !ok {
			l.Warnf("skip build %s under %s", ex.name, curOSArch)
			continue
		}

		osarch := goos + "/" + goarch
		if _, ok := ex.osarchs[osarch]; !ok {
			l.Warnf("skip build %s under %s", ex.name, osarch)
			continue
		}

		out := ex.name

		switch strings.ToLower(ex.lang) {
		case "go", "golang":

			switch osarch {
			case "windows/amd64", "windows/386":
				out = out + ".exe"
			default: // pass
			}

			args := []string{
				"go", "build",
				"-o", filepath.Join(outdir, "externals", out),
				"-ldflags",
				"-w -s",
				filepath.Join("plugins/externals", ex.name, ex.entry),
			}

			env := append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)

			msg, err := runEnv(args, env)
			if err != nil {
				l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
			}

		default: // for python, just copy source code into build dir
			args := append(ex.buildArgs, filepath.Join(outdir, "externals"))
			cmd := exec.Command(ex.buildCmd, args...)
			if ex.envs != nil {
				cmd.Env = append(os.Environ(), ex.envs...)
			}

			res, err := cmd.CombinedOutput()
			if err != nil {
				l.Fatalf("failed to build python(%s %s): %s, err: %s", ex.buildCmd, strings.Join(args, " "), res, err.Error())
			}
		}
	}
}

func buildInstaller(outdir, goos, goarch string) {

	l.Debugf("building %s-%s/installer...", goos, goarch)

	args := []string{
		"go", "build",
		"-o", filepath.Join(outdir, installerExe),
		"-ldflags",
		fmt.Sprintf("-w -s -X main.DataKitBaseURL=%s -X main.DataKitVersion=%s", *flagDownloadAddr, git.Version),
		"cmd/installer/installer.go",
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
	}

	msg, err := runEnv(args, env)
	if err != nil {
		l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
	}
}

func pubAgent() {

	var archs []string
	if *flagArchs == "all" {
		archs = OSArches
	} else {
		archs = strings.Split(*flagArchs, "|")
	}

	start := time.Now()
	var ak, sk, bucket, ossHost string
	objPath := "datakit/telegraf"

	// 在你本地设置好这些 oss-key 环境变量
	switch *flagRelease {
	case `test`, `local`, `release`, `preprod`:
		tag := strings.ToUpper(*flagRelease)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		l.Fatalf("unknown release type: %s", *flagRelease)
	}

	if ak == "" || sk == "" {
		l.Fatalf("oss access key or secret key missing, tag=%s", strings.ToUpper(*flagRelease))
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
			path.Join(*flagPubDir, gz),
			path.Join(*flagPubDir, relPath)}...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			l.Fatal(err)
		}

		ossfiles[path.Join(*flagPubDir, gz)] = path.Join(objPath, gz)
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

func main() {

	var err error

	logger.SetGlobalRootLogger("",
		logger.DEBUG,
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER|logger.OPT_COLOR)

	l = logger.SLogger("make")

	flag.Parse()

	if *flagPubAgent {
		pubAgent()
		return
	}

	if *flagPub {
		releaseAgent()
		return
	}

	// create version info
	vd := &versionDesc{
		Version:  strings.TrimSpace(git.Version),
		Date:     git.BuildAt,
		Uploader: git.Uploader,
		Branch:   git.Branch,
		Commit:   git.Commit,
		Go:       git.Golang,
	}

	versionInfo, err := json.MarshalIndent(vd, "", "    ")
	if err != nil {
		l.Fatal(err)
	}

	if err := ioutil.WriteFile(path.Join(*flagPubDir, *flagRelease, "version"), versionInfo, 0666); err != nil {
		l.Fatal(err)
	}

	os.RemoveAll(*flagBuildDir)
	_ = os.MkdirAll(*flagBuildDir, os.ModePerm)
	compile()
}
