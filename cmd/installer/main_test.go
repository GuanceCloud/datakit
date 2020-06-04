package main

import (
	//"os"
	"log"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kardianos/service"
)

func TestUpdateLagacyConfig(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	updateLagacyConfig("/usr/local/cloudcare/forethought/datakit")
}

func TestServiceInstall(t *testing.T) {

	installDir := ""

	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "windows/amd64":
		installDir = `C:\Program Files\DataFlux\` + ServiceName
	case "windows/386":
		installDir = `C:\Program Files (x86)\DataFlux\` + ServiceName
	case "linux/amd64", "linux/386", "linux/arm", "linux/arm64",
		"darwin/amd64", "darwin/386",
		"freebsd/amd64", "freebsd/386":
		installDir = `/usr/local/cloudcare/DataFlux/` + ServiceName

	default:
		// TODO
	}

	datakitExe := filepath.Join(installDir, "datakit")
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
		t.Fatalf("New %s service failed: %s", runtime.GOOS, err.Error())
	}

	//if err := installDatakitService(dkservice); err != nil {
	//	t.Errorf("Fail to register service %s: %s", ServiceName, err.Error())
	//}

	//serviceFile := "/etc/systemd/system/datakit.service"
	//if _, err := os.Stat(serviceFile); err == nil {
	//	t.Logf("file %s exits", serviceFile)
	//} else {
	//	t.Errorf("file %s missing", serviceFile)
	//}

	uninstallDataKitService(dkservice)
	//if _, err := os.Stat(serviceFile); err == nil {
	//	t.Errorf("file %s still exist", serviceFile)
	//} else {
	//	t.Logf("file %s cleaned ok", serviceFile)
	//}
}
