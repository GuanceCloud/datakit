package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (dw *DataWayCfg) GetLogFilter() ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.endPoints[0].getLogFilter()
}

func (dw *DataWayCfg) GetPipelinePull(ts int64) (*PullPipelineReturn, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.endPoints[0].getPipelinePull(ts)
}

func (dc *endPoint) getLogFilter() ([]byte, error) {
	url, ok := dc.categoryURL[datakit.LogFilter]
	if !ok {
		return nil, fmt.Errorf("LogFilter API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getLogFilter failed with status code %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type HTTPError struct {
	ErrCode  string `json:"error_code,omitempty"`
	Err      error  `json:"-"`
	HTTPCode int    `json:"-"`
}

type bodyResp struct {
	*HTTPError
	Message string              `json:"message,omitempty"`
	Content *PullPipelineReturn `json:"content,omitempty"` // 注意与 kodo 中的不一样
}

type PipelineUnit struct {
	Name       string `json:"name"`
	Base64Text string `json:"base64text"`
}

type PullPipelineReturn struct {
	UpdateTime int64           `json:"update_time"`
	Pipelines  []*PipelineUnit `json:"pipelines"`
}

func (dc *endPoint) getPipelinePull(ts int64) (*PullPipelineReturn, error) {
	url, ok := dc.categoryURL[datakit.PipelinePull]
	if !ok {
		return nil, fmt.Errorf("PipelinePull API missing, should not been here")
	}

	url += "&ts=" + fmt.Sprintf("%d", ts)

	log.Debugf("PipelinePull GET: %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getPipelinePull failed with status code %d, body: %s", resp.StatusCode, string(body))
	}

	var br bodyResp
	err = json.Unmarshal(body, &br)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return br.Content, err
}

func (dw *DataWayCfg) WorkspaceQuery(body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.Workspace]
	if !ok {
		return nil, fmt.Errorf("no workspace query URL available")
	}

	log.Debugf("NewRequest: %s", requrl)
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return dw.sendReq(req)
}

func (dw *DataWayCfg) DQLQuery(body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.QueryRaw]
	if !ok {
		return nil, fmt.Errorf("no DQL query URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return dw.sendReq(req)
}

func (dw *DataWayCfg) Election(namespace, id string) ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	log.Debugf("election sending %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		log.Debugf("election %s ok", requrl)
		return body, nil
	default:
		log.Debugf("election failed: %d", resp.StatusCode)
		return nil, fmt.Errorf("election failed: %s", string(body))
	}
}

func (dw *DataWayCfg) ElectionHeartbeat(namespace, id string) ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0] // 选举相关接口只只发送给第一个 dataway

	requrl, ok := dc.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	log.Debugf("election sending heartbeat %s", requrl)
	resp, err := dw.httpCli.Post(requrl, "", nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		return body, nil
	default:
		return nil, fmt.Errorf("election heartbeat failed: %s", string(body))
	}
}

func (dc *endPoint) heartBeat(data []byte) (int, error) {
	requrl, ok := dc.categoryURL[datakit.HeartBeat]
	if !ok {
		return heartBeatIntervalDefault, fmt.Errorf("HeartBeat API missing, should not been here")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))
	if err != nil {
		return heartBeatIntervalDefault, err
	}

	if dc.ontest {
		return heartBeatIntervalDefault, nil
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		return heartBeatIntervalDefault, err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("heart beat resp err: %+#v", resp)
		return heartBeatIntervalDefault, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return heartBeatIntervalDefault, err
	}
	type heartBeat struct {
		Content struct {
			Interval int `json:"interval"`
		} `json:"content"`
	}

	var hb heartBeat
	if err := json.Unmarshal(body, &hb); err != nil {
		log.Errorf(`%s, body: %s`, err, string(body))
		return heartBeatIntervalDefault, err
	}
	return hb.Content.Interval, nil
}

func (dw *DataWayCfg) DatawayList() ([]string, int, error) {
	if len(dw.endPoints) == 0 {
		return nil, datawayListIntervalDefault, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ListDataWay]
	if !ok {
		return nil, datawayListIntervalDefault, fmt.Errorf("dataway list API not available")
	}

	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	resp, err := dw.sendReq(req)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	type dataways struct {
		Content struct {
			DatawayList []string `json:"dataway_list"`
			Interval    int      `json:"interval"`
		} `json:"content"`
	}

	var dws dataways
	if err := json.Unmarshal(body, &dws); err != nil {
		log.Errorf(`%s, body: %s`, err, string(body))
		return nil, datawayListIntervalDefault, err
	}

	log.Debugf(`available dataways; %+#v,body: %s`, dws.Content, string(body))
	return dws.Content.DatawayList, dws.Content.Interval, nil
}

func (dw *DataWayCfg) HeartBeat() (int, error) {
	body := map[string]interface{}{
		"dk_uuid":   dw.Hostname, // 暂用 hostname 代之, 后将弃用该字段
		"heartbeat": time.Now().Unix(),
		"host":      dw.Hostname,
	}

	if dw.httpCli == nil {
		if err := dw.initHTTP(); err != nil {
			return heartBeatIntervalDefault, err
		}
	}

	bodyByte, err := json.Marshal(body)
	if err != nil {
		err := fmt.Errorf("[error] heartbeat json marshal err: %w", err)
		return heartBeatIntervalDefault, err
	}
	var interval int
	for _, dc := range dw.endPoints {
		interval, err = dc.heartBeat(bodyByte)
		if err != nil {
			return heartBeatIntervalDefault, err
		}
	}

	return interval, nil
}

// UpsertObjectLabels , dw api create or update object labels.
func (dw *DataWayCfg) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return dw.sendReq(req)
}

// DeleteObjectLabels , dw api delete object labels.
func (dw *DataWayCfg) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	requrl, ok := dc.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	rBody := bytes.NewReader(body)
	req, err := http.NewRequest("DELETE", requrl, rBody)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return dw.sendReq(req)
}

func (dw *DataWayCfg) UploadLog(r io.Reader, hostName string) (*http.Response, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.endPoints[0]
	reqURL, ok := dc.categoryURL[datakit.LogUpload]
	if !ok {
		return nil, fmt.Errorf("no file upload URL available")
	}

	req, err := http.NewRequest("POST", reqURL, r)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	req.Header.Add("Host-Name", hostName)
	return dw.sendReq(req)
}
