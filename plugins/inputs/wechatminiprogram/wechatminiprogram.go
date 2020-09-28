package wechatminiprogram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type WxClient struct {
	Appid     string            `toml:"appid"`
	Secret    string            `toml:"secret"`
	Tags      map[string]string `toml:"tags,omitempty"`
	Analysis  *Analysis         `toml:"analysis,omitempty"`
	Operation *Operation        `toml:"operation,omitempty"`
	RunTime string   `toml:"runtime,omitempty"`

	ctx       context.Context
	cancelFun context.CancelFunc
	wg sync.WaitGroup
	subModules []subModule

}
type subModule interface {
	run(wx *WxClient)
}

type Analysis struct {
	Name    []string `toml:"name,omitempty"`
}

type Operation struct {
	Name []string `toml:"name,omitempty"`
}




func (wx *WxClient) SampleConfig() string {
	return sampleCfg
}

func (wx *WxClient) Catalog() string {
	return "wechat"
}

func (an *Analysis) run(wx *WxClient) {
	interval, err := time.ParseDuration("24h")
	if err != nil {
		l.Error(err)
	}
	if wx.RunTime != "" {
		sleepTime := wx.formatRuntime()
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	for {
		token := wx.GetAccessToken()
		wxClient := reflect.ValueOf(wx)
		params := []reflect.Value{
			reflect.ValueOf(token),
		}
		for _,c :=range wx.Analysis.Name {
			if wxClient.MethodByName(c).IsValid() {
				l.Info(c)
				wxClient.MethodByName(c).Call(params)
			}
		}
		select {
		case <-wx.ctx.Done():
			return
		default:
		}
		datakit.SleepContext(wx.ctx,interval)
	}
}

func (op *Operation) run(wx *WxClient) {
	l.Info("operation run")
	interval, err := time.ParseDuration("24h")
	if err != nil {
		l.Error(err)
	}
	for {
		token := wx.GetAccessToken()
		wxClient := reflect.ValueOf(wx)

		for _, c := range wx.Operation.Name {
			params := []reflect.Value{
				reflect.ValueOf(token),
			}
			if c == "JsErrSearch" {
				var pageNum ,pageSize int64 = 1 ,100
				params = append(params, reflect.ValueOf(pageNum))
				params = append(params, reflect.ValueOf(pageSize))
			}
			if wxClient.MethodByName(c).IsValid() {
				l.Info(c)
				wxClient.MethodByName(c).Call(params)
			}
		}
		select {
		case <-wx.ctx.Done():
			return
		default:
		}
		datakit.SleepContext(wx.ctx,interval)
	}
}

func (wx *WxClient) addModule(m subModule) {
	wx.subModules = append(wx.subModules, m)
}



func (wx *WxClient) Run() {
	wx.ctx, wx.cancelFun = context.WithCancel(context.Background())

	l = logger.SLogger("wechat")
	l.Info("wechat input started...")
	if wx.Analysis != nil {
		wx.addModule(wx.Analysis)

	}
	if wx.Operation != nil {
		wx.addModule(wx.Operation)

	}

	for _, s := range wx.subModules {
		wx.wg.Add(1)
		go func(s subModule) {
			defer wx.wg.Done()
			s.run(wx)
		}(s)
	}

	wx.wg.Wait()

	l.Debugf("done")

}

type requestQueries map[string]string

func (wx *WxClient) DailySummary(accessToken string) {
	body, timeObj := wx.API(accessToken, DailySummaryURL)
	tags := map[string]string{}
	tags["appid"] = wx.Appid
	fields := map[string]interface{}{}
	for _, d := range gjson.Get(string(body), "list").Array() {
		fields["visit_total"] = d.Get("visit_total").Int()
		fields["share_pv"] = d.Get("share_pv").Int()
		fields["share_uv"] = d.Get("share_uv").Int()
	}
	wx.writeMetric("DailySummary", tags, fields, timeObj)
}

func (wx *WxClient) DailyVisitTrend(accessToken string) {
	body, timeObj := wx.API(accessToken, DailyVisitTrendURL)
	tags := map[string]string{}
	tags["appid"] = wx.Appid
	fields := map[string]interface{}{}
	for _, d := range gjson.Get(string(body), "list").Array() {
		fields["session_cnt"] = d.Get("session_cnt").Int()
		fields["visit_pv"] = d.Get("visit_pv").Int()
		fields["visit_uv"] = d.Get("visit_uv").Int()
		fields["visit_uv_new"] = d.Get("visit_uv_new").Int()
		fields["stay_time_uv"] = d.Get("stay_time_uv").Float()
		fields["stay_time_session"] = d.Get("stay_time_session").Float()
		fields["visit_depth"] = d.Get("visit_depth").Float()
	}
	wx.writeMetric("DailyVisitTrend", tags, fields, timeObj)

}

func (wx *WxClient) VisitDistribution(accessToken string) () {
	body, timeObj := wx.API(accessToken, VisitDistributionURL)
	tags := map[string]string{}
	fields := map[string]interface{}{}
	tags["appid"] = wx.Appid
	for _, v := range gjson.Get(string(body), "list").Array() {
		index := v.Get("index").String()
		for _, item := range v.Get("item_list").Array() {
			if itemValue, ok := Config[index]; ok {
				tags["index_key"] = index
				tags["index_value"] = itemValue[item.Get("key").String()]
				fields["visit_pv"] = item.Get("value").Int()
				wx.writeMetric("VisitDistribution", tags, fields, timeObj)
			}
		}
	}
}

func (wx *WxClient) UserPortrait(accessToken string) {
	body, timeObj := wx.API(accessToken, UserPortraitURL)
	wx.FormatUserPortraitData(string(body), "visit_uv_new", timeObj)
	wx.FormatUserPortraitData(string(body), "visit_uv", timeObj)
}

func (wx *WxClient) VisitPage(accessToken string) () {
	body, timeObj := wx.API(accessToken, VisitPageURL)
	tags := map[string]string{}
	tags["appid"] = wx.Appid
	fields := map[string]interface{}{}
	for _, v := range gjson.Get(string(body), "list").Array() {
		tags["page_path"] = v.Get("page_path").String()
		fields["page_visit_pv"] = v.Get("page_visit_pv").Int()
		fields["page_visit_uv"] = v.Get("page_visit_uv").Int()
		fields["page_staytime_pv"] = v.Get("page_staytime_pv").Float()
		fields["entrypage_pv"] = v.Get("entrypage_pv").Int()
		fields["exitpage_pv"] = v.Get("exitpage_pv").Int()
		fields["page_share_pv"] = v.Get("page_share_pv").Int()
		fields["page_share_uv"] = v.Get("page_share_uv").Int()

	}
	wx.writeMetric("VisitPage", tags, fields, timeObj)

}

func (wx *WxClient) GetUrl(api string, queries requestQueries) (url string) {
	url = EncodingUrl(baseURL+api, queries)
	return url
}

func (wx *WxClient) writeMetric(metric string, tags map[string]string, fields map[string]interface{}, timeObj time.Time) {
	if len(fields) == 0 {
		return
	}
	data, err := io.MakeMetric(metric, tags, fields, timeObj)

	if err != nil {
		l.Errorf("failed to make metric, err: %s,metric: %s, tags: %s ,fields: %s", err.Error(), metric, tags, fields)
		return
	}

	if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
		l.Errorf("failed to io Feed, err: %s", err.Error())
		return
	}

}

func (wx *WxClient) GetAccessToken() (token string) {

	queries := requestQueries{
		"appid":      wx.Appid,
		"secret":     wx.Secret,
		"grant_type": grantType,
	}
	url := wx.GetUrl(accessTokenURL, queries)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		l.Errorf("get token error %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("get token read err:%s", err)
	}

	defer resp.Body.Close()
	code := gjson.Get(string(body), "errcode").Int()
	if code >= errcode {
		fmt.Printf("error:%s", string(body))
		l.Errorf("config error: %s", gjson.Get(string(body), "errmsg").String())
		return wx.GetAccessToken()
	}
	return gjson.Get(string(body), "access_token").String()
}

func (wx *WxClient) FormatUserPortraitData(body, dataType string, timeObj time.Time) () {
	tags := map[string]string{}
	tags["appid"] = wx.Appid
	fields := map[string]interface{}{}
	for k, v := range gjson.Get(body, dataType).Map() {
		for _, value := range v.Array() {
			tags["index_key"] = k
			tags["index_value"] = value.Get("name").String()
			fields[dataType] = value.Get("value").Int()
			wx.writeMetric("UserPortrait", tags, fields, timeObj)
		}
	}
}

func (wx *WxClient) JsErrSearch(accessToken string, startPage, limit int64) () {
	queries := requestQueries{
		"access_token": accessToken,
	}
	url := wx.GetUrl(jsErrSearchURL, queries)
	var cstZone = time.FixedZone("CST", offect)
	d, _ := time.ParseDuration("-24h")
	date := time.Now().Add(d).In(cstZone).Format("20060102")
	start, _ := time.ParseInLocation("20060102 15:04:05", date+" 23:59:59", cstZone)
	end, _ := time.ParseInLocation("20060102 15:04:05", date+" 00:00:00", cstZone)
	bodyMap := map[string]interface{}{
		"start_time": start.Unix(),
		"end_time":   end.Unix(),
		"start":      startPage,
		"limit":      limit,
	}
	requestBody, err := json.Marshal(bodyMap)
	if err != nil {
		l.Errorf("wechat request body to json err: %s", err)
	}

	resp, err := http.Post(url, "application/json; encoding=utf-8", bytes.NewBuffer(requestBody)) //nolint:gosec
	if err != nil {
		l.Errorf("wechat http send url:%s body:%s err: %s", url, err, requestBody)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("failed to io Feed, err: %s", err.Error())

	}
	defer resp.Body.Close()

	if gjson.Get(string(body), "errcode").Int() != 0 {
		l.Errorf("jsErrSearch http error : %s", string(body))
		return
	}

	for _, result := range gjson.Get(string(body), "results").Array() {
		tags := map[string]string{
			"client_version": result.Get("client_version").String(),
			"app_version":    result.Get("app_version").String(),
			"appid":          wx.Appid,
		}
		fields := map[string]interface{}{
			"version_error_cnt": result.Get("version_error_cnt").Int(),
			"total_error_cnt":   result.Get("total_error_cnt").Int(),
			"__content":         result.Get("errmsg").String(),
		}
		timeStamp := result.Get("time").Int()
		data, err := io.MakeMetric("miniProgramJsErr", tags, fields, time.Unix(timeStamp, 0))
		if err != nil {
			l.Errorf("failed to make metric, err: %s,metric: %s, tags: %s ,fields: %s", err.Error(), "miniProgramJsErr", tags, fields)
			return
		}

		if err := io.NamedFeed(data, io.Logging, inputName); err != nil {
			l.Errorf("failed to io Feed, err: %s", err.Error())
			return
		}
	}

	if startPage*limit <= gjson.Get(string(body), "total").Int() {
		startPage++
		wx.JsErrSearch(accessToken, startPage, limit)
	}
}

func (wx *WxClient) Performance(accessToken string) () {
	queries := requestQueries{
		"access_token": accessToken,
	}
	url := wx.GetUrl(PerformanceURL, queries)
	var cstZone = time.FixedZone("CST", offect)
	d, _ := time.ParseDuration("-24h")
	date := time.Now().Add(d).In(cstZone).Format("20060102")
	end, _ := time.ParseInLocation("20060102 15:04:05", date+" 23:59:59", cstZone)
	start, _ := time.ParseInLocation("20060102 15:04:05", date+" 00:00:00", cstZone)
	for k, v := range Config["cost_time_type"] {
		key, _ := strconv.Atoi(k)
		bodyMap := map[string]interface{}{
			"default_start_time": start.Unix(),
			"default_end_time":   end.Unix(),
			"cost_time_type":     key,
		}
		requestBody, err := json.Marshal(bodyMap)
		if err != nil {
			l.Errorf("wechat request body to json err: %s", err)
		}

		resp, err := http.Post(url, "application/json; encoding=utf-8", bytes.NewBuffer(requestBody)) //nolint:gosec
		if err != nil {
			l.Errorf("wechat http send url:%s body:%s err: %s", url, err, requestBody)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Errorf("failed to io Feed, err: %s", err.Error())

		}
		defer resp.Body.Close()
		tags := map[string]string{
			"cost_time_type": v,
			"appid":          wx.Appid,
		}

		data := gjson.Get(string(body), "default_time_data").String()
		for _, v := range gjson.Get(data, "list").Array() {
			fields := map[string]interface{}{
				"cost_time": v.Get("cost_time").Int(),
			}
			line, err := io.MakeMetric("Performance", tags, fields, end)
			if err != nil {
				l.Errorf("failed to make metric, err: %s,metric: %s, tags: %s ,fields: %s", err.Error(), "Performance", tags, fields)
				return
			}

			if err := io.NamedFeed(line, io.Metric, inputName); err != nil {
				l.Errorf("failed to io Feed, err: %s", err.Error())
				return
			}
		}
	}
}

func (wx *WxClient) API(accessToken, apiUrl string) ([]byte, time.Time) {
	queries := requestQueries{
		"access_token": accessToken,
	}
	url := wx.GetUrl(apiUrl, queries) //nolint:gosec
	var cstZone = time.FixedZone("CST", offect)
	d, _ := time.ParseDuration("-24h")
	date := time.Now().Add(d).In(cstZone).Format("20060102")
	timeObj, err := time.ParseInLocation("20060102 15:04:05", date+" 23:59:59", cstZone)
	if err != nil {
		l.Errorf("API time format err: %s", err)
	}
	bodyMap := map[string]string{
		"begin_date": date,
		"end_date":   date,
	}
	requestBody, err := json.Marshal(bodyMap)
	if err != nil {
		l.Errorf("wechat request body to json err: %s", err)
	}

	resp, err := http.Post(url, "application/json; encoding=utf-8", bytes.NewBuffer(requestBody)) //nolint:gosec
	if err != nil {
		l.Errorf("wechat http send url:%s body:%s err: %s", url, err, requestBody)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("failed to io Feed, err: %s", err.Error())

	}
	defer resp.Body.Close()
	code := gjson.Get(string(body), "errcode").Int()
	if code >= errcode {
		l.Errorf("api error: %s %s", apiUrl, gjson.Get(string(body), "errmsg").String())
	}
	return body, timeObj
}

func (wx *WxClient) formatRuntime() int64 {

	var cstZone = time.FixedZone("CST", offect)
	now := time.Now().In(cstZone)

	format := fmt.Sprintf("%s %s:00", now.Format("20060102"), wx.RunTime)
	runTime, err := time.ParseInLocation("20060102 15:04:05", format, cstZone)
	if err != nil {
		return wx.formatRuntime()
	}
	var sleepTime float64
	if now.Unix() > runTime.Unix() {
		sleepTime = daySecond - now.Sub(runTime).Seconds()
	} else {
		sleepTime = runTime.Sub(now).Seconds()
	}
	return int64(sleepTime)
}

func EncodingUrl(api string, params requestQueries) string {
	url, err := url2.Parse(api)
	if err != nil {
		fmt.Println(err)
	}
	query := url.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	url.RawQuery = query.Encode()

	return url.String()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &WxClient{}
	})
}
