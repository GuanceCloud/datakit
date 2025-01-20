// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package check implements markdown basic checkings.
package check

type option func(*checkOption)

type checkOption struct {
	mddir, metadir string

	skips []string
	checkPunctuation,
	checkSection,
	checkLinks,
	checkMeta,
	autofix bool
}

func defaultOpt() *checkOption {
	return &checkOption{
		checkPunctuation: true,
		checkSection:     true,
		checkLinks:       true,
		checkMeta:        true,
		autofix:          true,
	}
}

func (o *checkOption) skipped(f string) bool {
	for _, x := range o.skips {
		if f == x {
			return true
		}
	}
	return false
}

// WithMarkdownDir set markdown path to checking.
func WithMarkdownDir(dir string) option {
	return func(o *checkOption) {
		if len(dir) > 0 {
			o.mddir = dir
		}
	}
}

// WithExcludeFiles set files that skip checking.
func WithExcludeFiles(files ...string) option {
	return func(o *checkOption) {
		o.skips = append(o.skips, files...)
	}
}

// WithMetaDir set markdown meta data path for checking.
func WithMetaDir(dir string) option {
	return func(o *checkOption) {
		if len(dir) > 0 {
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

func WithCheckSection(on bool) option {
	return func(o *checkOption) {
		o.checkSection = on
	}
}

func WithCheckPunctuation(on bool) option {
	return func(o *checkOption) {
		o.checkPunctuation = on
	}
}

func WithCheckLinks(on bool) option {
	return func(o *checkOption) {
		o.checkLinks = on
	}
}

func WithCheckMeta(on bool) option {
	return func(o *checkOption) {
		o.checkMeta = on
	}
}
