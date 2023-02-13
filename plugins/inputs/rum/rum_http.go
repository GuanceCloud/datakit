// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin/binding"
	influxm "github.com/influxdata/influxdb1-client/models"
	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		httpErr(resp, err)
	} else {
		httpOK(resp, nil)
	}
}

func (ipt *Input) handleRUM(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### RUM request from %s", req.URL.String())

	var (
		query                   = req.URL.Query()
		version, pipelineSource string
		precision               = dkhttp.DefaultPrecision
	)
	if x := query.Get(dkhttp.ArgVersion); x != "" {
		version = x
	}
	if x := query.Get(dkhttp.ArgPipelineSource); x != "" {
		pipelineSource = x
	}
	if x := query.Get(dkhttp.ArgPrecision); x != "" {
		precision = x
	}
	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		log.Warnf("### get invalid precision: %s", precision)
		httpErr(resp, dkhttp.ErrInvalidPrecision)

		return
	}

	body, err := uhttp.ReadBody(req)
	if err != nil {
		log.Error(err.Error())
		httpErr(resp, err)

		return
	}
	if len(body) == 0 {
		log.Debug(dkhttp.ErrEmptyBody.Err.Error())
		httpErr(resp, dkhttp.ErrEmptyBody)

		return
	}

	var (
		pts       []*point.Point
		apiConfig = config.Cfg.HTTPAPI
		isjson    = strings.Contains(req.Header.Get("Content-Type"), "application/json")
	)
	if pts, err = ipt.parseRUMBody(body, precision, isjson, geoTags(getSrcIP(apiConfig, req)), apiConfig.RUMAppIDWhiteList); err != nil {
		log.Error(err.Error())
		httpErr(resp, err)

		return
	}
	if len(pts) == 0 {
		log.Debug(dkhttp.ErrNoPoints.Err.Error())
		httpErr(resp, dkhttp.ErrNoPoints)

		return
	}

	log.Debugf("### received %d(%s) points from %s, pipeline source: %v", len(pts), req.URL.Path, inputName, pipelineSource)

	if err = dkio.Feed(inputName, req.URL.Path, pts, &dkio.Option{HighFreq: true, Version: version}); err != nil {
		log.Error(err.Error())
		httpErr(resp, err)

		return
	}

	if query.Get(dkhttp.ArgEchoLineProto) != "" {
		var res []*point.JSONPoint
		for _, pt := range pts {
			x, err := pt.ToJSON()
			if err != nil {
				log.Warnf("### ToJSON failed: %s, ignored", err.Error())
				continue
			}
			res = append(res, x)
		}
		httpOK(resp, res)
	}

	httpOK(resp, nil)
}

func (ipt *Input) parseRUMBody(body []byte, precision string, isjson bool,
	extraTags map[string]string, appIDWhiteList []string,
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
			if !contains(tags[rumMetricAppID], appIDWhiteList) {
				return nil, dkhttp.ErrRUMAppIDNotInWhiteList
			}
		}
		return rumpts, nil
	}

	rumpts, err := lp.ParsePoints(body,
		&lp.Option{
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
					err := ipt.parseSourcemap(p, sdkName)
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

func (ipt *Input) parseSourcemap(p influxm.Point, sdkName string) error {
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
			zipFileAbsPath := filepath.Join(getRumSourcemapDir(sdkName), zipFile)
			zipFileAbsDir := filepath.Join(getRumSourcemapDir(sdkName), strings.TrimSuffix(zipFile, filepath.Ext(zipFile)))

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
						if ipt.ProguardHome == "" {
							return fmt.Errorf("proguard home not set")
						}
						retraceCmd = filepath.Join(ipt.ProguardHome, "bin", "retrace.sh")
						if !path.IsFileExists(retraceCmd) {
							return fmt.Errorf("the script retrace.sh not found in the proguardHome [%s]", retraceCmd)
						}
					} else {
						if ipt.AndroidCmdLineHome == "" {
							return fmt.Errorf("android commandline tool home not set")
						}
						retraceCmd = filepath.Join(ipt.AndroidCmdLineHome, "bin/retrace")
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
					originStack = bytes.TrimLeft(originStack, "Waiting for stack-trace input...")
					originStack = bytes.TrimLeftFunc(originStack, func(r rune) bool {
						return r == '\r' || r == '\n'
					})
					originStackB64 := base64.StdEncoding.EncodeToString(originStack)
					p.AddTag("error_stack_source_base64", originStackB64)
				} else if errorType == NativeCrash {
					if ipt.NDKHome == "" {
						return fmt.Errorf("android ndk home not set")
					}

					ndkStack := filepath.Join(ipt.NDKHome, "ndk-stack")
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
				atosBinPath := ipt.AtosBinPath
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

func httpOK(w http.ResponseWriter, body interface{}) {
	if body == nil {
		if err := writeBody(w, dkhttp.OK.HttpCode, binding.MIMEJSON, nil); err != nil {
			log.Error(err.Error())
		}

		return
	}

	var (
		bodyBytes   []byte
		contentType string
		err         error
	)
	switch x := body.(type) {
	case []byte:
		bodyBytes = x
	default:
		resp := &uhttp.BodyResp{
			HttpError: dkhttp.OK,
			Content:   body,
		}
		contentType = `application/json`

		if bodyBytes, err = json.Marshal(resp); err != nil {
			log.Error(err.Error())
			jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)

			return
		}
	}

	if err := writeBody(w, dkhttp.OK.HttpCode, contentType, bodyBytes); err != nil {
		log.Error(err.Error())
	}
}

func httpErr(w http.ResponseWriter, err error) {
	switch e := err.(type) { // nolint:errorlint
	case *uhttp.HttpError:
		jsonReturnf(e, w, "")
	case *uhttp.MsgError:
		if e.Args != nil {
			jsonReturnf(e.HttpError, w, e.Fmt, e.Args)
		}
	default:
		jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "")
	}
}

func writeBody(w http.ResponseWriter, statusCode int, contentType string, body []byte) error {
	w.WriteHeader(statusCode)
	if body != nil {
		w.Header().Set("Content-Type", contentType)
		n, err := w.Write(body)
		if n != len(body) {
			return fmt.Errorf("send partial http response, full length(%d), send length(%d) ", len(body), n)
		}
		if err != nil {
			return fmt.Errorf("send http response popup err: %w", err)
		}
	}
	return nil
}

func jsonReturnf(httpErr *uhttp.HttpError, w http.ResponseWriter, format string, args ...interface{}) {
	resp := &uhttp.BodyResp{
		HttpError: httpErr,
	}

	if args != nil {
		resp.Message = fmt.Sprintf(format, args...)
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)

		return
	}

	if err := writeBody(w, httpErr.HttpCode, binding.MIMEJSON, buf); err != nil {
		log.Error(err.Error())
	}
}
