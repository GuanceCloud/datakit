package wechatminiprogram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"time"
)

type WxClient struct {
	Appid    string `toml:"appid"`
	Secret   string `toml:"secret"`
	RunTime string `toml:"runtime"`
}

func (wx *WxClient) SampleConfig() string {
	return sampleCfg
}

func (wx *WxClient) Catalog() string {
	return "wechat"
}

func (wx *WxClient) run() {
	token := wx.GetAccessToken()
	wx.GetDailySummary(token)
	wx.GetVisitDistribution(token)
	wx.GetUserPortrait(token)
	wx.GetDailyVisitTrend(token)
	wx.GetVisitPage(token)
}

func (wx *WxClient) Run() {

	l = logger.SLogger("wechat")
	l.Info("wechat input started...")
	interval, err := time.ParseDuration("24h")
	if err != nil {
		l.Error(err)
	}
	sleepTime := wx.formatRuntime()
	time.Sleep(time.Duration(sleepTime)*time.Second)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		wx.run()
		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}

}

type requestQueries map[string]string

func (wx *WxClient) GetDailySummary(accessToken string) {
	body, timeObj := wx.API(accessToken, DailySummaryURL)
	tags := map[string]string{}
	fields := map[string]interface{}{}
	for _, d := range gjson.Get(string(body), "list").Array() {
		fields["visit_total"] = d.Get("visit_total").Int()
		fields["share_pv"] = d.Get("share_pv").Int()
		fields["share_uv"] = d.Get("share_uv").Int()
	}
	wx.writeMetric("DailySummary", tags, fields, timeObj)
}

func (wx *WxClient) GetDailyVisitTrend(accessToken string) {
	body, timeObj := wx.API(accessToken, DailyVisitTrendURL)
	tags := map[string]string{}
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

func (wx *WxClient) GetVisitDistribution(accessToken string) () {
	body, timeObj := wx.API(accessToken, VisitDistributionURL)
	tags := map[string]string{}
	fields := map[string]interface{}{}
	for _, v := range gjson.Get(string(body), "list").Array() {
		index := v.Get("index").String()
		for _, item := range v.Get("item_list").Array() {
			if itemValue, ok := Config[index]; ok {
				tags[index] = itemValue[item.Get("key").String()]
				fields["visit_pv"] = item.Get("value").Int()
				wx.writeMetric("VisitDistribution", tags, fields, timeObj)
			}
		}
	}
}

func (wx *WxClient) GetUserPortrait(accessToken string) {
	body, timeObj := wx.API(accessToken, UserPortraitURL)
	wx.FormatUserPortraitData(string(body), "visit_uv_new", timeObj)
	wx.FormatUserPortraitData(string(body), "visit_uv", timeObj)
}

func (wx *WxClient) GetVisitPage(accessToken string) () {
	body, timeObj := wx.API(accessToken, VisitPageURL)
	tags := map[string]string{}
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
		l.Errorf("failed to make metric, err: %s,metric: %s, tags: %s ,fields: %s", err.Error(),metric,tags,fields)
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
	resp, err := http.Get(url)
	if err != nil {
		l.Errorf("get token error %s",err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("get token read err:%s",err)
	}

	defer resp.Body.Close()
	code := gjson.Get(string(body), "errcode").Int()
	if code >= 40000 {
		fmt.Printf("error:%s",string(body))
		l.Errorf("config error: %s", gjson.Get(string(body), "errmsg").String())
		return wx.GetAccessToken()
	}
	return gjson.Get(string(body), "access_token").String()
}

func (wx *WxClient) FormatUserPortraitData(body string, dataType string, timeObj time.Time) () {
	for k, v := range gjson.Get(body, dataType).Map() {
		for _, value := range v.Array() {
			tags := map[string]string{}
			fields := map[string]interface{}{}
			tags[k] = value.Get("name").String()
			fields[fmt.Sprintf("%s_by_%s", dataType, k)] = value.Get("value").Int()
			wx.writeMetric("UserPortrait", tags, fields, timeObj)
		}
	}
}




func (wx *WxClient) API(accessToken string, apiUrl string) ([]byte, time.Time) {
	queries := requestQueries{
		"access_token": accessToken,
	}
	url := wx.GetUrl(apiUrl, queries)
	var cstZone = time.FixedZone("CST", 8*3600)
	d, _ := time.ParseDuration("-24h")
	date := time.Now().Add(d).In(cstZone).Format("20060102")
	timeObj, err := time.ParseInLocation("20060102 15:04:05", date + " 23:59:59", cstZone)
	if err != nil {
		l.Errorf("API time format err: %s",err)
	}
	bodyMap := map[string]string{
		"begin_date": date,
		"end_date":   date,
	}
	requestBody, err := json.Marshal(bodyMap)
	if err != nil {
		l.Errorf("wechat request body to json err: %s",err)
	}

	resp, err := http.Post(url, "application/json; encoding=utf-8", bytes.NewBuffer(requestBody))
	if err != nil {
		l.Errorf("wechat http send url:%s body:%s err: %s",url,err,requestBody)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("failed to io Feed, err: %s", err.Error())

	}
	defer resp.Body.Close()
	code := gjson.Get(string(body), "errcode").Int()
	if code >= 40000 {
		l.Errorf("api error: %s %s", apiUrl, gjson.Get(string(body), "errmsg").String())
	}
	return body, timeObj
}

func (wx *WxClient)formatRuntime() int64{
	var cstZone = time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstZone)
	format := fmt.Sprintf("%s %s:00",now.Format("20060102"),wx.RunTime)
	runTime,err := time.ParseInLocation("20060102 15:04:05",format,cstZone)
	if err != nil {
		return wx.formatRuntime()
	}
	var sleepTime float64
	if now.Unix() > runTime.Unix() {
		sleepTime = 24 * 3600 - now.Sub(runTime).Seconds()

	}else {
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
