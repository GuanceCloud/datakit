// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestHandleSourcemapUpload(t *testing.T) {
	const Token = "xxxxxxxxxx"

	dw := &dataway.Dataway{URLs: []string{"http://localhost:9529?token=" + Token}}
	err := dw.Init()
	assert.NoError(t, err)
	config.Cfg.Dataway = dw

	dir, err := ioutil.TempDir("./", "tmp")
	if err != nil {
		t.Fatal("create tmp dir eror")
	}

	datakit.DataDir = dir
	datakit.DataRUMDir = filepath.Join(dir, "rum")
	defer os.RemoveAll(dir) //nolint: errcheck

	testCases := []struct {
		title       string
		appId       string
		env         string
		version     string
		fileContent string
		isOk        bool
	}{
		{
			title:       "upload ok",
			appId:       "app_1234",
			env:         "test",
			version:     "0.0.0",
			fileContent: "xxxxxx",
			isOk:        true,
		},
		{
			title:       "param missing",
			env:         "test",
			version:     "0.0.0",
			fileContent: "xxxxxx",
			isOk:        false,
		},
		{
			title:   "file missing",
			appId:   "app_123",
			env:     "test",
			version: "0.0.0",
			isOk:    false,
		},
	}

	ipt := defaultInput()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipt.handleSourcemapUpload(w, r)
	}))
	defer server.Close()

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appId := tc.appId
			env := tc.env
			version := tc.version

			url := fmt.Sprintf("%s/v1/rum/sourcemap?app_id=%s&env=%s&version=%s&platform=web&token=%s", server.URL, appId, env, version, Token)
			var body io.Reader
			boundary := "file"
			if len(tc.fileContent) > 0 {
				bodyBytes, err := createZipFileBytes(tc.fileContent)
				assert.NoError(t, err)

				data := &bytes.Buffer{}
				data.WriteString("--" + boundary + "\n")
				data.WriteString("Content-Disposition: form-data; name=\"file\";filename=\"1.zip\"\n\n")
				_, err = data.Write(bodyBytes)
				assert.NoError(t, err)
				data.WriteString("\n")
				data.WriteString("--" + boundary + "--")

				body = data
			}

			client := http.Client{}
			req, _ := http.NewRequest("POST", url, body)
			if body != nil {
				req.Header.Add("Content-Type", "multipart/form-data;boundary="+boundary)
			}

			resp, err := client.Do(req)

			assert.NoError(t, err)
			defer resp.Body.Close()

			bodyBytes, err := ioutil.ReadAll(resp.Body)

			assert.NoError(t, err)

			resBody := &sourcemapResponse{}
			err = json.Unmarshal(bodyBytes, resBody)
			assert.NoError(t, err)

			assert.Equal(t, tc.isOk, resBody.Success)
		})
	}
}

func createZipFileBytes(content string) ([]byte, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sourcemap")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	file, err := os.Create(filepath.Join(tmpDir, "1.zip"))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	wr := zip.NewWriter(file)

	f, err := wr.Create("1.txt")
	if err != nil {
		return nil, err
	}

	_, err = f.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	wr.Close()

	file.Seek(0, 0)

	return ioutil.ReadAll(file)
}

func TestHandleSourcemapCheck(t *testing.T) {
	var (
		appID      = "app_123"
		env        = "test"
		version    = "0.0.0"
		sdkName    = ""
		errorStack = url.QueryEscape("at <anonymous> @ http://localhost:5500/dist/bundle.js:1:821")
	)

	ipt := defaultInput()
	tmpDir := t.TempDir()

	ipt.rumDataDir = tmpDir
	ipt.initConfig()

	mapfile, err := ioutil.ReadFile("testdata/mapfile.json")
	assert.NoError(t, err)

	datakit.DataDir = path.Join(tmpDir, "data")

	sourcemapFileName := GetSourcemapZipFileName(appID, env, version)

	rumDir := ipt.getRumSourcemapDir(SdkWeb)

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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipt.handleSourcemapCheck(w, r)
	}))
	defer server.Close()

	testCases := []struct {
		title      string
		appId      string
		env        string
		version    string
		sdkName    string
		errorStack string
		isOk       bool
	}{
		{
			title:      "check ok",
			appId:      appID,
			env:        env,
			version:    version,
			sdkName:    sdkName,
			errorStack: errorStack,
			isOk:       true,
		},
		{
			title:      "check fail",
			appId:      appID,
			env:        "invalid_env",
			version:    version,
			sdkName:    sdkName,
			errorStack: errorStack,
			isOk:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appID := tc.appId
			env := tc.env
			version := tc.version
			sdkName := tc.sdkName
			errorStack := tc.errorStack

			url := fmt.Sprintf("%s/v1/sourcemap/check?app_id=%s&env=%s&version=%s&sdk_name=%s&error_stack=%s", server.URL, appID, env, version, sdkName, errorStack)

			client := http.Client{}
			req, _ := http.NewRequest("GET", url, nil)

			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			resBody := &sourcemapResponse{}
			err = json.Unmarshal(bodyBytes, resBody)
			assert.NoError(t, err)

			assert.Equal(t, tc.isOk, resBody.Success)
		})
	}
}

func TestHandleSourcemapDelete(t *testing.T) {
	const Token = "xxxxxxxxxxxxxxxxxxxxxxxxx"
	var (
		appID    = "app_123"
		env      = "test"
		version  = "0.0.0"
		platform = "web"
	)

	dw := &dataway.Dataway{URLs: []string{"http://localhost:9529?token=" + Token}}
	err := dw.Init()
	assert.NoError(t, err)
	config.Cfg.Dataway = dw

	ipt := defaultInput()
	tmpDir := t.TempDir()

	ipt.rumDataDir = tmpDir
	ipt.initConfig()

	mapfile, err := ioutil.ReadFile("testdata/mapfile.json")
	assert.NoError(t, err)

	datakit.DataDir = path.Join(tmpDir, "data")

	sourcemapFileName := GetSourcemapZipFileName(appID, env, version)

	rumDir := ipt.getRumSourcemapDir(SdkWeb)

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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipt.handleSourcemapDelete(w, r)
	}))
	defer server.Close()

	testCases := []struct {
		title    string
		appId    string
		env      string
		version  string
		platform string
		isOK     bool
	}{
		{
			title:    "delete ok",
			appId:    appID,
			env:      env,
			version:  version,
			platform: platform,
			isOK:     true,
		},
		{
			title:    "delete fail",
			appId:    appID,
			env:      "invalid_env",
			version:  version,
			platform: platform,
			isOK:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appID := tc.appId
			env := tc.env
			version := tc.version
			platform := tc.platform

			url := fmt.Sprintf("%s/v1/sourcemap?app_id=%s&env=%s&version=%s&platform=%s&token=%s", server.URL, appID, env, version, platform, Token)

			client := http.Client{}
			req, _ := http.NewRequest("DELETE", url, nil)

			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err)

			resBody := &sourcemapResponse{}
			err = json.Unmarshal(bodyBytes, resBody)
			assert.NoError(t, err)

			assert.Equal(t, tc.isOK, resBody.Success)
		})
	}
}
