// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"

	"github.com/BurntSushi/toml"
)

var (
	log           = logger.DefaultSLogger("flameshot")
	defaultConfig = `
  datakit_addr = "http://localhost:9529"
  ## profiling path
  profiling_path = "/profiling/v1/input"
  # The time interval for monitoring the program, in seconds
  monitor_interval = "1s"
  tags = ["globle_tag_1:xxx","other_tag:aaa"]
  auto_profiling = "10m"
  auto_profiling_duration = "30s"
  profiling_enabled = true
  oom_hprof_enabled = false
  oom_hprof_match_window = "2m"
  hprof_upload_enabled = false
  hprof_upload_provider = ""
  hprof_upload_auth_type = "static"
  hprof_upload_endpoint = ""
  hprof_upload_region = ""
  hprof_upload_bucket = ""
  hprof_upload_access_key_id = ""
  hprof_upload_access_key_secret = ""
  hprof_upload_security_token = ""
  hprof_upload_assume_role_arn = ""
  hprof_upload_assume_role_session_name = ""
  hprof_upload_assume_role_duration_seconds = 3600
  hprof_upload_assume_role_policy = ""
  hprof_upload_assume_role_external_id = ""
  hprof_upload_assume_role_sts_endpoint = ""
  hprof_upload_assume_role_source_access_key_id = ""
  hprof_upload_assume_role_source_access_key_secret = ""
  hprof_upload_assume_role_source_security_token = ""
  hprof_upload_path_template = "{service}/{pod_name}/{timestamp}/{filename}"
  hprof_download_url_template = ""
  hprof_upload_timeout = "5m"
  heap_dump_enabled = false
  heap_dump_path_template = "{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof"
  heap_dump_jmap_path = "jmap"
  heap_dump_timeout = "120s"
  heap_dump_cooldown = "10m"

  [[processes]]
    ## service name for profiling
    service = "default_service_name"
    command = '''^java\b.*xxx-name\.jar$'''
    # -e profiling event: cpu|alloc|nativemem|lock|cache-misses etc.
    # and 'all' for all events. 
	    events = "cpu,alloc,nativemem"
	    ## -d duration for profiling, in seconds.
	    duration = "30s"
	    ## duration for emergency memory trigger.
	    emergency_duration = "15s"
	    language = "java"
	    jdk_version = "-"
	    ## Go pprof HTTP base URL, required when language is go/golang.
	    # pprof_url = "http://127.0.0.1:6060"
	    ## Go pprof types: cpu, goroutine, heap, mutex, block.
	    # pprof_types = ["cpu", "goroutine", "heap", "mutex", "block"]
	    # pprof_timeout = "45s"
	    ## Python py-spy options, used when language is python.
	    # pyspy_path = "py-spy"
	    # pyspy_output_path = ""
	    # pyspy_rate = 100
	    # pyspy_subprocesses = false
	    # pyspy_idle = false
	    tags = ["env:env", "version:1.0.0"]
    ## cpu usage percent. 4C max is 400, 80% is 320
    cpu_usage_percent = 80
    ##memory usage percentage based limit, 0~100
    mem_usage_percent = 80
    ## emergency memory usage percentage based limit, 0~100
    mem_usage_percent_emergency = 95
    ## memory usage in MB
    mem_usage_mb = 1024
    ## emergency memory usage in MB
    mem_usage_mb_emergency = 2048
    ## trigger jmap heap dump when emergency memory threshold is reached
    heap_dump_on_memory_emergency = true

  [http]
    local_host = "localhost"
    local_port = "8089"

  [logging]
    level = "info"
    path = "/var/log/flameshot/log"
`
)

type HTTPConfig struct {
	LocalHost string `toml:"local_host"`
	LocalPort string `toml:"local_port"`
	// 其他接口上的配置等
}

type Logging struct {
	Level string `toml:"level"` // 日志级别
	Path  string `toml:"path"`  // 日志输出
}

type Process struct {
	Service                   string   `toml:"service" json:"service"`                                         // 服务名称
	Command                   string   `toml:"command" json:"command" `                                        // 命令支持正则
	Duration                  string   `toml:"duration" json:"duration"`                                       // 采集时长
	EmergencyDuration         string   `toml:"emergency_duration" json:"emergency_duration"`                   // 紧急触发时采集时长
	Events                    string   `toml:"events" json:"events"`                                           // 采集的事件 用逗号隔开 支持 'all'
	Language                  string   `toml:"language" json:"language"`                                       // 目标程序语言 java go
	JDKVersion                string   `toml:"jdk_version" json:"jdk_version"`                                 // jdk 版本
	PProfURL                  string   `toml:"pprof_url" json:"pprof_url"`                                     // Go pprof HTTP 地址
	PProfTypes                []string `toml:"pprof_types" json:"pprof_types"`                                 // Go pprof 类型
	PProfTimeout              string   `toml:"pprof_timeout" json:"pprof_timeout"`                             // Go pprof 请求超时
	PySpyPath                 string   `toml:"pyspy_path" json:"pyspy_path"`                                   // Python py-spy 可执行文件路径
	PySpyOutputPath           string   `toml:"pyspy_output_path" json:"pyspy_output_path"`                     // Python py-spy 本地输出路径
	PySpyRate                 int      `toml:"pyspy_rate" json:"pyspy_rate"`                                   // Python py-spy 采样频率
	PySpySubprocesses         bool     `toml:"pyspy_subprocesses" json:"pyspy_subprocesses"`                   // Python py-spy 是否采集子进程
	PySpyIdle                 bool     `toml:"pyspy_idle" json:"pyspy_idle"`                                   // Python py-spy 是否采集 idle 线程
	Tags                      []string `toml:"tags" json:"tags"`                                               // 自定义标签
	CPUUsagePercent           int      `toml:"cpu_usage_percent" json:"cpu_usage_percent"`                     // cpu 使用率
	MEMUsagePercent           int      `toml:"mem_usage_percent" json:"mem_usage_percent"`                     // 内存使用率平均值阈值
	MEMUsageMB                int      `toml:"mem_usage_mb" json:"mem_usage_mb"`                               // 内存使用量平均值阈值
	MEMUsagePercentEmergency  int      `toml:"mem_usage_percent_emergency" json:"mem_usage_percent_emergency"` // 内存使用率紧急瞬时阈值
	MEMUsageMBEmergency       int      `toml:"mem_usage_mb_emergency" json:"mem_usage_mb_emergency"`           // 内存使用量紧急瞬时阈值
	HeapDumpOnMemoryEmergency *bool    `toml:"heap_dump_on_memory_emergency,omitempty" json:"heap_dump_on_memory_emergency,omitempty"`
}

type Config struct {
	DataKitAddr                                string      `toml:"datakit_addr"`                                      // datakit 地址
	ProfilingPath                              string      `toml:"profiling_path"`                                    // 虚拟环境下必须保证是共享目录
	MonitorInterval                            string      `toml:"monitor_interval"`                                  // 监控间隔，单位 秒
	Tags                                       []string    `toml:"tags"`                                              // 全局自定义标签
	AutoProfiling                              string      `toml:"auto_profiling"`                                    // 开关定时自动执行, 配置 0 则关闭
	AutoProfileDuration                        string      `toml:"auto_profiling_duration"`                           // 定时自动采集时长
	ProfilingEnabled                           *bool       `toml:"profiling_enabled,omitempty"`                       // 开启 JFR profiling，nil 表示默认开启
	OOMHProfEnabled                            bool        `toml:"oom_hprof_enabled"`                                 // 开启 OOM hprof 摘要采集
	OOMHProfMatchWindow                        string      `toml:"oom_hprof_match_window"`                            // OOM 事件与 hprof 的时间匹配窗口
	HProfUploadEnabled                         bool        `toml:"hprof_upload_enabled"`                              // 开启 hprof 对象存储上传
	HProfUploadProvider                        string      `toml:"hprof_upload_provider"`                             // oss/s3
	HProfUploadAuthType                        string      `toml:"hprof_upload_auth_type"`                            // static/assume_role
	HProfUploadEndpoint                        string      `toml:"hprof_upload_endpoint"`                             // 对象存储 endpoint
	HProfUploadRegion                          string      `toml:"hprof_upload_region"`                               // S3 region
	HProfUploadBucket                          string      `toml:"hprof_upload_bucket"`                               // bucket
	HProfUploadAccessKeyID                     string      `toml:"hprof_upload_access_key_id"`                        // AK
	HProfUploadAccessKeySecret                 string      `toml:"hprof_upload_access_key_secret"`                    // SK
	HProfUploadSecurityToken                   string      `toml:"hprof_upload_security_token"`                       // STS security token
	HProfUploadAssumeRoleARN                   string      `toml:"hprof_upload_assume_role_arn"`                      // AssumeRole target role ARN
	HProfUploadAssumeRoleSessionName           string      `toml:"hprof_upload_assume_role_session_name"`             // AssumeRole session name
	HProfUploadAssumeRoleDurationSeconds       int         `toml:"hprof_upload_assume_role_duration_seconds"`         // AssumeRole duration seconds
	HProfUploadAssumeRolePolicy                string      `toml:"hprof_upload_assume_role_policy"`                   // AssumeRole inline policy
	HProfUploadAssumeRoleExternalID            string      `toml:"hprof_upload_assume_role_external_id"`              // AssumeRole external id
	HProfUploadAssumeRoleSTSEndpoint           string      `toml:"hprof_upload_assume_role_sts_endpoint"`             // AssumeRole STS endpoint
	HProfUploadAssumeRoleSourceAccessKeyID     string      `toml:"hprof_upload_assume_role_source_access_key_id"`     // AssumeRole source AK
	HProfUploadAssumeRoleSourceAccessKeySecret string      `toml:"hprof_upload_assume_role_source_access_key_secret"` // AssumeRole source SK
	HProfUploadAssumeRoleSourceSecurityToken   string      `toml:"hprof_upload_assume_role_source_security_token"`    // AssumeRole source STS token
	HProfUploadPathTemplate                    string      `toml:"hprof_upload_path_template"`                        // object key template
	HProfDownloadURLTemplate                   string      `toml:"hprof_download_url_template"`                       // download URL template
	HProfUploadTimeout                         string      `toml:"hprof_upload_timeout"`                              // upload timeout
	HProfUploadS3PathStyle                     *bool       `toml:"hprof_upload_s3_path_style,omitempty"`              // S3 path-style endpoint
	HeapDumpEnabled                            bool        `toml:"heap_dump_enabled"`                                 // 开启主动 heap dump
	HeapDumpPathTemplate                       string      `toml:"heap_dump_path_template"`                           // heap dump 文件路径模板
	HeapDumpJMapPath                           string      `toml:"heap_dump_jmap_path"`                               // jmap 路径
	HeapDumpTimeout                            string      `toml:"heap_dump_timeout"`                                 // jmap 超时
	HeapDumpCooldown                           string      `toml:"heap_dump_cooldown"`                                // 每进程 heap dump 冷却时间
	PodCPULimit                                string      `toml:"pod_cpu_limit"`                                     // pod resource limit
	PodMEMLimit                                string      `toml:"pod_mem_limit"`                                     // pod resource limit
	Processes                                  []*Process  `toml:"processes"`                                         // 监控的进程列表
	HTTPConfig                                 *HTTPConfig `toml:"http"`                                              // http 配置
	Log                                        *Logging    `toml:"logging"`                                           // 日志配置
}

func (c *Config) fromEnv() {
	if x := os.Getenv("FLAMESHOT_DATAKIT_ADDR"); x != "" {
		c.DataKitAddr = x
	}
	if x := os.Getenv("FLAMESHOT_PROFILING_PATH"); x != "" {
		c.ProfilingPath = x
	}
	if x := os.Getenv("FLAMESHOT_MONITOR_INTERVAL"); x != "" {
		c.MonitorInterval = x
	}
	if x := os.Getenv("FLAMESHOT_TAGS"); x != "" {
		kvs := strings.Split(x, ",")
		if c.Tags == nil {
			c.Tags = make([]string, 0)
		}

		c.Tags = append(c.Tags, kvs...)
	}
	if x := os.Getenv("FLAMESHOT_HTTP_LOCAL_IP"); x != "" {
		if c.HTTPConfig == nil {
			c.HTTPConfig = &HTTPConfig{}
		}
		c.HTTPConfig.LocalHost = x
	}
	if x := os.Getenv("FLAMESHOT_HTTP_LOCAL_PORT"); x != "" {
		if c.HTTPConfig == nil {
			c.HTTPConfig = &HTTPConfig{}
		}
		c.HTTPConfig.LocalPort = x
	}
	if x := os.Getenv("FLAMESHOT_LOG_LEVEL"); x != "" {
		if c.Log == nil {
			c.Log = &Logging{}
		}
		c.Log.Level = x
	}
	if x := os.Getenv("FLAMESHOT_LOG_PATH"); x != "" {
		if c.Log == nil {
			c.Log = &Logging{}
		}
		c.Log.Path = x
	}
	if x := os.Getenv("FLAMESHOT_AUTO_PROFILING"); x != "" {
		auto, err := time.ParseDuration(x)
		if err != nil {
			log.Warnf("parse %s=%s failed: %s", "FLAMESHOT_AUTO_PROFILING", x, err.Error())
		} else {
			if auto > 0 && auto < time.Minute {
				x = "5m"
				log.Warnf("parse %s=%s failed: %s", "FLAMESHOT_AUTO_PROFILING", x, "auto profiling time must >= 1 minute, use default 5 minute")
			}
			c.AutoProfiling = x
		}
	}
	if x := os.Getenv("FLAMESHOT_AUTO_PROFILING_DURATION"); x != "" {
		c.AutoProfileDuration = x
	}
	if x := os.Getenv("FLAMESHOT_PROFILING_ENABLED"); x != "" {
		c.ProfilingEnabled = boolPtr(parseBoolEnv(x))
	}
	if x := os.Getenv("FLAMESHOT_OOM_HPROF_ENABLED"); x != "" {
		c.OOMHProfEnabled = parseBoolEnv(x)
	}
	if x := os.Getenv("FLAMESHOT_OOM_HPROF_MATCH_WINDOW"); x != "" {
		c.OOMHProfMatchWindow = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ENABLED"); x != "" {
		c.HProfUploadEnabled = parseBoolEnv(x)
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_PROVIDER"); x != "" {
		c.HProfUploadProvider = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_AUTH_TYPE"); x != "" {
		c.HProfUploadAuthType = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ENDPOINT"); x != "" {
		c.HProfUploadEndpoint = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_REGION"); x != "" {
		c.HProfUploadRegion = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_BUCKET"); x != "" {
		c.HProfUploadBucket = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID"); x != "" {
		c.HProfUploadAccessKeyID = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET"); x != "" {
		c.HProfUploadAccessKeySecret = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_SECURITY_TOKEN"); x != "" {
		c.HProfUploadSecurityToken = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_ARN"); x != "" {
		c.HProfUploadAssumeRoleARN = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SESSION_NAME"); x != "" {
		c.HProfUploadAssumeRoleSessionName = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_DURATION_SECONDS"); x != "" {
		if seconds, err := strconv.Atoi(x); err == nil {
			c.HProfUploadAssumeRoleDurationSeconds = seconds
		} else {
			log.Warnf("parse %s=%s failed: %s", "FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_DURATION_SECONDS", x, err.Error())
		}
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_POLICY"); x != "" {
		c.HProfUploadAssumeRolePolicy = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_EXTERNAL_ID"); x != "" {
		c.HProfUploadAssumeRoleExternalID = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_STS_ENDPOINT"); x != "" {
		c.HProfUploadAssumeRoleSTSEndpoint = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_ID"); x != "" {
		c.HProfUploadAssumeRoleSourceAccessKeyID = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_ACCESS_KEY_SECRET"); x != "" {
		c.HProfUploadAssumeRoleSourceAccessKeySecret = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_ASSUME_ROLE_SOURCE_SECURITY_TOKEN"); x != "" {
		c.HProfUploadAssumeRoleSourceSecurityToken = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_PATH_TEMPLATE"); x != "" {
		c.HProfUploadPathTemplate = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_DOWNLOAD_URL_TEMPLATE"); x != "" {
		c.HProfDownloadURLTemplate = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_TIMEOUT"); x != "" {
		c.HProfUploadTimeout = x
	}
	if x := os.Getenv("FLAMESHOT_HPROF_UPLOAD_S3_PATH_STYLE"); x != "" {
		c.HProfUploadS3PathStyle = boolPtr(parseBoolEnv(x))
	}
	if x := os.Getenv("FLAMESHOT_HEAP_DUMP_ENABLED"); x != "" {
		c.HeapDumpEnabled = parseBoolEnv(x)
	}
	if x := os.Getenv("FLAMESHOT_HEAP_DUMP_PATH_TEMPLATE"); x != "" {
		c.HeapDumpPathTemplate = x
	}
	if x := os.Getenv("FLAMESHOT_HEAP_DUMP_JMAP_PATH"); x != "" {
		c.HeapDumpJMapPath = x
	}
	if x := os.Getenv("FLAMESHOT_HEAP_DUMP_TIMEOUT"); x != "" {
		c.HeapDumpTimeout = x
	}
	if x := os.Getenv("FLAMESHOT_HEAP_DUMP_COOLDOWN"); x != "" {
		c.HeapDumpCooldown = x
	}

	if x := os.Getenv("FLAMESHOT_POD_CPU_LIMIT"); x != "" {
		c.PodCPULimit = fmt.Sprintf("%sm", x)
	}

	if x := os.Getenv("FLAMESHOT_POD_MEM_LIMIT"); x != "" {
		c.PodMEMLimit = fmt.Sprintf("%sMi", x)
	}

	// 数组配置使用for循环
	if c.Processes == nil {
		c.Processes = make([]*Process, 0)
	}
	c.loadProcessesFromEnv()
}

func (c *Config) loadProcessesFromEnv() {
	for i := 0; ; i++ { // 通过索引i来遍历可能的进程配置
		serviceKey := fmt.Sprintf("FLAMESHOT_PROCESSES_%d_SERVICE", i)
		serviceVal := os.Getenv(serviceKey)

		// 如果找不到当前索引的SERVICE变量，认为数组结束
		if serviceVal == "" {
			break
		}

		// 为当前索引创建一个Process结构体
		process := &Process{
			Service: serviceVal,
		}

		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_DURATION", i)); val != nil {
			process.Duration = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_EMERGENCY_DURATION", i)); val != nil {
			process.EmergencyDuration = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_COMMAND", i)); val != nil {
			process.Command = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_EVENTS", i)); val != nil {
			process.Events = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_LANGUAGE", i)); val != nil {
			process.Language = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_JDK_VERSION", i)); val != nil {
			process.JDKVersion = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PPROF_URL", i)); val != nil {
			process.PProfURL = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PPROF_TYPES", i)); val != nil {
			list, err := parseEnvStringList(*val)
			if err != nil {
				log.Warnf("parse %s=%s failed: %s", "FLAMESHOT_PROCESSES_PPROF_TYPES", *val, err.Error())
			} else {
				process.PProfTypes = list
			}
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PPROF_TIMEOUT", i)); val != nil {
			process.PProfTimeout = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PYSPY_PATH", i)); val != nil {
			process.PySpyPath = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PYSPY_OUTPUT_PATH", i)); val != nil {
			process.PySpyOutputPath = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PYSPY_RATE", i)); val != nil {
			process.PySpyRate = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PYSPY_SUBPROCESSES", i)); val != nil {
			process.PySpySubprocesses = parseBoolEnv(*val)
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_PYSPY_IDLE", i)); val != nil {
			process.PySpyIdle = parseBoolEnv(*val)
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_CPU_USAGE_PERCENT", i)); val != nil {
			process.CPUUsagePercent = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_PERCENT", i)); val != nil {
			process.MEMUsagePercent = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_MB", i)); val != nil {
			process.MEMUsageMB = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_PERCENT_EMERGENCY", i)); val != nil {
			process.MEMUsagePercentEmergency = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_MB_EMERGENCY", i)); val != nil {
			process.MEMUsageMBEmergency = *val
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_HEAP_DUMP_ON_MEMORY_EMERGENCY", i)); val != nil {
			process.HeapDumpOnMemoryEmergency = boolPtr(parseBoolEnv(*val))
		}
		if val := getEnvString(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_TAGS", i)); val != nil {
			var list []string
			if err := json.Unmarshal([]byte(*val), &list); err != nil {
				log.Warnf("parse %s=%s failed: %s", "FLAMESHOT_PROCESSES_TAGS", *val, err.Error())
			} else {
				process.Tags = list
			}
		}

		c.Processes = append(c.Processes, process)
	}
	service := ""
	if val := getEnvString("FLAMESHOT_SERVICE"); val != nil {
		service = *val
	}

	if x := os.Getenv("FLAMESHOT_PROCESSES"); x != "" {
		ps := make([]*Process, 0)
		err := json.Unmarshal([]byte(x), &ps)
		if err != nil {
			log.Errorf("unmarshal process failed, err:%v  and config is:", err, x)
		} else {
			for i := range ps {
				if service != "" {
					ps[i].Service = service
				}
			}
			c.Processes = append(c.Processes, ps...)
		}
	}
	c.applyDefaults()
}

func boolPtr(v bool) *bool {
	return &v
}

func parseBoolEnv(v string) bool {
	return strings.EqualFold(v, "true") || strings.EqualFold(v, "yes") || strings.EqualFold(v, "on") || v == "1"
}

func parseEnvStringList(val string) ([]string, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}

	var list []string
	if strings.HasPrefix(val, "[") {
		if err := json.Unmarshal([]byte(val), &list); err != nil {
			return nil, err
		}
		return list, nil
	}

	parts := strings.Split(val, ",")
	list = make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			list = append(list, part)
		}
	}

	return list, nil
}

func getEnvString(key string) *string {
	if val := os.Getenv(key); val != "" {
		return &val
	}
	return nil
}

// 辅助函数：获取整数类型环境变量，如果不存在或解析失败则返回nil.
func getEnvInt(key string) *int {
	if valStr := os.Getenv(key); valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return &val
		} else {
			d, err := time.ParseDuration(valStr)
			if err == nil {
				val = int(d / time.Second)
				return &val
			}
		}
	}
	return nil
}

func InitConfig(logPath string) *Config {
	conf := defaultFlameshotConfig()
	bts, err := os.ReadFile(logPath) //nolint:gosec
	if err != nil {
		log.Errorf("read config file failed, err:%v", err)
	} else {
		err = toml.Unmarshal(bts, conf)
		if err != nil {
			log.Errorf("unmarshal config failed, err:%v", err)
		}
	}

	conf.fromEnv()
	conf.applyDefaults()
	conf.initLogging()
	// copy profiler files to share dir
	if conf.ProfilingPath != "" {
		src := asyncProfilePath
		dst := filepath.Join(conf.ProfilingPath, "profiler")
		if err = copyProfilerFiles(src, dst); err != nil {
			log.Errorf("copy profiler files failed, err:%v", err)
		} else {
			// 复制完成之后，再执行命令需要到共享目录执行
			asyncProfilePath = dst
			DefaultOutput = conf.ProfilingPath
		}
	}

	return conf
}

func defaultFlameshotConfig() *Config {
	return &Config{
		ProfilingEnabled:                     boolPtr(true),
		HProfUploadAuthType:                  "static",
		HProfUploadAssumeRoleDurationSeconds: 3600,
		HProfUploadPathTemplate:              "{service}/{pod_name}/{timestamp}/{filename}",
		HProfUploadTimeout:                   "5m",
		HProfUploadS3PathStyle:               boolPtr(true),
		HeapDumpPathTemplate:                 "{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof",
		HeapDumpJMapPath:                     "jmap",
		HeapDumpTimeout:                      "120s",
		HeapDumpCooldown:                     "10m",
	}
}

func (c *Config) applyDefaults() {
	if c == nil {
		return
	}
	if c.ProfilingEnabled == nil {
		c.ProfilingEnabled = boolPtr(true)
	}
	if c.HProfUploadAuthType == "" {
		c.HProfUploadAuthType = "static"
	}
	if c.HProfUploadAssumeRoleDurationSeconds == 0 {
		c.HProfUploadAssumeRoleDurationSeconds = 3600
	}
	if c.HProfUploadPathTemplate == "" {
		c.HProfUploadPathTemplate = "{service}/{pod_name}/{timestamp}/{filename}"
	}
	if c.HProfUploadTimeout == "" {
		c.HProfUploadTimeout = "5m"
	}
	if c.HProfUploadS3PathStyle == nil {
		c.HProfUploadS3PathStyle = boolPtr(true)
	}
	if c.HeapDumpPathTemplate == "" {
		c.HeapDumpPathTemplate = "{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof"
	}
	if c.HeapDumpJMapPath == "" {
		c.HeapDumpJMapPath = "jmap"
	}
	if c.HeapDumpTimeout == "" {
		c.HeapDumpTimeout = "120s"
	}
	if c.HeapDumpCooldown == "" {
		c.HeapDumpCooldown = "10m"
	}
	for _, p := range c.Processes {
		applyProcessDefaults(p)
	}
}

func applyProcessDefaults(p *Process) {
	if p == nil {
		return
	}
	if !isPythonLanguage(p.Language) {
		return
	}
	if p.PySpyPath == "" {
		p.PySpyPath = defaultPySpyPath
	}
	if p.PySpyRate == 0 {
		p.PySpyRate = defaultPySpyRate
	}
}

func copyProfilerFiles(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("file stat: err %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("%s not dir", src)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("mkdir err: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %w", err)
	}

	// 遍历并复制每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyProfilerFiles(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 如果是文件，复制文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile 复制单个文件.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src) //nolint:gosec
	if err != nil {
		return fmt.Errorf("os open file err: %w", err)
	}
	defer srcFile.Close() //nolint:errcheck,gosec

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("src file stat err: %w", err)
	}

	dstFile, err := os.Create(dst) //nolint:gosec
	if err != nil {
		return fmt.Errorf("creat file %w", err)
	}
	defer dstFile.Close() //nolint:errcheck,gosec

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("io copy err: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("set mod err: %w", err)
	}

	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("set modTime err: %w", err)
	}

	return nil
}

func (c *Config) initLogging() {
	lopt := &logger.Option{
		Level: "info",
		Flags: (logger.OPT_DEFAULT | logger.OPT_STDOUT),
	}

	if c.Log != nil && c.Log.Level == "debug" {
		lopt.Level = "debug"
	}

	if err := logger.InitRoot(lopt); err != nil {
		return
	}

	log = logger.SLogger("flameshot")
}
