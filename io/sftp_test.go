package io

import (
	"os"
	"testing"
)

func TestSFTPUPLoad(t *testing.T) {
	user := "parallels"
	password := "hjj940622"
	host := "10.211.55.6"
	port := 22
	cli, err := NewSFTPClient(user, password, host, "/home/parallels/Desktop/ccc", port)
	if err != nil {
		l.Fatal(err)
	}
	f, _ := os.Open("gen.sh")
	err = cli.SFTPUPLoad("/home/parallels/Desktop/ccc/cdir/ccc.sh", f)
	if err != nil {
		l.Fatal(err)
	}
}
