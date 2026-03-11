// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package installer

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

// validateInstallDir 验证安装目录的合法性.
func validateInstallDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("install dir is empty")
	}
	// 防止路径遍历攻击
	cleanDir := filepath.Clean(dir)
	if strings.HasPrefix(cleanDir, "..") {
		return fmt.Errorf("invalid install dir: %s", dir)
	}
	return nil
}

func Download(log *logger.Logger, opt ...Opt) error {
	var c config
	for _, fn := range opt {
		fn(&c)
	}

	if c.installDir == "" {
		c.installDir = dirDkInstall
	}

	if err := validateInstallDir(c.installDir); err != nil {
		return err
	}

	if c.launcherURL == "" {
		return fmt.Errorf("apm inject url is empty")
	}

	if c.cli == nil {
		// 注意：这里跳过 SSL 验证是为了兼容自签名证书的 internal 环境
		// 在生产环境中建议使用有效的 SSL 证书
		c.cli = httpcli.Cli(&httpcli.Options{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		})
	}

	cp.Printf("\n")
	dl.CurDownloading = "apm-inject"
	injTo := filepath.Join(c.installDir, DirInject, DirInjectSubInject)
	cp.Infof("Downloading %s => %s\n", c.launcherURL, injTo)
	if err := dl.Download(c.cli, c.launcherURL, injTo,
		true, false); err != nil {
		return err
	}

	if !(c.enableHostInject || c.enableDockerInject) {
		return nil
	}

	if c.ddJavaLibURL != "" {
		cp.Printf("\n")
		dl.CurDownloading = "apm-lib-java"
		injTo = filepath.Join(c.installDir, DirInject, DirInjectSubLib, "java", javaAgentFileName)
		cp.Infof("Downloading %s => %s\n", c.ddJavaLibURL, injTo)
		if err := dl.Download(c.cli, c.ddJavaLibURL, injTo, true, true); err != nil {
			log.Warn(err)
		}
	}

	if c.pyLib {
		installPythonLib(log)
	}

	if c.ddPhpLibURL != "" {
		installPhpLib(c.installDir, c.ddPhpLibURL, c.cli, log)
	}

	return nil
}

// installPythonLib 安装 Python ddtrace 库.
func installPythonLib(log *logger.Logger) {
	cp.Printf("\n")
	cp.Infof("Installing ddtrace python library\n")

	py := findPythonInterpreter()
	if py == "" {
		log.Warn("python not found")
		return
	}

	py3Ver, err := getPythonVersion(py)
	if err != nil {
		log.Warn(err)
		return
	}

	if py3Ver >= 7 {
		installDdtracePackage(py, log)
	}
}

// findPythonInterpreter 查找 Python 解释器路径.
func findPythonInterpreter() string {
	if py, err := exec.LookPath("python3"); err == nil {
		return py
	}
	if py, err := exec.LookPath("python"); err == nil {
		return py
	}
	return ""
}

// getPythonVersion 获取 Python 主版本号.
func getPythonVersion(pyPath string) (int, error) {
	//nolint:gosec
	ver, err := exec.Command(pyPath, "-V").CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("%s -V error: %w", pyPath, err)
	}

	v := py3Regexp.FindStringSubmatch(string(ver))
	if len(v) != 2 {
		return 0, fmt.Errorf("parse python version error: %s", string(ver))
	}

	py3Ver, err := strconv.Atoi(v[1])
	if err != nil {
		return 0, fmt.Errorf("convert python version error: %w", err)
	}

	return py3Ver, nil
}

// installDdtracePackage 安装 ddtrace 包.
func installDdtracePackage(pyPath string, log *logger.Logger) {
	// set env: PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
	//nolint:gosec
	output, err := exec.Command(pyPath, "-m", "pip", "install", "ddtrace").CombinedOutput()
	if err != nil {
		log.Warn(string(output))
		log.Warnf("pip install ddtrace error: %s", err.Error())
	} else {
		log.Info(string(output))
	}
}

// phpExtensionInfo 存储 PHP 扩展相关信息.
type phpExtensionInfo struct {
	phpPath      string // PHP 二进制路径
	version      string // PHP 版本 (如 "8.1")
	arch         string // 架构 (amd64/arm64)
	threadSafety bool   // 线程安全 (zts)
	extensionDir string // 扩展目录
	iniScanDir   string // INI 扫描目录
	iniMain      string // 主 php.ini 路径
	phpAPI       string // PHP API 版本
}

// installPhpLib 安装 PHP ddtrace 扩展.
func installPhpLib(installDir string, ddLibURL string, cli *http.Client, log *logger.Logger) {
	cp.Printf("\n")
	cp.Infof("Installing ddtrace PHP extension\n")

	// 查找所有可用的 PHP 二进制
	phpBinaries := findAllPhpBinaries()
	if len(phpBinaries) == 0 {
		if log != nil {
			log.Warn("no PHP interpreter found")
		}
		return
	}

	// 如果 client 为 nil，创建默认 client
	if cli == nil {
		cli = httpcli.Cli(&httpcli.Options{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint:gosec
		})
	}

	for _, php := range phpBinaries {
		cp.Infof("Processing PHP binary: %s\n", php)

		// 获取 PHP 扩展信息
		info, err := getPhpExtensionInfo(php)
		if err != nil {
			if log != nil {
				log.Warnf("get PHP info for %s failed: %v", php, err)
			}
			continue
		}

		// 下载并安装扩展
		if err := downloadAndInstallPhpDdtrace(installDir, ddLibURL, cli, info, log); err != nil {
			if log != nil {
				log.Warnf("install PHP ddtrace for %s failed: %v", php, err)
			}
			continue
		}

		cp.Infof("Successfully installed ddtrace for PHP %s (%s)\n", info.version, php)
	}
}

// findAllPhpBinaries 查找系统中所有 PHP 二进制文件.
func findAllPhpBinaries() []string {
	var binaries []string

	// 常见的 PHP 二进制路径
	commonPaths := []string{
		"php",
		"php8",
		"php8.3",
		"php8.2",
		"php8.1",
		"php8.0",
		"php7.4",
		"php7.3",
	}

	seen := make(map[string]bool)

	for _, name := range commonPaths {
		if path, err := exec.LookPath(name); err == nil {
			// 解析符号链接，获取真实路径
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				realPath = path
			}
			if !seen[realPath] {
				seen[realPath] = true
				binaries = append(binaries, path)
			}
		}
	}

	return binaries
}

// getPhpExtensionInfo 获取 PHP 扩展安装所需的所有信息.
func getPhpExtensionInfo(phpPath string) (*phpExtensionInfo, error) {
	info := &phpExtensionInfo{phpPath: phpPath}

	// 获取 PHP 版本
	ver, err := getPhpVersion(phpPath)
	if err != nil {
		return nil, err
	}
	info.version = ver

	// 获取架构
	arch, err := getPhpArch()
	if err != nil {
		return nil, err
	}
	info.arch = arch

	// 获取线程安全信息
	ts, err := getPhpThreadSafety(phpPath)
	if err != nil {
		// 默认使用 nts
		ts = false
	}
	info.threadSafety = ts

	// 获取扩展目录
	extDir, err := getPhpExtensionDir(phpPath)
	if err != nil {
		// 使用默认扩展目录
		extDir = ""
	}
	info.extensionDir = extDir

	// 获取 INI 扫描目录
	scanDir, _ := getPhpIniScanDir(phpPath)
	info.iniScanDir = scanDir

	// 获取主 php.ini 路径
	iniMain, _ := getPhpIniPath(phpPath)
	info.iniMain = iniMain

	// 获取 PHP API 版本
	apiVer, err := getPhpAPIVersion(phpPath)
	if err != nil {
		apiVer = ""
	}
	info.phpAPI = apiVer

	return info, nil
}

// getPhpExtensionDir 获取 PHP 扩展目录.
func getPhpExtensionDir(phpPath string) (string, error) {
	//nolint:gosec
	out, err := exec.Command(phpPath, "-i").CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "extension_dir") {
			parts := strings.Split(line, "=>")
			if len(parts) >= 2 {
				extDir := strings.TrimSpace(parts[len(parts)-1])
				if extDir != "" && extDir != "(none)" {
					return extDir, nil
				}
			}
		}
	}

	return "", fmt.Errorf("extension_dir not found")
}

// getPhpAPIVersion 获取 PHP API 版本（用于选择正确的扩展文件）.
func getPhpAPIVersion(phpPath string) (string, error) {
	//nolint:gosec
	out, err := exec.Command(phpPath, "-i").CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "PHP API") {
			parts := strings.Split(line, "=>")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("PHP API version not found")
}

// downloadAndInstallPhpDdtrace 下载并安装 PHP ddtrace 扩展.
func downloadAndInstallPhpDdtrace(installDir string, ddLibURL string, cli *http.Client, info *phpExtensionInfo, log *logger.Logger) error {
	// 检测 PHP 二进制文件链接的 libc 类型（musl 或 glibc）
	// 通过 ldd 查看 PHP 依赖的动态库，更准确地判断需要哪个版本的 ddtrace
	libcType := utils.DetectLibcTypeFromBinary(info.phpPath)

	// 输出调试信息
	cp.Infof("PHP info: version=%s, arch=%s, threadSafety=%v, libc=%s\n",
		info.version, info.arch, info.threadSafety, libcType)

	// 构建下载 URL
	// 新格式: dd-library-php-{version}-{arch}-linux-{libc}-{phpapi}{-zts}.tar.gz
	// arch: x86_64, aarch64
	// libc: gnu, musl
	// ts: zts, nts
	var fileName string
	var downloadURL string

	// 转换架构名称: amd64 -> x86_64, arm64 -> aarch64
	archName := info.arch
	switch archName {
	case "amd64":
		archName = "x86_64"
	case "arm64":
		archName = "aarch64"
	}

	// 确定 libc 类型
	var libcName string
	if libcType == utils.Muslc {
		libcName = "musl"
	} else {
		libcName = "gnu"
	}

	// 构建文件名
	ts := ""
	if info.threadSafety {
		ts = "-zts"
	}
	fileName = fmt.Sprintf("dd-library-php-%s-%s-linux-%s-%s%s.tar.gz",
		phpDdtraceVersion, archName, libcName, info.phpAPI, ts)

	downloadURL, err := url.JoinPath(ddLibURL, phpDdtraceVersion, fileName)
	if err != nil {
		return fmt.Errorf("construct download URL failed: %w", err)
	}

	cp.Printf("\n")
	dl.CurDownloading = "apm-lib-php"

	// 目标安装目录: {installDir}/apm_inject/lib/php/
	targetDir := filepath.Join(installDir, DirInject, DirInjectSubLib, "php")
	if err := os.MkdirAll(targetDir, 0o755); err != nil { // nolint:gosec
		return fmt.Errorf("create target dir failed: %w", err)
	}

	// 直接下载并解压到目标安装目录
	// dl.Download 的第4个参数表示是否显示进度，第5个参数 downloadOnly 为 false 时自动解压
	cp.Infof("Downloading and extracting %s => %s\n", downloadURL, targetDir)

	if err := dl.Download(cli, downloadURL, targetDir, false, false); err != nil {
		return fmt.Errorf("download and extract failed: %w", err)
	}

	// 安装扩展文件到系统 PHP 扩展目录
	if err := installPhpExtensionFiles(targetDir, info); err != nil {
		return fmt.Errorf("install extension files failed: %w", err)
	}

	// 配置 INI 文件
	if err := configurePhpIniFromInfo(info, log); err != nil {
		return fmt.Errorf("configure php.ini failed: %w", err)
	}

	return nil
}

// installPhpExtensionFiles 安装 PHP 扩展文件.
func installPhpExtensionFiles(extractDir string, info *phpExtensionInfo) error {
	// 查找扩展文件
	// 目录结构: dd-library-php/trace/ext/{php_api}/ddtrace.so
	extPattern := filepath.Join(extractDir, "dd-library-php", "trace", "ext", "*", "ddtrace*.so")
	matches, err := filepath.Glob(extPattern)
	if err != nil {
		return fmt.Errorf("glob extension failed: %w", err)
	}

	if len(matches) == 0 {
		// 尝试另一种目录结构
		extPattern = filepath.Join(extractDir, "trace", "ext", "*", "ddtrace*.so")
		matches, err = filepath.Glob(extPattern)
		if err != nil {
			return fmt.Errorf("glob extension failed: %w", err)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no ddtrace extension found in archive")
	}

	// 选择正确的扩展文件
	var extFile string

	for _, m := range matches {
		// 根据线程安全选择正确的文件
		if !info.threadSafety && !strings.Contains(m, "-zts") {
			extFile = m
			break
		} else if info.threadSafety && strings.Contains(m, "-zts") {
			extFile = m
			break
		}
	}

	if extFile == "" {
		extFile = matches[0] // 使用第一个找到的文件
	}

	// 目标路径
	extDir := info.extensionDir
	if extDir == "" {
		extDir = "/usr/lib/php/modules"
	}

	// 确保目标目录存在
	if err := os.MkdirAll(extDir, 0o755); err != nil { // nolint:gosec
		return fmt.Errorf("create extension dir failed: %w", err)
	}

	// 目标文件路径
	destFile := filepath.Join(extDir, "ddtrace.so")

	// 安全复制（使用临时文件+重命名，避免 segfault）
	if err := safeCopyFile(extFile, destFile); err != nil {
		return fmt.Errorf("copy extension failed: %w", err)
	}

	cp.Infof("Installed ddtrace.so to %s\n", destFile)

	return nil
}

// safeCopyFile 安全复制文件（使用临时文件+重命名，避免正在使用的文件导致 segfault）.
func safeCopyFile(src, dst string) error {
	// 读取源文件
	data, err := os.ReadFile(src) // nolint:gosec
	if err != nil {
		return err
	}

	// 先写入临时文件
	tmpFile := dst + ".tmp." + strconv.FormatInt(time.Now().UnixNano(), 36)
	if err := os.WriteFile(tmpFile, data, 0o755); err != nil { // nolint:gosec
		return err
	}

	// 重命名到目标文件
	if err := os.Rename(tmpFile, dst); err != nil {
		_ = os.Remove(tmpFile) // 清理临时文件
		return err
	}

	return nil
}

// configurePhpIniFromInfo 根据 PHP 信息配置 INI 文件.
func configurePhpIniFromInfo(info *phpExtensionInfo, log *logger.Logger) error {
	// 获取所有需要配置的 INI 文件路径
	iniPaths := getPhpIniPaths(info)

	if len(iniPaths) == 0 {
		return fmt.Errorf("no php.ini found")
	}

	for _, iniPath := range iniPaths {
		if err := updatePhpIniFile(iniPath, info); err != nil {
			if log != nil {
				log.Warnf("update %s failed: %v", iniPath, err)
			}
			continue
		}
		cp.Infof("Configured %s\n", iniPath)
	}

	return nil
}

// getPhpIniPaths 获取所有需要配置的 INI 文件路径
// 包括扫描目录和 Debian SAPI 分离目录.
func getPhpIniPaths(info *phpExtensionInfo) []string {
	var paths []string

	// 优先使用扫描目录
	if info.iniScanDir != "" {
		iniPath := filepath.Join(info.iniScanDir, phpDdtraceIniName)
		paths = append(paths, iniPath)

		// 检测 Debian 风格的 SAPI 分离目录
		// 例如: /etc/php/8.1/cli/conf.d -> /etc/php/8.1/apache2/conf.d
		if strings.Contains(info.iniScanDir, "/cli/conf.d") {
			// Apache2
			apacheDir := strings.Replace(info.iniScanDir, "/cli/conf.d", "/apache2/conf.d", 1)
			if _, err := os.Stat(apacheDir); err == nil {
				paths = append(paths, filepath.Join(apacheDir, phpDdtraceIniName))
			}

			// FPM
			fpmDir := strings.Replace(info.iniScanDir, "/cli/conf.d", "/fpm/conf.d", 1)
			if _, err := os.Stat(fpmDir); err == nil {
				paths = append(paths, filepath.Join(fpmDir, phpDdtraceIniName))
			}
		}
	}

	// 如果没有扫描目录，使用主 php.ini
	if len(paths) == 0 && info.iniMain != "" {
		paths = append(paths, info.iniMain)
	}

	return paths
}

// updatePhpIniFile 更新 PHP INI 文件.
func updatePhpIniFile(iniPath string, info *phpExtensionInfo) error {
	// 确保目录存在
	iniDir := filepath.Dir(iniPath)
	if err := os.MkdirAll(iniDir, 0o755); err != nil { // nolint:gosec
		return err
	}

	// 读取现有内容
	existingContent := ""
	if data, err := os.ReadFile(iniPath); err == nil { // nolint:gosec
		existingContent = string(data)
	}

	// 检查是否是 Datadog 专用 INI 文件
	isDatadogIni := filepath.Base(iniPath) == phpDdtraceIniName

	// 生成新的 INI 内容
	newContent := generatePhpIniContent(info, existingContent, isDatadogIni)

	// 备份原文件（如果存在）
	if _, err := os.Stat(iniPath); err == nil {
		backupPath := iniPath + ".bak." + strconv.FormatInt(time.Now().UnixNano(), 36)
		if data, err := os.ReadFile(iniPath); err == nil { // nolint:gosec
			_ = os.WriteFile(backupPath, data, 0o644) // nolint:gosec
		}
	}

	return os.WriteFile(iniPath, []byte(newContent), 0o644) // nolint:gosec
}

// generatePhpIniContent 生成 PHP INI 内容.
func generatePhpIniContent(info *phpExtensionInfo, existingContent string, isDatadogIni bool) string {
	var sb strings.Builder

	// 写入头部注释
	sb.WriteString("; DataDog PHP Tracing Extension\n")
	sb.WriteString("; Auto-generated by DataKit installer\n")
	sb.WriteString("; Priority: 98 - loaded after most extensions\n\n")

	// 扩展配置
	// 如果有扩展目录，使用简单名称；否则使用完整路径
	if info.extensionDir != "" {
		sb.WriteString("extension=ddtrace.so\n")
	} else {
		sb.WriteString("; extension=ddtrace.so\n")
		sb.WriteString("; Note: extension_dir not configured, please set it manually\n")
	}

	sb.WriteString("\n[datadog]\n")
	sb.WriteString("; Enable DataDog tracing\n")
	sb.WriteString("datadog.trace.enabled=1\n")
	sb.WriteString("\n")

	// 常用配置（注释状态）
	sb.WriteString("; Service name (customize as needed)\n")
	sb.WriteString("; datadog.service.name=my-service\n")
	sb.WriteString("\n")
	sb.WriteString("; Agent connection\n")
	sb.WriteString("; datadog.agent_host=localhost\n")
	sb.WriteString("; datadog.trace.agent_port=8126\n")
	sb.WriteString("\n")
	sb.WriteString("; Environment\n")
	sb.WriteString("; datadog.env=production\n")
	sb.WriteString("\n")
	sb.WriteString("; Sampling rate (1.0 = 100%%)\n")
	sb.WriteString("; datadog.trace.sample_rate=1.0\n")
	sb.WriteString("\n")

	// 如果是更新现有文件，保留用户的自定义配置
	if existingContent != "" && !isDatadogIni {
		// 检查是否已有 ddtrace 配置
		if !strings.Contains(existingContent, "ddtrace") &&
			!strings.Contains(existingContent, "datadog") {
			// 追加到现有内容
			return existingContent + "\n" + sb.String()
		}
	}

	return sb.String()
}

// findPhpInterpreter 查找 PHP 解释器路径（简单版本，用于测试）.
func findPhpInterpreter() string {
	if php, err := exec.LookPath("php"); err == nil {
		return php
	}
	return ""
}

// getPhpVersion 获取 PHP 版本信息.
func getPhpVersion(phpPath string) (string, error) {
	//nolint:gosec
	ver, err := exec.Command(phpPath, "-v").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s -v error: %w", phpPath, err)
	}

	// PHP 版本格式: PHP 8.1.2 (cli) ...
	// 提取主版本号和次版本号，如 8.1
	re := regexp.MustCompile(`PHP (\d+\.\d+)`)
	v := re.FindStringSubmatch(string(ver))
	if len(v) != 2 {
		return "", fmt.Errorf("parse php version error: %s", string(ver))
	}

	return v[1], nil
}

// getPhpArch 获取 PHP 架构信息
// 使用 runtime.GOARCH 获取当前系统架构，比执行 uname -m 更可靠.
func getPhpArch() (string, error) {
	// runtime.GOARCH 返回: amd64, arm64 等
	// 这正是我们需要的格式，无需转换
	return runtime.GOARCH, nil
}

// getPhpThreadSafety 检测 PHP 是否是线程安全版本.
func getPhpThreadSafety(phpPath string) (bool, error) {
	//nolint:gosec
	out, err := exec.Command(phpPath, "-i").CombinedOutput()
	if err != nil {
		return false, err
	}

	switch {
	case strings.Contains(string(out), "Thread Safety => disabled"):
		return false, nil
	case strings.Contains(string(out), "Thread Safety => enabled"):
		return true, nil
	}

	return false, nil
}

// getPhpIniPath 获取 PHP 配置文件路径.
func getPhpIniPath(phpPath string) (string, error) {
	//nolint:gosec
	out, err := exec.Command(phpPath, "-i").CombinedOutput()
	if err != nil {
		return "", err
	}

	// 查找 "Loaded Configuration File => /path/to/php.ini"
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Loaded Configuration File") {
			parts := strings.Split(line, "=>")
			if len(parts) == 2 {
				iniPath := strings.TrimSpace(parts[1])
				if iniPath != "" && iniPath != "(none)" {
					return iniPath, nil
				}
			}
		}
	}

	// 尝试查找 "Configuration File (php.ini) Path"
	for _, line := range lines {
		if strings.Contains(line, "Configuration File (php.ini) Path") {
			parts := strings.Split(line, "=>")
			if len(parts) == 2 {
				iniDir := strings.TrimSpace(parts[1])
				if iniDir != "" {
					// 尝试 php.ini
					iniPath := filepath.Join(iniDir, "php.ini")
					if _, err := os.Stat(iniPath); err == nil {
						return iniPath, nil
					}
					// 尝试 php.ini-development 或 php.ini-production
					for _, name := range []string{"php.ini-development", "php.ini-production"} {
						candidate := filepath.Join(iniDir, name)
						if _, err := os.Stat(candidate); err == nil {
							return candidate, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("php.ini not found")
}

// getPhpIniScanDir 获取 PHP 扫描目录（用于创建单独的 ddtrace.ini）.
func getPhpIniScanDir(phpPath string) (string, error) {
	//nolint:gosec
	out, err := exec.Command(phpPath, "-i").CombinedOutput()
	if err != nil {
		return "", err
	}

	// 查找 "Scan this dir for additional .ini files"
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Scan this dir for additional .ini files") {
			parts := strings.Split(line, "=>")
			if len(parts) == 2 {
				scanDir := strings.TrimSpace(parts[1])
				if scanDir != "" && scanDir != "(none)" {
					return scanDir, nil
				}
			}
		}
	}

	return "", fmt.Errorf("php scan dir not found")
}
