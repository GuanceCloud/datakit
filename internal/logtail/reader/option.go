// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package reader

type option struct {
	bufSize              int
	maxLineLength        int
	disablePreviousBlock bool
}

type Option func(*option)

func WithBufSize(n int) Option       { return func(opt *option) { opt.bufSize = n } }
func WithMaxLineLength(n int) Option { return func(opt *option) { opt.maxLineLength = n } }
func DisablePreviousBlock() Option   { return func(opt *option) { opt.disablePreviousBlock = true } }

func defaultOption() *option {
	return &option{
		bufSize:              1024 * 16,  // 16 KiB
		maxLineLength:        1024 * 128, // 128 KiB
		disablePreviousBlock: false,
	}
}
