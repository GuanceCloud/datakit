package io

import (
	"os"
	"testing"
)

func TestOSSUPLoad(t *testing.T) {
	region := "cn-shanghai"
	ak := "LTAI4G1E5j5QX5h1S4kT2qfg"
	sk := "aud5Bwb6tXExMoh5P1XEAinbZCH4kl"
	bucketName := "test20210223"
	cli, err := NewOSSClient(region, ak, sk, bucketName)
	if err != nil {
		l.Fatal(err)
	}
	f, _ := os.Open("gen.sh")
	err = cli.OSSUPLoad("123", f)
	if err != nil {
		l.Fatal(err)
	}
}
