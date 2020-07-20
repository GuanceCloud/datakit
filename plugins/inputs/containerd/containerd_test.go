// +build linux

package containerd

import (
	"testing"
)

func TestMain(t *testing.T) {

	con := Containerd{
		HostPath:  "/run/containerd/containerd.sock",
		Namespace: "moby",
		IDList:    []string{"*"},
		Cycle:     60,
		Tags:      make(map[string]string),
	}

	con.Run()
}
