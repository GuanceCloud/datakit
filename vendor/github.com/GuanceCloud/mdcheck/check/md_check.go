// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package check

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

type markdownRule struct {
	*regexp.Regexp
	info    string
	replace string
}

func newRule(re, replace, info string) *markdownRule {
	return &markdownRule{
		Regexp:  regexp.MustCompile(re),
		info:    info,
		replace: replace,
	}
}

var (

	// `grep` among macOS and Linux not compatible to match Chinese characters, but Golang's
	// regexp works well among these platforms. We use the regexp to match following patterns:
	//
	//   中文with english
	//   english with中文
	//   中文with english然后又是中文
	//   数字123然后又是中文
	//   行内代码 `abc_`然后又是中文
	//   行内代码`_abc` 然后又是中文
	//   ...
	// In one word: we must make sure there are space between ASCII character and Chinese characters.
	// But we ignore following, because they are standard markdown syntax:
	//
	//   [中文链接跳转](https://...)
	//   **加粗中文**
	//   *斜体中文*

	punctuationRules = []*markdownRule{
		// On Chinese characters followed by English punctuations, here we match a option space here,
		// to make there is no space after replace, for example:
		//
		//   中文, 然后...
		//
		// will replaced as(replace ',' as '，'):
		//
		//   中文，然后...
		//
		// and there is no space after `，'.
		newRule(`([\p{Han}]+)\. ?`, "$1。", "!!!invalid: English punctuation('.') after Chinese"),
		newRule(`([\p{Han}]+), ?`, "$1，", "!!!invalid: English punctuation(',') after Chinese"),
		newRule(`([\p{Han}]+); ?`, "$1；", "!!!invalid: English punctuation(';') after Chinese"),
		newRule(`([\p{Han}]+)! ?`, "$1！", "!!!invalid: English punctuation('!') after Chinese"),
		newRule(`([\p{Han}]+): ?`, "$1：", "!!!invalid: English punctuation(':') after Chinese"),

		// On English punctuation followed Chinese characters, here we do not match option space here,
		// because Markdown got list syntax like this:
		//
		//   1. 中文...
		//   1. 中文...
		//   1. 中文...
		//
		// and we hava to skip these case.
		newRule(`\.([\p{Han}]+)`, "。$1", "!!!invalid: English punctuation('.') before Chinese"),
		newRule(`,([\p{Han}]+)`, "，$1", "!!!invalid: English punctuation(',') before Chinese"),
		newRule(`;([\p{Han}]+)`, "；$1", "!!!invalid: English punctuation(';') before Chinese"),
		newRule(`!([\p{Han}]+)`, "！$1", "!!!invalid: English punctuation('!') before Chinese"),
		newRule(`:([\p{Han}]+)`, "：$1", "!!!invalid: English punctuation(':') before Chinese"),

		// Add space between English and Chinese
		//
		//  abc中文def
		//
		// will replaced as
		//
		//  abc 中文 def
		//
		newRule("([a-zA-Z0-9_`]+?)([\\p{Han}])", "$1 $2", "!!!invalid: need 1 space between English and Chinese"),
		newRule("([\\p{Han}])([a-zA-Z0-9_`]+?)", "$1 $2", "!!!invalid: need 1 space between Chinese and English"),
	}

	// Markdown Section checking, all sections must attach a tag like:
	//  ## This is a section {#some-section}
	regSection = newRule(
		`(?m)^\#{2,6}\s*(.*)$`,
		"", // can not fix invalid section
		"!!!invalid: section should have a tag")

	// Check external links.
	regExternalLink = newRule(`(?P<text>\[.*\])\((?P<link>http[a-zA-Z0-9_/\-\.:?=&+]+)\)(?P<suffix>\{:target="_blank"*\})?`,
		`$1($2){:target="_blank"}`,
		"!!!invalid: external links should suffixed with target=_blank")

	excludeExternalURLs = []string{
		"static.guance.com",
	}
)

type CheckResult struct {
	Path string `json:"path"`
	Text string `json:"text,omitempty"`
	Err  string `json:"err,omitempty"`
	Warn string `json:"warn,omitempty"`
}

func Check(opts ...option) (res []*CheckResult, err error) {
	co := defaultOpt()
	for _, opt := range opts {
		if opt != nil {
			opt(co)
		}
	}

	var mds []string
	if err = filepath.Walk(co.mddir, func(path string, info fs.FileInfo, err error) error {
		if info == nil {
			log.Printf("nothing under %s", path)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), "md") {
			return nil
		}

		if co.skipped(path) {
			return nil
		}

		mds = append(mds, path)
		return nil
	}); err != nil {
		log.Printf("[W] filepath.Walk: %s", err)
		return
	}

	for _, md := range mds {
		log.Printf("Checking %s ...", md)

		if co.metadir != "" && co.checkMeta {
			res = append(res, checkMarkdownMeta(md, co.metadir)...)
			continue // only check meta info
		}

		arr, fix := co.doMatch(md)
		if len(arr) > 0 { // we find some bad message
			if len(fix) > 0 {
				if err := ioutil.WriteFile(filepath.Clean(md), []byte(fix), 0o600); err != nil {
					panic(err.Error())
				}
			}

			res = append(res, arr...)
		}
	}

	log.Printf("matched %d files", len(mds))
	return res, nil
}

func getLineNumber(from, sub []byte) int {
	line := 1 // file line number start from 1
	if pos := bytes.Index(from, sub); pos > 0 {
		for i := 0; i < pos; i++ {
			if from[i] == '\n' {
				line++
			}
		}
	}

	return line
}

func (co *checkOption) doMatch(md string) (res []*CheckResult, fix string) {
	d, err := ioutil.ReadFile(filepath.Clean(md))
	if err != nil {
		log.Printf("[E] ReadFile: %s", err)
		return
	}

	// check punctuations
	for _, r := range punctuationRules {
		if !co.checkPunctuation {
			break
		}

		arr := r.Regexp.FindAllString(string(d), -1)
		if len(arr) > 0 {
			if co.autofix {
				fix = r.Regexp.ReplaceAllString(string(d), r.replace)
			}

			for _, item := range arr {
				res = append(res, &CheckResult{
					Path: fmt.Sprintf("%s:%d", md, getLineNumber(d, []byte(item))),
					Err:  r.info,
					Text: item,
				})
			}
		}
	}

	if co.checkSection {
		// check sections
		arr := regSection.Regexp.FindAllString(string(d), -1)
		if len(arr) > 0 {
			for _, item := range arr {
				if strings.Contains(item, "{{") { // this may be a template
					continue
				}

				if !strings.Contains(item, "{#") {
					res = append(res, &CheckResult{
						Path: fmt.Sprintf("%s:%d", md, getLineNumber(d, []byte(item))),
						Err:  regSection.info,
						Text: item,
					})
				}
			}
		}
	}

	if co.checkLinks {
		// check external links
		arr := regExternalLink.Regexp.FindAllString(string(d), -1)
		if len(arr) > 0 {
			for _, item := range arr {
				if urlExcluded(item, excludeExternalURLs) {
					continue
				}

				if strings.Contains(item, `{:target="_blank"}`) {
					continue
				}

				if co.autofix {
					fix = regExternalLink.ReplaceAllString(string(d), regExternalLink.replace)
				}

				res = append(res, &CheckResult{
					Path: fmt.Sprintf("%s:%d", md, getLineNumber(d, []byte(item))),
					Err:  regExternalLink.info,
					Text: item,
				})
			}
		}
	}

	return res, fix
}

func urlExcluded(url string, excluded []string) bool {
	for _, ex := range excluded {
		if strings.Contains(url, ex) {
			return true
		}
	}

	return false
}
