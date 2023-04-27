// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	lp "github.com/GuanceCloud/cliutils/lineproto"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/go-sourcemap/sourcemap"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestRUMHandleBody(t *testing.T) {
	cases := []struct {
		name           string
		body           []byte
		prec           string
		fail           bool
		js             bool
		npts           int
		appidWhiteList []string
	}{
		{
			name: `invalid json`,
			body: []byte(`[{
"measurement": "error",
"tags": {"t1": "tv1"},
"fields": 1.0, "f2": 2}
}]`),
			npts: 1,
			fail: true,
			js:   true,
		},

		{
			name: `valid app_id`,
			body: []byte(`[{
"measurement": "error",
"tags": {"app_id": "appid01"},
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid01"},
		},

		{
			name: `invalid app_id`,
			body: []byte(`[{
"measurement": "error",
"tags": {"app_id": "appid01"},
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `invalid json, no tags`,
			body: []byte(`[{
"measurement": "error",
"fields": {"f1": 1.0, "f2": 2}
}]`),
			npts:           1,
			js:             true,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `Precision ms`,
			prec: "ms",
			body: []byte(`error,app_id=appid01,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,app_id=appid01,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts:           2,
			appidWhiteList: []string{"appid01"},
		},

		{
			name: "app_id not in white-list",
			prec: "ms",
			body: []byte(`error,app_id=appid01,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,app_id=appid01,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts:           2,
			appidWhiteList: []string{"appid02"},
			fail:           true,
		},

		{
			name: `Precision ns`,
			prec: "n",
			body: []byte(`error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
			action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`),
			npts: 5,
		},

		{
			// 行协议指标带换行
			name: "line break in point",
			prec: "ms",
			body: []byte(`error,sdk_name=Web\ SDK,sdk_version=2.0.1,app_id=appid_16b35953792f4fcda0ca678d81dd6f1a,env=production,version=1.0.0,userid=60f0eae1-01b8-431e-85c9-a0b7bcb391e1,session_id=8c96307f-5ef0-4533-be8f-c84e622578cc,is_signin=F,os=Mac\ OS,os_version=10.11.6,os_version_major=10,browser=Chrome,browser_version=90.0.4430.212,browser_version_major=90,screen_size=1920*1080,network_type=4g,view_id=addb07a3-5ab9-4e30-8b4f-6713fc54fb4e,view_url=http://172.16.5.9:5003/,view_host=172.16.5.9:5003,view_path=/,view_path_group=/,view_url_query={},error_source=source,error_type=ReferenceError error_starttime=1621244127493,error_message="displayDate is not defined",error_stack="ReferenceError
  at onload @ http://172.16.5.9:5003/:25:30" 1621244127493`),
			npts: 1,
		},
	}

	ipt := &Input{}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := ipt.parseRUMBody(tc.body, tc.prec, tc.js, nil, tc.appidWhiteList)

			if tc.fail {
				tu.NotOk(t, err, "case[%d] expect fail, but ok", i)
				t.Logf("[%d] handle body failed: %s", i, err)
				return
			}

			if err != nil && !tc.fail {
				t.Errorf("[FAIL][%d] handle body failed: %s", i, err)
				return
			}

			tu.Equals(t, tc.npts, len(pts))

			for _, pt := range pts {
				lp := pt.String()
				t.Logf("\t%s", lp)
				_, err := models.ParsePointsWithPrecision([]byte(lp), time.Now(), "n")
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestHandleSourcemap(t *testing.T) {
	bodyStr := `resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=https://static.dataflux.cn/browser-sdk/v2/dataflux-rum.js,resource_url_host=static.dataflux.cn,resource_url_path=/browser-sdk/v2/dataflux-rum.js,resource_url_path_group=/browser-sdk/?/dataflux-rum.js,resource_url_query={},resource_type=js,resource_status=200,resource_status_group=2xx,resource_method=GET duration=0 1636524705407
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=158000000 1636524705407
	resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=http://localhost:5500/dist/,resource_url_host=localhost,resource_url_path=/dist/,resource_url_path_group=/dist/,resource_url_query={},resource_type=document,resource_method=GET duration=25300000,resource_ttfb=4900000,resource_trans=1300000,resource_first_byte=16100000 1636524705288
	resource,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},resource_url=http://localhost:5500/dist/bundle.js,resource_url_host=localhost,resource_url_path=/dist/bundle.js,resource_url_path_group=/dist/bundle.js,resource_url_query={},resource_type=js,resource_method=GET duration=15100000,resource_ttfb=2500000,resource_trans=1100000,resource_first_byte=14000000 1636524705407
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=118000000 1636524705588
	long_task,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={} duration=703000000 1636524705819
	error,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},error_source=source,error_type=TypeError error_message="l.aa is not a function",error_stack="TypeError
		at <anonymous> @ http://localhost:5500/dist/bundle.js:1:821" 1636524715601
	error,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},error_source=source,error_type=TypeError error_message="l.aa is not a function",error_stack="TypeError
		at <anonymous> @ http://localhost:5500/dist/bundle.js:1:821" 1636524725599
	view,sdk_name=df_web_rum_sdk,sdk_version=2.0.15,app_id=appid_e6208285ab6947dbaef25d1f1e4749bd,env=production,version=1.0.0,userid=5499c5f6-9b35-43dc-b752-d291a4677b07,session_id=e0e08c94-3096-419e-8eae-1273d045153c,session_type=user,is_signin=F,os=Mac\ OS,os_version=10.14.6,os_version_major=10,browser=Chrome,browser_version=95.0.4638.69,browser_version_major=95,screen_size=2560*1440,network_type=4g,view_id=6f272fe3-5ab1-430c-98b4-45d3d091126c,view_referrer=http://localhost:5500/dist/,view_url=http://localhost:5500/dist/,view_host=localhost:5500,view_path=/dist/,view_path_group=/dist/,view_url_query={},view_loading_type=initial_load,view_apdex_level=0,is_active=true view_error_count=2,view_resource_count=3,view_long_task_count=3,view_action_count=0,cumulative_layout_shift=0,loading_time=1252500000,dom_interactive=310000000,dom_content_loaded=331900000,dom_complete=1249000000,first_paint_time=17400000,resource_load_time=917200000,time_to_interactive=302100000,dom=939000000,dom_ready=324000000,time_spent=23318700000 1636524705288`
	body := []byte(bodyStr)
	mapFileContent := `{"version":3,"file":"bundle.js","mappings":"qBAeAA,EAAOC,QAfP,WACEC,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,sBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,gBACZD,QAAQC,IAAI,mBCXVC,EAA2B,GAG/B,SAASC,EAAoBC,GAE5B,IAAIC,EAAeH,EAAyBE,GAC5C,QAAqBE,IAAjBD,EACH,OAAOA,EAAaN,QAGrB,IAAID,EAASI,EAAyBE,GAAY,CAGjDL,QAAS,IAOV,OAHAQ,EAAoBH,GAAUN,EAAQA,EAAOC,QAASI,GAG/CL,EAAOC,QCpBfI,EAAoBK,EAAKV,IACxB,IAAIW,EAASX,GAAUA,EAAOY,WAC7B,IAAOZ,EAAiB,QACxB,IAAM,EAEP,OADAK,EAAoBQ,EAAEF,EAAQ,CAAEG,EAAGH,IAC5BA,GCLRN,EAAoBQ,EAAI,CAACZ,EAASc,KACjC,IAAI,IAAIC,KAAOD,EACXV,EAAoBY,EAAEF,EAAYC,KAASX,EAAoBY,EAAEhB,EAASe,IAC5EE,OAAOC,eAAelB,EAASe,EAAK,CAAEI,YAAY,EAAMC,IAAKN,EAAWC,MCJ3EX,EAAoBY,EAAI,CAACK,EAAKC,IAAUL,OAAOM,UAAUC,eAAeC,KAAKJ,EAAKC,G,yCCEhFrB,QAAQC,IAAI,KCEd,MAEAwB,aAAY,KAUVC,EAAEC,OARD,KAEH,IAAID,EAAI,I","sources":["webpack:///./src/func1.js","webpack:///webpack/bootstrap","webpack:///webpack/runtime/compat get default export","webpack:///webpack/runtime/define property getters","webpack:///webpack/runtime/hasOwnProperty shorthand","webpack:///./src/func.js","webpack:///./src/index.js"],"sourcesContent":["function add () {\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxx\\n\\n\\nxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n  console.log(\"xxxxxxxxxxxx\")\n}\n\nmodule.exports = add","// The module cache\nvar __webpack_module_cache__ = {};\n\n// The require function\nfunction __webpack_require__(moduleId) {\n\t// Check if module is in cache\n\tvar cachedModule = __webpack_module_cache__[moduleId];\n\tif (cachedModule !== undefined) {\n\t\treturn cachedModule.exports;\n\t}\n\t// Create a new module (and put it into the cache)\n\tvar module = __webpack_module_cache__[moduleId] = {\n\t\t// no module.id needed\n\t\t// no module.loaded needed\n\t\texports: {}\n\t};\n\n\t// Execute the module function\n\t__webpack_modules__[moduleId](module, module.exports, __webpack_require__);\n\n\t// Return the exports of the module\n\treturn module.exports;\n}\n\n","// getDefaultExport function for compatibility with non-harmony modules\n__webpack_require__.n = (module) => {\n\tvar getter = module && module.__esModule ?\n\t\t() => (module['default']) :\n\t\t() => (module);\n\t__webpack_require__.d(getter, { a: getter });\n\treturn getter;\n};","// define getter functions for harmony exports\n__webpack_require__.d = (exports, definition) => {\n\tfor(var key in definition) {\n\t\tif(__webpack_require__.o(definition, key) && !__webpack_require__.o(exports, key)) {\n\t\t\tObject.defineProperty(exports, key, { enumerable: true, get: definition[key] });\n\t\t}\n\t}\n};","__webpack_require__.o = (obj, prop) => (Object.prototype.hasOwnProperty.call(obj, prop))","export default function run() {\n  debugger\n  console.log(\"a\")\n}","import run from './func'\nimport run1 from './func1'\n\nrun()\nrun1()\n\nsetInterval(() => {\n  a()\n}, 10000)\n\nlet b = {}\nfunction a(){\n  c()\n}\n\nfunction c(){\n  b.aa()\n}"],"names":["module","exports","console","log","__webpack_module_cache__","__webpack_require__","moduleId","cachedModule","undefined","__webpack_modules__","n","getter","__esModule","d","a","definition","key","o","Object","defineProperty","enumerable","get","obj","prop","prototype","hasOwnProperty","call","setInterval","b","aa"],"sourceRoot":""}`

	tmpDir := "./tmp"
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Error(err)
		}
	})

	datakit.DataDir = path.Join(tmpDir, "data")

	sourcemapFileName := GetSourcemapZipFileName("appid_e6208285ab6947dbaef25d1f1e4749bd", "production", "1.0.0")
	rumDir := getRumSourcemapDir(srcMapDirWeb)

	err := os.MkdirAll(rumDir, os.ModePerm)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)

	f, err := w.Create("dist/bundle.js.map")

	assert.NoError(t, err)
	_, err = f.Write([]byte(mapFileContent))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	zipFilePath := filepath.Join(rumDir, sourcemapFileName)

	err = ioutil.WriteFile(zipFilePath, buf.Bytes(), os.ModePerm)

	assert.NoError(t, err)
	defer os.Remove(zipFilePath) //nolint:errcheck

	loadSourcemapFile()

	pts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: "ms",
		Callback: func(p models.Point) (models.Point, error) {
			if string(p.Name()) == "error" {
				ipt := &Input{}
				p, _ = ipt.parseSourcemap(p, SdkWeb)
			}
			return p, nil
		},
	})

	assert.NoError(t, err)
	assert.Greater(t, len(pts), 0)

	for _, p := range pts {
		if p.Name() == "error" {
			fields, err := p.Fields()
			if err != nil {
				continue
			}
			if _, ok := fields["error_stack"]; ok {
				fields, err := p.Fields()
				assert.NoError(t, err)
				if errorStackSource, ok := fields["error_stack_source_base64"]; !ok {
					assert.Fail(t, "error stack transform failed")
				} else {
					decodBytes, err := base64.StdEncoding.DecodeString(errorStackSource.(string))
					assert.NoError(t, err)
					assert.Contains(t, string(decodBytes), "webpack")
				}
			}
		}
	}

	t.Run("updateSourcemapCache", func(t *testing.T) {
		updateSourcemapCache("invalid")
		assert.Equal(t, len(sourcemapCache), 1)

		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		_, ok := sourcemapCache[fileName]
		assert.True(t, ok)

		deleteSourcemapCache(zipFilePath)
		_, ok = sourcemapCache[fileName]
		assert.False(t, ok)
	})

	t.Run("getSourceMapString", func(t *testing.T) {
		updateSourcemapCache(zipFilePath)
		fileName := filepath.Base(zipFilePath)
		cases := []struct {
			src           string
			dest          string
			desc          string
			sourceMapItem map[string]*sourcemap.Consumer
		}{
			{
				src:  "http://a.com/1.js",
				dest: "http://a.com/1.js",
				desc: "split with colon, length less than 3",
			},
			{
				src:  "http://a.com/1.js:a:1",
				dest: "http://a.com/1.js:a:1",
				desc: "invalid row number",
			},
			{
				src:  "http://a.com/1.js:1:a",
				dest: "http://a.com/1.js:1:a",
				desc: "invalid col number",
			},
			{
				src:           "http://localhost:5500/dist/bundle.js:1:821",
				dest:          "webpack:///./src/index.js:17:4",
				desc:          "it shourld work",
				sourceMapItem: sourcemapCache[fileName],
			},
			{
				src:           "http:///localhost:5500/dist/bundle.js:1:821",
				dest:          "http:///localhost:5500/dist/bundle.js:1:821",
				desc:          "invalid url",
				sourceMapItem: sourcemapCache[fileName],
			},
			{
				src:  "http://localhost:5500/dist/bundle.js:1:821",
				dest: "http://localhost:5500/dist/bundle.js:1:821",
				desc: "it shourld not work without sourcemap file",
			},
		}

		for _, item := range cases {
			assert.Equal(t, getSourceMapString(item.src, item.sourceMapItem), item.dest, item.desc)
		}
	})
}

func TestScanModuleSymbolFile(t *testing.T) {
	file, err := scanModuleSymbolFile("/Users/zy/software/source_map/iOS", "App")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(file)
	}
}

func TestScanIOSCrashAddress(t *testing.T) {
	originCrash := `

Hardware Model:  arm64
OS Version:   iPhone OS 15.2
Report Version:  104

Code Type:   ARM64

Last Exception Backtrace:
0   CoreFoundation                      0x00000001803e1188 __exceptionPreprocess + 236
1   libobjc.A.dylib                     0x0000000180193384 objc_exception_throw + 56
2   CoreFoundation                      0x00000001803f0530 +[NSObject(NSObject) instanceMethodSignatureForSelector:] + 0
3   CoreFoundation                      0x00000001803e53fc ___forwarding___ + 1408
4   CoreFoundation                      0x00000001803e743c _CF_forwarding_prep_0 + 92
5   App                                 0x0000000104486ef0 0x104480000 + 72
6   AFNetworking                        0x00000001049393f4 0x104934000 + 132
7   AFNetworking                        0x000000010494aeb0 0x104934000 + 148
8   libdispatch.dylib                   0x0000000104bb3e94 _dispatch_call_block_and_release + 24
9   libdispatch.dylib                   0x0000000104bb5694 _dispatch_client_callout + 16
10  libdispatch.dylib                   0x0000000104bc432c _dispatch_main_queue_callback_4CF + 1388
11  CoreFoundation                      0x000000018034fcd4 __CFRUNLOOP_IS_SERVICING_THE_MAIN_DISPATCH_QUEUE__ + 12
12  CoreFoundation                      0x000000018034a244 __CFRunLoopRun + 2448
13  CoreFoundation                      0x00000001803493a8 CFRunLoopRunSpecific + 572
14  GraphicsServices                    0x000000018c03c5ec GSEventRunModal + 160
15  UIKitCore                           0x0000000184d937ac -[UIApplication _run] + 992
16  UIKitCore                           0x0000000184d982e8 UIApplicationMain + 112
17  App                                 0x0000000104489940 0x104480000 + 96
18  dyld                                0x0000000104595ca0 start_sim + 20
19  ???                                 0x00000001046850f4 0x0 + 4368912628
20  ???                                 0xb841800000000000 0x0 + 13277033913953288192

Binary Images:
       0x104480000 -        0x1044dffff App arm64 <c5f567045f43313083662447212630b9> /Users/hulilei/Library/Developer/CoreSimulator/Devices/85A2FDD0-C469-4BFD-9628-5CC6F7A3CE13/data/Containers/Bundle/Application/5029205E-C195-480D-B112-626E8D614D67/App.app/App
       0x104934000 -        0x10495ffff AFNetworking arm64 <23b403d6f86e32349a498e63fa24055b> /Users/hulilei/Library/Developer/CoreSimulator/Devices/85A2FDD0-C469-4BFD-9628-5CC6F7A3CE13/data/Containers/Bundle/Application/5029205E-C195-480D-B112-626E8D614D67/App.app/Frameworks/AFNetworking.framework/AFNetworking
       0x180172000 -        0x1801a2fff libobjc.A.dylib arm64 <22330d97ba8130f1ba94b9879f9f23bf> /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes/iOS.simruntime/Contents/Resources/RuntimeRoot/usr/lib/libobjc.A.dylib
       0x1802cb000 -        0x180677fff CoreFoundation arm64 <3b05273d3bbf321d8274b848d167fd37> /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes/iOS.simruntime/Contents/Resources/RuntimeRoot/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation
       0x104bb0000 -        0x104bfbfff libdispatch.dylib arm64 <68675cfeb3333633bc16e0f5d5c052c2> /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes/iOS.simruntime/Contents/Resources/RuntimeRoot/usr/lib/system/introspection/libdispatch.dylib
       0x184189000 -        0x185613fff UIKitCore arm64 <08a5560d3f2b3b52a202add2a39b5d30> /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes/iOS.simruntime/Contents/Resources/RuntimeRoot/System/Library/PrivateFrameworks/UIKitCore.framework/UIKitCore
       0x18c039000 -        0x18c041fff GraphicsServices arm64 <01858b6bf55936239aafd03b35fccf3c> /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Library/Developer/CoreSimulator/Profiles/Runtimes/iOS.simruntime/Contents/Resources/RuntimeRoot/System/Library/PrivateFrameworks/GraphicsServices.framework/GraphicsServices

`

	address, err := scanIOSCrashAddress(originCrash)
	if err != nil {
		t.Fatal(err)
	}

	for moduleName, addresses := range address {
		fmt.Println(moduleName)
		for start, crashAddresses := range addresses {
			fmt.Println(start)
			for _, addr := range crashAddresses {
				fmt.Printf("%+v\n", addr)
			}
		}

		fmt.Println()
	}
}

func TestRunAtosTool(t *testing.T) {
	args := []string{
		"-o",
		"/Users/zy/software/source_map/iOS/App.app.dSYM/Contents/Resources/DWARF/App",
		"-l",
		"0x104f30000",
	}

	args = append(args, "0x0000000104fd0728", "0x0000000104fd00bc")

	cmd := exec.Command("/Users/zy/.cargo/bin/atosl", args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("err: ", err)
	}

	fmt.Println(string(output))
}

func TestScanABI(t *testing.T) {
	crash := `
*** *** *** *** *** *** *** *** *** *** *** *** *** *** *** ***

Crash type: 'native'
Device Uptime: '2022-07-08T13:31:33.131+0800'
OS Rooted: 'No'
API level: '31'
Kernel version: 'Linux version 5.10.66-android12-9-00021-g2c152aa32942-ab8087165 #1 SMP PREEMPT Fri Jan 14 17:35:16 UTC 2022 (aarch64)'
ABI list: 'arm64-v8a'
ABI: 'arm64'
Build fingerprint: 'Android/sdk_phone64_arm64/emulator64_arm64:12/SE1A.220203.002.A1/8151367:userdebug/test-keys'

--- --- --- --- --- --- --- --- --- --- --- --- --- --- --- ---

pid: 15733, tid: 15733, name: com.ft  >>> com.ft <<<
signal 5 (SIGTRAP), code 1 (TRAP_BRKPT), fault addr 0x72f0fab7fc
Abort message: 'abort message for ftNative internal testing'
    x0  000000760576f7cc  x1  000000760576f7cc  x2  00000073abbfda90  x3  0000007fee3cb748
    x4  0000000000000038  x5  00676e6974736574  x6  0000000000008000  x7  0000000000000000
    x8  6a1cc8a3100e9d5f  x9  6a1cc8a3100e9d5f  x10 00000073abbdd000  x11 0000000000000050
    x12 0000000000000010  x13 0000000000000001  x14 00000073684de91c  x15 0000000000000038
    x16 00000072f0fb8fb0  x17 00000076056d4b40  x18 000000761d302000  x19 0000000000000000
    x20 000000761c5d4000  x21 00000072f43c16c8  x22 0000000000000000  x23 b4000074dbbd70c0
    x24 00000072f43c16c8  x25 0000000000000001  x26 0000000000000037  x27 000000761c5d4000
    x28 0000007fee3cb850  x29 0000007fee3cb7d0
    sp  0000007fee3cb7d0  lr  00000072f0fab7fc  pc  00000072f0fab7fc

backtrace:
    #00 pc 00000000000057fc  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_4+12)
    #01 pc 00000000000058a4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_3+8)
    #02 pc 00000000000058b4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_2+12)
    #03 pc 00000000000058c4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_1+12)
    #04 pc 0000000000005938  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+112)
    #05 pc 00000000002d9a44  /apex/com.android.art/lib64/libart.so (art_quick_generic_jni_trampoline+148)
    #06 pc 00000000002d03e8  /apex/com.android.art/lib64/libart.so (art_quick_invoke_static_stub+568)
    #07 pc 00000000002f47cc  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter34ArtInterpreterToCompiledCodeBridgeEPNS_6ThreadEPNS_9ArtMethodEPNS_11ShadowFrameEtPNS_6JValueE+320)
    #08 pc 0000000000417a1c  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+820)
    #09 pc 000000000077691c  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+3812)
    #10 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #11 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #12 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #13 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #14 pc 00000000002c6858  /apex/com.android.art/lib64/libart.so (MterpInvokeVirtual+5380)
    #15 pc 00000000002ca894  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_virtual+20)
    #16 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #17 pc 000000000027cf1c  /apex/com.android.art/lib64/libart.so (artQuickToInterpreterBridge+1176)
    #18 pc 00000000002d9b78  /apex/com.android.art/lib64/libart.so (art_quick_to_interpreter_bridge+88)
    #19 pc 0000000000209188  /apex/com.android.art/lib64/libart.so (nterp_helper+152)
    #20 pc 00000000002d0164  /apex/com.android.art/lib64/libart.so (art_quick_invoke_stub+548)
    #21 pc 0000000000364cec  /apex/com.android.art/lib64/libart.so (_ZN3art12InvokeMethodILNS_11PointerSizeE8EEEP8_jobjectRKNS_33ScopedObjectAccessAlreadyRunnableES3_S3_S3_m+744)
    #22 pc 00000000003649dc  /apex/com.android.art/lib64/libart.so (_ZN3artL13Method_invokeEP7_JNIEnvP8_jobjectS3_P13_jobjectArray+52)
    #23 pc 00000000000b2f74  /apex/com.android.art/javalib/arm64/boot.oat (art_jni_trampoline+132)
    #24 pc 00000000002d0164  /apex/com.android.art/lib64/libart.so (art_quick_invoke_stub+548)
    #25 pc 00000000002f47c4  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter34ArtInterpreterToCompiledCodeBridgeEPNS_6ThreadEPNS_9ArtMethodEPNS_11ShadowFrameEtPNS_6JValueE+312)
    #26 pc 0000000000417a1c  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+820)
    #27 pc 00000000002c6858  /apex/com.android.art/lib64/libart.so (MterpInvokeVirtual+5380)
    #28 pc 00000000002ca894  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_virtual+20)
    #29 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #30 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #31 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #32 pc 000000000077691c  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+3812)
    #33 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #34 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #35 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #36 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #37 pc 000000000077691c  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+3812)
    #38 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #39 pc 0000000000776280  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+2120)
    #40 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #41 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #42 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #43 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #44 pc 00000000003e2b30  /apex/com.android.art/lib64/libart.so (MterpInvokeInterface+4912)
    #45 pc 00000000002caa94  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_interface+20)
    #46 pc 00000000002c5c48  /apex/com.android.art/lib64/libart.so (MterpInvokeVirtual+2292)
    #47 pc 00000000002ca894  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_virtual+20)
    #48 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #49 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #50 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #51 pc 0000000000416918  /apex/com.android.art/lib64/libart.so (MterpInvokeDirect+1580)
    #52 pc 00000000002ca994  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_direct+20)
    #53 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #54 pc 00000000003851d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter33ArtInterpreterToInterpreterBridgeEPNS_6ThreadERKNS_20CodeItemDataAccessorEPNS_11ShadowFrameEPNS_6JValueE+148)
    #55 pc 0000000000417c94  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+1452)
    #56 pc 000000000077691c  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+3812)
    #57 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #58 pc 00000000003e219c  /apex/com.android.art/lib64/libart.so (MterpInvokeInterface+2460)
    #59 pc 00000000002caa94  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_interface+20)
    #60 pc 0000000000776280  /apex/com.android.art/lib64/libart.so (MterpInvokeStatic+2120)
    #61 pc 00000000002caa14  /apex/com.android.art/lib64/libart.so (mterp_op_invoke_static+20)
    #62 pc 000000000027dd74  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreterL7ExecuteEPNS_6ThreadERKNS_20CodeItemDataAccessorERNS_11ShadowFrameENS_6JValueEbb.llvm.6649268296134209133+644)
    #63 pc 000000000027cf1c  /apex/com.android.art/lib64/libart.so (artQuickToInterpreterBridge+1176)
    #64 pc 00000000002d9b78  /apex/com.android.art/lib64/libart.so (art_quick_to_interpreter_bridge+88)
    #65 pc 00000000000057c4  /memfd:jit-cache (deleted)

build id:
    /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (BuildId: 763f64f5b706242d70cd233baaf0d75ed98a9c4f. FileSize: 76152. LastModified: 1981-01-01T01:01:02.000+0800. MD5: ab5776cc5beb400df1be1b816a922f8f)
    /apex/com.android.art/lib64/libart.so (BuildId: 428255a2676118ad9ec8a72f513b70db. FileSize: 11004168. LastModified: 1970-01-01T08:00:00.000+0800. MD5: bfd2f6c969c579ada2414591dd48349e)
    /apex/com.android.art/javalib/arm64/boot.oat (BuildId: 715d0a044ea13bcf499cd6094001f85c3246944e. FileSize: 4379936. LastModified: 1970-01-01T08:00:00.000+0800)
    /memfd:jit-cache (deleted) (BuildId: unknown. OPEN error: errno = 2, errmsg = No such file or directory)

stack:
         0000007fee3cb750  0000007fee3cb790  [stack]
         0000007fee3cb758  00000072f0faf2f4  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so
         0000007fee3cb760  0000000000000000
         0000007fee3cb768  00000072f43c16c8  [anon:dalvik-LinearAlloc]
         0000007fee3cb770  0000000000000000
         0000007fee3cb778  0000007fee3cb7a0  [stack]
         0000007fee3cb780  000000761c5d4000  [anon:stack_and_tls:main]
         0000007fee3cb788  000000761c5d4000  [anon:stack_and_tls:main]
         0000007fee3cb790  0000007fee3cb7c0  [stack]
         0000007fee3cb798  00000072f0fab87c  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so
         0000007fee3cb7a0  0000000000000000
         0000007fee3cb7a8  6a1cc8a3100e9d5f
         0000007fee3cb7b0  0000000000000000
         0000007fee3cb7b8  00000072f0fab900  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+56)
         0000007fee3cb7c0  0000007fee3cb7d0  [stack]
         0000007fee3cb7c8  00000072f0fab7fc  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_4+12)
    #00  0000007fee3cb7d0  0000007fee3cb7e0  [stack]
         0000007fee3cb7d8  00000072f0fab8a8  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_2)
    #01  0000007fee3cb7e0  0000007fee3cb7f0  [stack]
         0000007fee3cb7e8  00000072f0fab8b8  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_call_1)
    #02  0000007fee3cb7f0  0000007fee3cb800  [stack]
         0000007fee3cb7f8  00000072f0fab8c8  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash)
    #03  0000007fee3cb800  0000007fee3cb830  [stack]
         0000007fee3cb808  00000072f0fab93c  /data/app/~~Taci3mQyw7W7iWT7Jxo-ag==/com.ft-Q8m2flQFG1MbGImPiuAZmQ==/lib/arm64/libft_native_exp_lib.so (xc_test_crash+116)
    #04  0000007fee3cb810  0000007353e9ecb0  [anon:stack_and_tls:15847]
         0000007fee3cb818  6a1cc8a3100e9d5f
         0000007fee3cb820  0000000000000000
         0000007fee3cb828  b4000074dbbd7010
         0000007fee3cb830  0000007fee3cb850  [stack]
         0000007fee3cb838  00000073686d9a48  /apex/com.android.art/lib64/libart.so (art_quick_generic_jni_trampoline+152)
    #05  0000007fee3cb840  000000009d5ff7c3
         0000007fee3cb848  0000000700000000
         0000007fee3cb850  00000072f43c16c8  [anon:dalvik-LinearAlloc]
         0000007fee3cb858  0000000000000001
         0000007fee3cb860  0000000000000000
         0000007fee3cb868  0000000000000050
         0000007fee3cb870  0000000000000014
         0000007fee3cb878  000000000000001c
         0000007fee3cb880  fffffffffffffff4
         0000007fee3cb888  0000000200000000
         0000007fee3cb890  5410541504504140
         0000007fee3cb898  0000000044c62000
         0000007fee3cb8a0  0000000000000000
         0000007fee3cb8a8  0000000000000000
         0000007fee3cb8b0  b4000074dbbd7010
         0000007fee3cb8b8  0000007fee3cbca0  [stack]
         ........  ........
    #06  0000007fee3cb930  0000000000000000
         0000007fee3cb938  0000007200000000
         0000007fee3cb940  0000007fee3cbca0  [stack]
         0000007fee3cb948  00000072f65f848b  /data/data/com.ft/code_cache/.overlay/base.apk/classes.dex
         0000007fee3cb950  0000007fee3cbca0  [stack]
         0000007fee3cb958  b4000074dbbd7010
         0000007fee3cb960  0000007fee3cb9d0  [stack]
         0000007fee3cb968  00000073686f47d0  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter34ArtInterpreterToCompiledCodeBridgeEPNS_6ThreadEPNS_9ArtMethodEPNS_11ShadowFrameEtPNS_6JValueE+324)
    #07  0000007fee3cb970  fffffffffffffffe
         0000007fee3cb978  0000000000000001
         0000007fee3cb980  0000007fee3cbf40  [stack]
         0000007fee3cb988  0000007fee3cbda0  [stack]
         0000007fee3cb990  00000000133dcf80  [anon:dalvik-main space (region space)]
         0000007fee3cb998  000000006f495a40  [anon:dalvik-/apex/com.android.art/javalib/boot.art]
         0000007fee3cb9a0  0000000000000000
         0000007fee3cb9a8  0000007fee3cc1e0  [stack]
         0000007fee3cb9b0  0000007fee3cbdf0  [stack]
         0000007fee3cb9b8  0000007fee3cc008  [stack]
         0000007fee3cb9c0  0000007fee3cb9f0  [stack]
         0000007fee3cb9c8  6a1cc8a3100e9d5f
         0000007fee3cb9d0  0000007fee3cbb00  [stack]
         0000007fee3cb9d8  0000007368817a20  /apex/com.android.art/lib64/libart.so (_ZN3art11interpreter6DoCallILb0ELb0EEEbPNS_9ArtMethodEPNS_6ThreadERNS_11ShadowFrameEPKNS_11InstructionEtPNS_6JValueE+824)
         0000007fee3cb9e0  0000007fee3cbdf0  [stack]
         0000007fee3cb9e8  0000000000000000
         ........  ........
    #08  0000007fee3cba30  0000007fee3cbdf0  [stack]
         0000007fee3cba38  00000072f43c16c8  [anon:dalvik-LinearAlloc]
         0000007fee3cba40  0000000000000000
         0000007fee3cba48  0000000000000000
         0000007fee3cba50  0000000000000000
         0000007fee3cba58  0000000000000000
         0000007fee3cba60  0000000000000001
         0000007fee3cba68  0000000000000000
         0000007fee3cba70  0000000000000000
         0000007fee3cba78  00000072f43c1728  [anon:dalvik-LinearAlloc]
         0000007fee3cba80  0000007fee3cbbb0  [stack]
         0000007fee3cba88  00000002ee3cc2b0
         0000007fee3cba90  0000007fee3cba80  [stack]
         0000007fee3cba98  0000007fee3cbca0  [stack]
         0000007fee3cbaa0  b4000074dbbd7010
         0000007fee3cbaa8  000000761c5d4000  [anon:stack_and_tls:main]
         ........  ........
    #09  0000007fee3cbb60  b40000744bbdd950
         0000007fee3cbb68  00000072f62ef4a0  /data/data/com.ft/code_cache/.overlay/base.apk/classes.dex
         0000007fee3cbb70  000000761c5d4000  [anon:stack_and_tls:main]
         0000007fee3cbb78  000010711c5d4000
         0000007fee3cbb80  00000000133dca50  [anon:dalvik-main space (region space)]
         0000007fee3cbb88  0000007fee3cbbc4  [stack]
         0000007fee3cbb90  b40000744bbe1c60
         0000007fee3cbb98  0000000000008336
         0000007fee3cbba0  0000000000000000
         0000007fee3cbba8  133dcf8000000002
`

	abi := scanABI(crash)

	fmt.Println(abi)

	if abi != "arm64-v8a" {
		t.Errorf("unexpected ABI")
	}
}
