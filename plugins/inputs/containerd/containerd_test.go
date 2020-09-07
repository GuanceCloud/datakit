// +build linux

package containerd

import (
	"testing"
)

func TestMain(t *testing.T) {
	testAssert = true

	var con = Containerd{
		HostPath:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Interval:  "5s",
	}

	con.Run()
}
