// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package textparser

import (
	"bytes"
	"fmt"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

var (
	// timeFormatIn is the format for parsing timestamps from other logs.
	// timeFormatIn = "2006-01-02T15:04:05.999999999Z07:00".

	// delimiter is the delimiter for timestamp and stream type in log line.
	delimiter = []byte{' '}
	// tagDelimiter is the delimiter for log tags.
	tagDelimiter = []byte(runtimeapi.LogTagDelimiter)
)

// ParseCRILog parses logs in CRI log format. CRI Log format example:
//
//	2016-10-06T00:17:09.669794202Z stdout P log content 1
//	2016-10-06T00:17:09.669794203Z stderr F log content 2
//
// refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kuberuntime/logs/logs.go#L128
func ParseCRILog(log []byte, msg *LogMessage) error {
	idx := bytes.Index(log, delimiter)
	if idx < 0 {
		return fmt.Errorf("timestamp is not found")
	}

	// Parse timestamp
	// msg.Timestamp, err = time.Parse(timeFormatIn, string(log[:idx]))
	// if err != nil {
	// 	return fmt.Errorf("unexpected timestamp format %q: %w", timeFormatIn, err)
	// }

	// Parse stream type
	log = log[idx+1:]
	idx = bytes.Index(log, delimiter)
	if idx < 0 {
		return fmt.Errorf("stream type is not found")
	}
	msg.Stream = LogStreamType(log[:idx])
	if msg.Stream != Stdout && msg.Stream != Stderr {
		return fmt.Errorf("unexpected stream type %q", msg.Stream)
	}

	// Parse log tag
	log = log[idx+1:]
	idx = bytes.Index(log, delimiter)
	if idx < 0 {
		// log tag is not found, nil bytes
		return nil
	}

	// Keep this forward compatible.
	tags := bytes.Split(log[:idx], tagDelimiter)
	partial := runtimeapi.LogTag(tags[0]) == runtimeapi.LogTagPartial
	// Trim the tailing new line if this is a partial line.
	if partial && len(log) > 0 && log[len(log)-1] == '\n' {
		log = log[:len(log)-1]
	}
	msg.IsPartial = partial

	// Get log content
	msg.Log = log[idx+1:]
	return nil
}
