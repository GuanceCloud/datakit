// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	MarkdownMatch = ""

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
	hanENMix = regexp.MustCompile("[a-zA-Z0-9_`]+?[\\p{Han}]|[\\p{Han}]+[`_a-zA-Z0-9]+?")
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
		x, err := ioutil.ReadFile(filepath.Clean(md))
		if err != nil {
			l.Errorf("ReadFile: %s", err)
			continue
		}

		data := string(x)

		// regex find line-by-line
		for i, d := range strings.Split(data, "\n") {
			if len(d) == 0 {
				continue
			}

			arr := doMatch(d)

			if len(arr) > 0 {
				fmt.Printf("%s line %03d: %d matched =>\n",
					md,
					i+1, // file line starts from 1
					len(arr))

				totalMatched += len(arr)
				for _, matched := range arr {
					fmt.Printf("\t%q\n", matched)
				}
			}
		}
	}

	return totalMatched
}

func doMatch(d string) []string {
	return hanENMix.FindAllString(d, -1)
}
