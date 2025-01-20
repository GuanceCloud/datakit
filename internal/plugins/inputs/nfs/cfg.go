// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:deadcode,unused
package nfs

import (
	"errors"
	"time"
)

var (
	ErrFileParse  = errors.New("error parsing file")
	ErrFileRead   = errors.New("error reading file")
	ErrMountPoint = errors.New("error accessing mount point")

	float64Mantissa uint64 = 9007199254740992
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "nfs"
	metricName  = inputName

	deviceEntryLen = 8

	fieldBytesLen  = 8
	fieldEventsLen = 27

	statVersion10 = "1.0"
	statVersion11 = "1.1"

	fieldTransport10TCPLen = 10
	fieldTransport10UDPLen = 7

	fieldTransport11TCPLen = 13
	fieldTransport11UDPLen = 10

	sampleCfg = `
[[inputs.nfs]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ## Whether to enable NFSd metric collection
  # nfsd = true

  ## NFS mount point metric configuration
  [inputs.nfs.mountstats]
    ## Enable r/w statistics
    # rw = true
    ## Enable transport statistics
    # transport = true
    ## Enable event statistics
    # event = true
    ## Enable operation statistics
    # operations = true

  [inputs.nfs.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)
