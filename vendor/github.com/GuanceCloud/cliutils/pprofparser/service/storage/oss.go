// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package storage

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/pprofparser/cfg"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/parameter"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	timeZoneCST = time.FixedZone("CST", 3600*8)
	ossClient   *OSS
	ossInitLock sync.Mutex
)

type OSS struct {
	client *oss.Client
}

func InitOSS() (*OSS, error) {
	if ossClient != nil {
		return ossClient, nil
	}
	ossInitLock.Lock()
	defer ossInitLock.Unlock()
	if ossClient == nil {
		client, err := oss.New(cfg.Cfg.Oss.Host, cfg.Cfg.Oss.AccessKey,
			cfg.Cfg.Oss.SecretKey)
		if err != nil {
			return nil, fmt.Errorf("oss.New fail: %w", err)
		}
		ossClient = &OSS{client: client}
	}
	return ossClient, nil
}

func (o *OSS) selectBucket() (*oss.Bucket, error) {
	return o.client.Bucket(cfg.Cfg.Oss.ProfileBucket)
}

func (o *OSS) IsFileExists(path string) (bool, error) {
	bucket, err := o.selectBucket()
	if err != nil {
		return false, fmt.Errorf("call IsFileExists, path [%s]: %w", path, err)
	}
	return bucket.IsObjectExist(path)
}

func (o *OSS) ReadFile(path string) (io.ReadCloser, error) {
	bucket, err := o.selectBucket()
	if err != nil {
		return nil, fmt.Errorf("ReadFile selectBucket: %w", err)
	}
	return bucket.GetObject(path)
}

func (o *OSS) GetProfileDir(workspaceUUID string, profileID string, unixTimeNS int64) string {
	if unixTimeNS >= parameter.MinTimestampMicro && unixTimeNS <= parameter.MaxTimestampMicro {
		unixTimeNS *= 1000
	}

	date := time.Unix(0, unixTimeNS).In(timeZoneCST).Format("20060102")
	path := filepath.Join(cfg.Cfg.Oss.ProfileDir, date, workspaceUUID, profileID[:2], profileID)
	return strings.ReplaceAll(path, "\\", "/")
}

func (o *OSS) GetProfilePath(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string {
	if unixTimeNS >= parameter.MinTimestampMicro && unixTimeNS <= parameter.MaxTimestampMicro {
		unixTimeNS *= 1000
	}

	date := time.Unix(0, unixTimeNS).In(timeZoneCST).Format("20060102")
	path := filepath.Join(cfg.Cfg.Oss.ProfileDir, date, workspaceUUID, profileID[:2], profileID, ossFilename)
	return strings.ReplaceAll(path, "\\", "/")
}

func (o *OSS) GetProfilePathOld(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string {
	if unixTimeNS >= parameter.MinTimestampMicro && unixTimeNS <= parameter.MaxTimestampMicro {
		unixTimeNS *= 1000
	}
	date := time.Unix(0, unixTimeNS).In(timeZoneCST).Format("20060102")
	path := filepath.Join(cfg.Cfg.Oss.ProfileDir, date, workspaceUUID, profileID, ossFilename)
	return strings.ReplaceAll(path, "\\", "/")
}
