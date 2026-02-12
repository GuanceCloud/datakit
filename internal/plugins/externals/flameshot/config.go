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

  [[processes]]
    ## service name for profiling
    service = "default_service_name"
    command = '''^java\b.*xxx-name\.jar$'''
    # -e profiling event: cpu|alloc|nativemem|lock|cache-misses etc.
    # and 'all' for all events. 
    events = "cpu,alloc,nativemem"
    ## -d duration for profiling, in seconds.
    duration = "30s"
    language = "java"
    jdk_version = "-"
    tags = ["env:env", "version:1.0.0"]
    ## cpu usage percent. 4C max is 400, 80% is 320
    cpu_usage_percent = 80
    ##memory usage percentage based limit, 0~100
    mem_usage_percent = 80
    ## memory usage in MB
    mem_usage_mb = 1024

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
	Service         string   `toml:"service" json:"service"`                     // 服务名称
	Command         string   `toml:"command" json:"command" `                    // 命令支持正则
	Duration        string   `toml:"duration" json:"duration"`                   // 采集时长
	Events          string   `toml:"events" json:"events"`                       // 采集的事件 用逗号隔开 支持 'all'
	Language        string   `toml:"language" json:"language"`                   // 目标程序语言 java go
	JDKVersion      string   `toml:"jdk_version" json:"jdk_version"`             // jdk 版本
	Tags            []string `toml:"tags" json:"tags"`                           // 自定义标签
	CPUUsagePercent int      `toml:"cpu_usage_percent" json:"cpu_usage_percent"` // cpu 使用率
	MEMUsagePercent int      `toml:"mem_usage_percent" json:"mem_usage_percent"` // 内存使用率
	MEMUsageMB      int      `toml:"mem_usage_mb" json:"mem_usage_mb"`           // 内存使用量
}

type Config struct {
	DataKitAddr     string      `toml:"datakit_addr"`     // datakit 地址
	ProfilingPath   string      `toml:"profiling_path"`   // 虚拟环境下必须保证是共享目录
	MonitorInterval string      `toml:"monitor_interval"` // 监控间隔，单位 秒
	Tags            []string    `toml:"tags"`             // 全局自定义标签
	AutoProfiling   string      `toml:"auto_profiling"`   // 开关定时自动执行, 配置 0 则关闭
	PodCPULimit     string      `toml:"pod_cpu_limit"`    // pod resource limit
	PodMEMLimit     string      `toml:"pod_mem_limit"`    // pod resource limit
	Processes       []*Process  `toml:"processes"`        // 监控的进程列表
	HTTPConfig      *HTTPConfig `toml:"http"`             // http 配置
	Log             *Logging    `toml:"logging"`          // 日志配置
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
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_CPU_USAGE_PERCENT", i)); val != nil {
			process.CPUUsagePercent = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_PERCENT", i)); val != nil {
			process.MEMUsagePercent = *val
		}
		if val := getEnvInt(fmt.Sprintf("FLAMESHOT_PROCESSES_%d_MEM_USAGE_MB", i)); val != nil {
			process.MEMUsageMB = *val
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
	conf := &Config{}
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
