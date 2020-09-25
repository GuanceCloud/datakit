// +build linux

package nfsstat

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()

	var nfs = NFSstat{
		Interval: "3s",
	}

	nfs.Run()
}
