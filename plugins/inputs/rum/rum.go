// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"

	"github.com/go-sourcemap/sourcemap"
	influxm "github.com/influxdata/influxdb1-client/models"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

const iosDSYMFilePath = "Contents/Resources/DWARF"

const (
	SdkWeb     = "df_web_rum_sdk"
	SdkAndroid = "df_android_rum_sdk"
	SdkIOS     = "df_ios_rum_sdk"
)

const (
	srcMapDirWeb     = "web"
	srcMapDirAndroid = "android"
	srcMapDirIOS     = "ios"
)

const (
	JavaCrash   = "java_crash"
	NativeCrash = "native_crash"
	IOSCrash    = "ios_crash"
)

var NDKAvailableABI = map[string]struct{}{
	"armeabi-v7a": {},
	"arm64-v8a":   {},
	"x86":         {},
	"x86_64":      {},
}

var srcMapDirs = map[string]string{
	SdkWeb:     srcMapDirWeb,
	SdkAndroid: srcMapDirAndroid,
	SdkIOS:     srcMapDirIOS,
}

var (
	sourcemapCache      = make(map[string]map[string]*sourcemap.Consumer)
	sourcemapLock       sync.RWMutex
	latestCheckFileTime = time.Now().Unix()
	uncompressLock      sync.Mutex

	rumMetricNames = map[string]bool{
		`view`:      true,
		`resource`:  true,
		`error`:     true,
		`long_task`: true,
		`action`:    true,
	}

	rumMetricAppID = "app_id"
)

func sendToIO(input, category string, pts []*point.Point, opt *dkio.Option) error {
	return dkio.Feed(input, category, pts, opt)
}

func geoInfo(ip string) map[string]string {
	return geoTags(ip)
}

func geoTags(srcip string) (tags map[string]string) {
	// default set to be unknown
	tags = map[string]string{
		"city":     "unknown",
		"province": "unknown",
		"country":  "unknown",
		"isp":      "unknown",
		"ip":       srcip,
	}

	if srcip == "" {
		return
	}

	ipInfo, err := funcs.Geo(srcip)

	log.Debugf("ipinfo(%s): %+#v", srcip, ipInfo)

	if err != nil {
		log.Warnf("geo failed: %s, ignored", err)
		return
	}

	// avoid nil pointer error
	if ipInfo == nil {
		return tags
	}

	switch ipInfo.Country { // #issue 354
	case "TW":
		ipInfo.Country = "CN"
		ipInfo.Region = "Taiwan"
	case "MO":
		ipInfo.Country = "CN"
		ipInfo.Region = "Macao"
	case "HK":
		ipInfo.Country = "CN"
		ipInfo.Region = "Hong Kong"
	}

	if len(ipInfo.City) > 0 {
		tags["city"] = ipInfo.City
	}

	if len(ipInfo.Region) > 0 {
		tags["province"] = ipInfo.Region
	}

	if len(ipInfo.Country) > 0 {
		tags["country"] = ipInfo.Country
	}

	if len(srcip) > 0 {
		tags["ip"] = srcip
	}

	if isp := ip2isp.SearchIsp(srcip); len(isp) > 0 {
		tags["isp"] = isp
	}

	return tags
}

func doHandleRUMBody(body []byte,
	precision string,
	isjson bool,
	extraTags map[string]string,
	appIDWhiteList []string,
	input *Input,
) ([]*point.Point, error) {
	if isjson {
		opt := lp.NewDefaultOption()
		opt.Precision = precision
		opt.ExtraTags = extraTags
		rumpts, err := jsonPoints(body, opt)
		if err != nil {
			return nil, err
		}
		for _, p := range rumpts {
			tags := p.Tags()
			if tags != nil {
				if !contains(tags[rumMetricAppID], appIDWhiteList) {
					return nil, dkhttp.ErrRUMAppIDNotInWhiteList
				}
			}
		}
		return rumpts, nil
	}

	rumpts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extraTags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())

			if !contains(p.Tags().GetString(rumMetricAppID), appIDWhiteList) {
				return nil, dkhttp.ErrRUMAppIDNotInWhiteList
			}

			if _, ok := rumMetricNames[name]; !ok {
				return nil, uhttp.Errorf(dkhttp.ErrUnknownRUMMeasurement, "unknow RUM measurement: %s", name)
			}

			// handle sourcemap
			if name == "error" {
				sdkName := p.Tags().GetString("sdk_name")
				err := handleSourcemap(p, sdkName, input)
				if err != nil {
					log.Errorf("handle source map failed: %s", err.Error())
				}
			}

			return p, nil
		},
	})
	if err != nil {
		log.Warnf("doHandleRUMBody: %s", err)
		return nil, err
	}

	// 把error_stack_source_base64从tags中移到fields中
	for i, rumpt := range rumpts {
		fields, err := rumpt.Fields()
		if err != nil {
			log.Errorf("get client.Point Fields() err: %s", err)
			continue
		}
		_, ok1 := fields["error_stack"]
		_, ok2 := rumpt.Tags()["error_stack_source_base64"]
		if ok1 && ok2 {
			tags := rumpt.Tags()
			fields["error_stack_source_base64"] = tags["error_stack_source_base64"]
			delete(tags, "error_stack_source_base64")
			newPoint, err := client.NewPoint(rumpt.Name(), tags, fields, rumpt.Time())
			if err != nil {
				log.Errorf("client.NewPoint() err: %s", err)
				continue
			}
			rumpts[i] = newPoint
		}
	}

	return point.WrapPoint(rumpts), nil
}

func contains(str string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, a := range list {
		if a == str {
			return true
		}
	}
	return false
}

func getSrcIP(ac *dkhttp.APIConfig, req *http.Request) (ip string) {
	if ac != nil {
		ip = req.Header.Get(ac.RUMOriginIPHeader)
		log.Debugf("get ip from %s: %s", ac.RUMOriginIPHeader, ip)
		if ip == "" {
			for k, v := range req.Header {
				log.Debugf("%s: %s", k, strings.Join(v, ","))
			}
		}
	} else {
		log.Info("apiConfig not set")
	}

	if ip != "" {
		log.Debugf("header remote addr: %s", ip)
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			ip = parts[0] // 注意：此处只取第一个 IP 作为源 IP
			return
		}
	} else { // 默认取 http 框架带进来的 IP
		log.Debugf("gin remote addr: %s", req.RemoteAddr)
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			ip = host
			return
		} else {
			log.Warnf("net.SplitHostPort(%s): %s, ignored", req.RemoteAddr, err)
		}
	}

	return ip
}

func handleRUMBody(body []byte,
	precision string,
	isjson bool,
	geoInfo map[string]string,
	list []string,
	input *Input,
) ([]*point.Point, error) {
	return doHandleRUMBody(body, precision, isjson, geoInfo, list, input)
}

type iosCrashAddress struct {
	start     string
	end       string
	originStr string
}

func scanIOSCrashAddress(originErrStack string) (map[string]map[string][]iosCrashAddress, error) {
	// for match
	// 4   App                                         0x0000000104fd0728 0x104f30000 + 657192
	//
	// $1 "App"
	// $2 "0x0000000104fd0728"
	// $3 "0x104f30000 + 657192"  // 用于解析后对原始堆栈进行替换
	// $4 0x104f30000
	// $5 "657192"
	expStr := `(\S+)\s+(0x[0-9a-fA-F]+)\s+((0x[0-9a-fA-F]+)\s*\+\s*(\d+|0x[0-9a-fA-F]+))`
	re, err := regexp.Compile(expStr)
	if err != nil {
		return nil, fmt.Errorf("compile regexp [%s] fail: %w", expStr, err)
	}

	matches := re.FindAllStringSubmatch(originErrStack, -1)

	crashAddress := make(map[string]map[string][]iosCrashAddress)
	for _, match := range matches {
		if len(match) != 6 {
			continue
		}

		moduleName := match[1]
		startAddr := match[4]

		if _, ok := crashAddress[moduleName]; !ok {
			crashAddress[moduleName] = make(map[string][]iosCrashAddress)
		}

		crashAddress[moduleName][startAddr] = append(crashAddress[moduleName][startAddr], iosCrashAddress{
			start:     startAddr,
			end:       match[2],
			originStr: match[3],
		})
	}

	return crashAddress, nil
}

func scanABI(crashReport string) string {
	for _, fieldName := range []string{"ABI list:", "ABI:"} {
		idx := strings.Index(crashReport, fieldName)
		log.Infof("index of [%s] is %d", fieldName, idx)
		if idx > -1 {
			if eolIdx := strings.IndexAny(crashReport[idx:], "\r\n"); eolIdx > -1 {
				log.Infof("index of EOL is %d", eolIdx)
				abi := strings.Trim(strings.TrimSpace(crashReport[idx+len(fieldName):idx+eolIdx]), "'")
				abiList := []string{abi}
				if strings.ContainsRune(abi, ',') {
					abiList = strings.Split(abi, ",")
				}
				for _, abi := range abiList {
					if _, ok := NDKAvailableABI[abi]; ok {
						return abi
					}
				}
				log.Errorf("find illegal abi [%s], ignore", abi)
			}
		}
	}

	return ""
}

func checkJavaShrinkTool(mappingFile string) (string, error) {
	fp, err := os.Open(mappingFile) // nolint:gosec
	if err != nil {
		return "", fmt.Errorf("open mapping file [%s] fail: %w", mappingFile, err)
	}
	defer func() {
		_ = fp.Close()
	}()

	scanner := bufio.NewScanner(fp)

	rows := 0
	for scanner.Scan() && rows < 10 {
		rows++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") &&
			(strings.Contains(line, "compiler: R8") || strings.Contains(line, "com.android.tools.r8.mapping")) {
			return cmds.AndroidCommandLineTools, nil
		}
	}
	return cmds.Proguard, nil
}

func handleSourcemap(p influxm.Point, sdkName string, input *Input) error {
	fields, err := p.Fields()
	if err != nil {
		return fmt.Errorf("parse field error: %w", err)
	}
	errStack, ok := fields["error_stack"]

	// if error_stack exists
	if ok {
		errStackStr := fmt.Sprintf("%v", errStack)

		appID := p.Tags().GetString("app_id")
		env := p.Tags().GetString("env")
		version := p.Tags().GetString("version")

		if len(appID) > 0 && (len(env) > 0) && (len(version) > 0) {
			zipFile := GetSourcemapZipFileName(appID, env, version)
			zipFileAbsPath := filepath.Join(GetRumSourcemapDir(sdkName), zipFile)
			zipFileAbsDir := filepath.Join(GetRumSourcemapDir(sdkName), strings.TrimSuffix(zipFile, filepath.Ext(zipFile)))

			switch sdkName {
			case SdkWeb:
				sourcemapLock.RLock()
				_, ok := sourcemapCache[zipFile]
				sourcemapLock.RUnlock()

				if !ok {
					// 判断zip文件是否存在，存在则加载
					nowSecs := time.Now().Unix()
					// 2分钟检查一次
					if nowSecs-atomic.SwapInt64(&latestCheckFileTime, nowSecs) > 120 {
						if path.IsFileExists(zipFileAbsPath) {
							if err := updateSourcemapCache(zipFileAbsPath); err == nil {
								ok = true
							}
						}
					}
				}

				if ok {
					errorStackSource := getSourcemap(errStackStr, sourcemapCache[zipFile])
					errorStackSourceBase64 := base64.StdEncoding.EncodeToString([]byte(errorStackSource)) // tag cannot have '\n'
					p.AddTag("error_stack_source_base64", errorStackSourceBase64)
				}

			case SdkAndroid:

				errorType := p.Tags().GetString("error_type")

				if errorType == JavaCrash {
					if err := uncompressZipFile(zipFileAbsPath); err != nil {
						return fmt.Errorf("uncompress zip file fail: %w", err)
					}

					mappingFile := filepath.Join(zipFileAbsDir, "mapping.txt")
					if !path.IsFileExists(mappingFile) {
						return fmt.Errorf("java source mapping file not exists")
					}
					toolName, err := checkJavaShrinkTool(mappingFile)
					if err != nil {
						return fmt.Errorf("verify java shrink tool fail: %w", err)
					}
					retraceCmd := ""
					if toolName == cmds.Proguard {
						if input.ProguardHome == "" {
							return fmt.Errorf("proguard home not set")
						}
						retraceCmd = filepath.Join(input.ProguardHome, "bin", "retrace.sh")
						if !path.IsFileExists(retraceCmd) {
							return fmt.Errorf("the script retrace.sh not found in the proguardHome [%s]", retraceCmd)
						}
					} else {
						if input.AndroidCmdLineHome == "" {
							return fmt.Errorf("android commandline tool home not set")
						}
						retraceCmd = filepath.Join(input.AndroidCmdLineHome, "bin/retrace")
						if !path.IsFileExists(retraceCmd) {
							return fmt.Errorf("the cmdline tools [retrace] not found in the androidCmdLineHome [%s]", retraceCmd)
						}
					}

					token := sourceMapTokenBuckets.getToken()
					defer sourceMapTokenBuckets.sendBackToken(token)
					cmd := exec.Command("sh", retraceCmd, mappingFile) //nolint:gosec
					cmd.Stdin = strings.NewReader(errStackStr)
					originStack, err := cmd.Output()
					if err != nil {
						if ee, ok := err.(*exec.ExitError); ok { //nolint:errorlint
							return fmt.Errorf("run proguard retrace fail: %w, error_msg: %s", err, string(ee.Stderr))
						}
						return fmt.Errorf("run proguard retrace fail: %w", err)
					}
					originStackB64 := base64.StdEncoding.EncodeToString(originStack)
					p.AddTag("error_stack_source_base64", originStackB64)
				} else if errorType == NativeCrash {
					if input.NDKHome == "" {
						return fmt.Errorf("android ndk home not set")
					}

					ndkStack := filepath.Join(input.NDKHome, "ndk-stack")
					stat, err := os.Stat(ndkStack)
					if err != nil {
						return fmt.Errorf("ndk-stack command tool not found in the NDK HOME [%s]", ndkStack)
					}

					if !stat.Mode().IsRegular() {
						return fmt.Errorf("ndk-stack path is not a valid exectable program [%s]", ndkStack)
					}

					abi := scanABI(errStackStr)
					if abi == "" {
						return fmt.Errorf("no valid NDK ABI found")
					}

					if err := uncompressZipFile(zipFileAbsPath); err != nil {
						return fmt.Errorf("uncompress zip file fail: %w", err)
					}

					symbolObjDir := filepath.Join(zipFileAbsDir, abi)
					if !path.IsDir(symbolObjDir) {
						return fmt.Errorf("expected native objects dir [%s] not found", symbolObjDir)
					}

					cmd := exec.Command(ndkStack, "--sym", symbolObjDir) //nolint:gosec
					cmd.Stdin = strings.NewReader(errStackStr)
					originStack, err := cmd.Output()
					if err != nil {
						if ee, ok := err.(*exec.ExitError); ok { //nolint:errorlint
							return fmt.Errorf("run ndk-stack tool fail: %w, error_msg: %s", err, string(ee.Stderr))
						}
						return fmt.Errorf("run ndk-stack tool fail: %w", err)
					}

					re := regexp.MustCompile(`\s#00\s+`)
					idxRange := re.FindIndex(originStack)
					if len(idxRange) == 2 {
						originStack = originStack[idxRange[0]:]
					}

					originStack = bytes.ReplaceAll(originStack, []byte("Crash dump is completed"), nil)

					re = regexp.MustCompile(`backtrace:[\s\S]+?logcat:`)
					idxRange = re.FindStringIndex(errStackStr)
					if len(idxRange) == 2 {
						originStack = []byte(strings.ReplaceAll(errStackStr, errStackStr[idxRange[0]:idxRange[1]],
							fmt.Sprintf("backtrace:\n%slogcat:", string(originStack))))
					}

					originStackB64 := base64.StdEncoding.EncodeToString(originStack)
					p.AddTag("error_stack_source_base64", originStackB64)

					log.Infof("native crash source map 处理成功， appid: %s, creat time: %s", appID, p.Time().In(time.Local).Format(time.RFC3339))
				}

			case SdkIOS:
				atosBinPath := input.AtosBinPath
				if runtime.GOOS == "darwin" {
					if atosPath, err := exec.LookPath("atos"); err == nil && atosPath != "" {
						atosBinPath = atosPath
					}
				}
				if atosBinPath == "" {
					return fmt.Errorf("the path of atos/atosl bin not set")
				}
				if !path.IsFileExists(atosBinPath) {
					atosBinPath, err = exec.LookPath(cmds.Atosl)
					if err != nil || atosBinPath == "" {
						return fmt.Errorf("the atos tool/atosl not found")
					}
				}
				if err := uncompressZipFile(zipFileAbsPath); err != nil {
					return fmt.Errorf("uncompress zip file fail: %w", err)
				}
				crashAddress, err := scanIOSCrashAddress(errStackStr)
				if err != nil {
					return fmt.Errorf("scan crash address err: %w", err)
				}
				if len(crashAddress) == 0 {
					log.Infof("crashAddress length is 0")
					// do nothing
					return nil
				}
				originStackTrace := errStackStr
				token := sourceMapTokenBuckets.getToken()
				defer sourceMapTokenBuckets.sendBackToken(token)

				for moduleName, moduleCrashes := range crashAddress {
					symbolFile, err := scanModuleSymbolFile(zipFileAbsDir, moduleName)
					if err != nil {
						log.Errorf("scan symbol file fail: %s", err)
						continue
					}
					for startAddr, addresses := range moduleCrashes {
						for _, addr := range addresses {
							args := []string{
								"-o",
								symbolFile,
								"-l",
								startAddr,
								addr.end,
							}
							cmd := exec.Command(atosBinPath, args...) //nolint:gosec
							cmd.Env = []string{"HOME=/root"}          // run the tool "atosl" must set this env, why?
							log.Infof("%s %s", atosBinPath, strings.Join(args, " "))
							stdout, err := cmd.Output()
							if err != nil {
								if ee, ok := err.(*exec.ExitError); ok { // nolint:errorlint
									log.Errorf("run atos tool fail: %s, output: [%s], err: %s", string(ee.Stderr), string(stdout), err)
								} else {
									log.Errorf("run atos tool fail: %s", err)
								}
								continue
							}

							// adapt varies os newLine
							stdoutStr := strings.ReplaceAll(string(stdout), "\r\n", "\n")
							stdoutStr = strings.ReplaceAll(stdoutStr, "\r", "\n")
							stdoutStr = strings.Trim(stdoutStr, "\n")
							// symbols := strings.Split(stdoutStr, "\n")
							originStackTrace = strings.ReplaceAll(originStackTrace, addr.originStr, stdoutStr)
						}
					}
				}
				originStackB64 := base64.StdEncoding.EncodeToString([]byte(originStackTrace))
				p.AddTag("error_stack_source_base64", originStackB64)
			}
		}
	}

	return nil
}

type simpleQueue struct {
	buckets []string
}

func newSimpleQueue(initCap ...uint) *simpleQueue {
	if len(initCap) > 0 {
		return &simpleQueue{
			buckets: make([]string, 0, initCap[0]),
		}
	}
	return &simpleQueue{}
}

func (sq *simpleQueue) IsEmpty() bool {
	return len(sq.buckets) == 0
}

func (sq *simpleQueue) enqueue(item string) {
	sq.buckets = append(sq.buckets, item)
}

func (sq *simpleQueue) dequeue() string {
	if sq.IsEmpty() {
		return ""
	}
	item := sq.buckets[0]
	sq.buckets = sq.buckets[1:]
	return item
}

func scanModuleSymbolFile(dir string, moduleName string) (string, error) {
	queue := newSimpleQueue(4)
	queue.enqueue(dir)

	for !queue.IsEmpty() {
		dir = queue.dequeue()
		files, _ := os.ReadDir(dir)

		for _, entry := range files {
			if strings.HasPrefix(entry.Name(), moduleName) &&
				strings.HasSuffix(strings.ToLower(entry.Name()), ".dsym") {
				if entry.IsDir() {
					// for path SomeName.app.dSYM/Contents/Resources/DWARF/SomeName
					dSYMFile := filepath.Join(dir, entry.Name(), iosDSYMFilePath, moduleName)
					if path.IsFileExists(dSYMFile) {
						return dSYMFile, nil
					} else {
						log.Infof("file [%s] not exists, ignore...", dSYMFile)
					}
				} else if entry.Type().IsRegular() {
					return filepath.Join(dir, entry.Name()), nil
				}
			}

			if entry.IsDir() {
				// 放入队列继续查找
				queue.enqueue(filepath.Join(dir, entry.Name()))
			}
		}
	}

	return "", fmt.Errorf("IOS dSYM symbol file for module[%s] not found", moduleName)
}

func getSourcemap(errStack string, sourcemapItem map[string]*sourcemap.Consumer) string {
	reg := regexp.MustCompile(`@ .*:\d+:\d+`)

	replaceStr := reg.ReplaceAllStringFunc(errStack, func(str string) string {
		return str[0:2] + getSourceMapString(str[2:], sourcemapItem)
	})
	return replaceStr
}

func getSourceMapString(str string, sourcemapItem map[string]*sourcemap.Consumer) string {
	parts := strings.Split(str, ":")
	partsLen := len(parts)
	if partsLen < 3 {
		return str
	}
	rowNumber, err := strconv.Atoi(parts[partsLen-2])
	if err != nil {
		return str
	}
	colNumber, err := strconv.Atoi(parts[partsLen-1])
	if err != nil {
		return str
	}

	srcPath := strings.Join(parts[:partsLen-2], ":") // http://localhost:5500/dist/bundle.js

	urlObj, err := url.Parse(srcPath)
	if err != nil {
		log.Debugf("parse url failed, %s, %s", srcPath, err.Error())
		return str
	}

	urlPath := strings.TrimPrefix(urlObj.Path, "/")
	sourceMapFileName := urlPath + ".map"

	smap, ok := sourcemapItem[sourceMapFileName]
	if !ok {
		log.Debugf("no sourcemap: %s", sourceMapFileName)
		return str
	}

	file, _, line, col, ok := smap.Source(rowNumber, colNumber)

	if ok {
		return fmt.Sprintf("%s:%v:%v", file, line, col)
	}

	return str
}

// GetSourcemapZipFileName  zip file name.
func GetSourcemapZipFileName(applicatinID, env, version string) string {
	fileName := fmt.Sprintf("%s-%s-%s.zip", applicatinID, env, version)

	return strings.ReplaceAll(fileName, string(filepath.Separator), "__")
}

func GetRumSourcemapDir(sdkName string) string {
	dir, ok := srcMapDirs[sdkName]
	if !ok {
		dir = srcMapDirWeb
	}
	rumDir := filepath.Join(datakit.DataDir, "rum", dir)
	return rumDir
}

func loadSourcemapFile() {
	rumDir := GetRumSourcemapDir(srcMapDirWeb)
	files, err := ioutil.ReadDir(rumDir)
	if err != nil {
		log.Errorf("load rum sourcemap dir failed: %s", err.Error())
		return
	}

	sourcemapLock.Lock()
	defer sourcemapLock.Unlock()

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.HasSuffix(fileName, ".zip") {
				sourcemapItem, err := loadZipFile(filepath.Join(rumDir, fileName))
				if err != nil {
					log.Debugf("load zip file %s failed, %s", fileName, err.Error())
					continue
				}

				sourcemapCache[fileName] = sourcemapItem
			}
		}
	}
}

func copyZipItem(item *zip.File, dst string) error {
	reader, err := item.Open()
	if err != nil {
		return fmt.Errorf("open zip item fail: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	dstDir := filepath.Dir(dst)
	if !path.IsDir(dstDir) {
		if err := os.MkdirAll(dstDir, 0o750); err != nil {
			return fmt.Errorf("mkdir [%s] fail: %w", dstDir, err)
		}
	}

	writer, err := os.Create(dst) // nolint:gosec
	if err != nil {
		return fmt.Errorf("create dst file fail: %w", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	_, err = io.Copy(writer, reader) // nolint:gosec
	if err != nil {
		return fmt.Errorf("copy zip item fail: %w", err)
	}
	return nil
}

func uncompressZipFile(zipFileAbsPath string) error {
	zipFileAbsDir := strings.TrimSuffix(zipFileAbsPath, filepath.Ext(zipFileAbsPath))

	needUncompress := false

	dirStatInfo, dirErr := os.Stat(zipFileAbsDir)
	zipFileStat, fileErr := os.Stat(zipFileAbsPath)

	if os.IsNotExist(dirErr) || !dirStatInfo.IsDir() {
		if os.IsNotExist(fileErr) || !zipFileStat.Mode().IsRegular() {
			return fmt.Errorf("sourcemap zip file [%s] not exists", zipFileAbsPath)
		}
		needUncompress = true
	} else if fileErr == nil && zipFileStat.Mode().IsRegular() {
		if dirErr == nil && dirStatInfo.ModTime().Before(zipFileStat.ModTime()) {
			log.Infof("文件夹修改时间早于zip文件，重新解压缩zip包")
			needUncompress = true
		}
	}

	if needUncompress {
		// 加锁，避免消耗太多磁盘资源
		uncompressLock.Lock()
		defer uncompressLock.Unlock()

		zipFileDir := filepath.Dir(zipFileAbsPath)
		baseZipName := strings.TrimSuffix(filepath.Base(zipFileAbsPath), filepath.Ext(zipFileAbsPath))

		reader, err := zip.OpenReader(zipFileAbsPath)
		if err != nil {
			return fmt.Errorf("open zip file [%s] fail: %w", zipFileAbsPath, err)
		}
		defer func() {
			_ = reader.Close()
		}()

		for _, item := range reader.File {
			var absItemPath string
			if !strings.HasPrefix(filepath.Clean(item.Name), baseZipName) {
				absItemPath = filepath.Join(zipFileDir, baseZipName, item.Name) // nolint:gosec
			} else {
				absItemPath = filepath.Join(zipFileDir, item.Name) // nolint:gosec
			}

			if item.FileInfo().IsDir() {
				if err := os.MkdirAll(absItemPath, 0o750); err != nil {
					return fmt.Errorf("can not mkdir: %w", err)
				}
			} else if err := copyZipItem(item, absItemPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadZipFile(zipFile string) (map[string]*sourcemap.Consumer, error) {
	sourcemapItem := make(map[string]*sourcemap.Consumer)

	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close() //nolint:errcheck

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() && strings.HasSuffix(f.Name, ".map") {
			file, err := f.Open()
			if err != nil {
				log.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}
			defer file.Close() // nolint:errcheck

			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Debugf("ignore sourcemap %s, %s", f.Name, err.Error())
				continue
			}

			smap, err := sourcemap.Parse(f.Name, content)
			if err != nil {
				log.Debugf("sourcemap parse failed, %s", err.Error())
				continue
			}

			sourcemapItem[f.Name] = smap
		}
	}

	return sourcemapItem, nil
}

func updateSourcemapCache(zipFile string) error {
	fileName := filepath.Base(zipFile)
	if !strings.HasSuffix(fileName, ".zip") {
		return fmt.Errorf(`suffix name is not ".zip" [%s]`, zipFile)
	}
	sourcemapItem, err := loadZipFile(zipFile)
	if err != nil {
		log.Errorf("load zip file error: %s", err.Error())
		return fmt.Errorf("load zip file [%s] err: %w", zipFile, err)
	}
	sourcemapLock.Lock()
	defer sourcemapLock.Unlock()
	sourcemapCache[fileName] = sourcemapItem
	log.Debugf("load sourcemap success: %s", fileName)
	return nil
}

func deleteSourcemapCache(zipFile string) {
	fileName := filepath.Base(zipFile)
	if strings.HasSuffix(fileName, ".zip") {
		sourcemapLock.Lock()
		defer sourcemapLock.Unlock()
		delete(sourcemapCache, fileName)
	}
}
