// +build ignore

package main

import (
	"bytes"
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
	"text/template"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	flagBinary   = flag.String("binary", "", "binary name to build")
	flagName     = flag.String("name", *flagBinary, "same as -binary")
	flagBuildDir = flag.String("build-dir", "build", "output of build files")
	flagMain     = flag.String(`main`, `main.go`, `binary build entry`)
	flagCGO      = flag.Int(`cgo`, 0, `enable CGO or not`)

	flagDownloadAddr = flag.String("download-addr", "", "")
	flagPubDir       = flag.String("pub-dir", "pub", "")

	flagArchs = flag.String("archs", "linux/amd64", "os archs")

	flagRelease = flag.String(`release`, ``, `build for local/test/alpha/preprod/release`)

	flagPub = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/alpha/release/preprod`)

	curVersion string
	osarches   = []string{
		`windows/amd64`,
		`freebsd/amd64`,

		//`android/amd64`,
		//`android/arm64`,
		//`darwin/arm64`,
		`dragonfly/amd64`,
		`illumos/amd64`,
		//`js/wasm`,
		`linux/amd64`,
		`linux/arm64`,
		`linux/mips64`,
		`linux/mips64le`,
		//`linux/mipsle`,
		`linux/ppc64`,
		`linux/ppc64le`,
		//`linux/s390x`,
		//`nacl/amd64p32`,
		//`nacl/arm`,
		`netbsd/amd64`,
		`netbsd/arm64`,
		`openbsd/amd64`,
		//`openbsd/arm64`,
		//`plan9/amd64`,
		`solaris/amd64`,
		//`windows/arm`,

		`aix/ppc64`,
		`darwin/amd64`,
	}

	winInstallerExe = ""
)

type versionDesc struct {
	Version   string `json:"version"`
	Date      string `json:"date"`
	ChangeLog string `json:"changeLog"` // TODO: add release note
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

	// log.Printf("%s %s", strings.Join(env, " "), strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
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
	}

	//if goos == "windows" {
	//	env = append(env, []string{"CXX=g++-mingw-w64-x86-64", "CC=x86_64-w64-mingw32-gcc"}...)
	//}

	if *flagCGO == 1 && goos != "aix" {
		env = append(env, "CGO_ENABLED=1")
	} else {
		env = append(env, "CGO_ENABLED=0")
	}

	runEnv(args, env)
}

func compile() {
	start := time.Now()

	compileTask := func(bin, goos, goarch, dir string) {
		log.Printf("[debug] building %s/%s...", goos, goarch)
		compileArch(bin, goos, goarch, dir)
	}

	var archs []string

	if *flagArchs == "all" {
		archs = osarches
	} else {
		archs = strings.Split(*flagArchs, ",")
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

		if goos == "windows" { // build windows installer.exe
			winInstallerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, curVersion)
			buildWindowsInstall(path.Join(*flagPubDir, *flagRelease), goarch)
		}

		// generate install scripts & installer to pub-dir
		buildInstallScript(path.Join(*flagPubDir, *flagRelease), goos, goarch)
	}

	log.Printf("build elapsed %v", time.Since(start))
}

type installInfo struct {
	Name         string
	DownloadAddr string
	Version      string
}

func buildInstallScript(dir, goos, goarch string) {
	i := &installInfo{
		Name:         *flagName,
		DownloadAddr: *flagDownloadAddr,
		Version:      curVersion,
	}

	templateFile := "install.template.sh"
	installScript := "install.sh"

	if goos == "windows" {
		templateFile = "install.template.ps1"
		installScript = "install.ps1"
	}

	txt, err := ioutil.ReadFile(templateFile)
	if err != nil {
		log.Fatal(err)
	}

	t := template.New("")
	t, err = t.Parse(string(txt))
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, i)
	if err != nil {
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(path.Join(dir, installScript), buf.Bytes(), os.ModePerm); err != nil {
		log.Fatal(err)
	}
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
	installObj := "install.sh"

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

	objs := map[string]string{
		path.Join(*flagPubDir, *flagRelease, "version"): path.Join(objPath, "version"),
	}

	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			log.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, goos+"-"+goarch, curVersion)

		objs[path.Join(*flagPubDir, *flagRelease, gzName)] = path.Join(objPath, gzName)

		if goos == "windows" {
			objs[path.Join(*flagPubDir, *flagRelease, "install.ps1")] = path.Join(objPath, "install.ps1")
			winInstallerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, curVersion)
			objs[path.Join(*flagPubDir, *flagRelease, winInstallerExe)] = path.Join(objPath, winInstallerExe)
		} else {
			objs[path.Join(*flagPubDir, *flagRelease, "install.sh")] = path.Join(objPath, "install.sh")
		}

		if curVd != nil {
			if curVd.Version == git.Version {
				log.Printf("[warn] Current verison is the newest (%s <=> %s). Exit now.", curVd.Version, git.Version)
				os.Exit(0)
			}

			installObjOld := path.Join(objPath, fmt.Sprintf("install-%s.sh", curVd.withoutGitCommit()))
			if goos == "windows" {
				installObjOld = path.Join(objPath, fmt.Sprintf("install-%s.ps1", curVd.withoutGitCommit()))
				installObj = path.Join(objPath, "install.ps1")
			} else {
				installObj = path.Join(objPath, "install.sh")
			}

			// rename install script online, make it possible to install old version if required
			log.Printf("[debug] rename %s -> %s", installObj, installObjOld)
			if err := oc.Move(installObj, installObjOld); err != nil {
				log.Fatal(err)
			}
		}
	}

	for k, v := range objs {
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

	verinfo, err := exec.Command("git", []string{`describe`, `--always`, `--tags`}...).Output()
	if err != nil {
		log.Fatal(err)
	}
	curVersion = string(bytes.TrimSpace(verinfo))

	if *flagPub {
		releaseAgent()
		return
	}

	gitsha1, err := exec.Command("git", []string{`rev-parse`, `--short`, `HEAD`}...).Output()
	if err != nil {
		log.Fatal(err)
	}

	dateStr, err := exec.Command("date", []string{"+'%Y-%m-%d %H:%M:%S'"}...).Output()
	if err != nil {
		log.Fatal(err)
	}

	golang, err := exec.Command("go", []string{"version"}...).Output()
	if err != nil {
		log.Fatal(err)
	}

	buildInfo := fmt.Sprintf(`// THIS FILE IS GENERATED BY make.go, DO NOT EDIT IT.
package git
const (
	Sha1 string = "%s"
	BuildAt string = "%s"
	Version string = "%s"
	Golang string = "%s"
)`,
		bytes.TrimSpace(gitsha1),

		// 输出会带有 ' 字符, 剪掉之
		bytes.Replace(bytes.TrimSpace(dateStr), []byte("'"), []byte(""), -1),

		// 移除此处的 `v' 前缀.  前端的版本号判断机制容不下这个前缀
		strings.TrimSpace(curVersion),
		bytes.TrimSpace(golang),
	)

	// create git/git.go
	ioutil.WriteFile(`git/git.go`, []byte(buildInfo), 0666)

	// create version info
	vd := &versionDesc{
		Version: strings.TrimSpace(curVersion),
		Date:    string(bytes.TrimSpace(dateStr)),
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

func tarFiles(osname, arch string) {

	telegrafAgentName := "agent"
	if osname == "windows" {
		telegrafAgentName = "agent.exe"
	}

	args := []string{
		`czf`,
		path.Join(*flagPubDir, *flagRelease, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
			*flagName, osname, arch, string(curVersion))),
		path.Join("embed", telegrafAgentName),
		`-C`,
		// the whole build/datakit-<os>-<arch> dir
		path.Join(*flagBuildDir, fmt.Sprintf("%s-%s-%s", *flagName, osname, arch)), `.`,
	}

	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func buildWindowsInstall(outdir, goarch string) {

	args := []string{
		"go", "build",
		"-ldflags", "-w -s",
		"-o", path.Join(outdir, winInstallerExe),
		"win-installer.go",
	}

	env := []string{
		"GOOS=windows",
		"GOARCH=" + goarch,
	}

	runEnv(args, env)
}
