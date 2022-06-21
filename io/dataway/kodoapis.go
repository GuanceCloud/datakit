// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func (dw *DataWayDefault) GetLogFilter() ([]byte, error) {
	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.endPoints[0].getLogFilter()
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

func (dc *endPoint) datakitPull(args string) ([]byte, error) {
	url, ok := dc.categoryURL[datakit.DatakitPull]
	if !ok {
		return nil, fmt.Errorf("datakit pull API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url+"&"+args, nil)
	if err != nil {
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("datakitPull failed with status code %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (dw *DataWayDefault) WorkspaceQuery(body []byte) (*http.Response, error) {
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

func (dw *DataWayDefault) DQLQuery(body []byte) (*http.Response, error) {
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

func (dw *DataWayDefault) Election(namespace, id string) ([]byte, error) {
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

	req, err := http.NewRequest("POST", requrl, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
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

func (dw *DataWayDefault) ElectionHeartbeat(namespace, id string) ([]byte, error) {
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

	req, err := http.NewRequest("POST", requrl, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	resp, err := dc.dw.sendReq(req)
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

func (dw *DataWayDefault) DatawayList() ([]string, int, error) {
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

func (dw *DataWayDefault) HeartBeat() (int, error) {
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
func (dw *DataWayDefault) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
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
func (dw *DataWayDefault) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
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

func (dw *DataWayDefault) ProfileProxyURL() (*url.URL, error) {
	lastErr := fmt.Errorf("no dataway profile proxy endpoint available")

	if len(dw.endPoints) == 0 {
		return nil, lastErr
	}

	// copy endpoints
	endPoints := append([]*endPoint(nil), dw.endPoints...)

	for len(endPoints) > 0 {
		idx := big.NewInt(0)
		if len(dw.endPoints) > 1 {
			var err error
			idx.SetInt64(int64(len(dw.endPoints)))
			idx, err = rand.Int(rand.Reader, idx)
			if err != nil {
				return nil, fmt.Errorf("create random use crypto/rand fail: %w", err)
			}
		}
		ep := endPoints[idx.Int64()]
		// del the selected endpoint
		endPoints = append(endPoints[:idx.Int64()], endPoints[idx.Int64()+1:]...)

		rawURL := ep.categoryURL[datakit.ProfileUpload]
		if rawURL == "" {
			lastErr = fmt.Errorf("profile upload url empty")
			continue
		}

		URL, err := url.Parse(rawURL)
		if err != nil {
			lastErr = fmt.Errorf("profile upload url [%s] parse err:%w", rawURL, err)
			continue
		}
		return URL, nil
	}
	return nil, lastErr
}

func (dw *DataWayDefault) UploadProfile(profileID string, formFiles map[string][]*multipart.FileHeader) (*http.Response, error) {
	if profileID == "" {
		return nil, fmt.Errorf("empty profileID not allowed")
	}

	if len(formFiles) == 0 {
		return nil, fmt.Errorf("empty formFiles")
	}

	if len(dw.endPoints) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}
	dc := dw.endPoints[0]

	reqURL, ok := dc.categoryURL[datakit.Profile]
	if !ok {
		return nil, fmt.Errorf("no profile upload url available")
	}

	body := &bytes.Buffer{}
	mp := multipart.NewWriter(body)

	for fieldName, headers := range formFiles {
		for _, header := range headers {
			if header.Size <= 0 {
				continue
			}
			src, err := header.Open()
			if err != nil {
				return nil, fmt.Errorf("open upload file fail: %w", err)
			}
			partW, err := mp.CreateFormFile(fieldName, header.Filename)
			if err != nil {
				return nil, fmt.Errorf("multipart CreateFormFile fail:%w", err)
			}
			if n, err := io.Copy(partW, src); n != header.Size || err != nil {
				return nil, fmt.Errorf("copy uploaded file fail: %w", err)
			}
			_ = src.Close()
		}
	}

	if err := mp.Close(); err != nil {
		return nil, fmt.Errorf("multipart close fail: %w", err)
	}

	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("http newrequest fail: %w", err)
	}

	req.Header.Set("Content-Type", mp.FormDataContentType())
	req.Header.Set("Datakit-Profile-Id", profileID)
	return dw.sendReq(req)
}

func (dw *DataWayDefault) UploadLog(r io.Reader, hostName string) (*http.Response, error) {
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
