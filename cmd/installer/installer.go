package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/influxdata/toml"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	ServiceName    = `datakit`
	DataKitGzipUrl = ""
	installDir     = ""

	osarch           = runtime.GOOS + "/" + runtime.GOARCH
	dkservice        service.Service
	lagacyInstallDir = ""

	flagUpgrade         = flag.Bool("upgrade", false, ``)
	flagDataway         = flag.String("dataway", "", `address of dataway(ip:port), port default 9528`)
	flagVersion         = flag.Bool("version", false, "show installer version info")
	flagDownloadOnly    = flag.Bool("download-only", false, `download datakit only, not install`)
	flagDataKitGzipFile = flag.String("datakit-gzip", ``, `local path of datakit install files`)
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	config.InitDirs()

	applyFlags()

	// create install dir if not exists
	if err := os.MkdirAll(installDir, 0775); err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	datakitExe := filepath.Join(installDir, "datakit")
	if runtime.GOOS == "windows" {
		datakitExe += ".exe"
	}

	var err error
	prog := &program{}
	dkservice, err = service.New(prog, &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: `Collects data and upload it to DataFlux.`,
		Executable:  datakitExe,
		Arguments:   nil, // no args need here
	})

	if err != nil {
		log.Fatalf("E! new %s service failed: %s", runtime.GOOS, err.Error())
	}

	if *flagUpgrade { // upgrade new version

		migrateLagacyDatakit()

		if *flagDataKitGzipFile != "" {
			extractDatakit(*flagDataKitGzipFile, installDir)
		} else {
			downloadDatakit(DataKitGzipUrl, installDir)
		}

	} else { // install new datakit

		stopDataKitService(dkservice) // stop service if installed before

		if *flagDataKitGzipFile != "" {
			extractDatakit(*flagDataKitGzipFile, installDir)
		} else {
			downloadDatakit(DataKitGzipUrl, installDir)
		}

		var dwcfg *config.DataWayCfg
		if *flagDataway == "" {
			for {
				dw := readInput("Please set DataWay request URL(http://IP:Port/v1/write/metrics) > ")
				dwcfg, err = parseDataway(dw)
				if err == nil {
					break
				}

				fmt.Printf("%s\n", err.Error())
				continue
			}
		} else {
			dwcfg, err = parseDataway(*flagDataway)

			if err != nil {
				log.Fatalf("E! %s", err.Error())
			}
		}

		uninstallDataKitService(dkservice) // uninstall service if installed before

		if err := config.InitCfg(dwcfg); err != nil {
			log.Fatalf("E! failed to init datakit main config: %s", err.Error())
		}
	}

	log.Printf("I! install service %s...", ServiceName)
	if err := installDatakitService(dkservice); err != nil {
		log.Printf("I! fail to register service %s: %s, ignored", ServiceName, err.Error())
	}

	log.Printf("I! starting service %s...", ServiceName)
	if err := startDatakitService(dkservice); err != nil {
		log.Fatalf("E! fail to star service %s: %s", ServiceName, err.Error())
	}

	if *flagUpgrade { // upgrade new version
		log.Println("I! :) Upgrade Success!")
	} else {
		log.Println("I! :) Install Success!")
	}
}

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
   DataKitGzip: %s
`, git.Version, git.BuildAt, git.Golang, DataKitGzipUrl)
		os.Exit(0)
	}

	if *flagDownloadOnly {
		downloadDatakit(DataKitGzipUrl, "datakit.tar.gz")
		os.Exit(0)
	}

	switch osarch {

	case "windows/amd64":
		installDir = `C:\Program Files\DataFlux\` + ServiceName

	case "windows/386":
		installDir = `C:\Program Files (x86)\DataFlux\` + ServiceName

	case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386":
		installDir = `/usr/local/cloudcare/DataFlux/` + ServiceName

	default:
		// TODO
	}
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	txt, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	return strings.TrimSpace(txt)
}

func doDownload(r io.Reader, to string) {

	f, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	if _, err := io.Copy(f, r); err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	f.Close()
}

func doExtract(r io.Reader, to string) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return
		case err != nil:
			log.Fatalf("E! %s", err.Error())
		case hdr == nil:
			continue
		}

		target := filepath.Join(to, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					log.Fatalf("E! %s", err.Error())
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				log.Fatalf("E! %s", err.Error())
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				log.Fatalf("E! %s", err.Error())
			}

			if _, err := io.Copy(f, tr); err != nil {
				log.Fatalf("E! %s", err.Error())
			}

			f.Close()
		}
	}
}

func extractDatakit(gz, to string) {
	data, err := os.Open(gz)
	if err != nil {
		log.Fatalf("E! open file %s failed: %s", gz, err)
	}

	defer data.Close()

	doExtract(data, to)
}

type writeCounter struct {
	total   uint64
	current uint64
	last    float64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	return n, nil
}

func downloadDatakit(from, to string) {
	resp, err := http.Get(from)
	if err != nil {
		log.Fatalf("E! failed to download %s: %s", from, err)
	}

	defer resp.Body.Close()
	cnt := &writeCounter{
		total: uint64(resp.ContentLength),
	}

	if *flagDownloadOnly {
		doDownload(io.TeeReader(resp.Body, cnt), to)
	} else {
		doExtract(io.TeeReader(resp.Body, cnt), to)
	}
	fmt.Printf("\n")
}

func (wc *writeCounter) PrintProgress() {
	if wc.last > float64(wc.total)*0.01 || wc.current == wc.total { // update progress-bar each 1%
		fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Printf("\rDownloading... %s/%s", humanize.Bytes(wc.current), humanize.Bytes(wc.total))
		wc.last = 0.0
	}
}

func parseDataway(dw string) (*config.DataWayCfg, error) {

	dwcfg := &config.DataWayCfg{}

	if u, err := url.Parse(dw); err == nil {
		dwcfg.Scheme = u.Scheme
		dwcfg.Token = u.Query().Get("token")
		dwcfg.Host = u.Host
		dwcfg.DefaultPath = u.Path

		if dwcfg.Scheme == "https" {
			dwcfg.Host = dwcfg.Host + ":443"
		}
	} else {
		log.Printf("E! invalid dataway %s: %s", dw, err.Error())
		return nil, err
	}

	log.Printf("D! Testing DataWay(%s)...", dwcfg.Host)
	conn, err := net.DialTimeout("tcp", dwcfg.Host, time.Second*5)
	if err != nil {
		log.Printf("E! connect dataway %s failed: %s", dwcfg.Host, err.Error())
		return nil, err
	}

	if err := conn.Close(); err != nil {
		log.Printf("E! close dataway %s failed: %s, ignored", dwcfg.Host, err.Error())
	}

	return dwcfg, nil
}

type program struct{}

func (p *program) Start(s service.Service) error { go p.run(s); return nil }
func (p *program) run(s service.Service)         {}
func (p *program) Stop(s service.Service) error  { return nil }

func stopDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		log.Printf("I! stop service datakit failed: %s, ignored", err.Error())
	}

	return nil
}

func uninstallDataKitService(s service.Service) error {
	if err := service.Control(s, "uninstall"); err != nil {
		log.Printf("W! Stop service datakit failed: %s, ignored", err.Error())
	}

	return nil
}

func installDatakitService(s service.Service) error {
	return service.Control(s, "install")
}

func startDatakitService(s service.Service) error {
	return service.Control(s, "start")
}

func stopLagacyDatakit() {
	switch osarch {
	case "windows/amd64", "windows/386":
		stopDataKitService(dkservice)
	default:
		cmd := exec.Command(`stop`, []string{ServiceName}...)
		if _, err := cmd.Output(); err != nil {
			log.Printf("D! upstart stop datakit failed, try systemctl...")
		} else {
			return
		}

		cmd = exec.Command("systemctl", []string{"stop", ServiceName}...)
		if _, err := cmd.Output(); err != nil {
			log.Printf("D! systemctl stop datakit failed, ignored")
		}
	}
}

func updateLagacyConfig(dir string) {
	cfgdata, err := ioutil.ReadFile(filepath.Join(dir, "datakit.conf"))
	if err != nil {
		log.Fatalf("E! read lagacy datakit.conf failed: %s", err.Error())
	}

	var maincfg config.MainConfig
	if err := toml.Unmarshal(cfgdata, &maincfg); err != nil {
		log.Fatalf("E! TOML unmarshal failed: %s", err.Error())
	}

	maincfg.Log = filepath.Join(installDir, "datakit.log") // reset log path
	maincfg.ConfigDir = ""                                 // remove conf.d config: we use static conf.d dir, *not* configurable

	// split orgin ftdataway into dataway object
	dwcfg, err := parseDataway(maincfg.FtGateway)
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	maincfg.FtGateway = "" // deprecated
	maincfg.DataWay = dwcfg

	fd, err := os.OpenFile(filepath.Join(dir, "datakit.conf"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}

	defer fd.Close()

	tmp := template.New("")
	tmp, err = tmp.Parse(config.MainConfigTemplate)
	if err != nil {
		log.Fatalf("E! %s", err.Error())
	}
	if err := tmp.Execute(fd, maincfg); err != nil {
		log.Fatalf("E! %s", err.Error())
	}
}

func migrateLagacyDatakit() {

	var lagacyServiceFiles []string = nil

	switch osarch {

	case "windows/amd64", "windows/386":
		lagacyInstallDir = `C:\Program Files\Forethought\` + ServiceName
		if _, err := os.Stat(lagacyInstallDir); err != nil {
			lagacyInstallDir = `C:\Program Files (x86)\Forethought\` + ServiceName
		}

	case "linux/amd64", "linux/386",
		"linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386":
		lagacyInstallDir = `/usr/local/cloudcare/forethought/` + ServiceName
		lagacyServiceFiles = []string{"/lib/systemd/system/datakit.service", "/etc/systemd/system/datakit.service"}
	default:
		log.Fatalf("E! %s not support", osarch)
	}

	if _, err := os.Stat(lagacyInstallDir); err != nil {
		log.Printf("D! no lagacy datakit installed")
		return
	}

	stopLagacyDatakit()
	updateLagacyConfig(lagacyInstallDir)

	// uninstall service, remove old datakit.service file(for UNIX OS)
	uninstallDataKitService(dkservice)
	for _, sf := range lagacyServiceFiles {
		if _, err := os.Stat(sf); err == nil {
			if err := os.Remove(sf); err != nil {
				log.Fatalf("E! remove %s failed: %s", sf, err.Error())
			}
		}
	}

	os.RemoveAll(installDir) // clean new install dir if exists

	// move all lagacy datakit files to new install dir
	if err := os.Rename(lagacyInstallDir, installDir); err != nil {
		log.Fatalf("E! remove %s failed: %s", installDir, err.Error())
	}

	for _, dir := range []string{config.TelegrafDir, config.DataDir, config.LuaDir, config.ConfdDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Fatalf("E! create %s failed: %s", dir, err)
		}
	}
}
