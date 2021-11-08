package cmds

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

type successRes struct {
	Content string `json:"content"`
}

func uploadLog(urls []string) error {
	dw := dataway.DataWayCfg{URLs: urls, HTTPProxy: config.Cfg.DataWay.HTTPProxy}
	if err := dw.Apply(); err != nil {
		return err
	}

	logFileName, err := getLogFile()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(logFileName, os.TempDir()) {
		return fmt.Errorf("invalid tmp file: %s", logFileName)
	}

	fileReader, err := os.Open(filepath.Clean(logFileName))
	if err != nil {
		return err
	}

	defer os.Remove(logFileName) //nolint:errcheck

	hostName := getHostName()

	resp, err := dw.UploadLog(fileReader, hostName)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		resJSON := successRes{}
		if err := json.Unmarshal(respBody, &resJSON); err == nil {
			fmt.Printf("Upload to oss: %s\n", resJSON.Content)
		} else {
			return fmt.Errorf("upload failed: %w", err)
		}
		return nil
	} else {
		return fmt.Errorf("upload failed: %s", respBody)
	}
}

func getLogFile() (string, error) {
	var fileName string
	logPath := config.Cfg.Logging.Log
	logDir, logName := filepath.Split(logPath)
	if len(logDir) == 0 {
		return fileName, fmt.Errorf("log path is empty")
	}
	file, err := os.Open(filepath.Clean(logDir))
	if err != nil {
		return fileName, err
	}

	defer file.Close() //nolint:gosec,errcheck

	tmpFile, err := ioutil.TempFile(os.TempDir(), "datakit-log")
	if err != nil {
		return fileName, err
	}

	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close() // nolint:errcheck

	info, err := file.Stat()
	if err != nil {
		return fileName, err
	}
	if info.IsDir() {
		logNamePrefix := logName
		parts := strings.Split(logName, ".")
		partsLen := len(parts)
		if partsLen > 1 {
			logNamePrefix = strings.Join(parts[0:partsLen-1], ".")
		}
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return fileName, err
		}
		for _, fi := range fileInfos {
			name := fi.Name()
			if !strings.HasPrefix(name, logNamePrefix) {
				continue
			}
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return fileName, err
			}
			defer f.Close() //nolint:gosec,errcheck
			if fInfo, err := f.Stat(); err != nil {
				return fileName, err
			} else {
				// ignore dir
				if fInfo.IsDir() {
					continue
				}
				header, err := zip.FileInfoHeader(fInfo)
				if err != nil {
					return fileName, err
				}
				header.Name = fi.Name()
				header.Method = zip.Deflate
				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return fileName, err
				}
				_, err = io.Copy(writer, f)
				if err != nil {
					return fileName, err
				}
			}
		}
	} else {
		return fileName, fmt.Errorf("invalid log dir: %s", logPath)
	}
	return tmpFile.Name(), nil
}

func getHostName() string {
	var hostName string

	// 1. from datakit.conf
	if customHostName, ok := config.Cfg.Environments["ENV_HOSTNAME"]; ok {
		hostName = customHostName
	}

	// 2. default: os.Hostname()
	if len(hostName) == 0 {
		osHostName, err := os.Hostname()
		if err == nil && len(osHostName) > 0 {
			hostName = osHostName
		}
	}

	return hostName
}
