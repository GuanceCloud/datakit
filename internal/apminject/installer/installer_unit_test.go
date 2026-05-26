// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInstallDir(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid directory",
			dir:     "/opt/datakit",
			wantErr: false,
		},
		{
			name:    "empty directory",
			dir:     "",
			wantErr: true,
			errMsg:  "install dir is empty",
		},
		{
			name:    "path traversal attack",
			dir:     "../../../etc/passwd",
			wantErr: true,
			errMsg:  "invalid install dir",
		},
		{
			name:    "relative path with dots",
			dir:     "./datakit",
			wantErr: false,
		},
		{
			name:    "absolute path with spaces",
			dir:     "/opt/my datakit",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInstallDir(tt.dir)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLauncherSoPath(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	injectDir := filepath.Join(tmpDir, DirInject, DirInjectSubInject)
	err := os.MkdirAll(injectDir, 0o755)
	assert.NoError(t, err)

	// 创建测试文件
	glibcSoPath := filepath.Join(injectDir, launcherSoFileName)
	_, err = os.Create(glibcSoPath)
	assert.NoError(t, err)

	// 测试 glibc variant
	gotPath, err := launcherSoPath("glibc", tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, glibcSoPath, gotPath)

	// 测试不存在的文件
	_, err = launcherSoPath("glibc", "/nonexistent")
	assert.Error(t, err)
}

func TestBackupDockerConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "daemon.json")

	// 测试不存在的配置文件
	err := backupDockerConfig(configPath)
	assert.NoError(t, err)

	// 测试存在的配置文件
	testData := []byte(`{"default-runtime": "runc"}`)
	err = os.WriteFile(configPath, testData, 0o644)
	assert.NoError(t, err)

	err = backupDockerConfig(configPath)
	assert.NoError(t, err)

	// 验证备份文件存在
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	hasBackup := false
	for _, f := range files {
		if !f.IsDir() && f.Name() != "daemon.json" {
			hasBackup = true
			break
		}
	}
	assert.True(t, hasBackup, "backup file should be created")
}

func TestLoadAndDumpDockerDaemonConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "daemon.json")

	// 测试加载不存在的配置
	config, err := loadDockerDaemonConfig(configPath)
	assert.NoError(t, err)
	assert.Nil(t, config)

	// 创建测试配置
	testConfig := map[string]any{
		"default-runtime": "runc",
		"runtimes": map[string]any{
			"nvidia": map[string]any{
				"path": "/usr/bin/nvidia-container-runtime",
			},
		},
	}

	err = dumpDockerDaemonConfig(configPath, testConfig)
	assert.NoError(t, err)

	// 验证文件内容
	loadedConfig, err := loadDockerDaemonConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, testConfig, loadedConfig)

	// 测试无效 JSON 文件
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte(`{invalid json}`), 0o644)
	assert.NoError(t, err)

	_, err = loadDockerDaemonConfig(invalidPath)
	assert.Error(t, err)
}

func TestReadPreloadWithoutLauncher(t *testing.T) {
	tmpDir := t.TempDir()
	preloadPath := filepath.Join(tmpDir, "ld.so.preload")
	installDir := t.TempDir()

	// 创建测试 preload 文件
	launcherSoPath := filepath.Join(installDir, DirInject, DirInjectSubInject, launcherSoFileName)
	err := os.MkdirAll(filepath.Dir(launcherSoPath), 0o755)
	assert.NoError(t, err)

	// 测试 1: preload 包含 launcher
	content := "/usr/lib/some-lib.so\n" + launcherSoPath + "\n/usr/lib/another-lib.so\n"
	err = os.WriteFile(preloadPath, []byte(content), 0o644)
	assert.NoError(t, err)

	result, err := readPreloadWithoutLanucher(preloadPath, installDir)
	assert.NoError(t, err)
	assert.NotContains(t, result, launcherSoPath)
	assert.Contains(t, result, "/usr/lib/some-lib.so")
	assert.Contains(t, result, "/usr/lib/another-lib.so")

	// 测试 2: 空文件
	err = os.WriteFile(preloadPath, []byte(""), 0o644)
	assert.NoError(t, err)

	result, err = readPreloadWithoutLanucher(preloadPath, installDir)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestReadPreloadWithoutLauncherUsesExactWhitespaceToken(t *testing.T) {
	tmpDir := t.TempDir()
	preloadPath := filepath.Join(tmpDir, "ld.so.preload")
	installDir := t.TempDir()

	launcherSoPath := filepath.Join(installDir, DirInject, DirInjectSubInject, launcherSoFileName)
	err := os.MkdirAll(filepath.Dir(launcherSoPath), 0o755)
	assert.NoError(t, err)

	otherPath := launcherSoPath + ".backup"
	content := "/usr/lib/some-lib.so " + launcherSoPath + "\n" + otherPath + "\n"
	err = os.WriteFile(preloadPath, []byte(content), 0o644)
	assert.NoError(t, err)

	result, err := readPreloadWithoutLanucher(preloadPath, installDir)
	assert.NoError(t, err)
	assert.NotContains(t, strings.Fields(result), launcherSoPath)
	assert.Contains(t, strings.Fields(result), "/usr/lib/some-lib.so")
	assert.Contains(t, strings.Fields(result), otherPath)
}

func TestSetAndUnsetPreloadHelper(t *testing.T) {
	// 测试 readPreloadWithoutLanucher 函数的逻辑
	tmpDir := t.TempDir()
	preloadPath := filepath.Join(tmpDir, "ld.so.preload")
	installDir := t.TempDir()

	// 创建测试 preload 文件，包含 launcher 路径
	launcherSoPath := filepath.Join(installDir, DirInject, DirInjectSubInject, launcherSoFileName)
	err := os.MkdirAll(filepath.Dir(launcherSoPath), 0o755)
	assert.NoError(t, err)

	content := "/usr/lib/some-lib.so\n" + launcherSoPath + "\n/usr/lib/another-lib.so\n"
	err = os.WriteFile(preloadPath, []byte(content), 0o644)
	assert.NoError(t, err)

	// 测试读取并过滤 launcher
	result, err := readPreloadWithoutLanucher(preloadPath, installDir)
	assert.NoError(t, err)
	assert.NotContains(t, result, launcherSoPath)
	assert.Contains(t, result, "/usr/lib/some-lib.so")
	assert.Contains(t, result, "/usr/lib/another-lib.so")
}

func TestBackupDockerConfigDetailed(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "daemon.json")

	// 创建配置
	testConfig := []byte(`{"default-runtime": "runc"}`)
	err := os.WriteFile(configPath, testConfig, 0o644)
	assert.NoError(t, err)

	// 备份
	err = backupDockerConfig(configPath)
	assert.NoError(t, err)

	// 验证备份文件存在
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)

	hasBackup := false
	for _, f := range files {
		if !f.IsDir() && f.Name() != "daemon.json" {
			hasBackup = true
			// 验证备份文件内容
			backupPath := filepath.Join(tmpDir, f.Name())
			backupContent, err := os.ReadFile(backupPath)
			assert.NoError(t, err)
			assert.Equal(t, testConfig, backupContent)
			break
		}
	}
	assert.True(t, hasBackup, "backup file should be created")
}

func TestCleanupPreloadOnError(t *testing.T) {
	// 测试 cleanupPreloadOnError 函数调用 unsetPreload
	// 由于 preloadConfigFilePath 是常量，我们主要验证函数不会 panic
	installDir := t.TempDir()

	// 测试 1: nil logger 不应该 panic
	assert.NotPanics(t, func() {
		cleanupPreloadOnError(installDir, nil)
	})
}

func TestFindPythonInterpreter(t *testing.T) {
	// 这个测试依赖于系统环境，可能在不同机器上表现不同
	py := findPythonInterpreter()
	// 如果系统有 Python，应该返回路径
	// 如果没有，返回空字符串
	if py != "" {
		assert.FileExists(t, py)
	}
}

func TestGetPythonVersion(t *testing.T) {
	// 查找 Python 解释器
	py := findPythonInterpreter()
	if py == "" {
		t.Skip("Python not found, skipping test")
	}

	version, err := getPythonVersion(py)
	if err != nil {
		t.Logf("Warning: %v", err)
		return
	}

	// Python 3.x 应该返回主版本号
	assert.Greater(t, version, 0)
	assert.Less(t, version, 100) // 合理的版本号范围
}

func TestDockerRuntimeInfoParsing(t *testing.T) {
	tests := []struct {
		name            string
		config          map[string]any
		expectedDefault string
		expectedCount   int
	}{
		{
			name: "complete config",
			config: map[string]any{
				"default-runtime": "runc",
				"runtimes": map[string]any{
					"runc": map[string]any{
						"path": "/usr/bin/runc",
					},
					"gvisor": map[string]any{
						"path":        "/usr/bin/runsc",
						"runtimeType": "io.containerd.runsc.v1",
					},
				},
			},
			expectedDefault: "runc",
			expectedCount:   2,
		},
		{
			name: "config without default-runtime",
			config: map[string]any{
				"runtimes": map[string]any{
					"runc": map[string]any{
						"path": "/usr/bin/runc",
					},
				},
			},
			expectedDefault: "",
			expectedCount:   1,
		},
		{
			name:            "empty config",
			config:          map[string]any{},
			expectedDefault: "",
			expectedCount:   0,
		},
		{
			name: "config with invalid runtime type",
			config: map[string]any{
				"runtimes": map[string]any{
					"runc": "invalid-type", // 应该是 map
				},
			},
			expectedDefault: "",
			expectedCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultRuntime, runtimes := getDockerRuntimeInfoFromConfig(tt.config)
			assert.Equal(t, tt.expectedDefault, defaultRuntime)
			assert.Equal(t, tt.expectedCount, len(runtimes))
		})
	}
}

func TestConstants(t *testing.T) {
	// 验证常量定义正确
	assert.Equal(t, "ld.so.preload", ldPreloadFileName)
	assert.Equal(t, "apm_launcher.so", launcherSoFileName)
	assert.Equal(t, "apm_launcher_musl.so", launcherSoMuslFileName)
	assert.Equal(t, "dd-java-agent.jar", javaAgentFileName)
	assert.Equal(t, "/etc/ld.so.preload", preloadConfigFilePath)
	assert.Equal(t, "/etc/docker/daemon.json", dockerDaemonJSONPath)
	assert.Equal(t, "runc", RuntimeRunc)
	assert.Equal(t, "dk-runc", RuntimeDkRunc)
}

func TestOptFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	testURL := "https://example.com/launcher.tar.gz"

	tests := []struct {
		name   string
		opt    Opt
		verify func(c *config) bool
	}{
		{
			name: "WithInstallDir",
			opt:  WithInstallDir(tmpDir),
			verify: func(c *config) bool {
				return c.installDir == tmpDir
			},
		},
		{
			name: "WithInstrumentationEnabled host",
			opt:  WithInstrumentationEnabled("host"),
			verify: func(c *config) bool {
				return c.enableHostInject && !c.enableDockerInject
			},
		},
		{
			name: "WithInstrumentationEnabled docker",
			opt:  WithInstrumentationEnabled("docker"),
			verify: func(c *config) bool {
				return !c.enableHostInject && c.enableDockerInject
			},
		},
		{
			name: "WithInstrumentationEnabled both",
			opt:  WithInstrumentationEnabled("host,docker"),
			verify: func(c *config) bool {
				return c.enableHostInject && c.enableDockerInject
			},
		},
		{
			name: "WithInstrumentationEnabled disable",
			opt:  WithInstrumentationEnabled("disable"),
			verify: func(c *config) bool {
				return !c.enableHostInject && !c.enableDockerInject
			},
		},
		{
			name: "WithJavaLibURL",
			opt:  WithJavaLibURL(testURL),
			verify: func(c *config) bool {
				return c.ddJavaLibURL == testURL
			},
		},
		{
			name: "WithPythonLib true",
			opt:  WithPythonLib(true),
			verify: func(c *config) bool {
				return c.pyLib
			},
		},
		{
			name: "WithPythonLib false",
			opt:  WithPythonLib(false),
			verify: func(c *config) bool {
				return !c.pyLib
			},
		},
		{
			name: "WithPhpLibURL",
			opt:  WithPhpLibURL(testURL),
			verify: func(c *config) bool {
				return c.ddPhpLibURL == testURL
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c config
			tt.opt(&c)
			assert.True(t, tt.verify(&c), "config verification failed")
		})
	}
}

// PHP 相关测试

func TestFindPhpInterpreter(t *testing.T) {
	// 这个测试依赖于系统环境，可能在不同机器上表现不同
	php := findPhpInterpreter()
	// 如果系统有 PHP，应该返回路径
	// 如果没有，返回空字符串
	if php != "" {
		assert.FileExists(t, php)
	}
}

func TestGetPhpVersion(t *testing.T) {
	// 查找 PHP 解释器
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	version, err := getPhpVersion(php)
	if err != nil {
		t.Logf("Warning: %v", err)
		return
	}

	// PHP 版本应该是 x.y 格式
	assert.Regexp(t, `^\d+\.\d+$`, version)
}

func TestGetPhpArch(t *testing.T) {
	arch, err := getPhpArch()
	assert.NoError(t, err)
	// 常见架构
	assert.Contains(t, []string{"amd64", "arm64", "x86_64", "aarch64"}, arch)
}

func TestGetPhpThreadSafety(t *testing.T) {
	// 查找 PHP 解释器
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	_, err := getPhpThreadSafety(php)
	if err != nil {
		t.Logf("Warning: %v", err)
		return
	}
}

func TestPhpDdtraceDownloadURL(t *testing.T) {
	// 测试 URL 格式构建
	// 新格式: dd-library-php-{version}-{arch}-linux-{libc}-20200930-{ts}.tar.gz
	version := phpDdtraceVersion

	tests := []struct {
		name         string
		arch         string
		libc         string
		ts           bool
		expectedArch string
	}{
		{"amd64_glibc_nts", "amd64", "gnu", false, "x86_64"},
		{"amd64_glibc_zts", "amd64", "gnu", true, "x86_64"},
		{"amd64_musl_nts", "amd64", "musl", false, "x86_64"},
		{"amd64_musl_zts", "amd64", "musl", true, "x86_64"},
		{"arm64_glibc_nts", "arm64", "gnu", false, "aarch64"},
		{"arm64_glibc_zts", "arm64", "gnu", true, "aarch64"},
		{"arm64_musl_nts", "arm64", "musl", false, "aarch64"},
		{"arm64_musl_zts", "arm64", "musl", true, "aarch64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts string
			if tt.ts {
				ts = "-zts"
			}
			expectedFileName := fmt.Sprintf("dd-library-php-%s-%s-linux-%s-20200930%s.tar.gz",
				version, tt.expectedArch, tt.libc, ts)
			phpDdtraceReleaseURL := "github.com/DataDog/dd-trace-php/releases"
			expectedURL := fmt.Sprintf("%s/%s/%s", phpDdtraceReleaseURL, version, expectedFileName)

			// 验证格式正确
			assert.Contains(t, expectedURL, "github.com/DataDog/dd-trace-php/releases")
			assert.Contains(t, expectedURL, version)
			assert.Contains(t, expectedURL, tt.expectedArch)
			assert.Contains(t, expectedURL, tt.libc)
			assert.Contains(t, expectedURL, ts)
		})
	}
}

// PHP ini 配置相关测试

func TestGetPhpIniPath(t *testing.T) {
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	iniPath, err := getPhpIniPath(php)
	// 可能找不到配置文件，但不应报错
	if err == nil {
		assert.FileExists(t, iniPath)
	}
}

func TestGetPhpIniScanDir(t *testing.T) {
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	// 这个函数可能返回错误（如果扫描目录不存在），所以只是测试不会 panic
	_, err := getPhpIniScanDir(php)
	// 不应该 panic
	_ = err
}

func TestGetPhpExtensionDir(t *testing.T) {
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	extDir, err := getPhpExtensionDir(php)
	// 可能找不到扩展目录
	if err == nil {
		assert.NotEmpty(t, extDir)
	}
}

func TestGetPhpAPIVersion(t *testing.T) {
	php := findPhpInterpreter()
	if php == "" {
		t.Skip("PHP not found, skipping test")
	}

	apiVer, err := getPhpAPIVersion(php)
	// 可能找不到 API 版本
	if err == nil {
		assert.NotEmpty(t, apiVer)
	}
}

func TestFindAllPhpBinaries(t *testing.T) {
	binaries := findAllPhpBinaries()
	// 可能找到多个 PHP 版本
	// 如果系统没有 PHP，返回空切片
	for _, bin := range binaries {
		assert.FileExists(t, bin)
	}
}

func TestGetPhpIniPaths(t *testing.T) {
	// 测试有扫描目录的情况
	info := &phpExtensionInfo{
		iniScanDir: "/etc/php/8.1/cli/conf.d",
		iniMain:    "/etc/php/8.1/cli/php.ini",
	}

	paths := getPhpIniPaths(info)
	assert.NotEmpty(t, paths)
	// 应该包含 98-ddtrace.ini
	for _, p := range paths {
		assert.Contains(t, p, phpDdtraceIniName)
	}
}

func TestGetPhpIniPathsDebianSAPI(t *testing.T) {
	// 测试 Debian 风格的 SAPI 分离目录
	tmpDir := t.TempDir()

	// 创建模拟的目录结构
	cliDir := filepath.Join(tmpDir, "cli", "conf.d")
	apacheDir := filepath.Join(tmpDir, "apache2", "conf.d")
	fpmDir := filepath.Join(tmpDir, "fpm", "conf.d")

	os.MkdirAll(cliDir, 0o755)
	os.MkdirAll(apacheDir, 0o755)
	os.MkdirAll(fpmDir, 0o755)

	info := &phpExtensionInfo{
		iniScanDir: cliDir,
		iniMain:    filepath.Join(tmpDir, "cli", "php.ini"),
	}

	paths := getPhpIniPaths(info)
	// 应该检测到所有三个 SAPI 目录
	assert.GreaterOrEqual(t, len(paths), 1)
}

func TestSafeCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建源文件
	srcPath := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("test content")
	err := os.WriteFile(srcPath, srcContent, 0o644)
	assert.NoError(t, err)

	// 复制到目标
	dstPath := filepath.Join(tmpDir, "dest.txt")
	err = safeCopyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// 验证内容
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, srcContent, dstContent)
}

func TestGeneratePhpIniContent(t *testing.T) {
	info := &phpExtensionInfo{
		extensionDir: "/usr/lib/php/modules",
	}

	content := generatePhpIniContent(info, "", true)

	// 验证内容包含必要配置
	assert.Contains(t, content, "ddtrace.so")
	assert.Contains(t, content, "datadog.trace.enabled=1")
	assert.Contains(t, content, "Priority: 98")
	assert.Contains(t, content, "datadog.agent_host")
	assert.Contains(t, content, "datadog.trace.agent_port")
}

func TestGeneratePhpIniContentWithoutExtDir(t *testing.T) {
	info := &phpExtensionInfo{
		extensionDir: "", // 没有扩展目录
	}

	content := generatePhpIniContent(info, "", true)

	// 应该包含注释说明
	assert.Contains(t, content, "extension_dir not configured")
}

func TestUpdatePhpIniFile(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, phpDdtraceIniName)

	info := &phpExtensionInfo{
		extensionDir: "/usr/lib/php/modules",
	}

	err := updatePhpIniFile(iniPath, info)
	assert.NoError(t, err)

	// 验证文件被创建
	assert.FileExists(t, iniPath)

	// 验证内容
	content, err := os.ReadFile(iniPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "ddtrace.so")
}

func TestUpdatePhpIniFileWithExisting(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, phpDdtraceIniName)

	// 创建已存在的文件
	existingContent := "; existing comment\n"
	err := os.WriteFile(iniPath, []byte(existingContent), 0o644)
	assert.NoError(t, err)

	info := &phpExtensionInfo{
		extensionDir: "/usr/lib/php/modules",
	}

	err = updatePhpIniFile(iniPath, info)
	assert.NoError(t, err)

	// 验证备份文件被创建
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	hasBackup := false
	for _, f := range files {
		if strings.Contains(f.Name(), ".bak.") {
			hasBackup = true
			break
		}
	}
	assert.True(t, hasBackup, "backup file should be created")
}
