package io

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
)

type OSSClient struct {
	EndPoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	BucketName      string `toml:"bucket_name"`

	Cli *oss.Client
}

func NewOSSClient(endpoint, ak, sk, bucket string) (*OSSClient, error) {
	var oc = &OSSClient{
		EndPoint:        endpoint,
		AccessKeyID:     ak,
		AccessKeySecret: sk,
		BucketName:      bucket,
	}
	cli, err := oc.GetOSSCli()
	if err != nil {
		return nil, err
	}
	oc.Cli = cli
	return oc, nil
}

func (oc *OSSClient) GetOSSCli() (*oss.Client, error) {
	cli, err := oss.New(oc.EndPoint, oc.AccessKeyID, oc.AccessKeySecret)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (oc *OSSClient) OSSUPLoad(objectName string, f *os.File) error {
	bucket, err := oc.Cli.Bucket(oc.BucketName)
	if err != nil {
		return err
	}
	err = bucket.PutObject(objectName, f)
	return err
}
