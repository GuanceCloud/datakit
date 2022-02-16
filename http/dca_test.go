package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const TOKEN string = "123456"

func getResponse(req *http.Request, config *DCAConfig) *httptest.ResponseRecorder {
	dcaConfig = &DCAConfig{}
	if config != nil {
		dcaConfig = config
	}
	dw = &dataway.DataWayCfg{URLs: []string{"http://localhost:9529?token=123456"}}
	dw.Apply() //nolint: errcheck
	router := setupDcaRouter()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func getResponseBody(w *httptest.ResponseRecorder) (*dcaResponse, error) {
	res := &dcaResponse{}
	err := json.Unmarshal(w.Body.Bytes(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func TestCors(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	w := getResponse(req, nil)
	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Headers"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Credentials"))
	assert.NotEmpty(t, w.Header().Values("Access-Control-Allow-Methods"))
}

type TestCase struct {
	Title          string
	Method         string
	URL            string
	Header         map[string]string
	IsCorrectToken bool
	IsLoopback     bool
	RemoteAddr     string
	Expected       *dcaResponse
	ExpectedNot    *dcaResponse
	DcaConfg       *DCAConfig
	SubCases       []TestCase
}

func runTestCases(t *testing.T, cases []TestCase) {
	t.Helper()
	for _, tc := range cases {
		title := "test"
		if len(tc.Title) > 0 {
			title = tc.Title
		}
		t.Run(title, func(t *testing.T) {
			if len(tc.SubCases) > 0 {
				runTestCases(t, tc.SubCases)
				return
			}
			method := "GET"
			url := "/"
			if len(tc.Method) > 0 {
				method = tc.Method
			}
			if len(tc.URL) > 0 {
				url = tc.URL
			}

			req, _ := http.NewRequest(method, url, nil)

			for k, v := range tc.Header {
				req.Header.Add(k, v)
			}

			if tc.IsCorrectToken {
				req.Header.Add("X-Token", TOKEN)
			}

			if len(tc.RemoteAddr) > 0 {
				req.RemoteAddr = tc.RemoteAddr
			}

			if tc.IsLoopback {
				req.RemoteAddr = "127.0.0.1:10000"
			}

			dcaConfig := &DCAConfig{}
			if tc.DcaConfg != nil {
				dcaConfig = tc.DcaConfg
			}

			w := getResponse(req, dcaConfig)

			res, _ := getResponseBody(w)

			if tc.Expected != nil {
				if tc.Expected.Code > 0 {
					assert.Equal(t, tc.Expected.Code, res.Code)
				}

				if len(tc.Expected.ErrorCode) > 0 {
					assert.Equal(t, tc.Expected.ErrorCode, res.ErrorCode)
				}
			}

			if tc.ExpectedNot != nil {
				if tc.ExpectedNot.Code > 0 {
					assert.NotEqual(t, tc.ExpectedNot.Code, res.Code)
				}

				if len(tc.ExpectedNot.ErrorCode) > 0 {
					assert.NotEqual(t, tc.ExpectedNot.ErrorCode, res.ErrorCode)
				}
			}
		})
	}
}

func TestDca(t *testing.T) {
	testCases := []TestCase{
		{
			Title:          "test default route",
			URL:            "/invalid/url",
			IsCorrectToken: true,
			IsLoopback:     true,
			Expected:       &dcaResponse{Code: 404, ErrorCode: "route.not.found"},
		},
		{
			Title: "test dcaAuthMiddleware",
			SubCases: []TestCase{
				{
					Title:          "correct token",
					IsCorrectToken: true,
					IsLoopback:     true,
					ExpectedNot:    &dcaResponse{ErrorCode: "auth.failed", Code: 401},
				},
				{
					Title:          "wrong token",
					IsCorrectToken: false,
					IsLoopback:     true,
					Expected:       &dcaResponse{ErrorCode: "auth.failed", Code: 401},
				},
			},
		},
		{
			Title: "white list",
			SubCases: []TestCase{
				{
					Title:    "white list check error",
					DcaConfg: &DCAConfig{WhiteList: []string{"111.111.111.111"}},
					Expected: &dcaResponse{Code: 401, ErrorCode: "whiteList.check.error"},
				},
				{
					Title:       "ignore loopback ip",
					IsLoopback:  true,
					ExpectedNot: &dcaResponse{ErrorCode: "whiteList.check.error"},
				},
				{
					Title:       "client ip in whitelist",
					RemoteAddr:  "111.111.111.111:10000",
					DcaConfg:    &DCAConfig{WhiteList: []string{"111.111.111.111"}},
					ExpectedNot: &dcaResponse{ErrorCode: "whiteList.check.error"},
				},
			},
		},
		{
			Title: "api test",
			SubCases: []TestCase{
				{
					Title: "dcaStats",
					URL:   "/v1/dca/stats",
				},
			},
		},
	}

	runTestCases(t, testCases)
}

func TestDcaStats(t *testing.T) {
	req, _ := http.NewRequest("GET", "/v1/dca/stats", nil)
	req.Header.Add("X-Token", TOKEN)
	hostName := "XXXXX"

	// mock
	dcaAPI.GetStats = func() (*DatakitStats, error) {
		return &DatakitStats{HostName: hostName}, nil
	}

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.Equal(t, 200, res.Code)
	content, ok := res.Content.(map[string]interface{})
	assert.True(t, ok)
	hostNameValue, ok := content["hostname"]
	assert.True(t, ok)
	assert.Equal(t, hostName, hostNameValue)
}

func TestDcaInputDoc(t *testing.T) {
	mdContent := "this is demo content"
	dcaAPI.GetMarkdownContent = func(s string) ([]byte, error) {
		return []byte(mdContent), nil
	}

	// no query parameter "inputName"
	req, _ := http.NewRequest("GET", "/v1/dca/inputDoc", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.Equal(t, 500, res.Code)

	// has "inputName"
	req, _ = http.NewRequest("GET", "/v1/dca/inputDoc?inputName=elasticsearch", nil)
	req.Header.Add("X-Token", TOKEN)

	w = getResponse(req, nil)
	res, _ = getResponseBody(w)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, mdContent, res.Content)
}

func TestDcaReload(t *testing.T) {
	// reload ok
	dcaAPI.RestartDataKit = func() error {
		return nil
	}

	req, _ := http.NewRequest("GET", "/v1/dca/reload", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.Equal(t, 200, res.Code)

	// reload fail
	dcaAPI.RestartDataKit = func() error {
		return errors.New("restart error")
	}

	w = getResponse(req, nil)
	res, _ = getResponseBody(w)
	assert.Equal(t, 500, res.Code)
	assert.Equal(t, "system.restart.error", res.ErrorCode)
}

func TestDcaSaveConfig(t *testing.T) {
	inputName := "demo-input"
	inputs.ConfigInfo[inputName] = &inputs.Config{}

	confDir := os.TempDir()
	datakit.ConfdDir = confDir
	f, err := ioutil.TempFile(confDir, "new-conf*.conf")
	assert.NoError(t, err)

	defer os.Remove(f.Name()) //nolint: errcheck //nolint: errcheck

	bodyTemplate := `{"path": "%s","config":"%s", "isNew":%s, "inputName": "%s"}`
	config := "[input]"
	body := strings.NewReader(fmt.Sprintf(bodyTemplate, f.Name(), config, "true", inputName))
	req, _ := http.NewRequest("POST", "/v1/dca/saveConfig", body)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)

	res, _ := getResponseBody(w)

	content, ok := res.Content.(map[string]interface{})
	assert.True(t, ok)
	path, ok := content["path"]
	assert.True(t, ok)
	assert.Equal(t, f.Name(), path)

	confContent, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, config, string(confContent))

	configPaths := inputs.ConfigInfo[inputName].ConfigPaths
	assert.Equal(t, 1, len(configPaths))
	assert.Equal(t, &inputs.ConfigPathStat{Loaded: 2, Path: f.Name()}, configPaths[0])
}

func TestGetConfig(t *testing.T) {
	// no path
	req, _ := http.NewRequest("GET", "/v1/dca/getConfig", nil)
	req.Header.Add("X-Token", TOKEN)
	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	assert.False(t, res.Success)

	// invalid path
	req, _ = http.NewRequest("GET", "/v1/dca/getConfig?path=xxxxxxx.conf", nil)
	req.Header.Add("X-Token", TOKEN)
	w = getResponse(req, nil)
	res, _ = getResponseBody(w)

	assert.False(t, res.Success)
	assert.Equal(t, "params.invalid.path_invalid", res.ErrorCode)

	// get config ok
	confDir := os.TempDir()
	datakit.ConfdDir = confDir
	f, err := ioutil.TempFile(confDir, "new-conf*.conf")
	assert.NoError(t, err)
	defer os.Remove(f.Name()) //nolint: errcheck

	config := "[input]"

	err = ioutil.WriteFile(f.Name(), []byte(config), os.ModePerm)
	assert.NoError(t, err)

	req, _ = http.NewRequest("GET", "/v1/dca/getConfig?path="+f.Name(), nil)
	req.Header.Add("X-Token", TOKEN)
	w = getResponse(req, nil)
	res, _ = getResponseBody(w)

	assert.True(t, res.Success)
	assert.Equal(t, config, res.Content)
}

func TestDcaGetPipelines(t *testing.T) {
	pipelineDir, err := ioutil.TempDir(os.TempDir(), "pipeline")
	datakit.PipelineDir = pipelineDir

	defer os.RemoveAll(pipelineDir) //nolint: errcheck
	assert.NoError(t, err)

	f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/v1/dca/pipelines", nil)
	req.Header.Add("X-Token", TOKEN)

	w := getResponse(req, nil)
	res, _ := getResponseBody(w)

	content, ok := res.Content.([]interface{})
	assert.True(t, ok)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, 1, len(content))
	pipelineInfo, ok := content[0].(map[string]interface{})
	assert.True(t, ok)
	fileName, ok := pipelineInfo["fileName"]
	assert.True(t, ok)
	fileDir, ok := pipelineInfo["fileDir"]
	assert.True(t, ok)

	assert.Equal(t, f.Name(), filepath.Join(fmt.Sprintf("%v", fileDir), fmt.Sprintf("%v", fileName)))
}

func TestDcaGetPipelinesDetail(t *testing.T) {
	pipelineDir, err := ioutil.TempDir(os.TempDir(), "pipeline")
	datakit.PipelineDir = pipelineDir

	defer os.RemoveAll(pipelineDir) //nolint: errcheck
	assert.NoError(t, err)

	f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
	assert.NoError(t, err)

	pipelineContent := "this is demo pipeline"
	fileName := filepath.Base(f.Name())

	err = ioutil.WriteFile(f.Name(), []byte(pipelineContent), os.ModePerm)
	assert.NoError(t, err)

	testCases := []struct {
		Title    string
		IsOk     bool
		FileName string
	}{
		{
			Title: "no query parameter `fileName`",
		},
		{
			Title:    "invalid `fileName` format",
			FileName: "xxxxxx",
		},
		{
			Title:    "file not exist",
			FileName: "invalid.p",
		},
		{
			Title:    "get pipeline ok",
			FileName: fileName,
			IsOk:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Title, func(t *testing.T) {
			url := "/v1/dca/pipelines/detail"
			if len(tc.FileName) > 0 {
				url += "?fileName=" + tc.FileName
			}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Add("X-Token", TOKEN)

			w := getResponse(req, nil)
			res, _ := getResponseBody(w)

			assert.Equal(t, tc.IsOk, res.Success)
		})
	}
}

func TestDcaTestPipelines(t *testing.T) {
	testCases := []struct {
		Title        string
		TestPipeline func(string, string) (string, error)
		Body         string
		IsOk         bool
	}{
		{
			Title: "test ok",
			IsOk:  true,
		},
		{
			Title:        "test pipeline failed",
			TestPipeline: func(s1, s2 string) (string, error) { return "", errors.New("pipeline error") },
		},
		{
			Title: "invalid body",
			Body:  "xxxxxxx",
		},
		{
			Title: "invalid body",
			Body:  `{"fileName": "xxxx"}`,
		},
		{
			Title: "invalid body",
			Body:  `{"text": "xxxx"}`,
		},
	}
	for _, tc := range testCases {
		parsedPipeline := "parse text"
		if tc.TestPipeline != nil {
			dcaAPI.TestPipeline = tc.TestPipeline
		} else {
			dcaAPI.TestPipeline = func(s1, s2 string) (string, error) {
				return parsedPipeline, nil
			}
		}
		pipelineDir, err := ioutil.TempDir(os.TempDir(), "pipeline")
		datakit.PipelineDir = pipelineDir

		defer os.RemoveAll(pipelineDir) //nolint: errcheck
		assert.NoError(t, err)

		f, err := ioutil.TempFile(pipelineDir, "pipeline*.p")
		assert.NoError(t, err)

		pipelineContent := "this is demo pipeline"
		fileName := filepath.Base(f.Name())

		err = ioutil.WriteFile(f.Name(), []byte(pipelineContent), os.ModePerm)
		assert.NoError(t, err)

		var body *strings.Reader
		if len(tc.Body) > 0 {
			body = strings.NewReader(tc.Body)
		} else {
			bodyTemplate := `{"fileName": "%s", "text": "%s"}`
			body = strings.NewReader(fmt.Sprintf(bodyTemplate, fileName, pipelineContent))
		}
		req, _ := http.NewRequest("POST", "/v1/dca/pipelines/test", body)
		req.Header.Add("X-Token", TOKEN)

		w := getResponse(req, nil)
		res, _ := getResponseBody(w)

		if tc.IsOk {
			assert.True(t, res.Success)
		} else {
			assert.False(t, res.Success)
		}
	}
}

func TestDcaCreatePipeline(t *testing.T) {
	testCases := []struct {
		Title          string
		IsOk           bool
		FileName       string
		IsContentEmpty bool
		Body           string
	}{
		{
			Title: "create ok",
			IsOk:  true,
		},
		{
			Title: "invalid body format",
			Body:  "invalid",
		},
		{
			Title:          "content is empty",
			IsContentEmpty: true,
		},
		{
			Title:    "invalid fileName",
			FileName: "pipeline.p",
		},
	}

	for _, tc := range testCases {
		pipelineDir, err := ioutil.TempDir(os.TempDir(), "pipeline")
		datakit.PipelineDir = pipelineDir

		defer os.RemoveAll(pipelineDir) //nolint: errcheck
		assert.NoError(t, err)

		pipelineContent := "this is demo pipeline"

		if tc.IsContentEmpty {
			pipelineContent = ""
		}

		fileName := "custom_pipeline.p"

		if len(tc.FileName) > 0 {
			fileName = tc.FileName
		}

		var body *strings.Reader
		if len(tc.Body) > 0 {
			body = strings.NewReader(tc.Body)
		} else {
			bodyTemplate := `{"fileName":"%s", "fileDir": "%s","content": "%s"}`
			body = strings.NewReader(fmt.Sprintf(bodyTemplate, fileName, pipelineDir, pipelineContent))
		}
		req, _ := http.NewRequest("POST", "/v1/dca/pipelines", body)
		req.Header.Add("X-Token", TOKEN)

		w := getResponse(req, nil)
		res, _ := getResponseBody(w)

		if tc.IsOk {
			assert.True(t, res.Success)
		} else {
			assert.False(t, res.Success)
		}
	}
}

func TestDcaUploadSourcemap(t *testing.T) {
	datakit.DataDir = os.TempDir()
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

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appId := tc.appId
			env := tc.env
			version := tc.version

			url := fmt.Sprintf("/v1/rum/sourcemap?app_id=%s&env=%s&version=%s", appId, env, version)
			var body io.Reader
			data := ""
			boundary := "file"
			if len(tc.fileContent) > 0 {
				data += "--" + boundary + "\n"
				data += "Content-Disposition: form-data; name=\"file\";filename=\"1.zip\"\n\n"
				data += tc.fileContent + "\n"
				data += "--" + boundary + "--"

				body = strings.NewReader(data)
			}

			req, _ := http.NewRequest("POST", url, body)
			req.Header.Add("X-Token", TOKEN)
			if body != nil {
				req.Header.Add("Content-Type", "multipart/form-data;boundary="+boundary)
			}

			w := getResponse(req, nil)
			res, _ := getResponseBody(w)

			assert.Equal(t, tc.isOk, res.Success)
		})
	}
}

func TestDcaDeleteSourcemap(t *testing.T) {
	datakit.DataDir = "./"

	appId := "app_1234"
	env := "test"
	version := "0.0.0"
	GetSourcemapZipFileName(appId, env, version)
	zipFilePath := filepath.Clean(filepath.Join(GetRumSourcemapDir(), GetSourcemapZipFileName(appId, env, version)))

	if err := os.MkdirAll(filepath.Dir(zipFilePath), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(filepath.Dir(zipFilePath))

	err := ioutil.WriteFile(filepath.Join(zipFilePath), []byte(""), os.ModePerm)
	if err != nil {
		t.Fatal("create temp zip file failed", err)
	}

	testCases := []struct {
		title   string
		appId   string
		env     string
		version string
		isOk    bool
	}{
		{
			title:   "delete ok",
			appId:   appId,
			env:     env,
			version: version,
			isOk:    true,
		},
		{
			title:   "param missing",
			env:     "test",
			version: "0.0.0",
			isOk:    false,
		},
		{
			title:   "invalid file path",
			appId:   "invalid_app",
			env:     "test",
			version: "0.0.0",
			isOk:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			appId := tc.appId
			env := tc.env
			version := tc.version

			url := fmt.Sprintf("/v1/rum/sourcemap?app_id=%s&env=%s&version=%s", appId, env, version)

			req, _ := http.NewRequest("DELETE", url, nil)
			req.Header.Add("X-Token", TOKEN)

			w := getResponse(req, nil)
			res, _ := getResponseBody(w)

			assert.Equal(t, tc.isOk, res.Success)
		})
	}
}
