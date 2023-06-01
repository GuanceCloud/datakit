// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
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
	MarkdownCheck = ""
	Autofix       = "off"

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
	regSection = newRule(`^##+.*`,
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

func Match(dir string) int {
	var mds []string
	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), "md") {
			return nil
		}

		mds = append(mds, path)
		return nil
	}); err != nil {
		l.Errorf("filepath.Walk: %s", err)
		return 0
	}

	totalMatched := 0

	for _, md := range mds {
		fmt.Printf("Checking %s ...\n", md)

		x, err := ioutil.ReadFile(filepath.Clean(md))
		if err != nil {
			l.Errorf("ReadFile: %s", err)
			continue
		}

		arr, fix := doMatch(string(x), Autofix == "on")
		if len(arr) > 0 { // we find some bad message
			totalMatched += len(arr)

			if len(fix) > 0 {
				if err := ioutil.WriteFile(filepath.Clean(md), []byte(fix), 0o600); err != nil {
					panic(err.Error())
				}
			}

			fmt.Printf("--------------------\n%s: mached =>\n", md)
			for _, item := range arr {
				if strings.HasPrefix(item, "!!!invalid") {
					if Autofix == "on" {
						fmt.Printf("%s [auto-fixed]\n", item)
					} else {
						fmt.Printf("%s\n", item)
					}
				} else {
					fmt.Printf("\t%s:%d %q\n", md, getLineNumber(x, []byte(item)), item)
				}
			}
		}
	}

	fmt.Printf("matched %d files, total find %d issues.\n", len(mds), totalMatched)

	// when autofix enabled, we do not fire any error, makes it ok.
	// NOTE: Do not enable autofix during CI, you should autofix errors
	// on your local machine.
	if Autofix == "on" {
		return 0
	} else {
		return totalMatched
	}
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

func doMatch(d string, autofix bool) ([]string, string) {
	var matches []string

	// check punctuations
	for _, r := range punctuationRules {
		arr := r.Regexp.FindAllString(d, -1)
		if len(arr) > 0 {
			if autofix {
				d = r.Regexp.ReplaceAllString(d, r.replace)
			}
			matches = append(matches, r.info)
			matches = append(matches, arr...)
		}
	}

	// check sections
	arr := regSection.Regexp.FindAllString(d, -1)
	if len(arr) > 0 {
		for _, item := range arr {
			if !strings.Contains(item, "{#") {
				matches = append(matches, item)
			}
		}
	}

	// check external links
	arr = regExternalLink.Regexp.FindAllString(d, -1)
	if len(arr) > 0 {
		for _, item := range arr {
			if urlExcluded(item, excludeExternalURLs) {
				continue
			}

			if strings.Contains(item, `{:target="_blank"}`) {
				continue
			}

			if autofix {
				d = regExternalLink.ReplaceAllString(d, regExternalLink.replace)
			}

			matches = append(matches, item)
		}
	}

	return matches, d
}

func urlExcluded(url string, excluded []string) bool {
	for _, ex := range excluded {
		if strings.Contains(url, ex) {
			return true
		}
	}

	return false
}
