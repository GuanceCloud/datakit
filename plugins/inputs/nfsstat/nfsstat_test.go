// +build linux

package nfsstat

import (
	"testing"
)

func TestMain(t *testing.T) {
	testAssert = true

	var nfs = NFSstat{
		Interval: "3s",
	}

	nfs.Run()
}
