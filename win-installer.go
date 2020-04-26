//+build ignore

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
	"time"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	ServiceName = `datakit`

	flagUpgrade     = flag.Bool("upgrade", false, ``)
	flagDataway     = flag.String("dataway", "", `address of dataway`)
	flagInstallDir  = flag.String("installdir", `C:\Program Files (x86)\Forethought\`+ServiceName, `directory to install`)
	flagDownloadURL = flag.String("downloadurl", "", "base download path")
	flagGZPath      = flag.String("gzpath", "", "datakit gzip path")
	flagVersion     = flag.Bool("version", false, "show installer version info")
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

func downloadDatakit(url, to string) {
	log.Println("start downloading...")

	client := &http.Client{}
	resp, err := client.Get(*flagDownloadURL)
	if err != nil {
		log.Fatalf("[error] download %s failed: %s", url, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		log.Fatalf("[error] download %s failed: %s", url, resp.Status)
	}

	doExtract(resp.Body, to)
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

	if *flagUpgrade {
		log.Printf("[info] start upgrading %s under %s", ServiceName, *flagInstallDir)

		// stop exist service
		log.Printf("[info] stopping %s", ServiceName)
		cmd := exec.Command(`sc`, `stop`, ServiceName)
		cmd.CombinedOutput()
	}

	//if *flagDownloadURL == "" {
	//	log.Fatalf("[error] download URL not set")
	//}

	//downloadDatakit(*flagDownloadURL, *flagInstallDir)
	extractDatakit(*flagGZPath, *flagInstallDir)

	datakitExe := filepath.Join(*flagInstallDir, "datakit.exe")
	var err error

	deleteSvr()
	if *flagUpgrade {

		log.Printf("[debug] stop and delete old datakit service...")

		// upgrade new version
		cmd := exec.Command(datakitExe, "-upgrade")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		log.Printf("[debug] upgrading...")
		if err = cmd.Run(); err != nil {
			os.Exit(1)
		}
	} else {
		// install new datakit
		cmd := exec.Command(datakitExe, "-init", "-dataway", *flagDataway)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		log.Printf("[debug] initing datakit...")
		if err = cmd.Run(); err != nil {
			os.Exit(1)
		}

		cfgpath := filepath.Join(*flagInstallDir, fmt.Sprintf("%s.conf", ServiceName))

		log.Printf("[info] try install service %s", ServiceName)
		for index := 0; index < 3; index++ {
			err = regSvr(datakitExe, cfgpath, false)
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}

		if err != nil {
			log.Fatalf("[error] fail to register service %s: %s", ServiceName, err.Error())
			return
		}
	}

	log.Printf("[debug] install service ok")

	log.Println(":)Success!")
}

type program struct {
}

func (p *program) Start(s service.Service) error {
	go p.run(s)
	return nil
}

func (p *program) run(s service.Service) {
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func deleteSvr() error {
	cmd := exec.Command(`sc`, "stop", `datakit`)
	cmd.CombinedOutput()

	time.Sleep(time.Millisecond * 200)

	cmd = exec.Command(`sc`, "delete", `datakit`)
	cmd.CombinedOutput()

	return nil
}

func regSvr(exepath, cfgpath string, remove bool) error {
	svcConfig := &service.Config{
		Name:        ServiceName,
		DisplayName: ServiceName,
		Description: `Collects data and publishes it to dataway.`,
		Arguments:   []string{"/config", cfgpath},
		Executable:  exepath,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		return err
	}

	if remove {
		service.Control(s, "stop")
		time.Sleep(time.Millisecond * 100)
		return service.Control(s, "uninstall")
	} else {
		return service.Control(s, "install")
	}
}
