// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	jsoniter "github.com/json-iterator/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
)

const (
	Proguard                = "proguard"
	AndroidCommandLineTools = "cmdline-tools"
	AndroidNDK              = "android-ndk"
	Atosl                   = "atosl"
)

var (
	idxRangeRegexp  = regexp.MustCompile(`\s#00\s+`)
	backtraceRegexp = regexp.MustCompile(`backtrace:[\s\S]+?logcat:`)

	ErrJSONUnmarshal = errors.New("")
	archiveDictFile  = filepath.Join(datakit.DataRUMDir, ".--source-map-archive-dict.json")
)

type ArchiveMeta struct {
	ModTime     time.Time `json:"mod_time"`     // latest modification time
	ExtractTime time.Time `json:"extract_time"` // latest extract time
}

type ArchiveMetaDict map[string]*ArchiveMeta

type SourceMapArchive struct {
	ModTime  time.Time `json:"mod_time"`
	Filepath string    `json:"filepath"`
}

func isFile(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func getExtractDir(archivePath string) string {
	return strings.TrimSuffix(archivePath, httpapi.ZipExt)
}

func readDict(file string, loose bool) (ArchiveMetaDict, error) {
	body, err := os.ReadFile(file) // nolint:gosec
	if err != nil {
		_, statErr := os.Stat(file)
		if statErr != nil && errors.Is(statErr, fs.ErrNotExist) {
			return ArchiveMetaDict{}, nil
		}
		return nil, fmt.Errorf("unable to read source map dict: %w", err)
	}

	var dict ArchiveMetaDict
	if err := jsoniter.Unmarshal(body, &dict); err != nil {
		if loose {
			if e := os.Remove(file); e != nil {
				log.Warnf("unable to remove archive dict file: %s", e)
			}
			return ArchiveMetaDict{}, nil
		}
		return nil, fmt.Errorf("unable to unmarshal from dict content: %s%w", err, ErrJSONUnmarshal)
	}
	return dict, nil
}

func writeDict(file string, dict ArchiveMetaDict) error {
	jsonBytes, err := jsoniter.MarshalIndent(dict, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to Marshal archive meta dict: %w", err)
	}
	if err := os.WriteFile(file, jsonBytes, 0o644); err != nil { //nolint: gosec
		return fmt.Errorf("unable to write archive meta dict file: %w", err)
	}
	return nil
}

func extractArchives(loose bool) error {
	if !ExtractZipLock.TryLock() {
		log.Warnf("unable to get lock, skip this interval")
		return nil
	}
	defer ExtractZipLock.Unlock()

	sourceMapDirs := make(map[string]struct{}, 3)

	for _, sdkName := range []string{SdkAndroid, SdkIOS} {
		sourceMapDirs[getRumSourcemapDir(sdkName)] = struct{}{}
	}

	var totalArchives []*SourceMapArchive

	for dir := range sourceMapDirs {
		archives, err := scanArchives(dir)
		if err != nil {
			log.Warnf("scan .zip file from dir [%s] encounter err: %s", dir, err)
		}
		if len(archives) > 0 {
			totalArchives = append(totalArchives, archives...)
		}
	}

	oldDict, err := readDict(archiveDictFile, loose)
	if err != nil {
		return err
	}

	newDict := make(ArchiveMetaDict, len(oldDict))
	gaugeMap := make(map[string]int, 2)

	for _, archive := range totalArchives {
		meta, ok := oldDict[archive.Filepath]
		extractTo := getExtractDir(archive.Filepath)

		extractTime := time.Time{}
		if ok {
			extractTime = meta.ExtractTime
		}
		if (ok && meta.ExtractTime.Before(archive.ModTime)) || !isDir(extractTo) {
			log.Infof("extract zip archive: %s", archive.Filepath)
			if err := extractZipFile(archive.Filepath); err != nil {
				log.Warnf("unable to extract zip file[%s]: %s", archive.Filepath, err)
			} else {
				extractTime = time.Now()
			}
		}

		gaugeMap[filepath.Base(filepath.Dir(archive.Filepath))] += 1
		if ok {
			meta.ModTime = archive.ModTime
			meta.ExtractTime = extractTime
			newDict[archive.Filepath] = meta
			delete(oldDict, archive.Filepath)
		} else {
			newDict[archive.Filepath] = &ArchiveMeta{
				ModTime:     archive.ModTime,
				ExtractTime: extractTime,
			}
		}
	}

	for platform, cnt := range gaugeMap {
		loadedZipGauge.WithLabelValues(platform).Set(float64(cnt))
	}

	for archivePath := range oldDict {
		extractDir := getExtractDir(archivePath)
		if !isFile(archivePath) && isDir(extractDir) {
			if err := os.RemoveAll(extractDir); err != nil {
				log.Warnf("unable to clean dir for removed zip [%s]: %s", archivePath, err)
			}
		}
	}

	if err := writeDict(archiveDictFile, newDict); err != nil {
		return err
	}

	return nil
}

func scanArchives(dir string) ([]*SourceMapArchive, error) {
	f, err := os.Open(dir) // nolint: gosec
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	archives := make([]*SourceMapArchive, 0, 8)

	for {
		entries, err := f.Readdir(40)

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), httpapi.ZipExt) {
				continue
			}

			archives = append(archives, &SourceMapArchive{
				Filepath: filepath.Join(dir, entry.Name()),
				ModTime:  entry.ModTime(),
			})
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return archives, fmt.Errorf("unable to read dir entry: %w", err)
			}
		}
	}
	return archives, nil
}

func (ipt *Input) parseSourcemap(p *point.Point, sdkName string, status *sourceMapStatus) (*point.Point, error) {
	switch sdkName {
	case SdkWeb, SdkWebMiniApp, SdkWebUniApp:
		return ipt.resolveWebSourceMap(p, sdkName, status)
	case SdkAndroid:
		return ipt.resolveAndroidSourceMap(p, sdkName, status)
	case SdkIOS:
		return ipt.resolveIOSSourceMap(p, sdkName, status)
	}
	return p, nil
}

func runAtosCMD(atosCMDPath, symbolFile, loadAddress string, addresses []string) ([]string, error) {
	args := []string{
		"-o", symbolFile, "-l", loadAddress,
	}
	args = append(args, addresses...)
	cmd := exec.Command(atosCMDPath, args...) //nolint:gosec
	cmd.Env = []string{"HOME=/root"}          // run the tool "atosl" must set this env, why?
	log.Infof("%s %s", atosCMDPath, strings.Join(args, " "))
	stdout, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			log.Errorf("run atos tool fail: %s, err: %s", string(ee.Stderr), err)
		} else {
			log.Errorf("run atos tool fail: %s", err)
		}
		return nil, fmt.Errorf("unable to run atosl command: %w, stdout: %s", err, string(stdout))
	}

	// adapt different os newLine
	stdoutStr := strings.ReplaceAll(string(stdout), "\r\n", "\n")
	stdoutStr = strings.ReplaceAll(stdoutStr, "\r", "\n")
	stdoutStr = strings.Trim(stdoutStr, "\n")
	symbols := strings.Split(stdoutStr, "\n")
	return symbols, nil
}

// miniAppZipStat First find in "miniapp", and then find in "web" if not exists.
func miniAppZipStat(sdkName, zipFile string) (string, os.FileInfo, error) {
	zipFileAbsPath := filepath.Join(getRumSourcemapDir(sdkName), zipFile)
	statInfo, err := os.Stat(zipFileAbsPath)
	if sdkName != SdkWebMiniApp && sdkName != SdkWebUniApp {
		return zipFileAbsPath, statInfo, err
	}
	if err != nil && errors.Is(err, os.ErrNotExist) {
		zipFileAbsPath = filepath.Join(getRumSourcemapDir(SdkWeb), zipFile)
		statInfo, err = os.Stat(zipFileAbsPath)
		if err == nil {
			return zipFileAbsPath, statInfo, err
		}
	}
	return zipFileAbsPath, statInfo, err
}

func (ipt *Input) resolveWebSourceMap(p *point.Point, sdkName string, status *sourceMapStatus) (*point.Point, error) {
	fields := p.InfluxFields()
	errStack, ok := fields["error_stack"]
	if !ok {
		status.status = StatusLackField
		return p, nil
	}

	// if error_stack exists
	errStackStr := fmt.Sprintf("%v", errStack)

	tags := p.InfluxTags()
	appID := tags["app_id"]
	env := tags["env"]
	version := tags["version"]

	if appID != "" {
		zipFile := httpapi.GetSourcemapZipFileName(appID, env, version)
		webSourcemapLock.RLock()
		_, cacheExists := webSourcemapCache[zipFile]
		loadTime := webSourceCacheLoadTime[zipFile]
		webSourcemapLock.RUnlock()

		if !cacheExists || loadTime.Add(time.Minute*5).Before(time.Now()) {
			zipFileAbsPath, statInfo, err := miniAppZipStat(sdkName, zipFile)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) && cacheExists {
					deleteSourcemapCache(zipFileAbsPath)
					cacheExists = false
				}
			} else {
				if !cacheExists || loadTime.Before(statInfo.ModTime()) {
					if err := updateSourcemapCache(zipFileAbsPath); err == nil {
						cacheExists = true
					}
				}
			}
		}
		if cacheExists {
			start := time.Now()
			errorStackSource := getSourcemap(errStackStr, webSourcemapCache[zipFile])
			sourceMapDurationSummary.WithLabelValues(sdkName, appID, env, version).Observe(float64(time.Since(start)) / promDurationUnit)
			errorStackSourceBase64 := base64.StdEncoding.EncodeToString([]byte(errorStackSource))
			status.status = StatusOK
			p.MustAdd([]byte("error_stack_source_base64"), errorStackSourceBase64)
		} else {
			status.status = StatusZipNotFound
			log.Warnf("source map file [%s] not exists", filepath.Join(getRumSourcemapDir(sdkName), zipFile))
		}
	} else {
		status.status = StatusLackField
	}

	return p, nil
}

func (ipt *Input) resolveAndroidSourceMap(p *point.Point, sdkName string, status *sourceMapStatus) (*point.Point, error) {
	fields := p.InfluxFields()
	errStack, ok := fields["error_stack"]
	if !ok {
		status.status = StatusLackField
		return p, nil
	}

	// if error_stack exists
	errStackStr := fmt.Sprintf("%v", errStack)

	tags := p.InfluxTags()
	appID := tags["app_id"]
	env := tags["env"]
	version := tags["version"]

	if appID != "" {
		zipFile := httpapi.GetSourcemapZipFileName(appID, env, version)
		zipFileAbsDir := filepath.Join(getRumSourcemapDir(sdkName), strings.TrimSuffix(zipFile, httpapi.ZipExt))

		errorType := tags["error_type"]
		if errorType == JavaCrash {
			mappingFile := filepath.Join(zipFileAbsDir, "mapping.txt")
			if !path.IsFileExists(mappingFile) {
				status.status = StatusZipNotFound
				return p, fmt.Errorf("java source mapping file [%s] not exists", mappingFile)
			}
			toolName, err := checkJavaShrinkTool(mappingFile)
			if err != nil {
				return p, fmt.Errorf("verify java shrink tool fail: %w", err)
			}
			retraceCmd := ""
			if toolName == Proguard {
				if ipt.ProguardHome == "" {
					return p, fmt.Errorf("proguard home not set")
				}
				retraceCmd = filepath.Join(ipt.ProguardHome, "bin", "retrace.sh")
				if !path.IsFileExists(retraceCmd) {
					status.status = StatusToolNotFound
					return p, fmt.Errorf("the script retrace.sh not found in the proguardHome [%s]", retraceCmd)
				}
			} else {
				if ipt.AndroidCmdLineHome == "" {
					return p, fmt.Errorf("android commandline tool home not set")
				}
				retraceCmd = filepath.Join(ipt.AndroidCmdLineHome, "bin/retrace")
				if !path.IsFileExists(retraceCmd) {
					status.status = StatusToolNotFound
					return p, fmt.Errorf("the cmdline tools [retrace] not found in the androidCmdLineHome [%s]", retraceCmd)
				}
			}

			token := sourceMapTokenBuckets.getToken()
			defer sourceMapTokenBuckets.sendBackToken(token)
			cmd := exec.Command("sh", retraceCmd, mappingFile) //nolint:gosec
			cmd.Stdin = strings.NewReader(errStackStr)
			start := time.Now()
			originStack, err := cmd.Output()
			sourceMapDurationSummary.WithLabelValues(sdkName, appID, env, version).Observe(float64(time.Since(start)) / promDurationUnit)
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok { //nolint:errorlint
					return p, fmt.Errorf("run proguard retrace fail: %w, error_msg: %s", err, string(ee.Stderr))
				}
				return p, fmt.Errorf("run proguard retrace fail: %w", err)
			}
			originStack = bytes.TrimLeft(originStack, "Waiting for stack-trace input...")
			originStack = bytes.TrimLeftFunc(originStack, func(r rune) bool {
				return r == '\r' || r == '\n'
			})
			originStackB64 := base64.StdEncoding.EncodeToString(originStack)
			status.status = StatusOK
			p.MustAdd([]byte("error_stack_source_base64"), originStackB64)
			return p, nil
		} else if errorType == NativeCrash {
			if ipt.NDKHome == "" {
				return p, fmt.Errorf("android ndk home not set")
			}

			ndkStack := filepath.Join(ipt.NDKHome, "ndk-stack")
			stat, err := os.Stat(ndkStack)
			if err != nil {
				status.status = StatusToolNotFound
				return p, fmt.Errorf("ndk-stack command tool not found in the NDK HOME [%s]", ndkStack)
			}

			if !stat.Mode().IsRegular() {
				status.status = StatusToolNotFound
				return p, fmt.Errorf("ndk-stack path is not a valid exectable program [%s]", ndkStack)
			}

			abi := scanABI(errStackStr)
			if abi == "" {
				return p, fmt.Errorf("no valid NDK ABI found")
			}

			symbolObjDir := filepath.Join(zipFileAbsDir, abi)
			if !path.IsDir(symbolObjDir) {
				status.status = StatusZipNotFound
				return p, fmt.Errorf("expected native objects dir [%s] not found", symbolObjDir)
			}

			token := sourceMapTokenBuckets.getToken()
			defer sourceMapTokenBuckets.sendBackToken(token)
			cmd := exec.Command(ndkStack, "--sym", symbolObjDir) //nolint:gosec
			cmd.Stdin = strings.NewReader(errStackStr)
			start := time.Now()
			originStack, err := cmd.Output()
			sourceMapDurationSummary.WithLabelValues(sdkName, appID, env, version).Observe(float64(time.Since(start)) / promDurationUnit)
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok { //nolint:errorlint
					return p, fmt.Errorf("run ndk-stack tool fail: %w, error_msg: %s", err, string(ee.Stderr))
				}
				return p, fmt.Errorf("run ndk-stack tool fail: %w", err)
			}

			idxRange := idxRangeRegexp.FindIndex(originStack)
			if len(idxRange) == 2 {
				originStack = originStack[idxRange[0]:]
			}

			originStack = bytes.ReplaceAll(originStack, []byte("Crash dump is completed"), nil)
			idxRange = backtraceRegexp.FindStringIndex(errStackStr)
			if len(idxRange) == 2 {
				originStack = []byte(strings.ReplaceAll(errStackStr, errStackStr[idxRange[0]:idxRange[1]],
					fmt.Sprintf("backtrace:\n%slogcat:", string(originStack))))
			}

			originStackB64 := base64.StdEncoding.EncodeToString(originStack)
			log.Infof("native crash source map 处理成功， appid: %s, creat time: %s", appID, p.Time().In(time.Local).Format(time.RFC3339))
			p.MustAdd([]byte("error_stack_source_base64"), originStackB64)
			status.status = StatusOK
			return p, nil
		}
	} else {
		status.status = StatusLackField
	}

	return p, nil
}

func (ipt *Input) resolveIOSSourceMap(p *point.Point, sdkName string, status *sourceMapStatus) (*point.Point, error) {
	fields := p.InfluxFields()
	errStack, ok := fields["error_stack"]

	if !ok {
		status.status = StatusLackField
		return p, nil
	}

	// if error_stack exists
	errStackStr := fmt.Sprintf("%v", errStack)

	tags := p.InfluxTags()
	appID := tags["app_id"] // nolint:ifshort
	env := tags["env"]
	version := tags["version"]

	if appID != "" {
		zipFile := httpapi.GetSourcemapZipFileName(appID, env, version)
		zipFileAbsDir := filepath.Join(getRumSourcemapDir(sdkName), strings.TrimSuffix(zipFile, httpapi.ZipExt))

		atosBinPath := ipt.AtosBinPath
		if runtime.GOOS == "darwin" {
			if atosPath, err := exec.LookPath("atos"); err == nil && atosPath != "" {
				atosBinPath = atosPath
			}
		}
		if atosBinPath == "" {
			return p, fmt.Errorf("the path of atos/atosl bin not set")
		}
		if !path.IsFileExists(atosBinPath) {
			var err error
			atosBinPath, err = exec.LookPath(Atosl)
			if err != nil || atosBinPath == "" {
				return p, fmt.Errorf("the atos tool/atosl not found")
			}
		}
		crashAddress, err := scanIOSCrashAddress(errStackStr)
		if err != nil {
			return p, fmt.Errorf("scan crash address err: %w", err)
		}
		if len(crashAddress) == 0 {
			log.Infof("crashAddress length is 0")
			// do nothing
			return p, nil
		}
		originStackTrace := errStackStr
		token := sourceMapTokenBuckets.getToken()
		defer sourceMapTokenBuckets.sendBackToken(token)

		start := time.Now()
		for moduleName, moduleCrashes := range crashAddress {
			symbolFile, err := scanModuleSymbolFile(zipFileAbsDir, moduleName)
			if err != nil {
				log.Warnf("scan symbol file fail: %s", err)
				continue
			}
			for loadAddress, addresses := range moduleCrashes {
				if len(addresses) == 0 {
					continue
				}

				addrs := make([]string, 0, len(addresses))
				for _, addr := range addresses {
					addrs = append(addrs, addr.end)
				}

				// try resolve batch
				symbols, err := runAtosCMD(atosBinPath, symbolFile, loadAddress, addrs)
				if err != nil {
					continue
				}

				if len(symbols) == len(addresses) {
					for i, addr := range addresses {
						originStackTrace = strings.ReplaceAll(originStackTrace, addr.originStr, symbols[i])
					}
				} else if len(symbols) > 0 && len(addresses) > 1 {
					log.Errorf("resolved symbols count[%d] not equals addresses count[%d], try one by one", len(symbols), len(addresses))
					// try resolve one by one
					for _, addr := range addresses {
						addrs = addrs[:0]
						addrs = append(addrs, addr.end)
						symbols, err := runAtosCMD(atosBinPath, symbolFile, loadAddress, addrs)
						if err != nil || len(symbols) != 1 {
							continue
						}
						originStackTrace = strings.ReplaceAll(originStackTrace, addr.originStr, symbols[0])
					}
				}
			}
		}
		sourceMapDurationSummary.WithLabelValues(sdkName, appID, env, version).Observe(float64(time.Since(start)) / promDurationUnit)
		originStackB64 := base64.StdEncoding.EncodeToString([]byte(originStackTrace))
		p.MustAdd([]byte("error_stack_source_base64"), originStackB64)
		status.status = StatusOK
		return p, nil
	}
	return p, nil
}
