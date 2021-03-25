// +build linux

package containerd

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var con = Containerd{
		Location:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Interval:  "5s",
	}

	con.Run()
}

func TestTesting(t *testing.T) {
	var con = Containerd{
		Location:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Interval:  "5s",
	}

	data, err := con.Test()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", data)
}
