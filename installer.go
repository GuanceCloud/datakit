package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	ServiceName    = `datakit`
	DataKitGzipUrl = ""

	flagUpgrade         = flag.Bool("upgrade", false, ``)
	flagDataway         = flag.String("dataway", "", `address of dataway(ip:port), port default 9528`)
	flagInstallDir      = flag.String("install-dir", "", `directory to install`)
	flagVersion         = flag.Bool("version", false, "show installer version info")
	flagDownloadOnly    = flag.Bool("download-only", false, `download datakit only, not install`)
	flagDataKitGzipFile = flag.String("datakit-gzip", ``, `local path of datakit install files`)
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	applyFlags()

	// create install dir if not exists
	if err := os.MkdirAll(*flagInstallDir, 0775); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	datakitExe := filepath.Join(*flagInstallDir, "datakit")
	if runtime.GOOS == "windows" {
		datakitExe += ".exe"
	}

	prog := &program{}
	dkservice, err := service.New(prog, &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: `Collects data and upload it to DataFlux.`,
		Executable:  datakitExe,
		Arguments:   nil, // no args need here
	})

	if err != nil {
		log.Fatal("New %s service failed: %s", runtime.GOOS, err.Error())
	}

	if err := stopDataKitService(dkservice); err != nil {
		// ignore
	}

	if *flagDataKitGzipFile != "" {
		extractDatakit(*flagDataKitGzipFile, *flagInstallDir)
	} else {
		downloadDatakit(DataKitGzipUrl, *flagInstallDir)
	}

	if *flagUpgrade { // upgrade new version
		// do nothing
	} else { // install new datakit

		if *flagDataway == "" {
			for {
				dw := readInput("Please set DataWay(IP:Port or OpenWay link) > ")
				*flagDataway, err = parseDataway(dw)
				if err == nil {
					break
				}

				fmt.Printf("%s\n", err.Error())
				continue
			}
		} else {
			*flagDataway, err = parseDataway(*flagDataway)
			if err != nil {
				log.Fatal("%s", err.Error())
			}
		}

		uninstallDataKitService(dkservice)

		if err := config.InitCfg(*flagDataway); err != nil {
			log.Fatalf("Failed to init datakit main config: %s", err.Error())
		}

		log.Printf("install service %s...", ServiceName)
		if err := installDatakitService(dkservice); err != nil {
			log.Fatalf("Fail to register service %s: %s", ServiceName, err.Error())
		}
	}

	log.Printf("starting service %s...", ServiceName)
	if err := startDatakitService(dkservice); err != nil {
		log.Fatalf("Fail to register service %s: %s", ServiceName, err.Error())
	}

	log.Println(":) Success!")
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

	if *flagInstallDir == "" {
		switch runtime.GOOS + "/" + runtime.GOARCH {

		case "windows/amd64":
			*flagInstallDir = `C:\Program Files\DataFlux\` + ServiceName

		case "windows/386":
			*flagInstallDir = `C:\Program Files (x86)\DataFlux\` + ServiceName

		case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
			"darwin/amd64", "darwin/386",
			"freebsd/amd64", "freebsd/386":
			*flagInstallDir = `/usr/local/cloudcare/DataFlux/` + ServiceName

		default:
			// TODO
		}
	}
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	txt, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(txt)
}

func doDownload(r io.Reader, to string) {

	f, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	if _, err := io.Copy(f, r); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	f.Close()
}

func doExtract(r io.Reader, to string) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return
		case err != nil:
			log.Fatalf("[error] %s", err.Error())
		case hdr == nil:
			continue
		}

		target := filepath.Join(to, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					log.Fatalf("[error] %s", err.Error())
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			if _, err := io.Copy(f, tr); err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			f.Close()
		}
	}
}

func extractDatakit(gz, to string) {
	data, err := os.Open(gz)
	if err != nil {
		log.Fatalf("[error] open file %s failed: %s", gz, err)
	}

	defer data.Close()

	doExtract(data, to)
}

type WriteCounter struct {
	total   uint64
	current uint64
	last    float64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	return n, nil
}

func downloadDatakit(from, to string) {
	resp, err := http.Get(from)
	if err != nil {
		log.Fatalf("failed to download %s: %s", from, err)
	}

	defer resp.Body.Close()
	cnt := &WriteCounter{
		total: uint64(resp.ContentLength),
	}

	if *flagDownloadOnly {
		doDownload(io.TeeReader(resp.Body, cnt), to)
	} else {
		doExtract(io.TeeReader(resp.Body, cnt), to)
	}
	fmt.Printf("\n")
}

func (wc *WriteCounter) PrintProgress() {
	if wc.last > float64(wc.total)*0.01 || wc.current == wc.total { // update progress-bar each 1%
		fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Printf("\rDownloading... %s/%s", humanize.Bytes(wc.current), humanize.Bytes(wc.total))
		wc.last = 0.0
	}
}

func parseDataway(dw string) (string, error) {

	tkn := ""
	host := ""
	scheme := "http"

	if u, err := url.Parse(dw); err == nil {
		scheme = u.Scheme
		tkn = u.Query().Get("token")
		host = u.Host

		if scheme == "https" {
			host = u.Host + ":443"
		}
	} else {
		if _, _, err := net.SplitHostPort(dw); err != nil { // parse dataway IP:Port
			return "", err
		}

		host = dw
	}

	log.Printf("Testing DataWay(%s)...", host)
	conn, err := net.DialTimeout("tcp", host, time.Second*5)
	if err != nil {
		return "", err
	}
	conn.Close()

	if tkn == "" {
		return fmt.Sprintf("%s://%s/v1/write/metrics", scheme, host), nil
	} else {
		return fmt.Sprintf("%s://%s/v1/write/metrics?token=%s", scheme, host, tkn), nil
	}
}

type program struct{}

func (p *program) Start(s service.Service) error { go p.run(s); return nil }
func (p *program) run(s service.Service)         {}
func (p *program) Stop(s service.Service) error  { return nil }

func stopDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		log.Printf("Stop service datakit failed: %s, ignored", err.Error())
	}

	return nil
}

func uninstallDataKitService(s service.Service) error {
	if err := service.Control(s, "uninstall"); err != nil {
		log.Printf("Stop service datakit failed: %s, ignored", err.Error())
	}

	return nil
}

func installDatakitService(s service.Service) error {
	return service.Control(s, "install")
}

func startDatakitService(s service.Service) error {
	return service.Control(s, "start")
}

func initDatakit(exe, dw string) error {

	cmd := exec.Command(exe, "-init", "-dataway", dw)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	log.Printf("Initing datakit...")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
