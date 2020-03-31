//+build ignore

package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/kardianos/service"
)

var (
	serviceName = `datakit`

	flagUpgrade    = flag.Bool("upgrade", false, ``)
	flagTest       = flag.Bool("test", false, ``)
	flagFtDataway  = flag.String("ftdataway", "", `address of ftdataway`)
	flagInstallDir = flag.String("installdir", fmt.Sprintf(`C:\Program Files (x86)\Forethought\%s`, serviceName), `directory to install`)

	baseDownloadUrl = `https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit`

	installDir = fmt.Sprintf(`C:\Program Files (x86)\Forethought\%s`, serviceName)
)

type VerSt struct {
	Version string `json:"version"`
}

func main() {

	flag.Parse()

	if *flagTest {
		baseDownloadUrl = `https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/datakit`
	}

	if err := os.MkdirAll(installDir, 0775); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	if !*flagUpgrade {
		log.Printf("start installing %s in %s", serviceName, installDir)
	}

	//stop
	log.Printf("stopping %s", serviceName)
	cmd := exec.Command(`sc`, `stop`, serviceName)
	cmd.CombinedOutput()

	tarDownloadUrl := ""

	//log.Println("check version...")
	verUrl := baseDownloadUrl + `/version_win`
	verResp, err := http.Get(verUrl)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}
	defer verResp.Body.Close()
	js, err := ioutil.ReadAll(verResp.Body)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}
	var vt VerSt
	err = json.Unmarshal(js, &vt)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	osarch := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	tarDownloadUrl = fmt.Sprintf("%s/datakit-%s-%s.tar.gz", baseDownloadUrl, osarch, vt.Version)

	//download
	log.Println("start downloading...")

	client := &http.Client{}
	resp, err := client.Get(tarDownloadUrl)
	if err != nil {
		log.Fatalf("[error] download %s failed:  %s", tarDownloadUrl, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		log.Fatalf("[error] download %s failed: %s", tarDownloadUrl, resp.Status)
	}

	fsize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	destPath := os.TempDir()
	tmpName := fmt.Sprintf("%s-%v", serviceName, time.Now().Unix())
	destPath = filepath.Join(destPath, tmpName)

	file, err := os.Create(destPath)
	if err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	defer func() {
		file.Close()
		os.Remove(destPath)
	}()

	buf := make([]byte, 32*1024)
	var written int64
	for {
		nr, err := resp.Body.Read(buf)
		if nr > 0 {
			nw, err := file.Write(buf[:nr])
			if err != nil {
				log.Fatalf("[error] %s", err.Error())
			}

			written += int64(nw)

			percent := int(written * 100 / fsize)
			pro := Progress(percent)
			pro.Show(fsize)
		}

		if err != nil {
			if err == io.EOF {
				pro := Progress(100)
				pro.Show(fsize)
				break
			}
			log.Fatalf("[error] %s", err.Error())
		}
	}

	fmt.Printf("\r\n")
	if !*flagUpgrade {
		log.Println("installing...")
	} else {
		log.Println("upgrading...")
	}

	if err = deCompress(destPath, installDir); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	platformDir := filepath.Join(installDir, fmt.Sprintf("%s-%s-%s", serviceName, runtime.GOOS, runtime.GOARCH))
	_, err = os.Stat(platformDir)
	if err != nil {
		log.Fatalf("[error] unsupport for os=%s and arch=%s", runtime.GOOS, runtime.GOARCH)
	}

	binName := serviceName + ".exe"
	destbin := filepath.Join(installDir, binName)
	if err = os.Rename(filepath.Join(platformDir, binName), destbin); err != nil {
		log.Fatalf("[error] %s", err.Error())
	}

	os.Remove(platformDir)

	if *flagUpgrade {
		//upgrade
		cmd = exec.Command(destbin, "-upgrade")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err = cmd.Run(); err != nil {
			os.Exit(1)
		}
	} else {
		//init
		cmd = exec.Command(destbin, "-init", "-ftdataway", *flagFtDataway)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err = cmd.Run(); err != nil {
			os.Exit(1)
		}

		cfgpath := filepath.Join(installDir, fmt.Sprintf("%s.conf", serviceName))

		deleteSvr()
		time.Sleep(time.Millisecond * 500)

		err = nil
		for index := 0; index < 3; index++ {
			err = regSvr(destbin, cfgpath, false)
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			log.Fatalf("[error] fail to register as service: %s", err.Error())
			return
		}
	}

	log.Println(":)Success!")

}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(name), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}

func deCompress(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		filename := filepath.Join(dest, hdr.Name)
		if !hdr.FileInfo().IsDir() && !strings.HasPrefix(hdr.FileInfo().Name(), ".") {
			file, err := createFile(filename)
			if err != nil {
				return err
			}
			if _, err = io.Copy(file, tr); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}

	}
	return nil
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
		Name:        serviceName,
		DisplayName: serviceName,
		Description: `Collects data and publishes it to ftdataway.`,
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

type Progress int

func (x Progress) Show(filesize int64) {
	percent := int(x)

	total := 50
	middle := int(percent * total / 100.0)

	arr := make([]string, total)
	for j := 0; j < total; j++ {
		if j < middle-1 {
			arr[j] = "-"
		} else if j == middle-1 {
			arr[j] = ">"
		} else {
			arr[j] = " "
		}
	}
	bar := fmt.Sprintf("%vbytes(%s) [%s]", filesize, convFilesize(filesize), strings.Join(arr, ""))
	fmt.Printf("\r%s %%%d", bar, percent)
}

func convFilesize(filesize int64) string {
	if filesize < 1024 {
		return fmt.Sprintf("%vB", filesize)
	} else if filesize < 1024*1024 {
		return fmt.Sprintf("%vKB", filesize/1024)
	} else if filesize < 1024*1024*1024 {
		return fmt.Sprintf("%vMB", filesize/(1024*1024))
	} else {
		return fmt.Sprintf("%vGB", filesize/(1024*1024*1024))
	}
}
