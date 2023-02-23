// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cliutils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	ErrOssFileTooLarge = errors.New(`oss file too large`)

	DefaultPartSize = int64(32 * 1024 * 1024) //nolint:gomnd // 32MB
	DefaultTimeout  = uint(30)                //nolint:gomnd // seconds
	DefaultWorkers  = 8
)

const (
	ossMaxParts = 10000
)

type OssCli struct {
	Host, AccessKey, SecretKey, BucketName, WorkDir string
	Timeout                                         uint
	PartSize                                        int64
	Workers                                         int

	bkt *oss.Bucket

	ReconnectCnt  int
	FailedCnt     int
	UploadedFiles int
	UploadedBytes int64
}

func (oc *OssCli) getBucket() error {
	cli, err := oss.New(oc.Host, oc.AccessKey, oc.SecretKey)
	if err != nil {
		return err
	}

	cli.Config.Timeout = oc.Timeout

	cli.Config.HTTPTimeout.ConnectTimeout = time.Second * 3
	cli.Config.HTTPTimeout.HeaderTimeout = time.Duration(oc.Timeout) * time.Second
	cli.Config.HTTPTimeout.ReadWriteTimeout = time.Duration(oc.Timeout) * time.Second
	cli.Config.HTTPTimeout.LongTimeout = time.Duration(oc.Timeout) * time.Second

	bkt, err := cli.Bucket(oc.BucketName)
	if err != nil {
		return err
	}

	oc.bkt = bkt
	return nil
}

func (oc *OssCli) Init() error {
	if oc.PartSize == 0 {
		oc.PartSize = DefaultPartSize
	}

	if oc.Timeout == 0 {
		oc.Timeout = DefaultTimeout
	}

	if oc.Workers == 0 {
		oc.Workers = DefaultWorkers
	}

	if err := oc.getBucket(); err != nil {
		return err
	}

	oc.ReconnectCnt = 0
	oc.FailedCnt = 0
	oc.UploadedFiles = 0
	oc.UploadedBytes = int64(0)

	return nil
}

func (oc *OssCli) Reconnect() error {
	if err := oc.getBucket(); err != nil {
		return err
	}

	oc.ReconnectCnt++
	return nil
}

func (oc *OssCli) Stat() string {
	return fmt.Sprintf("uploaded %d files, total %s, reconnect: %d, FailedCnt: %d",
		oc.UploadedFiles, SizeFmt(oc.UploadedBytes), oc.ReconnectCnt, oc.FailedCnt)
}

func (oc *OssCli) Upload(from, to string) error {
	var err error

	st, err := os.Stat(from)
	if err != nil {
		return err
	}

	if size := st.Size(); size <= oc.PartSize { // 小文件直接上传
		if err := oc.bkt.PutObjectFromFile(to, from); err != nil {
			oc.FailedCnt++
			return err
		}

		oc.UploadedFiles++
		oc.UploadedBytes += size

		return nil
	}

	return oc.multipartUpload(from, to)
}

func (oc *OssCli) SetMeta(obj string, meta map[string]string) error {
	options := []oss.Option{}
	for k, v := range meta {
		options = append(options, oss.Meta(k, v))
	}

	return oc.bkt.SetObjectMeta(obj, options...)
}

func (oc *OssCli) GetMeta(obj string) (map[string][]string, error) {
	return oc.bkt.GetObjectDetailedMeta(obj)
}

func (oc *OssCli) ListObjects(prefix, marker string, maxKeys int) (oss.ListObjectsResult, error) {
	res, err := oc.bkt.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.MaxKeys(maxKeys))
	return res, err
}

func (oc *OssCli) Move(from, to string) error {
	if _, err := oc.bkt.CopyObject(from, to); err != nil {
		return err
	}

	return oc.bkt.DeleteObject(from)
}

func (oc *OssCli) mpworker(imur *oss.InitiateMultipartUploadResult,
	c *oss.FileChunk, from string, exit chan interface{},
) (p oss.UploadPart, err error) {
	select {
	case <-exit:
		return

	default:

		for i := 0; i < 3; i++ {
			p, err = oc.bkt.UploadPartFromFile(*imur, from, c.Offset, c.Size, c.Number)
			if err == nil {
				return p, nil
			}
			time.Sleep(time.Second)
		}
		return
	}
}

func (oc *OssCli) multipartUpload(from, to string) error {
	st, err := os.Stat(from)
	if err != nil {
		return err
	}

	size := st.Size()

	if size > oc.PartSize*ossMaxParts { // 最大只支持 10k 个分片
		return ErrOssFileTooLarge
	}

	partCnt := size / oc.PartSize

	chunks, err := oss.SplitFileByPartNum(from, int(partCnt))
	if err != nil {
		return err
	}

	// 新建分片上传
	imur, err := oc.bkt.InitiateMultipartUpload(to, nil)
	if err != nil {
		return err
	}

	resCh := make(chan *oss.UploadPart, len(chunks)) // 接受返回结果
	failedCh := make(chan error)
	exit := make(chan interface{}) // 勒令退出

	defer close(exit)

	// 启动上传任务
	parts := []oss.UploadPart{}
	nrWorkers := 0
	idx := 0
	for {
		select {
		case p := <-resCh:
			parts = append(parts, *p)
			nrWorkers--

		case err = <-failedCh:
			// ignore abort error
			_ = oc.bkt.AbortMultipartUpload(imur)
			return err

		default:
			time.Sleep(time.Second)
		}

		if len(parts) == len(chunks) { // 所有分片全部完成
			break
		}
		if nrWorkers >= oc.Workers || idx >= len(chunks) { // 控制并发个数
			continue
		}

		// 每个分片都用一个新的 goroutine 上传
		go func() {
			if p, err := oc.mpworker(&imur, &chunks[idx], from, exit); err != nil { //nolint:govet
				failedCh <- err
			} else {
				resCh <- &p
			}
		}()

		idx++
		nrWorkers++
	}

	_, err = oc.bkt.CompleteMultipartUpload(imur, parts)
	if err != nil {
		_ = oc.bkt.AbortMultipartUpload(imur)

		oc.FailedCnt++
		return err
	}

	oc.UploadedFiles++
	oc.UploadedBytes += size
	return nil
}

func (oc *OssCli) Download(obj, to string) error {
	return oc.bkt.GetObjectToFile(obj, to)
}
