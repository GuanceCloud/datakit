// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GuanceCloud/cliutils"
)

func GetOSSClient() (*cliutils.OssCli, error) {
	var ak, sk, bucket, ossHost string

	switch ReleaseType {
	case ReleaseTesting, ReleaseProduction, ReleaseLocal:
		tag := strings.ToUpper(ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		if ak == "" {
			return nil, errors.New("env " + tag + "_OSS_ACCESS_KEY" + " is not configured")
		}
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		if sk == "" {
			return nil, errors.New("env " + tag + "_OSS_SECRET_KEY" + " is not configured")
		}
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		if bucket == "" {
			return nil, errors.New("env " + tag + "_OSS_BUCKET" + " is not configured")
		}
		ossHost = os.Getenv(tag + "_OSS_HOST")
		if ossHost == "" {
			return nil, errors.New("env " + tag + "_OSS_HOST" + " is not configured")
		}
	default:
		return nil, fmt.Errorf("unknown release type: %s", ReleaseType)
	}

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    "datakit",
	}
	if err := oc.Init(); err != nil {
		return nil, err
	}
	return oc, nil
}
