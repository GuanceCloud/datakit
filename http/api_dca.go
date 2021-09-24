package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/toml"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var dcaErrorMessage = map[string]string{
	"server.error": "server error",
}

func getBody(c *gin.Context, data interface{}) error {
	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, data)

	if err != nil {
		return err
	}

	return nil
}

func dcaGetMessage(errCode string) string {
	if errMsg, ok := dcaErrorMessage[errCode]; ok {
		return errMsg
	} else {
		return "server error"
	}
}

type dcaResponse struct {
	Code      int         `json:"code"`
	Content   interface{} `json:"content"`
	ErrorCode string      `json:"errorCode"`
	ErrorMsg  string      `json:"errorMsg"`
	Success   bool        `json:"success"`
}

type dcaError struct {
	Code      int
	ErrorCode string
	ErrorMsg  string
}

type dcaContext struct {
	c    *gin.Context
	data interface{}
}

func (d *dcaContext) send(response dcaResponse) {
	body, err := json.Marshal(response)
	if err != nil {
		d.fail(dcaError{})
		return
	}
	d.c.Data(http.StatusOK, "application/json", body)
}

func (d *dcaContext) success(datas ...interface{}) {
	var data interface{}

	if len(datas) > 0 {
		data = datas[0]
	}

	if data == nil {
		data = d.data
	}

	response := dcaResponse{
		Code:    200,
		Content: data,
		Success: true,
	}

	d.send(response)
}

func (d *dcaContext) fail(errors ...dcaError) {
	var e dcaError
	if len(errors) > 0 {
		e = errors[0]
	} else {
		e = dcaError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "server.error",
			ErrorMsg:  "",
		}
	}

	code := e.Code
	errorCode := e.ErrorCode
	errorMsg := e.ErrorMsg

	if code == 0 {
		code = http.StatusInternalServerError
	}

	if errorCode == "" {
		errorCode = "server.error"
	}

	if errorMsg == "" {
		errorMsg = dcaGetMessage(errorCode)
	}

	response := dcaResponse{
		Code:      code,
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
		Success:   false,
	}

	d.send(response)
}

// dca reload
func dcaReload(c *gin.Context) {
	context := getContext(c)
	context.success()

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s",
			runtime.GOOS, err.Error())
		return
	}

	l.Info("new datakit servier ok...")

	if err := service.Control(svc, "restart"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}
}

func dcaStats(c *gin.Context) {
	s, err := GetStats()
	context := dcaContext{c: c}

	if err != nil {
		context.fail()
		return
	}

	context.success(s)
}

type saveConfigParam struct {
	Path      string `json:"path"`
	Config    string `json:"config"`
	IsNew     bool   `json:"isNew"`
	InputName string `json:"inputName"`
}

// auth middleware
func dcaAuthMiddleware(f func(*gin.Context)) func(*gin.Context) {
	return func(c *gin.Context) {
		tokens := c.Request.Header["X-Token"]
		context := &dcaContext{c: c}
		if len(tokens) == 0 {
			context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
			return
		}

		token := tokens[0]
		localTokens := dw.GetToken()
		if len(token) == 0 || len(localTokens) == 0 || (token != localTokens[0]) {
			context.fail(dcaError{Code: 401, ErrorCode: "auth.failed", ErrorMsg: "auth failed"})
			return
		}
		f(c)
	}
}

func dcaGetConfig(c *gin.Context) {
	context := getContext(c)
	path := c.Query("path")

	if errMsg, err := checkPath(path); err != nil {
		context.fail(errMsg)
		return
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		l.Error(err)
		context.fail(dcaError{ErrorCode: "invalid.path", ErrorMsg: "invalid path"})
		return
	}
	context.success(string(content))
}

// save config
func dcaSaveConfig(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	context := &dcaContext{c: c}
	if err != nil {
		l.Error(err)
		context.fail()
		return
	}

	param := saveConfigParam{}

	if err := json.Unmarshal(body, &param); err != nil {
		l.Error(err)
		context.fail()
		return
	}

	if errMsg, err := checkPath(param.Path); err != nil {
		context.fail(errMsg)
		return
	}

	configContent := []byte(param.Config)

	// add new config
	if param.IsNew {
		if _, err := os.Stat(param.Path); err == nil { // exist
			var content []byte
			var err error

			if content, err = ioutil.ReadFile(param.Path); err != nil {
				l.Error(err)
				context.fail()
				return
			}
			configContent = append(content, configContent...)
		}
	}

	// check toml
	_, err = toml.Parse(configContent)
	if err != nil {
		l.Errorf("parse toml %s failed", string(configContent))
		context.fail(dcaError{ErrorCode: "toml.format.error", ErrorMsg: "toml format error"})
		return
	}

	// create and save
	err = ioutil.WriteFile(param.Path, configContent, datakit.ConfPerm)
	if err != nil {
		l.Error(err)
		context.fail()
		return
	}

	// update configInfo
	if len(param.InputName) > 0 {
		if c, ok := inputs.ConfigInfo[param.InputName]; ok { // update
			existed := false
			for _, configPath := range c.ConfigPaths {
				if configPath.Path == param.Path {
					existed = true
					configPath.Loaded = int8(2)
					break
				}
			}
			if !existed {
				c.ConfigPaths = append(c.ConfigPaths, &inputs.ConfigPathStat{
					Loaded: int8(2),
					Path:   param.Path,
				})
			}

		} else { // add new info
			if creator, ok := inputs.Inputs[param.InputName]; ok {
				inputs.ConfigInfo[param.InputName] = &inputs.Config{
					ConfigPaths: []*inputs.ConfigPathStat{
						{
							Loaded: int8(2),
							Path:   param.Path,
						},
					},
					SampleConfig: creator().SampleConfig(),
					Catalog:      creator().Catalog(),
					ConfigDir:    datakit.ConfdDir,
				}
			}
		}
	}

	context.success(map[string]string{
		"path": param.Path,
	})
}

func dcaInputDoc(c *gin.Context) {
	context := getContext(c)
	inputName := c.Query("inputName")
	l.Debug("api_doc", inputName)
	if inputName == "" {
		context.fail()
		return
	}

	md, err := man.BuildMarkdownManual(inputName, &man.Option{})
	if err != nil {
		l.Error(err)
		context.fail(dcaError{ErrorCode: "record.not.exist", ErrorMsg: "record not exist", Code: http.StatusNotFound})
		return
	}

	context.success(string(md))
}

func getContext(c *gin.Context) dcaContext {
	return dcaContext{c: c}
}

func checkPath(path string) (dcaError, error) {
	err := errors.New("invalid conf")

	// path should under conf.d
	dir := filepath.Dir(path)

	pathReg := regexp.MustCompile(`\.conf$`)

	if pathReg == nil {
		return dcaError{}, err
	}

	// check path
	if !strings.Contains(path, datakit.ConfdDir) || !pathReg.Match([]byte(path)) {
		return dcaError{ErrorCode: "params.invalid.path_invalid", ErrorMsg: "invalid param 'path'"}, err
	}

	// check dir
	if _, err := os.Stat(dir); err != nil {
		return dcaError{ErrorCode: "params.invalid.dir_not_exist", ErrorMsg: "dir not exist"}, err
	}

	return dcaError{}, nil
}

func isValidPipelineFileName(name string) bool {
	pipelineFileRegxp := regexp.MustCompile(`.+\.p$`)

	return pipelineFileRegxp.Match([]byte(name))
}

func isValidCustomPipelineName(name string) bool {
	pipelineFileRegxp := regexp.MustCompile(`^custom_.+\.p$`)

	return pipelineFileRegxp.Match([]byte(name))
}

type pipelineInfo struct {
	FileName string `json:"fileName"`
	FileDir  string `json:"fileDir"`
	Content  string `json:"content"`
}

func dcaGetPipelines(c *gin.Context) {
	context := getContext(c)

	allFiles, err := ioutil.ReadDir(datakit.PipelineDir)
	if err != nil {
		context.fail()
		return
	}

	pipelines := []pipelineInfo{}

	// filter pipeline file
	for _, file := range allFiles {
		if !file.IsDir() {
			name := file.Name()
			if isValidPipelineFileName(name) {
				pipelines = append(pipelines, pipelineInfo{FileName: name, FileDir: datakit.PipelineDir})
			}
		}
	}

	context.success(pipelines)
}

func dcaGetPipelinesDetail(c *gin.Context) {
	context := getContext(c)
	fileName := c.Query("fileName")
	if len(fileName) == 0 {
		context.fail(dcaError{ErrorCode: "params.required", ErrorMsg: fmt.Sprintf("param %s is required", "fileName")})
		return
	}

	if !isValidPipelineFileName(fileName) {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	path := filepath.Join(datakit.PipelineDir, fileName)

	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		l.Error(err)
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	context.success(string(contentBytes))
}

func dcaCreatePipeline(c *gin.Context) {
	dcaSavePipeline(c, false)
}

func dcaUpdatePipeline(c *gin.Context) {
	dcaSavePipeline(c, true)
}

func dcaSavePipeline(c *gin.Context, isUpdate bool) {
	var filePath string
	var fileName string
	context := getContext(c)

	pipeline := pipelineInfo{}
	err := getBody(c, &pipeline)
	if err != nil {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "make sure parameter is in correct format"})
		return
	}

	fileName = pipeline.FileName
	filePath = filepath.Join(datakit.PipelineDir, fileName)

	if len(pipeline.Content) == 0 {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "'content' should not be empty"})
		return
	}

	// update pipeline
	if isUpdate {
		if !path.IsFileExists(filePath) {
			context.fail(dcaError{ErrorCode: "file.not.exist", ErrorMsg: "file not exist"})
			return
		}
	} else { // new pipeline
		if path.IsFileExists(filePath) {
			context.fail(dcaError{ErrorCode: "param.invalid.duplicate", ErrorMsg: fmt.Sprintf("current name '%s' is duplicate", pipeline.FileName)})
			return
		}
	}

	// only save or update custom pipeline!
	if !isValidCustomPipelineName(fileName) {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "fileName is not valid custom pipeline name"})
		return
	}

	err = ioutil.WriteFile(filePath, []byte(pipeline.Content), datakit.ConfPerm)

	if err != nil {
		l.Error(err, filePath)
		context.fail()
		return
	}
	pipeline.FileDir = datakit.PipelineDir
	context.success(pipeline)
}

func pipelineTest(pipelineFile string, text string) (string, error) {
	if err := pipeline.Init(datakit.DataDir); err != nil {
		return "", err
	}

	pl, err := pipeline.NewPipelineFromFile(filepath.Join(datakit.PipelineDir, pipelineFile))
	if err != nil {
		return "", err
	}

	res, err := pl.Run(text).Result()
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		l.Debug("No data extracted from pipeline")
		return "", nil
	}

	if j, err := json.Marshal(res); err != nil {
		return "", err
	} else {
		return string(j), nil
	}
}

func dcaTestPipelines(c *gin.Context) {
	context := getContext(c)

	body := map[string]string{}

	err := getBody(c, &body)
	if err != nil {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter format error"})
		return
	}

	text, ok := body["text"]

	if !ok {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter 'text' is required"})
		return
	}

	fileName, ok := body["fileName"]
	if !ok {
		context.fail(dcaError{ErrorCode: "param.invalid", ErrorMsg: "parameter 'fileName' is required"})
		return
	}

	parsedText, err := pipelineTest(fileName, text)
	if err != nil {
		l.Error(err)
		context.fail(dcaError{ErrorCode: "pipeline.parse.error", ErrorMsg: err.Error()})
		return
	}

	context.success(parsedText)
}
