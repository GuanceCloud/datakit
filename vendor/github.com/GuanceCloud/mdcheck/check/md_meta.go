// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package check

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// checkMeta checks if markdown source get expected headers:
//
// ---
// title: 'title name'
// summary: 'summary text'
// __int_icon: 'dir/to/icon'
// dashboard:
//   - desc: 'dashboard name'
//     path: 'dir/to/dashboard-json'
// monitor:
//   - desc: 'monitor name'
//     path: 'dir/to/monitor-json'
// ---.
func checkMarkdownMeta(file string, metaDataDir string) []*CheckResult {
	source, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		log.Printf("[E] ReadFile: %s", err)
		return nil
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			meta.New(
				meta.WithStoresInDocument(),
			),
			extension.GFM,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	node := md.Parser().Parse(text.NewReader(source))
	metaData := node.OwnerDocument().Meta()

	delete(metaData, "icon") // the icon used for side bar

	if len(metaData) == 0 {
		// no meta expected
		return nil
	}

	return checkMeta(file, metaData, metaDataDir)
}

func checkMeta(md string, meta map[string]interface{}, metaDataDir string) (res []*CheckResult) {
	for _, k := range requiredKeys {
		if _, ok := meta[k]; !ok {
			res = append(res, &CheckResult{
				Path: md,
				Err:  fmt.Sprintf("missing meta %q", k),
			})
			return
		}
	}

	for k, v := range meta {
		switch k {
		case "title":
			if x, ok := v.(string); !ok || len(x) == 0 {
				// TODO: should we add max-length check?
				res = append(res, &CheckResult{
					Path: md,
					Err:  fmt.Sprintf("invalid title %q", k),
				})
			}
		case "summary":
			if x, ok := v.(string); !ok || len(x) == 0 {
				// TODO: should we add max-length check?
				res = append(res, &CheckResult{
					Path: md,
					Err:  fmt.Sprintf("missing summary %q", k),
				})
			}

		case "__int_icon":
			x, ok := v.(string)
			if !ok || len(x) == 0 {
				res = append(res, &CheckResult{
					Path: md,
					Err:  "invalid __int_icon",
				})
			}

			if err := fileExist(
				filepath.Join(metaDataDir, x, "icon.png"),
				filepath.Join(metaDataDir, x, "icon-dark.png"),
			); err != nil {
				res = append(res, &CheckResult{
					Path: md,
					Err:  fmt.Sprintf("icon: %s", err.Error()),
				})
			}

		case "monitor", "dashboard": // they are the same
			// arr: [map{"desc": "...", "path": "..."}, {}, {}]
			arr, ok := v.([]any)
			if !ok || len(arr) == 0 {
				res = append(res, &CheckResult{
					Path: md,
					Err:  fmt.Sprintf("%s not found: %q, type: %s", k, v, reflect.TypeOf(v)),
				})
			}

			for _, elem := range arr {
				dict, ok := elem.(map[any]any)
				if !ok {
					res = append(res, &CheckResult{
						Path: md,
						Err:  fmt.Sprintf("invalid %s, expect map, got %s", k, reflect.TypeOf(elem)),
					})
				}

				var path string
				for subKey, v := range dict {
					switch subKey.(string) {
					case "desc": // pass
					case "path":
						path = v.(string)
					default:
						res = append(res, &CheckResult{
							Path: md,
							Err:  fmt.Sprintf("invalid %s key %q", k, subKey),
						})
					}
				}

				if path == "-" {
					log.Printf("[W] %s not set", k)
					continue
				}

				if err := fileExist(
					filepath.Join(metaDataDir, path, "meta.json"),
				); err != nil {
					res = append(res, &CheckResult{
						Path: md,
						Err:  fmt.Sprintf("invalid %s json: %s", k, err.Error()),
					})
				}

				// both dashboard and monitor are JSON
				if err := jsonOK(
					filepath.Join(metaDataDir, path, "meta.json"),
				); err != nil {
					res = append(res, &CheckResult{
						Path: md,
						Err:  fmt.Sprintf("invalid %s json: %s", k, err.Error()),
					})
				}
			}
		}
	}

	return res
}

var requiredKeys = []string{
	// icon used only with integrations, to avoid key conflict with mkdocs key `icon'
	"__int_icon",
	"dashboard",
	"monitor",
	"summary",
	"title",
}

// fileExist check if files exist and are nomal file, not dir or else.
func fileExist(files ...string) error {
	for _, f := range files {
		if fi, err := os.Stat(f); err != nil {
			return fmt.Errorf("file %q not exist", f)
		} else if fi.IsDir() {
			return fmt.Errorf("invalid file %q, expect file, got dir", f)
		}
	}

	return nil
}

func jsonOK(files ...string) error {
	for _, f := range files {
		if j, err := ioutil.ReadFile(filepath.Clean(f)); err != nil {
			return err
		} else if !json.Valid(j) {
			return fmt.Errorf("invalid json: %s", f)
		}
	}

	return nil
}
