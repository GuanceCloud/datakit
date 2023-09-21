// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

type datakitInfo struct {
	tmpDir string
}

func (info *datakitInfo) init() error {
	tmpDir, err := ioutil.TempDir("", "datakit-info")
	if err != nil {
		return fmt.Errorf("create temporary dir error: %w", err)
	}

	info.tmpDir = tmpDir
	return nil
}

func (info *datakitInfo) clean() error {
	if len(info.tmpDir) > 0 {
		return os.RemoveAll(info.tmpDir)
	}

	return nil
}

func (info *datakitInfo) collect() error {
	num := 1
	cp.Infof("%d. collect log files\n", num)
	if err := info.collectLog(); err != nil {
		cp.Warnf("collect log files error: %s\n", err.Error())
	}
	num++

	cp.Infof("%d. collect config files\n", num)
	if err := info.collectConfig(); err != nil {
		cp.Warnf("collect config files error: %s\n", err.Error())
	}
	num++

	cp.Infof("%d. collect env\n", num)
	if err := info.collectEnv(); err != nil {
		cp.Warnf("collect env error: %s\n", err.Error())
	}
	num++

	if !*flagDebugDisableProfile {
		cp.Infof("%d. collect profile\n", num)
		if err := info.collectProfile(); err != nil {
			cp.Warnf("collect profile data error: %s, ignored\n", err.Error())
		}
		num++
	}

	cp.Infof("%d. collect metrics\n", num)
	if err := info.collectMetrics(); err != nil {
		cp.Warnf("collect metrics error: %s\n", err.Error())
	}
	num++

	if runtime.GOOS == "linux" {
		cp.Infof("%d. collect systemd log\n", num)
		if err := info.collectSystemdLog(); err != nil {
			cp.Warnf("collect systemd log error: %s\n", err.Error())
		}
	}

	return nil
}

func (info *datakitInfo) collectSystemdLog() error {
	sysLogDir, err := info.makeDir("syslog")
	if err != nil {
		return err
	}
	cmd := exec.Command("journalctl", "-u", "datakit.service", "-n", "10000", "--no-pager") // last 10000 lines
	res, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(sysLogDir, fmt.Sprintf("syslog-%d", time.Now().UnixMilli())), res, os.ModePerm)
}

func (info *datakitInfo) collectMetrics() error {
	metricsDir, err := info.makeDir("metrics")
	if err != nil {
		return err
	}
	dkHost := config.Cfg.HTTPAPI.Listen
	for i := 0; i < 3; i++ {
		if i != 0 {
			time.Sleep(5 * time.Second)
		}
		resp, err := http.Get(fmt.Sprintf("http://%s/metrics", dkHost))
		if err != nil {
			return err
		}
		defer resp.Body.Close() //nolint:errcheck
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			cp.Warnf("read metrics body error: %s\n", err.Error())
			continue
		}

		err = ioutil.WriteFile(filepath.Join(metricsDir, fmt.Sprintf("metric-%d", time.Now().UnixMilli())), bodyBytes, os.ModePerm)

		if err != nil {
			cp.Warnf("write metric file error: %s\n", err.Error())
			continue
		}
	}

	return nil
}

func (info *datakitInfo) collectProfile() error {
	if !config.Cfg.EnablePProf {
		return nil
	}

	profileDir, err := info.makeDir("profile")
	if err != nil {
		return err
	}

	profileTypes := []string{
		"profile",
		"heap",
		"allocs",
		"goroutine",
		"block",
		"mutex",
	}

	for _, name := range profileTypes {
		params := ""
		if name == "profile" {
			params = "duration=10"
		}
		resp, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/%s?%s", config.Cfg.PProfListen, name, params))
		if err != nil {
			cp.Warnf("request profile for %s error: %s\n", name, err.Error())
			continue
		}

		defer resp.Body.Close() //nolint:errcheck

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			cp.Warnf("read profile data for %s error: %s\n", name, err.Error())
			continue
		}

		err = ioutil.WriteFile(filepath.Join(profileDir, name), bodyBytes, os.ModePerm)

		if err != nil {
			cp.Warnf("write profile file %s error: %s\n", name, err.Error())
		}
	}

	return nil
}

func (info *datakitInfo) collectEnv() error {
	envs := []string{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ENV_") {
			parts := strings.Split(env, "=")
			if len(parts) >= 2 {
				key := parts[0]
				value := strings.Join(parts[1:], "=")
				if info.containString(key, []string{
					"password",
					"token",
					"key",
					"key_pw",
					"secret",
				}) {
					value = "******"
				}
				if key == "ENV_DATAWAY" {
					value = info.escapeString(value, []string{"dataway"})
				}
				envs = append(envs, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}

	envsString := strings.Join(envs, "\n")
	return ioutil.WriteFile(filepath.Join(info.tmpDir, "env.txt"), []byte(envsString), os.ModePerm)
}

func (info *datakitInfo) collectConfig() error {
	configDir, err := info.makeDir("config")
	if err != nil {
		return err
	}

	err = info.copyDir(datakit.ConfdDir, configDir, func(s string) string {
		return info.escapeString(s, []string{"dataway", "password", "uri"})
	})

	if err != nil {
		return err
	}

	return nil
}

func (info *datakitInfo) escapeString(str string, kinds []string) string {
	for _, kind := range kinds {
		switch kind {
		case "dataway":
			str = regexp.MustCompile(`token=tkn_[A-Za-z0-9_]+`).ReplaceAllString(str, `token=******`)
		case "password":
			str = regexp.MustCompile(`(pass|password|bearer_token_string|sk|token)\s*=\s*(".*")`).ReplaceAllString(str, `${1} = "******"`)
			str = regexp.MustCompile(`('--password'\s*,\s*)'.*'\s*,`).ReplaceAllString(str, `${1}'******',`)
		case "uri":
			str = regexp.MustCompile(`(["']?[A-Za-z0-9]+)\:\/\/([A-Za-z0-9_]+)\:(.+)\@`).ReplaceAllString(str, `${1}://${2}:******@`)
		default:
		}
	}

	return str
}

func (info *datakitInfo) containString(str string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(strings.ToLower(str), strings.ToLower(substr)) {
			return true
		}
	}

	return false
}

func (info *datakitInfo) makeDir(name string) (string, error) {
	logDir := filepath.Join(info.tmpDir, name)
	err := os.Mkdir(logDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("create log dir error: %w", err)
	}
	return logDir, nil
}

func (info *datakitInfo) collectLog() error {
	log := config.Cfg.Logging
	logDir, err := info.makeDir("log")
	if err != nil {
		return err
	}
	// copy main log
	if len(log.Log) > 0 && log.Log != "stdout" {
		if err := info.copyFile(log.Log, filepath.Join(logDir, "log"), nil); err != nil {
			cp.Warnf("Collect log error: %s\n", err.Error())
		}
	}

	// copy gin log
	if len(log.GinLog) > 0 && log.GinLog != "stdout" {
		if err := info.copyFile(log.GinLog, filepath.Join(logDir, "gin.log"), nil); err != nil {
			cp.Warnf("Collect gin.log error: %s\n", err.Error())
		}
	}

	return nil
}

type transformFunc func(string) string

func (info *datakitInfo) copyFile(src, dst string, transform transformFunc) error {
	if transform == nil {
		srcFile, err := os.Open(filepath.Clean(src))
		if err != nil {
			return err
		}
		defer func() {
			if err := srcFile.Close(); err != nil {
				cp.Warnf("close src file error: %s\n", err.Error())
			}
		}()

		dstFile, err := os.Create(filepath.Clean(dst))
		if err != nil {
			return err
		}
		defer func() {
			if err := dstFile.Close(); err != nil {
				cp.Warnf("close dst file error: %s\n", err.Error())
			}
		}()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}
	} else {
		file, err := os.Open(filepath.Clean(src))
		if err != nil {
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				cp.Warnf("close file error: %s", err.Error())
			}
		}()

		newFile, err := os.Create(filepath.Clean(dst))
		if err != nil {
			return err
		}
		defer func() {
			if err := newFile.Close(); err != nil {
				cp.Warnf("close file error: %s", err.Error())
			}
		}()

		scanner := bufio.NewScanner(file)
		writer := bufio.NewWriter(newFile)

		for scanner.Scan() {
			line := scanner.Text()
			newLine := transform(line)

			fmt.Fprintln(writer, newLine)
		}

		err = writer.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

func (info *datakitInfo) compressDir() (string, error) {
	srcDir := info.tmpDir
	date := time.Now().UnixMilli()
	fileName := fmt.Sprintf("info-%d", date)
	zipPath := fmt.Sprintf("%s.zip", fileName)
	// Open a file to write the compressed data to
	zipFile, err := os.Create(filepath.Clean(zipPath))
	if err != nil {
		return "", fmt.Errorf("error creating file %s: %w", zipPath, err)
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			cp.Warnf("close zip file error: %s", err.Error())
		}
	}()

	// Create a new zip archive
	zipWriter := zip.NewWriter(zipFile)

	defer func() {
		if err := zipWriter.Close(); err != nil {
			cp.Warnf("close zip writer error: %s", err.Error())
		}
	}()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path: %w", err)
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %w", err)
		}

		filePath := filepath.Clean(filepath.Join(fileName, relPath))
		// Add the file to the zip archive
		zipEntry, err := zipWriter.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating entry for file %s in zip archive: %w", relPath, err)
		}

		// Open the file and copy its contents to the zip entry
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return fmt.Errorf("error opening file %s: %w", path, err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				cp.Warnf("close file error: %s", err.Error())
			}
		}()

		_, err = io.Copy(zipEntry, file)
		if err != nil {
			return fmt.Errorf("error copying file %s to zip entry: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return zipPath, err
}

func (info *datakitInfo) copyDir(srcDir string, dstDir string, transform transformFunc) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		fileName := file.Name()

		srcFilePath := filepath.Join(srcDir, fileName)

		if file.IsDir() {
			dstFilePath := filepath.Join(dstDir, fileName)
			err = os.MkdirAll(dstFilePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", dstFilePath, err)
			}

			err = info.copyDir(srcFilePath, dstFilePath, transform)
			if err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(fileName, ".conf") {
				continue
			}
			dstName := fmt.Sprintf("%s.copy", fileName)
			dstFilePath := filepath.Join(dstDir, dstName)
			err := info.copyFile(srcFilePath, dstFilePath, transform)
			if err != nil {
				fmt.Printf("error writing file %s: %s\n", dstFilePath, err)
				continue
			}
		}
	}

	return nil
}

func bugReport() error {
	infoInstance := &datakitInfo{}

	if err := infoInstance.init(); err != nil {
		return err
	}

	defer func() {
		err := infoInstance.clean()
		if err != nil {
			cp.Warnf("delete tmp file error: %s\n", err.Error())
		}
	}()

	if err := infoInstance.collect(); err != nil {
		return err
	}

	if zipPath, err := infoInstance.compressDir(); err != nil {
		cp.Errorf("compress zip file failed: %s\n", err.Error())
	} else {
		cp.Infof("DataKit info saved into %s\n", zipPath)
	}

	return nil
}
