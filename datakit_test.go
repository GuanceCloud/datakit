package datakit

import (
	"testing"
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
