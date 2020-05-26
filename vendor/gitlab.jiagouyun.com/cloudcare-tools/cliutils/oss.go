package cliutils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	ErrOssFileTooLarge = errors.New(`oss file too large`)

	DefaultPartSize = int64(32 * 1024 * 1024) // 32MB
	DefaultTimeout  = uint(30)                // seconds
	DefaultWorkers  = 8
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
		log.Printf("[error] %s", err.Error())
		return err
	}

	cli.Config.Timeout = oc.Timeout

	cli.Config.HTTPTimeout.ConnectTimeout = time.Second * 3
	cli.Config.HTTPTimeout.HeaderTimeout = time.Duration(oc.Timeout) * time.Second
	cli.Config.HTTPTimeout.ReadWriteTimeout = time.Duration(oc.Timeout) * time.Second
	cli.Config.HTTPTimeout.LongTimeout = time.Duration(oc.Timeout) * time.Second

	bkt, err := cli.Bucket(oc.BucketName)
	if err != nil {
		log.Printf("[error] %s", err.Error())
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
		log.Printf("[error] %s", err.Error())
		return err
	}

	size := st.Size()

	if size <= oc.PartSize { // 小文件直接上传
		if err := oc.bkt.PutObjectFromFile(to, from); err != nil {
			oc.FailedCnt++
			log.Printf("[error] %s", err.Error())
			return err
		}

		oc.UploadedFiles++
		oc.UploadedBytes += size

		return nil
	} else {
		return oc.multipartUpload(from, to)
	}
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

func (oc *OssCli) multipartUpload(from, to string) error {
	// 大文件分片上传

	var err error

	st, err := os.Stat(from)
	if err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	size := st.Size()

	if size > int64(oc.PartSize*10000) { // 最大只支持 10k 个分片
		log.Printf("[error] too large file %s(%d)", from, size)
		return ErrOssFileTooLarge
	}

	partCnt := size / int64(oc.PartSize)

	chunks, err := oss.SplitFileByPartNum(from, int(partCnt))
	if err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	// 新建分片上传
	imur, err := oc.bkt.InitiateMultipartUpload(to, nil)
	if err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	log.Printf("[debug]: %+#v", imur)

	resCh := make(chan *oss.UploadPart, len(chunks)) // 接受返回结果
	failedCh := make(chan error)
	die := make(chan int) // 勒令退出

	defer close(die)

	worker := func(c *oss.FileChunk) {
		select {
		case <-die:
			return
		default:

			for i := 0; i < 3; i++ {
				p, err := oc.bkt.UploadPartFromFile(imur, from, c.Offset, c.Size, c.Number)
				if err == nil {
					resCh <- &p
					return
				} else {
					time.Sleep(time.Second)
				}
			}

			if err != nil {
				failedCh <- err
			}
		}
	}

	// 启动上传任务
	parts := []oss.UploadPart{}
	nrWorkers := 0
	idx := 0
	for {
		select {
		case p := <-resCh:
			parts = append(parts, *p)
			log.Printf("[debug] part-%05d (%s) uplaod OK(%d/%d)", p.PartNumber, from, len(parts), len(chunks))
			nrWorkers--

		case err := <-failedCh:
			if err != nil {
				log.Printf("[error] %s", err.Error())
			}

			oc.bkt.AbortMultipartUpload(imur)
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
		go worker(&chunks[idx])
		idx++
		nrWorkers++
	}

	log.Printf("[debug] complete %s...", to)
	_, err = oc.bkt.CompleteMultipartUpload(imur, parts)
	if err != nil {
		log.Printf("[error] %s", err.Error())
		oc.bkt.AbortMultipartUpload(imur)

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
