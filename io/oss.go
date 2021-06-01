package io

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OSSClient struct {
	EndPoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	BucketName      string `toml:"bucket_name"`
	DomainName      string `toml:"domain_name,omitempty"`

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

func (oc *OSSClient) OSSUPLoad(objectName string, reader io.Reader) error {
	bucket, err := oc.Cli.Bucket(oc.BucketName)
	if err != nil {
		return err
	}
	err = bucket.PutObject(objectName, reader)
	return err
}

func (oc *OSSClient) GetOSSUrl(remotePath string) string {
	if oc.DomainName == "" {
		return fmt.Sprintf("https://%s.%s/%s", oc.BucketName, oc.EndPoint, remotePath)
	}
	return fmt.Sprintf("https://%s/%s", oc.DomainName, remotePath)
}

func (oc *OSSClient) ObjectExist(remotePath string) (http.Header, error) {
	bucket, err := oc.Cli.Bucket(oc.BucketName)
	if err != nil {
		return nil, err
	}
	header, err := bucket.GetObjectMeta(remotePath)
	if err != nil {
		return nil, err
	}
	return header, nil
}
