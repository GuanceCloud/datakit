package datakit

import (
	"testing"

	"github.com/influxdata/toml"
	"github.com/kardianos/service"
)

func TestMarshalMainCfg(t *testing.T) {

	if Cfg.MainCfg.Hostname == "" {
		Cfg.setHostname()
	}

	data, err := toml.Marshal(Cfg.MainCfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(data))
}

func TestLocalIP(t *testing.T) {
	ip, err := LocalIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("IP: %s", ip)
}

func TestGetFirstGlobalUnicastIP(t *testing.T) {
	ip, err := GetFirstGlobalUnicastIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("IP: %s", ip)
}

func TestServiceInstall(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatal(err)
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		t.Log(err)
	}

	if err := service.Control(svc, "install"); err != nil {
		t.Fatal(err)
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		t.Fatal(err)
	}
}
