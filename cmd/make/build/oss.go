// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"errors"
	"os"

	"github.com/GuanceCloud/cliutils"
)

func getOSSInfo() (*cliutils.OssCli, error) {
	var ak, sk, bucket, ossHost, path string

	ak = os.Getenv("OSS_ACCESS_KEY")
	if ak == "" {
		return nil, errors.New("env OSS_ACCESS_KEY is not configured")
	}

	sk = os.Getenv("OSS_SECRET_KEY")
	if sk == "" {
		return nil, errors.New("env OSS_SECRET_KEY is not configured")
	}

	bucket = os.Getenv("OSS_BUCKET")
	if bucket == "" {
		return nil, errors.New("env OSS_BUCKET is not configured")
	}

	ossHost = os.Getenv("OSS_HOST")
	if ossHost == "" {
		return nil, errors.New("env OSS_HOST is not configured")
	}

	path = os.Getenv("OSS_PATH")
	if ossHost == "" {
		return nil, errors.New("env OSS_PATH is not configured")
	}

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    path,
	}

	if err := oc.Init(); err != nil {
		return nil, err
	}

	return oc, nil
}
