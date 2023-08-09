// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

const (
	DefaultInstallDir       = "/usr/local/datakit/data/rum/tools"
	DefaultArch             = "default"
	Java                    = "java"
	Proguard                = "proguard"
	AndroidCommandLineTools = "cmdline-tools"
	AndroidNDK              = "android-ndk"
	Atosl                   = "atosl"
)

var downloadClient = &http.Client{
	Timeout: time.Minute * 30,
}

func setupLinks(baseURL string) map[string]map[string]string {
	baseURL = CanonicalInstallBaseURL(baseURL)

	return map[string]map[string]string{
		Java: {
			"darwin/amd64": baseURL + "sourcemap/jdk/OpenJDK11U-jdk_x64_mac_hotspot_11.0.16_8.tar.gz",
			"darwin/arm64": baseURL + "sourcemap/jdk/OpenJDK11U-jdk_aarch64_mac_hotspot_11.0.15_10.tar.gz",
			"linux/amd64":  baseURL + "sourcemap/jdk/OpenJDK11U-jdk_x64_linux_hotspot_11.0.16_8.tar.gz",
			"linux/arm64":  baseURL + "sourcemap/jdk/OpenJDK11U-jdk_aarch64_linux_hotspot_11.0.16_8.tar.gz",
		},

		AndroidCommandLineTools: {
			"darwin/amd64": baseURL + "sourcemap/R8/commandlinetools-mac-8512546_simplified.tar.gz",
			"darwin/arm64": baseURL + "sourcemap/R8/commandlinetools-mac-8512546_simplified.tar.gz",
			"linux/amd64":  baseURL + "sourcemap/R8/commandlinetools-linux-8512546_simplified.tar.gz",
			"linux/arm64":  baseURL + "sourcemap/R8/commandlinetools-linux-8512546_simplified.tar.gz",
		},

		Proguard: {
			"default": baseURL + "sourcemap/proguard/proguard-7.2.2.tar.gz",
		},

		AndroidNDK: {
			"darwin/amd64": baseURL + "sourcemap/ndk/android-ndk-r22b-x64-mac-simplified.tar.gz",
			"darwin/arm64": baseURL + "sourcemap/ndk/android-ndk-r22b-x64-mac-simplified.tar.gz",
			"linux/amd64":  baseURL + "sourcemap/ndk/android-ndk-r25-x64-linux-simplified.tar.gz",
			"linux/arm64":  baseURL + "sourcemap/ndk/android-ndk-r25-x64-linux-simplified.tar.gz",
		},
		Atosl: {
			"darwin/amd64": baseURL + "sourcemap/atosl/atosl-darwin-x64",
			"darwin/arm64": baseURL + "sourcemap/atosl/atosl-darwin-arm64",
			"linux/amd64":  baseURL + "sourcemap/atosl/atosl-linux-x64",
			"linux/arm64":  baseURL + "sourcemap/atosl/atosl-linux-arm64",
		},
	}
}

var downloadLink = setupLinks(OnlineBaseURL)

func IsDir(name string) (bool, error) {
	fileInfo, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		// probably it is a permission error
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func InstallSymbolTools() error {
	if installBaseURL := datakit.GetEnv("DK_INSTALLER_BASE_URL"); installBaseURL != "" {
		cp.Warnf("setup base URL to %s \n", installBaseURL)
		downloadLink = setupLinks(installBaseURL)
	}

	ok, err := IsDir(DefaultInstallDir)
	if err != nil {
		return fmt.Errorf("check the dir %s fail: %w", DefaultInstallDir, err)
	}
	if !ok {
		if err := os.MkdirAll(DefaultInstallDir, 0o750); err != nil {
			return fmt.Errorf("mkdir [%s] fail: %w", DefaultInstallDir, err)
		}
	}

	if err := installAndroidCmdLineTool(); err != nil {
		// nolint:lll
		cp.Errorf(`install android commandline tool fail: %s, you may need to install "https://developer.android.com/studio#cmdline-tools" manually and modify the rum config item "android_cmdline_home" correspondingly%s`, err, "\n")
	} else {
		cp.Infof("install android commandline tool success\n\n")
	}

	if err := installProguard(); err != nil {
		// nolint:lll
		cp.Errorf(`install proguard fail: %s,you may need to install "https://github.com/Guardsquare/proguard" manually and modify the rum config item "proguard_home" correspondingly%s`, err, "\n")
	} else {
		cp.Infof("install proguard success\n\n")
	}

	if err := installAndroidNDK(); err != nil {
		// nolint:lll
		cp.Errorf(`install android-ndk fail: %s,you may need to install "https://developer.android.com/ndk/downloads" manually and modify the rum config item "proguard_home" correspondingly %s`, err, "\n")
	} else {
		cp.Infof("install android-ndk success\n\n")
	}

	if err := installAtosl(); err != nil {
		// nolint:lll
		cp.Errorf(`install tool atosl fail: %s,you may need to install "https://github.com/Br4ndonZhang/atosl" manually and modify the rum config item "atos_bin_path" correspondingly %s`, err, "\n")
	} else {
		cp.Infof("install atosl success\n\n")
	}

	cp.Infof("installation complete, you may need to open a new shell and restart your datakit\n")

	return nil
}

func checkToolInstalled(tool string) (string, bool) {
	cp.Infof("checking %s installation status \n", tool)
	binPath, err := exec.LookPath(tool)
	return binPath, err == nil
}

func readTgzRootDir(tgz string) (string, error) {
	r, err := os.Open(tgz) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("open tar.gz file [%s] fail: %w", tgz, err)
	}
	unGzip, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("read gzip file [%s] fail: %w", tgz, err)
	}
	unTar := tar.NewReader(unGzip)

	for {
		entry, err := unTar.Next()
		if err != nil {
			return "", fmt.Errorf("read tar file [%s] fail: %w", tgz, err)
		}

		// clear the leading "./"
		entryName := strings.TrimLeft(entry.Name, "./")
		if strings.ContainsRune(entryName, filepath.Separator) {
			return rootDir(entryName), nil
		}
	}
}

func rootDir(path string) string {
	path = filepath.Clean(path)
	idx := strings.IndexByte(path, filepath.Separator)
	if idx > -1 {
		return path[:idx]
	}
	return path
}

func downloadFileToTmpDir(link string, filename ...string) (string, error) {
	cp.Infof("downloading software %s\n", link)

	resp, err := downloadClient.Get(link)
	if err != nil {
		return "", fmt.Errorf("download file fail: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	tmpDir := os.TempDir()
	baseName := filepath.Base(link)
	if len(filename) > 0 {
		baseName = filename[0]
	}
	fullPath := filepath.Join(tmpDir, baseName)

	progressBar := &downloader.WriteCounter{
		Total: uint64(resp.ContentLength),
	}
	downloader.CurDownloading = baseName

	fp, err := os.Create(fullPath) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("save to local file [%s] fail:%w", fullPath, err)
	}
	defer func(fp *os.File) {
		_ = fp.Close()
	}(fp)

	if _, err := io.Copy(io.MultiWriter(fp, progressBar), resp.Body); err != nil {
		return "", fmt.Errorf("save to local file [%s] fail:%w", fullPath, err)
	}
	return fullPath, nil
}

func installJDK() error {
	link, err := toolDownloadURL(Java)
	if err != nil {
		return fmt.Errorf("get jdk download link fail: %w, please install jdk manually", err)
	}

	tgzFile, err := downloadFileToTmpDir(link)
	if err != nil {
		return fmt.Errorf("download jdk from [%s] fail:%w", link, err)
	}

	if _, err := execCmd("tar", "-zxf", tgzFile, "-C", "/usr/local"); err != nil {
		return fmt.Errorf("untar file [%s] fail: %w", tgzFile, err)
	}

	unTarDir, err := readTgzRootDir(tgzFile)
	if err != nil {
		return fmt.Errorf("read tar.gz file [%s] root dir fail:%w", tgzFile, err)
	}

	jdkHome := filepath.Join("/usr/local", unTarDir)

	ok, err := IsDir(jdkHome)
	if err != nil {
		return fmt.Errorf("check the dir [%s] fail: %w", jdkHome, err)
	}
	if !ok {
		return fmt.Errorf("untar jdk to /usr/local fail")
	}

	jdkBinDir, err := scanJDKBinPath(jdkHome)
	if err != nil {
		return fmt.Errorf("jdk bin path not found in [%s]: %w", jdkHome, err)
	}

	if _, err := execCmd("sudo", "ln", "-s", "-f", filepath.Join(jdkBinDir, "java"), "/usr/local/bin/java"); err != nil {
		return fmt.Errorf("create symbol link fail: %w", err)
	}
	cp.Infof("install jdk success to %s :)\n", jdkHome)
	return nil
}

func scanJDKBinPath(homeDir string) (string, error) {
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return "", fmt.Errorf("scan dir [%s] fail:%w", homeDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "bin" && path.IsFileExists(filepath.Join(homeDir, "bin/java")) {
				return filepath.Abs(filepath.Join(homeDir, "bin"))
			}
			if dir, err := scanJDKBinPath(filepath.Join(homeDir, entry.Name())); err == nil {
				return filepath.Abs(dir)
			}
		}
	}
	return "", fmt.Errorf("jdk bin dir not found in [%s]", homeDir)
}

//nolint:unused
func scanProguardBinPath(homeDir string) (string, error) {
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return "", fmt.Errorf("open dir [%s] fail: %w", homeDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "bin" && path.IsFileExists(filepath.Join(homeDir, "/bin/retrace.sh")) {
				return filepath.Abs(filepath.Join(homeDir, "bin"))
			}
			if dir, err := scanProguardBinPath(filepath.Join(homeDir, entry.Name())); err == nil {
				return filepath.Abs(dir)
			}
		}
	}
	return "", fmt.Errorf("bin dir not found in [%s]", homeDir)
}

func execCmd(name string, args ...string) ([]byte, error) {
	shellCmd := name + " " + strings.Join(args, " ")
	cp.Infof("%s%s", shellCmd, "\n")
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok { // nolint:errorlint
			return nil, fmt.Errorf("exec cmd [%s] fail: %s", shellCmd, ee.Stderr)
		}
		return nil, fmt.Errorf("exec cmd [%s] fail: %w", shellCmd, err)
	}
	return out, nil
}

func toolDownloadURL(software string) (string, error) {
	arch := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	jdkURL, ok := downloadLink[software][arch]
	if !ok {
		jdkURL, ok = downloadLink[software][DefaultArch]
		if !ok {
			return "", fmt.Errorf("the tool [%s] doesnot have downloading link for architeture [%s] yet", software, arch)
		}
	}
	return jdkURL, nil
}

func installTool(software string) (string, error) {
	link, err := toolDownloadURL(software)
	if err != nil {
		return "", fmt.Errorf("can not find the download link of tool %s: %w, you may should install it manually", software, err)
	}

	tgzFile, err := downloadFileToTmpDir(link)
	if err != nil {
		return "", fmt.Errorf("download tool [%s] from link [%s] fail: %w", software, link, err)
	}

	if _, err := execCmd("sudo", "tar", "-zxf", tgzFile, "-C", DefaultInstallDir); err != nil {
		return "", fmt.Errorf("untar tgz file [%s] fail: %w", tgzFile, err)
	}

	unTarRootDir, err := readTgzRootDir(tgzFile)
	if err != nil {
		return "", fmt.Errorf("read tar.gz file [%s] root dir fail:%w", tgzFile, err)
	}
	toolHomeDir := filepath.Join(DefaultInstallDir, unTarRootDir)
	ok, err := IsDir(toolHomeDir)
	if err != nil {
		return "", fmt.Errorf("check the dir [%s] fail: %w", toolHomeDir, err)
	}
	if !ok {
		return "", fmt.Errorf("untar tar.gz file [%s] fail", tgzFile)
	}
	installPrefix := filepath.Join(DefaultInstallDir, software)

	if installPrefix != toolHomeDir {
		if err := os.RemoveAll(installPrefix); err != nil {
			return "", fmt.Errorf("remove dir [%s] fail: %w", installPrefix, err)
		}

		if _, err := execCmd("sudo", "mv", toolHomeDir, installPrefix); err != nil {
			return "", fmt.Errorf("rename [%s] to [%s] fail: %w", toolHomeDir, installPrefix, err)
		}
	}

	return installPrefix, nil
}

// installProguard for detail of proguard see https://github.com/Guardsquare/proguard
func installProguard() error {
	if _, ok := checkToolInstalled(Java); !ok {
		if err := installJDK(); err != nil {
			return fmt.Errorf("install jdk fail: %w", err)
		}
	}

	_, err := installTool(Proguard)
	return err
}

func installAndroidNDK() error {
	_, err := installTool(AndroidNDK)
	return err
}

func ScanToolAtosPath() (string, bool) {
	if runtime.GOOS == "darwin" {
		// macOS use builtin tool "atos"
		if atosPath, err := exec.LookPath("atos"); err == nil && atosPath != "" {
			return atosPath, true
		}

		stat, err := os.Stat("/usr/bin/atos")
		if err == nil && (stat.Mode().Perm()&0o111) > 0 {
			return "/usr/bin/atos", true
		}
	}

	return "", false
}

func installAtosl() error {
	if _, ok := ScanToolAtosPath(); ok {
		return nil
	}

	link, err := toolDownloadURL(Atosl)
	if err != nil {
		return fmt.Errorf("atosl download link not found: %w", err)
	}

	binaryFile, err := downloadFileToTmpDir(link)
	if err != nil {
		return fmt.Errorf("unable to download atosl from [%s], err: %w", link, err)
	}

	targetBin := filepath.Join(DefaultInstallDir, "atosl")
	if _, err := execCmd("sudo", "cp", binaryFile, targetBin); err != nil {
		return fmt.Errorf("cp bin %s to dir [%s] fail:%w", binaryFile, DefaultInstallDir, err)
	}

	if err := os.Chmod(targetBin, 0o775); err != nil { // nolint:gosec
		return fmt.Errorf("unable to chmod +x to %s", targetBin)
	}

	return nil
}

func installAndroidCmdLineTool() error {
	if _, ok := checkToolInstalled(Java); !ok {
		if err := installJDK(); err != nil {
			return fmt.Errorf("install jdk fail: %w", err)
		}
	}
	_, err := installTool(AndroidCommandLineTools)
	return err
}
