// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func (dw *Dataway) GetLogFilter() ([]byte, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("[error] dataway url empty")
	}

	return dw.eps[0].getLogFilter()
}

type checkTokenResult struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

func (dw *Dataway) CheckToken(token, scheme, host string) (bool, error) {
	if len(dw.eps) == 0 {
		return false, fmt.Errorf("no dataway available")
	}

	dc := dw.eps[0]

	if len(scheme) == 0 {
		scheme = dc.scheme
	}

	if len(host) == 0 {
		host = dc.host
	}
	reqURL := fmt.Sprintf("%s://%s%s/%s", scheme, host, datakit.TokenCheck, token)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := dc.sendReq(req)
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return false, err
	}

	defer resp.Body.Close() //nolint:errcheck

	result := checkTokenResult{}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("invalid JSON body content")
	}

	if result.Code == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

func (dw *Dataway) WorkspaceQuery(body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.eps[0]
	requrl, ok := dc.categoryURL[datakit.Workspace]
	if !ok {
		return nil, fmt.Errorf("no workspace query URL available")
	}

	log.Debugf("NewRequest: %s", requrl)
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return dw.eps[0].sendReq(req)
}

func (dw *Dataway) DQLQuery(body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	dc := dw.eps[0]
	requrl, ok := dc.categoryURL[datakit.QueryRaw]
	if !ok {
		return nil, fmt.Errorf("no DQL query URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return dw.eps[0].sendReq(req)
}

func (dw *Dataway) Election(namespace, id string, reqBody io.Reader) ([]byte, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]

	requrl, ok := ep.categoryURL[datakit.Election]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	log.Debugf("election sending %s", requrl)

	req, err := http.NewRequest("POST", requrl, reqBody)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	resp, err := ep.sendReq(req)
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

func (dw *Dataway) ElectionHeartbeat(namespace, id string, reqBody io.Reader) ([]byte, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]

	requrl, ok := ep.categoryURL[datakit.ElectionHeartbeat]
	if !ok {
		return nil, fmt.Errorf("no election URL available")
	}

	if strings.Contains(requrl, "?token") {
		requrl += fmt.Sprintf("&namespace=%s&id=%s", namespace, id)
	} else {
		return nil, fmt.Errorf("token missing")
	}

	log.Debugf("election sending heartbeat %s", requrl)

	req, err := http.NewRequest("POST", requrl, reqBody)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	resp, err := ep.sendReq(req)
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

func (dw *Dataway) DatawayList() ([]string, int, error) {
	if len(dw.eps) == 0 {
		return nil, datawayListIntervalDefault, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ListDataWay]
	if !ok {
		return nil, datawayListIntervalDefault, fmt.Errorf("dataway list API not available")
	}

	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, datawayListIntervalDefault, err
	}

	resp, err := ep.sendReq(req)
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

// UpsertObjectLabels used to create or update object labels.
func (dw *Dataway) UpsertObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return ep.sendReq(req)
}

// DeleteObjectLabels used to delete object labels.
func (dw *Dataway) DeleteObjectLabels(tkn string, body []byte) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	requrl, ok := ep.categoryURL[datakit.ObjectLabel]
	if !ok {
		return nil, fmt.Errorf("no object labels URL available")
	}

	rBody := bytes.NewReader(body)
	req, err := http.NewRequest("DELETE", requrl, rBody)
	if err != nil {
		return nil, fmt.Errorf("delete object label error: %w", err)
	}

	return ep.sendReq(req)
}

func (dw *Dataway) GetAvailableEndpoints() []*endPoint {
	return dw.eps
}

func (dw *Dataway) UploadLog(r io.Reader, hostName string) (*http.Response, error) {
	if len(dw.eps) == 0 {
		return nil, fmt.Errorf("no dataway available")
	}

	ep := dw.eps[0]
	reqURL, ok := ep.categoryURL[datakit.LogUpload]
	if !ok {
		return nil, fmt.Errorf("no file upload URL available")
	}

	req, err := http.NewRequest("POST", reqURL, r)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	req.Header.Add("Host-Name", hostName)
	return ep.sendReq(req)
}

func (dw *Dataway) Pull(args string) ([]byte, error) {
	if dw.ClientsCount() == 0 {
		return nil, fmt.Errorf("dataway URL not set")
	}

	return dw.eps[0].datakitPull(args)
}
