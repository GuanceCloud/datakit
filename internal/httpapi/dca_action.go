// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
	"github.com/gorilla/websocket"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	dk "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var HostActionHandlerMap map[string]ws.ActionHandler

func SetDCALogger(log *logger.Logger) {
	if log != nil {
		l = log
	}
}

func getDatakitStatsAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	if stats, err := getDatakitStats(); err != nil {
		l.Warnf("get datakit stats failed: %s", err.Error())
		response.SetError()
	} else {
		response.SetSuccess(stats)
	}
}

func checkPath(path string) *ws.ResponseError {
	// path should under conf.d
	dir := filepath.Dir(path)

	pathReg := regexp.MustCompile(`\.conf$`)

	if pathReg == nil {
		return &ws.ResponseError{Code: 400, ErrorCode: "params.invalid.path_invalid", ErrorMsg: "invalid param 'path'"}
	}

	// check path
	if !strings.Contains(path, dk.ConfdDir) || !pathReg.Match([]byte(path)) {
		return &ws.ResponseError{ErrorCode: "params.invalid.path_invalid", ErrorMsg: "invalid param 'path'"}
	}

	// check dir
	if _, err := os.Stat(dir); err != nil {
		return &ws.ResponseError{ErrorCode: "params.invalid.dir_not_exist", ErrorMsg: "dir not exist"}
	}

	return nil
}

func doGetConfig(path string) (string, *ws.ResponseError) {
	if err := checkPath(path); err != nil {
		return "", err
	} else if content, err := os.ReadFile(filepath.Clean(path)); err != nil {
		l.Errorf("Read config file %s error: %s", path, err.Error())
		return "", &ws.ResponseError{ErrorCode: "invalid.path", ErrorMsg: "invalid path"}
	} else {
		return string(content), nil
	}
}

func getDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	path := data.Query.Get("path")
	if content, err := doGetConfig(path); err != nil {
		response.SetError(err)
	} else {
		response.SetSuccess(content)
	}
}

type saveConfigParam struct {
	Path      string `json:"path"`
	Config    string `json:"config"`
	IsNew     bool   `json:"isNew"`
	InputName string `json:"inputName"`
}

func doSaveConfig(param *saveConfigParam) *ws.ResponseError {
	if err := checkPath(param.Path); err != nil {
		return err
	}

	configContent := []byte(param.Config)

	// add new config
	if param.IsNew {
		if _, err := os.Stat(param.Path); err == nil { // exist
			var content []byte
			var err error

			if content, err = os.ReadFile(param.Path); err != nil {
				l.Errorf("Read file %s error: %s", param.Path, err.Error())
				return &ws.ResponseError{Code: 500, ErrorMsg: "read file error"}
			}
			configContent = append(content, configContent...)
		}
	}

	// check toml
	var v any
	if err := toml.Unmarshal(configContent, &v); err != nil {
		l.Errorf("parse toml failed: %s", err.Error())
		return &ws.ResponseError{ErrorCode: "toml.format.error", ErrorMsg: "toml format error"}
	}

	// create and save
	err := os.WriteFile(param.Path, configContent, dk.ConfPerm)
	if err != nil {
		l.Errorf("Write file %s failed: %s", param.Path, err.Error())
		return &ws.ResponseError{ErrorCode: "save.file.failed", ErrorMsg: "save file failed"}
	}

	// update configInfo
	if len(param.InputName) > 0 {
		inputs.AddConfigInfoPath(param.InputName, param.Path, 2)
	} else if param.Path == dk.MainConfPath {
		inputs.UpdateDatakitConfigInfo(2)
	}

	return nil
}

func saveDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	param := saveConfigParam{}

	if err := json.Unmarshal([]byte(data.Body), &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "params.invalid.body_invalid", ErrorMsg: "invalid body"})
		return
	}

	if err := doSaveConfig(&param); err != nil {
		response.SetError(err)
	} else {
		response.SetSuccess(map[string]string{"path": param.Path})
	}
}

func doDeleteConfig(inputName, filePath string) *ws.ResponseError {
	if filePath == dk.MainConfPath {
		return &ws.ResponseError{Code: 400, ErrorCode: "file.path.invalid", ErrorMsg: "Not supporet to delete main conf file"}
	}
	if err := checkPath(filePath); err != nil {
		return err
	}

	if !path.IsFileExists(filePath) {
		return &ws.ResponseError{Code: 400, ErrorCode: "file.path.invalid", ErrorMsg: "The file to be deleted is not existed!"}
	}

	if err := os.Remove(filePath); err != nil {
		l.Errorf("Delete conf file [%s] failed, %s", filePath, err.Error())
		return &ws.ResponseError{Code: 500, ErrorCode: "file.delete.failed", ErrorMsg: "Fail to delete conf file"}
	} else {
		inputs.DeleteConfigInfoPath(inputName, filePath)
	}

	return nil
}

func deleteDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	param := &struct {
		Path      string `json:"path"`
		InputName string `json:"inputName"`
	}{}

	if err := json.Unmarshal([]byte(data.Body), &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if err := doDeleteConfig(param.InputName, param.Path); err != nil {
		response.SetError(err)
	} else {
		response.SetSuccess()
	}
}

type pipelineInfo struct {
	FileName string `json:"fileName"`
	FileDir  string `json:"fileDir"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

func isValidPipelineFileName(name string) bool {
	pipelineFileRegxp := regexp.MustCompile(`.+\.p$`)

	return pipelineFileRegxp.Match([]byte(name))
}

func getDatakitPipelineAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	pipelines := []pipelineInfo{}

	allFiles, err := os.ReadDir(datakit.DataKitRuntimeInfo.PipelineDir)
	if err != nil {
		response.SetError()
		return
	}

	// filter pipeline file
	for _, file := range allFiles {
		if !file.IsDir() {
			name := file.Name()
			if isValidPipelineFileName(name) {
				pipelines = append(pipelines, pipelineInfo{FileName: name, FileDir: datakit.DataKitRuntimeInfo.PipelineDir})
			}
		} else {
			pipelines = append(pipelines, pipelineInfo{Category: file.Name()})
			allFiles, err := os.ReadDir(filepath.Join(datakit.DataKitRuntimeInfo.PipelineDir, file.Name()))
			if err != nil {
				response.SetError()
				return
			}
			for _, subFile := range allFiles {
				if !subFile.IsDir() {
					name := subFile.Name()
					if isValidPipelineFileName(name) {
						pipelines = append(pipelines,
							pipelineInfo{
								FileName: name,
								FileDir:  filepath.Join(datakit.DataKitRuntimeInfo.PipelineDir, file.Name()),
								Category: file.Name(),
							},
						)
					}
				}
			}
		}
	}

	response.SetSuccess(pipelines)
}

type pipelineDetailResponse struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func getDatakitPipelineDetailAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	var fileName, category, path string
	var contentBytes []byte
	var err error

	fileName = data.Query.Get("fileName")
	if len(fileName) == 0 {
		response.SetError(&ws.ResponseError{ErrorCode: "params.required", ErrorMsg: fmt.Sprintf("param %s is required", "fileName")})
		return
	}

	category = data.Query.Get("category")

	if !isValidPipelineFileName(fileName) {
		response.SetError(&ws.ResponseError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	path = filepath.Join(datakit.DataKitRuntimeInfo.PipelineDir, category, fileName)

	contentBytes, err = os.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf("Read pipeline file %s failed: %s", path, err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "param.invalid", ErrorMsg: fmt.Sprintf("param %s is not valid", fileName)})
		return
	}

	response.SetSuccess(pipelineDetailResponse{
		Path:    path,
		Content: string(contentBytes),
	})
}

func saveDatakitPipelineAction(isUpdate bool) ws.HandlerFunc {
	return func(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
		var filePath string
		var fileName string

		pipeline := pipelineInfo{}
		if err := json.Unmarshal([]byte(data.Body), &pipeline); err != nil {
			l.Errorf("Json unmarshal error: %s", err.Error())
			response.SetError(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
			return
		}

		if pipeline.Category == "default" {
			pipeline.Category = ""
		}

		fileName = pipeline.FileName
		category := pipeline.Category
		filePath = filepath.Join(datakit.DataKitRuntimeInfo.PipelineDir, category, fileName)

		if isUpdate && len(pipeline.Content) == 0 {
			response.SetError(&ws.ResponseError{
				ErrorCode: "param.invalid",
				ErrorMsg:  "'content' should not be empty",
			})
			return
		}

		// update pipeline
		if isUpdate {
			if !path.IsFileExists(filePath) {
				response.SetError(&ws.ResponseError{ErrorCode: "file.not.exist", ErrorMsg: "file not exist"})
				return
			}
		} else { // new pipeline
			if path.IsFileExists(filePath) {
				response.SetError(&ws.ResponseError{
					ErrorCode: "param.invalid.duplicate",
					ErrorMsg:  fmt.Sprintf("current name '%s' is duplicate", pipeline.FileName),
				})
				return
			}
		}

		if !isValidPipelineFileName(fileName) {
			response.SetError(&ws.ResponseError{
				ErrorCode: "param.invalid",
				ErrorMsg:  "fileName is not valid pipeline name",
			})
			return
		}

		err := os.WriteFile(filePath, []byte(pipeline.Content), dk.ConfPerm)
		if err != nil {
			l.Errorf("Write pipeline file %s failed: %s", filePath, err.Error())
			response.SetError()
			return
		}
		pipeline.FileDir = datakit.DataKitRuntimeInfo.PipelineDir
		response.SetSuccess(pipeline)
	}
}

func deleteDatakitPipelineAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	param := &struct {
		Category string `json:"category"`
		FileName string `json:"fileName"`
	}{}

	if err := json.Unmarshal([]byte(data.Body), &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if param.Category == "default" {
		param.Category = ""
	}

	fp := filepath.Join(datakit.DataKitRuntimeInfo.PipelineDir, param.Category, param.FileName)

	if !path.IsFileExists(fp) {
		response.SetError(&ws.ResponseError{Code: 400, ErrorCode: "file.path.invalid", ErrorMsg: "The file to be deleted is not existed!"})
		return
	}

	if err := os.Remove(fp); err != nil {
		response.SetError(&ws.ResponseError{Code: 500, ErrorCode: "file.delete.failed", ErrorMsg: "Fail to delete conf file"})
		l.Errorf("Delete pipeline file [%s] failed, %s", fp, err.Error())
	} else {
		response.SetSuccess()
	}
}

type dcaTestParam struct {
	Pipeline   map[string]map[string]string `json:"pipeline"`
	ScriptName string                       `json:"script_name"`
	Category   string                       `json:"category"`
	Data       []string                     `json:"data"`
}

func testDatakitPipelineAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	body := dcaTestParam{}

	if err := json.Unmarshal([]byte(data.Body), &body); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}
	if len(body.Category) == 0 {
		body.Category = "logging"
	}

	// deal with default
	if body.Category == "default" {
		body.Pipeline["logging"] = body.Pipeline["default"]
		body.Category = "logging"
	}

	category := point.CatString(body.Category)

	pls, errs := pipeline.NewPipelineMulti(category, body.Pipeline[body.Category], nil)
	if err, ok := errs[body.ScriptName]; ok && err != nil {
		response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: fmt.Sprintf("pipeline parse error: %s", err.Error())})
		return
	}

	pl, ok := pls[body.ScriptName]

	if !ok {
		response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: "pipeline is not valid"})
		return
	}

	var pts []*point.Point
	var pointErr error

	dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
	defer point.PutDecoder(dec)

	for _, data := range body.Data {
		switch category {
		case point.Logging:
			kvs := point.NewTags(datakit.DataKitRuntimeInfo.GlobalHostTags)
			kvs = append(kvs, point.NewKVs(map[string]interface{}{
				constants.FieldMessage: data,
			})...)
			pts = append(pts, point.NewPointV2(
				body.ScriptName, kvs, point.DefaultLoggingOptions()...))

		case point.CustomObject,
			point.DynamicDWCategory,
			point.KeyEvent,
			point.MetricDeprecated,
			point.Metric,
			point.Network,
			point.Object,
			point.ObjectChange, // Deprecated.
			point.Profiling,
			point.RUM,
			point.Security,
			point.Tracing,
			point.DialTesting,
			point.UnknownCategory:

			arr, err := dec.Decode([]byte(data))
			if err != nil {
				l.Warnf("make point error: %s", err.Error())
				pointErr = err
				break
			}
			pts = append(pts, arr...)
		}
	}

	if pointErr != nil {
		response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: fmt.Sprintf("invalid sample: %s", pointErr.Error())})
		return
	}

	var runResult []*pipelineResult

	for _, pt := range pts {
		plpt := ptinput.WrapPoint(category, pt)
		err := pl.Run(plpt, newPlTestSingal(), nil)
		if err != nil {
			plerr, ok := err.(*errchain.PlError) //nolint:errorlint
			if !ok {
				plerr = errchain.NewErr(body.ScriptName+".p", token.LnColPos{
					Pos: 0,
					Ln:  1,
					Col: 1,
				}, err.Error())
			}
			runResult = append(runResult, &pipelineResult{
				RunError: plerr,
			})
		} else {
			dropFlag := plpt.Dropped()

			plpt.KeyTime2Time()
			runResult = append(runResult, &pipelineResult{
				Point: &PlRetPoint{
					Dropped: dropFlag,
					Name:    plpt.GetPtName(),
					Tags:    plpt.Tags(),
					Fields:  plpt.Fields(),
					Time:    plpt.PtTime().Unix(),
					TimeNS:  int64(plpt.PtTime().Nanosecond()),
				},
			})
		}
	}
	response.SetSuccess(&runResult)
}

// newWebsocketConnectionAction create a new websocket connection.
// It applys to the log related actions.
func newWebsocketConnectionAction(client *ws.Client, id int64, data any) error {
	messageData := &ws.ActionData{}
	msg := ws.WebsocketMessage{
		Data: &messageData,
	}

	if err := json.Unmarshal(data.([]byte), &msg); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	connID := messageData.Query.Get(ws.HeaderNewWebSocketConnectionID)
	if connID == "" {
		l.Errorf("newWebsocketConnectionAction: connID is empty")
	}

	header := make(http.Header)
	header.Set(ws.HeaderNewWebSocketConnectionID, connID)

	if conn, _, err := websocket.DefaultDialer.Dial(client.GetWebsocketAddress(), header); err != nil {
		l.Errorf("dial failed: %s", err.Error())
	} else {
		g.Go(func(ctx context.Context) error {
			defer conn.Close() //nolint:errcheck
			_, message, err := conn.ReadMessage()
			if err != nil {
				l.Errorf("read message failed: %s", err.Error())
			}
			msg := ws.WebsocketMessage{
				Data: &ws.ActionData{},
			}
			if err := json.Unmarshal(message, &msg); err != nil {
				l.Errorf("json unmarshal failed: %s", err.Error())
				return nil
			}
			dk := client.GetDatakit()
			if msg.Action == ws.GetDatakitLogTailAction {
				getDatakitLogTailAction(conn, message, dk)
			} else if msg.Action == ws.GetDatakitLogDownloadAction {
				getDatakitLogDownloadAction(conn, message, dk)
			}

			return nil
		})
	}

	return nil
}

func getDatakitLogDownloadAction(conn *websocket.Conn, data []byte, dk *ws.DataKit) {
	var (
		logFile   string
		r         *bufio.Reader
		buf       []byte
		stat      fs.FileInfo
		sentBytes int64
		f         *os.File
		err       error
	)

	response := &ws.DCAResponse{}
	message := &ws.WebsocketMessage{
		Action: ws.GetDatakitLogDownloadAction,
	}
	messageData := &ws.ActionData{}
	msg := ws.WebsocketMessage{
		Data: &messageData,
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		l.Errorf("failed to unmarshal data: %w", err)
	}

	logFile = dk.DataKitRuntimeInfo.Log
	if messageData.Query.Get("type") == "gin.log" {
		logFile = dk.DataKitRuntimeInfo.GinLog
	}

	// if bytes, err := os.ReadFile(filepath.Clean(logFile)); err != nil {
	// 	l.Errorf("read file failed: %s", err.Error())
	// 	response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: "read file failed"})
	// 	goto Send
	// } else {
	// 	if err := conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
	// 		l.Errorf("write message failed: %s", err.Error())
	// 	}
	// 	return
	// }
	f, err = os.Open(filepath.Clean(logFile))
	if err != nil {
		l.Errorf("open log file failed: %s", err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: "open log file failed"})
		goto Send
	}

	stat, err = f.Stat()
	if err != nil {
		l.Errorf("get log file stat failed: %s", err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "400", ErrorMsg: "get log file stat failed"})
		goto Send
	}

	r = bufio.NewReader(f)
	buf = make([]byte, 1024)
	for {
		if sentBytes > stat.Size() {
			response.SetSuccess()
			goto Send
		}
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			l.Errorf("read log file failed: %s", err.Error())
			response.SetError(&ws.ResponseError{ErrorCode: "500", ErrorMsg: "read log file failed"})
			goto Send
		}
		if n == 0 {
			response.SetSuccess()
			goto Send
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			l.Errorf("write message failed: %s", err.Error())
			return
		}

		sentBytes += int64(n)
	}

Send:
	message.Data = response
	if err := conn.WriteMessage(websocket.TextMessage, message.Bytes()); err != nil {
		l.Errorf("send message failed: %s", err.Error())
	}
}

func getDatakitLogTailAction(conn *websocket.Conn, data any, dk *ws.DataKit) {
	var (
		buffer   []byte
		fileSeek int64
		fileSize int64
		file     *os.File
		err      error
		logFile  string
	)

	response := &ws.DCAResponse{}
	message := &ws.WebsocketMessage{
		Action: ws.GetDatakitLogTailAction,
	}
	messageData := &ws.ActionData{}
	msg := ws.WebsocketMessage{
		Data: &messageData,
	}

	if err := json.Unmarshal(data.([]byte), &msg); err != nil {
		l.Errorf("failed to unmarshal data: %w", err)
		return
	}

	logFile = dk.DataKitRuntimeInfo.Log

	if messageData.Query.Get("type") == "gin.log" {
		logFile = dk.DataKitRuntimeInfo.GinLog
	}

	if logFile == "stdout" {
		l.Infof("DCA log file is stdout, not supported now")
		response.SetError(&ws.ResponseError{ErrorCode: "stdout.not.supported", ErrorMsg: "stdout not supported"})
		goto Send
	}

	file, err = os.Open(filepath.Clean(logFile))
	if err != nil {
		l.Errorf("DCA open log file %s failed: %s", logFile, err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "dca.log.file.invalid", ErrorMsg: "datakit log file is not valid"})
		goto Send
	}
	defer file.Close() //nolint: errcheck,gosec

	if info, err := file.Stat(); err == nil {
		fileSize = info.Size()
	}

	if fileSize >= 2000 {
		fileSeek = -2000
	} else {
		fileSeek = -1 * fileSize
	}

	buffer = make([]byte, 1024)

	_, err = file.Seek(fileSeek, io.SeekEnd)
	if err != nil {
		l.Errorf("Seek offset %v error: %s", fileSeek, err.Error())
		response.SetError(&ws.ResponseError{ErrorCode: "dca.log.file.seek.error", ErrorMsg: "seek datakit log file error"})
		goto Send
	}

	for {
		n, err := file.Read(buffer)
		if errors.Is(err, io.EOF) {
			time.Sleep(5 * time.Second)
			continue
		}
		if err != nil {
			break
		}
		if n > 0 {
			response.SetSuccess(buffer[0:n])
			message.Data = response

			if err := conn.WriteMessage(websocket.TextMessage, message.Bytes()); err != nil {
				l.Errorf("send message failed: %s", err.Error())
				return
			}
		} else {
			time.Sleep(5 * time.Second)
		}
	}

Send:
	message.Data = response
	if err := conn.WriteMessage(websocket.TextMessage, message.Bytes()); err != nil {
		l.Errorf("send message failed: %s", err.Error())
	}
}

type filterInfo struct {
	Content  string `json:"content"`  // file content string
	FilePath string `json:"filePath"` // file path
}

//	getDatakitFilterAction return filter file content, which is located at data/.pull.
//
// if the file not existed, return empty content.
func getDatakitFilterAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	dataDir := datakit.DataKitRuntimeInfo.DataDir
	pullFilePath := filepath.Join(dataDir, ".pull")
	pullFileBytes, err := os.ReadFile(pullFilePath) //nolint: gosec
	if err != nil {
		response.SetSuccess(filterInfo{Content: "", FilePath: ""})
		return
	}

	response.SetSuccess(filterInfo{Content: string(pullFileBytes), FilePath: pullFilePath})
}

func reloadDatakitAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	if err := ReloadDataKit(ctx); err != nil {
		l.Error("reloadDataKit: %s", err)
		response.SetError()
		return
	}

	response.SetSuccess()
}

// nolint:gochecknoinits
func init() {
	HostActionHandlerMap = map[string]ws.ActionHandler{
		ws.GetDatakitStatsAction:          ws.GetActionHandler(ws.GetDatakitStatsAction, getDatakitStatsAction),
		ws.GetDatakitConfigAction:         ws.GetActionHandler(ws.GetDatakitConfigAction, getDatakitConfigAction),
		ws.SaveDatakitConfigAction:        ws.GetActionHandler(ws.SaveDatakitConfigAction, saveDatakitConfigAction),
		ws.DeleteDatakitConfigAction:      ws.GetActionHandler(ws.DeleteDatakitConfigAction, deleteDatakitConfigAction),
		ws.GetDatakitPipelineAction:       ws.GetActionHandler(ws.GetDatakitPipelineAction, getDatakitPipelineAction),
		ws.PatchDatakitPipelineAction:     ws.GetActionHandler(ws.PatchDatakitPipelineAction, saveDatakitPipelineAction(true)),
		ws.CreateDatakitPipelineAction:    ws.GetActionHandler(ws.CreateDatakitPipelineAction, saveDatakitPipelineAction(false)),
		ws.GetDatakitPipelineDetailAction: ws.GetActionHandler(ws.GetDatakitPipelineDetailAction, getDatakitPipelineDetailAction),
		ws.TestDatakitPipelineAction:      ws.GetActionHandler(ws.TestDatakitPipelineAction, testDatakitPipelineAction),
		ws.DeleteDatakitPipelineAction:    ws.GetActionHandler(ws.DeleteDatakitPipelineAction, deleteDatakitPipelineAction),
		ws.GetDatakitFilterAction:         ws.GetActionHandler(ws.GetDatakitFilterAction, getDatakitFilterAction),
		ws.ReloadDatakitAction:            ws.GetActionHandler(ws.ReloadDatakitAction, reloadDatakitAction),
		ws.NewWebsocketConnectionAction:   newWebsocketConnectionAction,
	}
}
