package io

import (
	"testing"
	"os"
)

func TestOSSUPLoad(t *testing.T) {
	region := "cn-shanghai"
	ak := "LTAI4G1E5j5QX5h1S4kT2qfg"
	sk := "aud5Bwb6tXExMoh5P1XEAinbZCH4kl"
	bucketName := "test20210223"
	cli,err := GetOSSCli(region,ak,sk)
	if err != nil {
		l.Fatal(err)
	}
	f,_ := os.Open("gen.sh")
	err = OSSUPLoad(cli,bucketName,"123.sh",f)
	if err != nil {
		l.Fatal(err)
	}
}