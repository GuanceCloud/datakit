// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

// checkMeta checks if markdown source get expected headers:
//
// ---
// title: 'title name'
// summary: 'summary text'
// dashboard:
//   - desc: 'dashboard name'
//     path: 'dir/to/dashboard-json'
// monitor:
//   - desc: 'monitor name'
//     path: 'dir/to/monitor-json'
// ---.
func checkMarkdownMeta(source []byte, metaDataDir string) error {
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

	return checkMeta(metaData, metaDataDir)
}

func checkMeta(meta map[string]interface{}, metaDataDir string) error {
	errMsgs := []string{}

	for _, k := range requiredKeys {
		if _, ok := meta[k]; !ok {
			errMsgs = append(errMsgs, fmt.Sprintf("missing meta %q", k))
		}
	}

	for k, v := range meta {
		switch k {
		case "title":
			if x, ok := v.(string); !ok || len(x) == 0 {
				// TODO: should we add max-length check?
				errMsgs = append(errMsgs, fmt.Sprintf("invalid title %q", k))
			}
		case "summary":
			if x, ok := v.(string); !ok || len(x) == 0 {
				// TODO: should we add max-length check?
				errMsgs = append(errMsgs, fmt.Sprintf("missing summary %q", k))
			}

		case "icon":
			x, ok := v.(string)
			if !ok || len(x) == 0 {
				errMsgs = append(errMsgs, "invalid icon")
			}

			if err := fileExist(
				filepath.Join(metaDataDir, x, "icon.png"),
				filepath.Join(metaDataDir, x, "icon-dark.png"),
			); err != nil {
				// TODO: icon not in current git repo, we ignore the error.
				cp.Warnf("icon: %s, it's ok\n", err.Error())
			}

		case "monitor", "dashboard": // they are the same
			// arr: [map{"desc": "...", "path": "..."}, {}, {}]
			arr, ok := v.([]any)
			if !ok || len(arr) == 0 {
				errMsgs = append(errMsgs, fmt.Sprintf("%s not found: %q, type: %s", k, v, reflect.TypeOf(v)))
			}

			for _, elem := range arr {
				dict, ok := elem.(map[any]any)
				if !ok {
					errMsgs = append(errMsgs, fmt.Sprintf("invalid %s, expect map, got %s", k, reflect.TypeOf(elem)))
				}

				var path string
				for subKey, v := range dict {
					switch subKey.(string) {
					case "desc": // pass
					case "path":
						path = v.(string)
					default:
						errMsgs = append(errMsgs, fmt.Sprintf("invalid %s key %q", k, subKey))
					}
				}

				if path == "-" {
					cp.Warnf("[W] %s not set\n", k)
					continue
				}

				if err := fileExist(
					filepath.Join(metaDataDir, path, "meta.json"),
				); err != nil {
					errMsgs = append(errMsgs, fmt.Sprintf("invalid %s json: %s", k, err.Error()))
				}

				// both dashboard and monitor are JSON
				if err := jsonOK(
					filepath.Join(metaDataDir, path, "meta.json"),
				); err != nil {
					errMsgs = append(errMsgs, fmt.Sprintf("invalid %s json: %s", k, err.Error()))
				}
			}
		}
	}

	if len(errMsgs) > 0 {
		return &mderr{
			level: mdError,
			msg:   strings.Join(errMsgs, "\n"),
		}
	}

	return nil
}

var requiredKeys = []string{
	"dashboard",
	"icon",
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

type mdErrorLevel int

const (
	mdWarn mdErrorLevel = iota
	mdError
)

func (l mdErrorLevel) String() string {
	switch l {
	case mdWarn:
		return "WARN"
	case mdError:
		return "ERROR"
	default:
		return "ERROR"
	}
}

type mderr struct {
	level mdErrorLevel
	msg   string
}

func (e *mderr) Error() string {
	return e.msg
}
