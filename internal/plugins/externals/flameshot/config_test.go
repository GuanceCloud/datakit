// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/toml"
)

func TestConfigMarshal(t *testing.T) {
	c := &Config{
		DataKitAddr:     "http://localhost:9529",
		ProfilingPath:   "/profiling/v1/input",
		MonitorInterval: "1s",
		Tags:            []string{"host:my_host"},
		AutoProfiling:   "5m",
		Processes: []*Process{
			{
				Service:         "tmall",
				Command:         "java -jar tmall.jar",
				Events:          "all",
				Language:        "java",
				JDKVersion:      "",
				Tags:            []string{"env:env", "version:1.0.0"},
				CPUUsagePercent: 80,
				MEMUsagePercent: 80,
				MEMUsageMB:      1024,
			},
			{
				Service:         "springboot-server",
				Command:         "springbooot-server.jar",
				Events:          "cpu",
				Duration:        defaultConfig,
				Language:        "java",
				JDKVersion:      "",
				Tags:            []string{"env:env", "version:1.0.0"},
				CPUUsagePercent: 80,
				MEMUsagePercent: 80,
				MEMUsageMB:      1024,
			},
		},
		HTTPConfig: &HTTPConfig{
			LocalHost: "localhost",
			LocalPort: "8089",
		},
		Log: &Logging{
			Level: "debug",
			Path:  "log",
		},
	}

	bts, err := toml.Marshal(c)
	assert.NoError(t, err)
	t.Logf("%s", string(bts))

	// --- test unmarshal ---
	c2 := &Config{}

	err = toml.NewDecoder(bytes.NewBuffer([]byte(defaultConfig))).Decode(c2)
	assert.NoError(t, err)
	assert.NotNil(t, c2)
	assert.NotEmpty(t, c2.DataKitAddr)
	assert.NotEmpty(t, c2.ProfilingPath)
	assert.NotEmpty(t, c2.MonitorInterval)
	assert.NotEmpty(t, c2.AutoProfiling)
	assert.NotEmpty(t, c2.Tags)
	assert.NotEmpty(t, c2.Processes)
	assert.NotEmpty(t, c2.HTTPConfig)
	assert.NotEmpty(t, c2.HTTPConfig.LocalHost)
	assert.NotEmpty(t, c2.HTTPConfig.LocalPort)
	assert.NotEmpty(t, c2.Processes[0].Service)
	assert.NotEmpty(t, c2.Processes[0].Command)
	assert.NotEmpty(t, c2.Processes[0].Events)
	assert.NotEmpty(t, c2.Processes[0].Language)
	assert.NotEmpty(t, c2.Processes[0].JDKVersion)
	assert.NotEmpty(t, c2.Processes[0].Tags)
	assert.NotEmpty(t, c2.Processes[0].CPUUsagePercent)
	assert.NotEmpty(t, c2.Processes[0].MEMUsagePercent)
	assert.NotEmpty(t, c2.Processes[0].MEMUsageMB)
	assert.NotEmpty(t, c2.Log)
}

func TestInitConfig(t *testing.T) {
	c := &Config{}
	err := toml.Unmarshal([]byte(defaultConfig), c)
	assert.NoError(t, err)
	t.Logf("%s", c.toString())
}

func (c *Config) toString() string {
	msg := ""
	for _, process := range c.Processes {
		msg += fmt.Sprintf("process:{Service:%s,Command:%s,Events:%s,Language:%s,JDKVersion:%s,Tags:%+v,CPUUsagePercent:%d,MEMUsagePercent:%d,MEMUsageBytes:%d}\n",
			process.Service, process.Command, process.Events, process.Language, process.JDKVersion, process.Tags, process.CPUUsagePercent, process.MEMUsagePercent, process.MEMUsageMB)
	}
	return fmt.Sprintf("config:{DataKitAddr:%s,MonitorInterval:%s, \n Processes:%s,HTTPConfig:%+v}", c.DataKitAddr, c.MonitorInterval, msg, c.HTTPConfig)
}

func TestConfig_loadProcessesFromEnv(t *testing.T) {
	c := &Config{Processes: make([]*Process, 0)}
	// setenv
	t.Setenv("FLAMESHOT_PROCESSES_0_SERVICE", "tmall")
	t.Setenv("FLAMESHOT_PROCESSES_0_COMMAND", "java -jar tmall.jar")
	t.Setenv("FLAMESHOT_PROCESSES_0_EVENTS", "all")
	t.Setenv("FLAMESHOT_PROCESSES_0_LANGUAGE", "java")
	t.Setenv("FLAMESHOT_PROCESSES_0_JDK_VERSION", "")
	t.Setenv("FLAMESHOT_PROCESSES_0_TAGS", "[\"env:env\",\"version:1.0.0\"]")
	t.Setenv("FLAMESHOT_PROCESSES_0_CPU_USAGE_PERCENT", "80")
	t.Setenv("FLAMESHOT_PROCESSES_0_MEM_USAGE_PERCENT", "80")
	t.Setenv("FLAMESHOT_PROCESSES_0_MEM_USAGE_MB", "1024")
	t.Setenv("FLAMESHOT_PROCESSES_1_SERVICE", "tmall_server")
	t.Setenv("FLAMESHOT_PROCESSES_1_COMMAND", "^java\\b.*tmall\\.jar")
	t.Setenv("FLAMESHOT_PROCESSES_1_EVENTS", "cpu,mem")
	t.Setenv("FLAMESHOT_PROCESSES_1_LANGUAGE", "java")
	t.Setenv("FLAMESHOT_PROCESSES_1_JDK_VERSION", "")
	t.Setenv("FLAMESHOT_PROCESSES_1_TAGS", "[\"env:env\",\"version:1.0.0\"]")
	t.Setenv("FLAMESHOT_PROCESSES_1_CPU_USAGE_PERCENT", "80")
	t.Setenv("FLAMESHOT_PROCESSES_1_MEM_USAGE_PERCENT", "80")
	t.Setenv("FLAMESHOT_PROCESSES_1_MEM_USAGE_MB", "1024")

	p := &Process{
		Service:         "tmall",
		Command:         "java -jar tmall.jar",
		Duration:        "60s",
		Events:          "--all",
		Language:        "java",
		JDKVersion:      "-",
		Tags:            []string{"env:testing", "version:1.0.0"},
		CPUUsagePercent: 80,
		MEMUsagePercent: 80,
		MEMUsageMB:      1024,
	}
	bts, err := json.Marshal(p)
	assert.NoError(t, err)
	t.Logf("process json %s", string(bts))
	t.Setenv("FLAMESHOT_PROCESSES_0", string(bts))
	t.Setenv("FLAMESHOT_PROCESSES_1", "{\"service\":\"tmall\",\"command\":\"^.*org\\\\.springframework\\\\.boot\\\\.loader\\\\.JarLauncher$\",\"duration\":\"60s\",\"events\":\"--all\",\"language\":\"java\",\"jdk_version\":\"-\",\"tags\":[\"env:testing\",\"version:1.0.0\"],\"cpu_usage_percent\":80,\"mem_usage_percent\":80,\"mem_usage_mb\":1024}")

	// 数组形式
	t.Setenv("FLAMESHOT_PROCESSES", "[{\"service\":\"jfr-parser\",\"command\":\"^.*org\\\\.springframework\\\\.boot\\\\.loader\\\\.JarLauncher$\",\"duration\":\"60s\",\"events\":\"--all\",\"language\":\"java\",\"jdk_version\":\"-\",\"tags\":[\"env:testing\",\"version:1.0.0\"],\"cpu_usage_percent\":80,\"mem_usage_percent\":80,\"mem_usage_mb\":1024}]")

	c.loadProcessesFromEnv()
	if len(c.Processes) < 2 {
		t.Errorf("loadProcessesFromEnv failed, processes length is %d", len(c.Processes))
	}

	for i, process := range c.Processes {
		t.Logf("%d process %+v", i, process)
		assert.NotEmpty(t, process.Service)
		assert.NotEmpty(t, process.Command)
		t.Logf("command :%s", process.Command)
		assert.NotEmpty(t, process.Events)
		assert.NotEmpty(t, process.Language)
		assert.NotEmpty(t, process.Tags)
		assert.NotEmpty(t, process.CPUUsagePercent)
		assert.NotEmpty(t, process.MEMUsagePercent)
		assert.NotEmpty(t, process.MEMUsageMB)
		assert.Equal(t, process.MEMUsageMB, 1024)
		assert.Equal(t, process.MEMUsagePercent, 80)
		assert.Equal(t, process.CPUUsagePercent, 80)
	}
}

func TestConfig_FromEnv(t *testing.T) {
	c := &Config{}
	t.Setenv("FLAMESHOT_DATAKIT_ADDR", "http://127.0.0.1:9529/profile/v1/input")
	t.Setenv("FLAMESHOT_PROFILING_PATH", "/data")
	t.Setenv("FLAMESHOT_MONITOR_INTERVAL", "10s")
	t.Setenv("FLAMESHOT_LOG_LEVEL", "debug")
	t.Setenv("FLAMESHOT_LOG_PATH", "/tmp/flameshot.log")
	t.Setenv("FLAMESHOT_HTTP_LOCAL_IP", "0.0.0.0")
	t.Setenv("FLAMESHOT_HTTP_LOCAL_PORT", "8089")
	t.Setenv("FLAMESHOT_TAGS", "env:env,version:1.0.0")
	t.Setenv("FLAMESHOT_AUTO_PROFILING", "30s")

	c.fromEnv()
	assert.NotEmpty(t, c.DataKitAddr)
	assert.NotEmpty(t, c.ProfilingPath)
	assert.NotEmpty(t, c.MonitorInterval)
	t.Log(c.MonitorInterval)
	assert.NotEmpty(t, c.Tags)
	assert.NotEmpty(t, c.Log.Path)
	assert.NotEmpty(t, c.Log.Level)
	assert.NotEmpty(t, c.HTTPConfig.LocalPort)
	assert.NotEmpty(t, c.HTTPConfig.LocalHost)
	t.Logf("config AutoProfiling %+v", c.AutoProfiling)
	assert.Equal(t, c.AutoProfiling, "5m")
}

func TestRegex(t *testing.T) {
	rex := "^.*org\\.springframework\\.boot\\.loader\\.JarLauncher$"

	command := "java -Xmx1920m -Xms1536m -javaagent:dd-java-agent.jar org.springframework.boot.loader.JarLauncher"
	re, err := regexp.Compile(rex)
	assert.NoError(t, err)
	if re.MatchString(command) {
		t.Logf("matched")
	} else {
		t.Errorf("not matched")
	}
}
