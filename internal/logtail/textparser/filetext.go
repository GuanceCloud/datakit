// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package textparser

func ParseFileText(log []byte, msg *LogMessage) error {
	msg.Stream = Stdout
	msg.Log = log
	return nil
}
