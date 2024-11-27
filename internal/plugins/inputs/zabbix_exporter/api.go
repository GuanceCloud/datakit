// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ZabbixAPI struct {
	server   string
	user, pw string

	token string
	// todo 增加时间限制，比如5秒才能请求一次。
}

func (za *ZabbixAPI) getToken() {
	authPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]interface{}{
			"user":     za.user,
			"password": za.pw,
		},
		"id": 1,
	}

	authData, err := json.Marshal(authPayload)
	if err != nil {
		log.Errorf("Error marshaling auth payload: %v", err)
		return
	}

	authResp, err := http.Post(za.server, "application/json-rpc", strings.NewReader(string(authData)))
	if err != nil {
		log.Errorf("Error sending auth request: %v", err)
		return
	}
	defer authResp.Body.Close() //nolint

	authBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		log.Errorf("Error reading auth response body: %v", err)
		return
	}
	log.Infof("get auth body is %s", string(authBody))
	var authResult struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  string      `json:"result"`
		Error   interface{} `json:"error"`
		ID      int         `json:"id"`
	}

	err = json.Unmarshal(authBody, &authResult)
	if err != nil {
		log.Errorf("Error unmarshaling auth response: %v", err)
		return
	}

	if authResult.Error != nil {
		log.Errorf("Auth error: %v", authResult.Error)
		return
	}
	log.Infof("zabix server token is %s", authResult.Result)
	za.token = authResult.Result
}

type ItemResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  []struct {
		Itemid string `json:"itemid"`
		TypeC  string `json:"type"`
		Name   string `json:"name"`
		Key_   string `json:"key_"`
		Hostid string `json:"hostid"`
		Units  string `json:"units"`
	} `json:"result"`
	ID int `json:"id"`
}

func (za *ZabbixAPI) getItemByID(itemID int64) *ItemC {
	// Prepare the request payload
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "item.get",
		"params": map[string]interface{}{
			"output":  []string{"itemid", "type", "name", "key_", "hostid", "units"},
			"itemids": strconv.FormatInt(itemID, 10),
		},
		"id":   1,
		"auth": "", // Auth token will be added after the first request for login
	}
	payload["auth"] = za.token
	itemData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling item payload: %v", err)
		return nil
	}

	itemResp, err := http.Post(za.server, "application/json-rpc", strings.NewReader(string(itemData)))
	if err != nil {
		log.Errorf("Error sending item request: %v", err)
		return nil
	}
	defer itemResp.Body.Close() //nolint

	itemBody, err := io.ReadAll(itemResp.Body)
	if err != nil {
		log.Errorf("Error reading item response body: %v", err)
		return nil
	}
	log.Infof("item body = %s", string(itemBody))
	var itemResult ItemResponse

	err = json.Unmarshal(itemBody, &itemResult)
	if err != nil {
		log.Errorf("Error unmarshaling item response: %v", err)
		return nil
	}

	if len(itemResult.Result) == 0 {
		log.Errorf("No item found with itemid: %d", itemID)
		return nil
	}

	item := itemResult.Result[0]
	log.Debugf("Item Name: %s,type:%s, Key: %s,hostID:%s, Units: %s \n", item.Name, item.TypeC, item.Key_, item.Hostid, item.Units)
	id, err := strconv.ParseInt(item.Itemid, 10, 64)
	if err != nil {
		log.Errorf("itemID:%s to int64 parseInt err %v", item.Itemid, err)
		return nil
	}
	hostID, err := strconv.ParseInt(item.Hostid, 10, 64)
	if err != nil {
		log.Errorf("itemID:%s to int64 parseInt err %v", item.Itemid, err)
		return nil
	}
	t, err := strconv.Atoi(item.TypeC)
	if err != nil {
		log.Errorf("can not parse type:%s to int", item.TypeC)
		t = 0
	}
	ic := &ItemC{
		ItemID: id,
		Name:   item.Name,
		TypeC:  t,
		HostID: hostID,
		Units:  item.Units,
		Key_:   item.Key_,
	}
	ic.setValues()
	return ic
}
