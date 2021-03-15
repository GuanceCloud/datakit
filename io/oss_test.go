package io

import (
	"testing"
)

func TestOSSUPLoad(t *testing.T) {
	endpoint := "oss-cn-shanghai.aliyuncs.com"
	ak := "LTAI4G1E5j5QX5h1S4kT2qfg"
	sk := "aud5Bwb6tXExMoh5P1XEAinbZCH4kl"
	bucketName := "test20210223"
	cli, err := NewOSSClient(endpoint, ak, sk, bucketName)
	if err != nil {
		l.Fatal(err)
	}
	//f, _ := os.Open("gen.sh")
	//err = cli.OSSUPLoad("123", f)
	//if err != nil {
	//	l.Fatal(err)
	//}
	_, err = cli.ObjectExist("tkn_12595c1a660711ebb18e46cf65a67f12/122.conf")
	if err != nil {
		l.Fatal(err)
	}

}
