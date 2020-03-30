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
	"runtime"
	"strings"
	"text/template"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	flagParallel = flag.Int("parallel", runtime.NumCPU(), "number of commands to run in parallel")
	flagBinary   = flag.String("binary", "", "binary name to build")
	flagName     = flag.String("name", *flagBinary, "same as -binary")
	flagBuildDir = flag.String("build-dir", "build", "output of build files")
	flagMain     = flag.String(`main`, `main.go`, `binary build entry`)
	flagCGO      = flag.Bool(`cgo`, true, `enable CGO or not`)
	flagTargetOS = flag.String(`os`, `linux`, `linux/mac/windows`)

	flagKodoHost     = flag.String("kodo-host", "", "")
	flagDownloadAddr = flag.String("download-addr", "", "")
	flagSsl          = flag.Int("ssl", 0, "")
	flagPort         = flag.Int("port", 0, "")
	flagPubDir       = flag.String("pub-dir", "pub", "")
	flagCsHost       = flag.String("cs-host", "corestone host", "")

	flagArchs    = flag.String("archs", "linux/amd64", "os archs")
	flagArchAll  = flag.Bool("all-arch", false, "build for all OS")
	flagShowArch = flag.Bool(`show-arch`, false, `show all OS`)

	flagRelease = flag.String(`release`, ``, `build for local/test/alpha/preprod/release`)

	flagPub = flag.Bool(`pub`, false, `publish binaries to OSS: local/test/alpha/release/preprod`)

	workDir string
	homeDir string

	curVersion []byte

	osarches = []string{
		"linux/386",
		"linux/amd64",

		"windows/386",
		"windows/amd64",
		"darwin/386",
		"darwin/amd64",

		"linux/arm",
		"linux/arm64",
		"freebsd/386",
		"freebsd/amd64",
		"freebsd/arm",
		"netbsd/386",
		"netbsd/amd64",
		"netbsd/arm",
		"openbsd/386",
		"openbsd/amd64",
		"plan9/386",
		"plan9/amd64",
		"solaris/amd64",
		"linux/mips",
		"linux/mipsle",
	}
)

type versionDesc struct {
	Version   string `json:"version"`
	Date      string `json:"date"`
	ChangeLog string `json:"changeLog"` // TODO: add release note
}

func init() {

	var err error
	workDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("%v", err)
	}

	workDir, err = filepath.Abs(workDir)
	if err != nil {
		log.Fatalf("%v", err)
	}
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
		log.Fatalf("failed to run %v: %v", args, err)
	}
}

func run(args ...string) {
	runEnv(args, nil)
}

func compileArch(bin, goos, goarch, dir string) {
	// log.Printf("building %s.%s/%s(%s)...", bin, goos, goarch, *flagMain)

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
	}

	if *flagCGO {
		env = append(env, "CGO_ENABLED=1")
	} else {
		env = append(env, "CGO_ENABLED=0")
	}

	runEnv(args, env)
}

func compile() {
	start := time.Now()

	compileTask := func(bin, goos, goarch, dir string) {
		compileArch(bin, goos, goarch, dir)
	}

	var archs []string

	if *flagArchAll {
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

		tarFiles(fmt.Sprintf("%s-%s", goos, goarch))

	}

	log.Printf("build elapsed %v", time.Since(start))
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

func getPudirByRelease() string {
	prefix := path.Join(*flagPubDir, *flagRelease)
	switch *flagTargetOS {
	case "windows":
		prefix += "_win"
	case "mac":
		prefix += "_mac"
	}

	return prefix

}

func publishAgent() {
	var ak, sk, bucket, ossHost string
	objPath := *flagName

	// 在你本地设置好这些 oss-key 环境变量
	switch *flagRelease {
	case `test`, `local`, `release`, `preprod`, `alpha`:
		tag := "DK_" + strings.ToUpper(*flagRelease)
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

	switch *flagTargetOS {
	case "windows":
		versionFile = "version_win"
		installObj = "install.exe"
	case "mac":
		versionFile = "version_mac"
		installObj = "install-mac.sh"
	}

	// 请求线上版本信息
	url := fmt.Sprintf("http://%s.%s/%s/%s", bucket, ossHost, *flagName, versionFile)
	curVd := getCurrentVersionInfo(url)

	if curVd != nil {
		if curVd.Version == git.Version {
			log.Printf("[warn] Current verison is the newest (%s <=> %s). Exit now.", curVd.Version, git.Version)
			os.Exit(0)
		}

		installObjOld := path.Join(objPath, fmt.Sprintf("install-%s.sh", curVd.Version))
		switch *flagTargetOS {
		case "windows":
			installObjOld = path.Join(objPath, fmt.Sprintf("install-%s.exe", curVd.Version))
		case "mac":
			installObjOld = path.Join(objPath, fmt.Sprintf("install-mac-%s.sh", curVd.Version))
		}

		oc.Move(installObj, installObjOld)
	}

	pubdir := getPudirByRelease()

	// upload all build archs
	archs := []string{}
	switch *flagArchs {
	case "all":
		archs = osarches
	default:
		archs = strings.Split(*flagArchs, ",")
	}

	objs := map[string]string{}

	switch *flagTargetOS {
	case "windows":
		objs[path.Join(pubdir, `version`)] = path.Join(objPath, `version_win`)
		objs[path.Join(pubdir, `install.exe`)] = path.Join(objPath, `install.exe`)
	case "mac":
		objs[path.Join(pubdir, `version`)] = path.Join(objPath, `version_mac`)
		objs[path.Join(pubdir, `install.sh`)] = path.Join(objPath, `install-mac.sh`)
	default:
		objs[path.Join(pubdir, `version`)] = path.Join(objPath, `version`)
		objs[path.Join(pubdir, `install.sh`)] = path.Join(objPath, `install.sh`)
	}

	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			log.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, goos+"-"+goarch, string(curVersion))

		objs[path.Join(pubdir, gzName)] = path.Join(objPath, gzName)
	}

	for k, v := range objs {
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

	if *flagShowArch {
		fmt.Printf("available archs:\n\t%s\n", strings.Join(osarches, "\n\t"))
		return
	}

	// 获取当前版本信息, 形如: v3.0.0-42-g3ed424a
	curVersion, err = exec.Command("git", []string{`describe`, `--always`, `--tags`}...).Output()
	if err != nil {
		log.Fatal(err)
	}

	curVersion = bytes.TrimSpace(curVersion)

	if *flagPub {
		publishAgent()
		return
	}

	if *flagBinary == "" {
		log.Fatal("-binary required")
	}

	if *flagTargetOS != "windows" && *flagTargetOS != "linux" && *flagTargetOS != "mac" {
		log.Fatal("invalid target os")
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

	lastNCommits, err := exec.Command("git", []string{`log`, `-n`, `8`}...).Output()
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
		bytes.TrimSpace(curVersion),
		bytes.TrimSpace(golang),
	)

	// create git/git.go
	ioutil.WriteFile(`git/git.go`, []byte(buildInfo), 0666)

	// create version info
	vd := &versionDesc{
		Version:   string(bytes.TrimSpace(curVersion)),
		Date:      string(bytes.TrimSpace(dateStr)),
		ChangeLog: string(bytes.TrimSpace(lastNCommits)),
	}

	outdir := getPudirByRelease()

	versionInfo, _ := json.Marshal(vd)
	ioutil.WriteFile(path.Join(outdir, `version`), versionInfo, 0666)

	switch *flagTargetOS {
	case "windows":
		archs := []string{}
		switch *flagArchs {
		case "all":
			archs = osarches
		default:
			archs = strings.Split(*flagArchs, ",")
		}

		for _, arch := range archs {
			parts := strings.Split(arch, "/")
			if len(parts) != 2 {
				log.Fatalf("invalid arch %q", parts)
			}

			buildWindowsInstall(outdir, parts[1])
		}
	default:
		// create install.sh script
		type Install struct {
			Name         string
			DownloadAddr string
			Version      string
		}

		install := &Install{
			Name:         *flagName,
			DownloadAddr: *flagDownloadAddr,
			Version:      string(curVersion),
		}

		txt, err := ioutil.ReadFile("install.template.sh")
		if err != nil {
			log.Fatal(err)
		}

		t := template.New("")
		t, err = t.Parse(string(txt))
		if err != nil {
			log.Fatal(err)
		}

		var byts bytes.Buffer

		err = t.Execute(&byts, install)
		if err != nil {
			log.Fatal(err)
		}

		results := bytes.Replace(byts.Bytes(), []byte{'\r', '\n'}, []byte{'\n'}, -1)
		if err = ioutil.WriteFile(path.Join(outdir, `install.sh`), results, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	os.RemoveAll(*flagBuildDir)
	_ = os.MkdirAll(*flagBuildDir, os.ModePerm)
	compile()
}

func buildWindowsInstall(outdir, goarch string) {

	output := path.Join(outdir, `install.exe`)

	//gzName := fmt.Sprintf("%s-%s.tar.gz", *flagName, string(curVersion))

	//downloadUrl := *flagDownloadAddr + "/" + gzName

	//log.Printf("downloadUrl=%s", downloadUrl)

	args := []string{
		"go", "build",
		"-ldflags", fmt.Sprintf(`-s -w -X main.serviceName=%s`, *flagName),
		"-o", output,
		"install.go",
	}

	env := []string{
		"GOOS=windows",
		"GOARCH=" + goarch,
	}

	runEnv(args, env)
}

func tarFiles(osarch string) {

	pubdir := getPudirByRelease()

	args := []string{
		`czf`,
		path.Join(pubdir, fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, osarch, string(curVersion))),
		`agent`,
		`-C`,
		path.Join(*flagBuildDir, fmt.Sprintf("%s-%s", *flagName, osarch)),
		`.`,
	}

	agentBinaryDir := "agent-binary/linux"
	agentName := "agent"
	switch *flagTargetOS {
	case "windows":
		agentBinaryDir = "agent-binary/windows"
		agentName = "agent.exe"
	case "mac":
		agentBinaryDir = "agent-binary/mac"
	}

	args = []string{
		`czf`,
		path.Join(pubdir, fmt.Sprintf("%s-%s-%s.tar.gz", *flagName, osarch, string(curVersion))),
		`-C`,
		agentBinaryDir,
		agentName,
		`-C`,
		fmt.Sprintf(`../../%s/%s-%s`, *flagBuildDir, *flagName, osarch),
		`.`,
	}

	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
