// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-sourcemap/sourcemap"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

func TestHandleSourcemap(t *testing.T) {
	body, err := ioutil.ReadFile("testdata/body")
	assert.NoError(t, err)

	mapfile, err := ioutil.ReadFile("testdata/mapfile.json")
	assert.NoError(t, err)

	tmpDir := t.TempDir()

	// setup input
	ipt := defaultInput()
	ipt.rumDataDir = tmpDir
	ipt.initMeasurementMap()

	datakit.DataDir = path.Join(tmpDir, "data")

	sourcemapFileName := httpapi.GetSourcemapZipFileName(
		"appid_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"production",
		"1.0.0")
	rumDir := ipt.getRumSourcemapDir(SdkWebMiniApp)

	assert.NoError(t, os.MkdirAll(rumDir, os.ModePerm))

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("dist/bundle.js.map")
	assert.NoError(t, err)

	_, err = f.Write(mapfile)
	assert.NoError(t, err)
	assert.NoError(t, w.Close())

	zipFilePath := filepath.Join(rumDir, sourcemapFileName)
	assert.NoError(t, ioutil.WriteFile(zipFilePath, buf.Bytes(), os.ModePerm))
	defer os.Remove(zipFilePath) //nolint:errcheck

	assert.NoError(t, ipt.loadSourcemapFile())

	opts := []point.Option{
		point.WithPrecision(point.MS),
		point.WithTime(time.Now()),
		point.WithCallback(ipt.parseCallback),
	}

	pts, err := httpapi.HandleWriteBody(body, false, opts...)

	assert.NoError(t, err)
	assert.Greater(t, len(pts), 0)

	for _, p := range pts {
		if string(p.Name()) == "error" {
			fields := p.InfluxFields()
			if _, ok := fields["error_stack"]; ok {
				if errorStackSource, ok := fields["error_stack_source_base64"]; !ok {
					assert.Fail(t, "error stack transform failed")
				} else {
					decodBytes, err := base64.StdEncoding.DecodeString(errorStackSource.(string))
					assert.NoError(t, err)
					assert.Contains(t, string(decodBytes), "webpack")
				}
			}
		}
	}

	t.Run("updateSourcemapCache", func(t *testing.T) {
		updateSourcemapCache("invalid")
		assert.Equal(t, len(webSourcemapCache), 1)

		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		_, ok := webSourcemapCache[fileName]
		assert.True(t, ok)

		deleteSourcemapCache(zipFilePath)
		_, ok = webSourcemapCache[fileName]
		assert.False(t, ok)
	})

	t.Run("getSourceMapString", func(t *testing.T) {
		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		cases := []struct {
			src           string
			dest          string
			desc          string
			sourceMapItem map[string]*sourcemap.Consumer
		}{
			{
				src:  "http://a.com/1.js",
				dest: "http://a.com/1.js",
				desc: "split with colon, length less than 3",
			},
			{
				src:  "http://a.com/1.js:a:1",
				dest: "http://a.com/1.js:a:1",
				desc: "invalid row number",
			},
			{
				src:  "http://a.com/1.js:1:a",
				dest: "http://a.com/1.js:1:a",
				desc: "invalid col number",
			},
			{
				src:           "http://localhost:5500/dist/bundle.js:1:821",
				dest:          "webpack:///./src/index.js:17:4",
				desc:          "it shourld work",
				sourceMapItem: webSourcemapCache[fileName],
			},
			{
				src:           "http:///localhost:5500/dist/bundle.js:1:821",
				dest:          "http:///localhost:5500/dist/bundle.js:1:821",
				desc:          "invalid url",
				sourceMapItem: webSourcemapCache[fileName],
			},
			{
				src:  "http://localhost:5500/dist/bundle.js:1:821",
				dest: "http://localhost:5500/dist/bundle.js:1:821",
				desc: "it shourld not work without sourcemap file",
			},
		}

		for _, item := range cases {
			assert.Equal(t, getSourceMapString(item.src, item.sourceMapItem), item.dest, item.desc)
		}
	})
}

func TestScanModuleSymbolFile(t *testing.T) {
	t.Skip("skipped: this case should rewrite")
	file, err := scanModuleSymbolFile("/Users/zy/software/source_map/iOS", "App")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(file)
	}
}

func TestScanIOSCrashAddress(t *testing.T) {
	originCrash, err := ioutil.ReadFile("testdata/origin-crash")
	assert.NoError(t, err)

	address, err := scanIOSCrashAddress(string(originCrash))
	assert.NoError(t, err)

	for moduleName, addresses := range address {
		fmt.Printf("module-name: %s\n", moduleName)
		for start, crashAddresses := range addresses {
			fmt.Printf("start: %+#v\n", start)
			for _, addr := range crashAddresses {
				fmt.Printf("addr: %+#v\n", addr)
			}
		}
	}
}

func TestRunAtosTool(t *testing.T) {
	t.Skip("skipped: this case should rewrite")
	args := []string{
		"-o",
		"/Users/zy/software/source_map/iOS/App.app.dSYM/Contents/Resources/DWARF/App",
		"-l",
		"0x104f30000",
	}

	args = append(args, "0x0000000104fd0728", "0x0000000104fd00bc")

	cmd := exec.Command("/Users/zy/.cargo/bin/atosl", args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("err: ", err)
	}

	fmt.Println(string(output))
}

func TestScanABI(t *testing.T) {
	crash, err := ioutil.ReadFile("testdata/crash")
	assert.NoError(t, err)

	abi := scanABI(string(crash))

	fmt.Println(abi)

	assert.Equal(t, abi, "arm64-v8a")
}
