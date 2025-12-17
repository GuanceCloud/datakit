// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	asyncProfilePath    = "/opt/async-profiler"
	asyncProfileVersion = "2.8.1"
	DufaultEvents       = "all"
	DefaultDuration     = 60
	DefaultOutputFormat = "jfr"
	DefaultOutput       = "/tmp/"
)

type triggerStats struct {
	CommandName  string
	PID          int32
	Triggered    bool
	Reason       []string
	Service      string
	ProfilerPath string // async-profiler 的安装目录（例如：/opt/async-profiler）
	Event        string // 采集的事件类型：cpu, alloc, lock, wall 支持 all
	Duration     int    // 采集持续时间
	OutputFile   string // 输出文件路径
	OutputFormat string // output format: flat|traces|collapsed|flamegraph|tree|jfr|otlp
	startTime    string
	endTime      string
}

func newTriggerStats(e string, d string, tags []string) *triggerStats {
	timestamp := time.Now().Format("20060102_150405")
	outputName := fmt.Sprintf("profiler_%s.jfr", timestamp)
	jfrFile := filepath.Join(DefaultOutput, outputName)
	duration := DefaultDuration
	if val, err := time.ParseDuration(d); err == nil {
		duration = int(val.Seconds())
	}
	if e == "" {
		e = DufaultEvents
	}
	return &triggerStats{
		ProfilerPath: asyncProfilePath,
		Event:        e,
		Duration:     duration,
		OutputFormat: DefaultOutputFormat,
		Reason:       tags,
		OutputFile:   jfrFile,
	}
}

// runProfiling 执行一次性能分析采集并生成报告文件.
func runProfiling(ctx context.Context, stats *triggerStats) error {
	// 1. 基础检查
	if stats.ProfilerPath == "" {
		return fmt.Errorf("ProfilerPath is required")
	}
	if stats.PID <= 0 {
		return fmt.Errorf("PID (%d) must be > 0", stats.PID)
	}
	stats.startTime = time.Now().Format(time.RFC3339Nano)

	// 2. 确定使用 profiler.sh 还是 asprof
	// 优先尝试 asprof（通常位于 async-profiler 的 bin 目录下），如果不存在则使用 profiler.sh

	asprofPath := filepath.Join(stats.ProfilerPath, "bin", "asprof")

	if !pathExists(asprofPath) {
		return fmt.Errorf("at %s cannot fine  asprof or profiler.sh", stats.ProfilerPath)
	}

	// 3. 构建命令参数
	// 默认动作为 collect，即采集指定时间后自动停止
	var args []string

	// 设置采集持续时间
	if stats.Duration > 0 {
		args = append(args, "-d", fmt.Sprintf("%d", stats.Duration))
	}

	// 设置采集事件类型
	switch stats.Event {
	case "--all", "all":
		args = append(args, "--all")
	case "":
		args = append(args, "-e", "cpu") // 默认使用 CPU 分析
	default:
		args = append(args, "-e", stats.Event)
	}

	// 设置输出文件
	if stats.OutputFile != "" {
		args = append(args, "-f", stats.OutputFile)
	} else {
		// 如果未指定输出文件，生成一个带时间戳的默认文件
		timestamp := time.Now().Format("20060102_150405")
		defaultOutput := fmt.Sprintf(DefaultOutput+"/profile_%d_%s.jfr", stats.PID, timestamp)
		args = append(args, "-f", defaultOutput)
		stats.OutputFile = defaultOutput
	}

	// 设置输出格式 (如不指定，通常根据输出文件后缀自动判断)
	if stats.OutputFormat != "" {
		args = append(args, "-o", stats.OutputFormat)
	}

	args = append(args, fmt.Sprintf("%d", stats.PID))
	log.Infof("exec command: %s %v", asprofPath, args)

	cmd := exec.CommandContext(ctx, asprofPath, args...) //nolint

	// 捕获标准错误和标准输出，有助于调试
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// 启动命令并等待完成
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd start err: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		// 如果命令执行失败，返回标准错误的内容
		return fmt.Errorf("cmd exec err: %w, and stderr is %s", err, stderr.String())
	}

	// 5. 成功输出
	log.Infof("ok, output file is  %s", stats.OutputFile)
	log.Infof("stdout: %s", stdout.String())

	stats.endTime = time.Now().Format(time.RFC3339Nano)

	return nil
}

// pathExists 检查文件路径是否存在.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// -- upload file to DataKit --

func uploadFileToDataKit(stats *triggerStats, datakitURL string) error {
	// 创建 multipart 写入缓冲区
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 1. 上传硬盘中的 JFR 文件
	if err := addJFRFile(writer, stats.OutputFile); err != nil {
		return fmt.Errorf("add jfr file to request body err: %w", err)
	}

	// 2. 添加上传手动组装的JSON数据
	if err := addJSONConfig(writer, stats); err != nil {
		return fmt.Errorf("add event.json err: %w", err)
	}

	// 关闭写入器，完成multipart数据构造
	if err := writer.Close(); err != nil {
		return fmt.Errorf("write close err: %w", err)
	}

	// 4. 创建HTTP请求
	req, err := http.NewRequest("POST", datakitURL, &requestBody)
	if err != nil {
		return fmt.Errorf("newRequest err: %w", err)
	}

	// 设置Content-Type头，必须使用writer.FormDataContentType()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 5. 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do reqeuest err: %w", err)
	}
	defer resp.Body.Close() // nolint

	// 6. 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		uploadToDK.WithLabelValues(stats.Service, fmt.Sprintf("%d", resp.StatusCode))
		return fmt.Errorf("response code: %d, body: %s", resp.StatusCode, string(body))
	} else {
		uploadToDK.WithLabelValues(stats.Service, "200")
		log.Infof("upload ok, %s", stats.OutputFile)
		return nil
	}
}

func deleteFile(stats *triggerStats) {
	if err := os.Remove(stats.OutputFile); err != nil {
		log.Errorf("delete file err: %w", err)
	} else {
		log.Infof("delete file ok, %s", stats.OutputFile)
	}
}

type Event struct {
	TagProfiler string `json:"tags_profiler"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Family      string `json:"family"`
	Format      string `json:"format"`
}

func addJSONConfig(writer *multipart.Writer, stats *triggerStats) error {
	// 序列化JSON配置
	event := &Event{
		Start:  stats.startTime,
		End:    stats.endTime,
		Family: "java",
		Format: "jfr",
	}

	service := stats.getKeyFromTags("service", "unknown_service")
	host := stats.getKeyFromTags("host", "unknown_host")
	env := stats.getKeyFromTags("env", "unknown_env")
	version := stats.getKeyFromTags("version", "unknown_version")
	tags := fmt.Sprintf("library_version:%s,library_type:async_profiler,process_id:%d,process_name:%s,service:%s,host:%s,env:%s,version:%s",
		asyncProfileVersion, stats.PID, stats.CommandName, service, host, env, version)
	// 添加 trigger tag
	tags += "," + strings.Join(stats.Reason, ",")

	event.TagProfiler = tags
	jsonData, _ := json.Marshal(event)

	// 创建自定义MIME类型的文件部分
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="event"; filename="event.json"`)
	header.Set("Content-Type", "application/json")

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("creat header err: %w", err)
	}

	// 写入JSON数据
	_, err = part.Write(jsonData)
	if err != nil {
		return fmt.Errorf("write json data err: %w", err)
	}

	log.Debugf("add json 'event': %s", string(jsonData))
	return nil
}

func addJFRFile(writer *multipart.Writer, filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("JFR file not exist %s", filePath)
	}

	// 打开文件
	file, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("open file err: %w", err)
	}

	// 创建文件部分
	part, err := writer.CreateFormFile("main", "main.jfr")
	if err != nil {
		return fmt.Errorf("create file err: %w", err)
	}
	_, err = part.Write(file)
	if err != nil {
		return fmt.Errorf("write file err: %w", err)
	}
	log.Debugf("JFR add to part ok, %s", filePath)
	return nil
}

func (stat *triggerStats) getKeyFromTags(key, defaultVal string) string {
	for _, s := range stat.Reason {
		keyVal := strings.Split(s, ":")
		if len(keyVal) == 2 && keyVal[0] == key {
			return keyVal[1]
		}
	}

	return defaultVal
}
