// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package plmap used to store data during pl script execution
package plmap

// type PtSource [3]string // source,sub_source,tmp_source

const FeedName = "pl_agg"

type UploadFunc func(ptID string, elem any) error

type PlMapIface interface {
	SetUploadFunc(UploadFunc)
}
