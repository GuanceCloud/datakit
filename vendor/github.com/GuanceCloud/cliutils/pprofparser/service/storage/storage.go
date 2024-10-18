// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package storage implements a series of storage kinds.
package storage

import (
	"fmt"
	"io"
)

type Type string

const (
	AliOSS    Type = "oss"
	LocalDisk Type = "disk"
)

var DefaultDiskStorage = &Disk{}

type Storage interface {
	IsFileExists(path string) (bool, error)
	GetProfilePath(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string
	GetProfileDir(workspaceUUID string, profileID string, unixTimeNS int64) string
	GetProfilePathOld(workspaceUUID string, profileID string, unixTimeNS int64, ossFilename string) string
	ReadFile(path string) (io.ReadCloser, error)
}

func GetStorage(typ Type) (Storage, error) {
	switch typ {
	case AliOSS:
		return InitOSS()
	case LocalDisk:
		return DefaultDiskStorage, nil
	}
	return nil, fmt.Errorf("storage type [%s] not supported yet", typ)
}
