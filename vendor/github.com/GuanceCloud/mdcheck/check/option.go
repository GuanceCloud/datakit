// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package check

type option func(*checkOption)

type checkOption struct {
	mddir, metadir string
	autofix        bool
}

// WithMarkdownDir set markdown path to checking.
func WithMarkdownDir(dir string) option {
	return func(o *checkOption) {
		if len(dir) >= 0 {
			o.mddir = dir
		}
	}
}

// WithMetaDir set markdown meta data path for checking.
func WithMetaDir(dir string) option {
	return func(o *checkOption) {
		if len(dir) >= 0 {
			o.metadir = dir
		}
	}
}

// WithAutofix set auto fix during checking.
func WithAutofix(on bool) option {
	return func(o *checkOption) {
		o.autofix = on
	}
}
