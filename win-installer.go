package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	ServiceName = `datakit`

	flagUpgrade    = flag.Bool("upgrade", false, ``)
	flagDataway    = flag.String("dataway", "", `address of dataway(ip:port)`)
	flagInstallDir = flag.String("install-dir", `C:\Program Files (x86)\Forethought\`+ServiceName, `directory to install`)
	flagGZPath     = flag.String("gzpath", "", "datakit gzip path")
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

			log.Printf("[debug] create %s ok, extract file %s", filepath.Dir(target), target)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			if _, err := io.Copy(f, tr); err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			f.Close()
			log.Printf("[debug] extract file %s ok", target)
		}
	}
}

func extractDatakit(gz, to string) {
	data, err := os.Open(gz)
	if err != nil {
		log.Fatal(err)
	}

	defer data.Close()

	doExtract(data, to)
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	if *flagVersion {
		fmt.Printf(`Version:        %s
Build At:       %s
Golang Version: %s
`, git.Version, git.BuildAt, git.Golang)
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

	if err := deleteDataKitService(dkservice); err != nil {
		// ignore
	}

	extractDatakit(*flagGZPath, *flagInstallDir)

	if *flagUpgrade { // upgrade new version
		if err := upgradeDatakit(datakitExe); err != nil {
			log.Fatal("[error] upgrade datakit failed: %s", err.Error())
		}
	} else { // install new datakit
		if err := initDatakit(datakitExe, fmt.Sprintf("http://%s/v1/write/metrics", *flagDataway)); err != nil {
			log.Fatal("[error] init datakit failed: %s", err.Error())
		}

		log.Printf("[info] try install service %s", ServiceName)
		if err := installDatakitService(dkservice); err != nil {
			log.Fatalf("[error] fail to register service %s: %s", ServiceName, err.Error())
		}

		log.Printf("[info] try start service %s", ServiceName)
		if err := startDatakitService(dkservice); err != nil {
			log.Fatalf("[error] fail to register service %s: %s", ServiceName, err.Error())
		}
	}

	log.Println(":)Success!")
}

type program struct{}

func (p *program) Start(s service.Service) error { go p.run(s); return nil }
func (p *program) run(s service.Service)         {}
func (p *program) Stop(s service.Service) error  { return nil }

func deleteDataKitService(s service.Service) error {

	if err := service.Control(s, "stop"); err != nil {
		log.Printf("[warn] stop service datakit failed: %s, ignored", err.Error())
	}

	if err := service.Control(s, "uninstall"); err != nil {
		log.Printf("[warn] uninstall service datakit failed: %s, ignored", err.Error())
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
