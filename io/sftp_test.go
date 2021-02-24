package io

import (
	"testing"
	"os"
)

func TestSFTPUPLoad(t *testing.T) {
	user := "parallels"
	password := "hjj940622"
	host := "10.211.55.6"
	port := 22
	cli ,err := GetSFTPClient(user,password,host,port)
	if err != nil {
		l.Fatal(err)
	}
	f,_ := os.Open("gen.sh")
	err = SFTPUPLoad(cli,f,"/home/parallels/Desktop/ccc/123.sh")
	if err != nil {
		l.Fatal(err)
	}
}