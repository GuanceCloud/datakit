package datakit

import (
	"testing"

	"github.com/kardianos/service"
)

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
