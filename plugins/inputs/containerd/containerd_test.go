// +build linux

package containerd

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var con = Containerd{
		HostPath:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Interval:  "5s",
	}

	con.Run()
}
