package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	ServiceName    = `datakit`
	DataKitGzipUrl = ""

	flagUpgrade    = flag.Bool("upgrade", false, ``)
	flagDataway    = flag.String("dataway", "", `address of dataway(ip:port)`)
	flagInstallDir = flag.String("install-dir", `C:\Program Files (x86)\Forethought\`+ServiceName, `directory to install`)
	flagVersion    = flag.Bool("version", false, "show installer version info")
)

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

	doExtract(io.TeeReader(resp.Body, cnt), to)
	fmt.Printf("\n")
}

func (wc *WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces

	if wc.last > float64(wc.total)*0.01 {
		fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Printf("\rDownloading... %s/%s", humanize.Bytes(wc.current), humanize.Bytes(wc.total))
		wc.last = 0.0
	}
}

func main() {

	//log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	if *flagVersion {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
   DataKitGzip: %s
`, git.Version, git.BuildAt, git.Golang, DataKitGzipUrl)
		return
	}

	// create install dir if not exists
	if err := os.MkdirAll(*flagInstallDir, 0775); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	datakitExe := filepath.Join(*flagInstallDir, "datakit.exe")

	prog := &program{}
	cfgpath := filepath.Join(*flagInstallDir, fmt.Sprintf("%s.conf", ServiceName))
	dkservice, err := service.New(prog, &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: `Collects data and upload it to DataFlux.`,
		Arguments:   []string{"/config", cfgpath},
		Executable:  datakitExe,
	})

	if err != nil {
		log.Fatal("[error] new service failed: %s", err.Error())
	}

	if err := stopDataKitService(dkservice); err != nil {
		// ignore
	}

	downloadDatakit(DataKitGzipUrl, *flagInstallDir)

	if *flagUpgrade { // upgrade new version
		if err := upgradeDatakit(datakitExe); err != nil {
			log.Fatal("[error] upgrade datakit failed: %s", err.Error())
		}
	} else { // install new datakit

		if *flagDataway == "" {
			log.Fatal("DataWay IP:Port required")
		}

		uninstallDataKitService(dkservice)

		if err := initDatakit(datakitExe, fmt.Sprintf("http://%s/v1/write/metrics", *flagDataway)); err != nil {
			log.Fatal("[error] init datakit failed: %s", err.Error())
		}

		log.Printf("install service %s...", ServiceName)
		if err := installDatakitService(dkservice); err != nil {
			log.Fatalf("[error] fail to register service %s: %s", ServiceName, err.Error())
		}
	}

	log.Printf("starting service %s...", ServiceName)
	if err := startDatakitService(dkservice); err != nil {
		log.Fatalf("[error] fail to register service %s: %s", ServiceName, err.Error())
	}

	log.Println(":) Success!")
}

type program struct{}

func (p *program) Start(s service.Service) error { go p.run(s); return nil }
func (p *program) run(s service.Service)         {}
func (p *program) Stop(s service.Service) error  { return nil }

func stopDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		log.Printf("[warn] stop service datakit failed: %s, ignored", err.Error())
	}

	return nil
}

func uninstallDataKitService(s service.Service) error {
	if err := service.Control(s, "uninstall"); err != nil {
		log.Printf("[warn] stop service datakit failed: %s, ignored", err.Error())
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

	log.Printf("[debug] initing datakit...")
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func upgradeDatakit(exe string) error {
	cmd := exec.Command(exe, "-upgrade")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	log.Printf("[debug] datakit upgrading...")
	return cmd.Run()
}
