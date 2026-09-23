// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows || !amd64

package mem

import "github.com/GuanceCloud/cliutils/point"

func (*Input) collectWindowsCounters(kvs point.KVs) point.KVs { return kvs }
