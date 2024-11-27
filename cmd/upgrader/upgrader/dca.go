// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"

	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

var ActionHandlerMap = map[string]ws.ActionHandler{}

type DCAClient struct {
	*ws.Client
	info map[string]*httpapi.DCAInfo
}

var dcaClient = &DCAClient{info: make(map[string]*httpapi.DCAInfo)}

func (c *DCAClient) Exit() {
	if c.Client != nil {
		c.Client.Stop()
		c.Client = nil

		l.Info("dca exit")
	}
}

func (c *DCAClient) UpdateDatakitStatus(status ws.DataKitStatus, dk *ws.DataKit) {
	if c.Client == nil {
		return
	}
	query := url.Values{"status": []string{status.String()}}
	if dk != nil {
		query.Add("runtimeID", dk.RunTimeID)
	}
	if err := c.SendMessage(&ws.WebsocketMessage{
		Action: ws.UpdateDatakitStatus,
		Data:   ws.ActionData{Query: query},
	}); err != nil {
		l.Warnf("send message failed: %s", err.Error())
	}
}

func (c *DCAClient) UpdateDatakit(datakit *ws.DataKit, updateRemote bool) {
	if c.Client == nil || datakit == nil {
		return
	}

	// update websocket header
	c.Client.SetDatakit(datakit)

	if updateRemote {
		if err := c.SendMessage(&ws.WebsocketMessage{
			Action: ws.UpdateDatakit,
			Data:   ws.ActionData{Body: string(datakit.Bytes())},
		}); err != nil {
			l.Warnf("send message failed: %s", err.Error())
		}
	}
}

func (c *DCAClient) isInitialized() bool {
	return c.Client != nil
}

func (c *DCAClient) restartClient(websocketURL string, datakit *ws.DataKit, actionHandlers map[string]ws.ActionHandler) error {
	var err error

	if c.Client != nil {
		c.Client.Stop()
		c.Client = nil
	}

	if c.Client, err = ws.NewClient(
		ws.WithWebsocketAddress(websocketURL),
		ws.WithLogger(l),
		ws.WithActionHandlers(actionHandlers),
		ws.WithDataKit(datakit),
		ws.WithOnInitialized(func() { c.UpdateDatakit(datakit, true) }),
	); err != nil {
		return fmt.Errorf("create websocket client failed: %w", err)
	} else {
		c.Client.Start()
	}

	return nil
}

type DataKitClient interface {
	Init() error
	SyncDataKit()
	Request(method, path string, dest any, url string, data *ws.ActionData) error
}

type baseDatakitClient struct {
	SyncCh chan struct{}
}

func (c *baseDatakitClient) Request(method, path string, dest any, urlAddress string, data *ws.ActionData) error {
	if urlAddress == "" {
		return fmt.Errorf("empty url")
	}

	u, err := url.Parse(fmt.Sprintf("%s/v1/dca%s", urlAddress, path))
	if err != nil {
		return fmt.Errorf("parse url: %s, error: %w", u, err)
	}

	var body io.Reader
	if data != nil {
		if data.Query != nil {
			u.RawQuery = data.Query.Encode()
		}
		if data.Body != "" {
			body = strings.NewReader(data.Body)
		}
	}

	cli := cmds.GetHTTPClient("")
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("unable to request: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body result: %w", err)
	}

	if err := json.Unmarshal(respBody, &dest); err != nil {
		return fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return err
}

type hostDatakitClient struct {
	baseDatakitClient
	url  string
	info *httpapi.DCAInfo
}

func (c *hostDatakitClient) Init() error {
	schema := "http"
	if Cfg.DatakitAPIHTTPS {
		schema = "https"
	}
	c.url = fmt.Sprintf("%s://%s", schema, Cfg.DatakitAPIListen)

	return nil
}

func (c *hostDatakitClient) SyncDataKit() {
	l.Debug("sync datakit info")
	dcaInfo := httpapi.DCAInfo{}
	res := ws.DCAResponse{Content: &dcaInfo}

	if err := c.Request(http.MethodGet, "/info", &res, c.url, nil); err != nil {
		l.Warnf("get dca info failed: %s", err.Error())

		c.UpdateDatakitStatus(ws.StatusStopped)
		return
	}

	if !res.Success {
		l.Warnf("get dca info failed: %s", res.Message)
		return
	}

	if dcaInfo.DataKit != nil {
		dcaInfo.DataKit.URL = c.url
		dcaInfo.DataKit.Status = ws.StatusRunning

		c.LoadConfig(&dcaInfo)
	}
}

func (c *hostDatakitClient) UpdateDatakitStatus(status ws.DataKitStatus) {
	if c.info != nil && c.info.DataKit != nil {
		c.info.DataKit.Status = status
		dcaClient.UpdateDatakitStatus(status, c.info.DataKit)
	}
}

func (c *hostDatakitClient) LoadConfig(dcaInfo *httpapi.DCAInfo) {
	if dcaInfo == nil || dcaInfo.DataKit == nil {
		l.Warn("dca info or datakit is empty, ignore")
		if c.info != nil && c.info.DataKit != nil {
			c.info.DataKit.Status = ws.StatusStopped
			dcaClient.UpdateDatakit(c.info.DataKit, false)
		}
		return
	}

	// dca not enabled
	if dcaInfo.Config.DCAConfig == nil || !dcaInfo.Config.DCAConfig.Enable {
		l.Debug("dca not enabled, ignore, exit")
		dcaClient.Exit()
		return
	}

	if !dcaClient.isInitialized() || (c.info.DataKit.ConnID != dcaInfo.DataKit.ConnID) { // create new client
		if err := dcaClient.restartClient(
			dcaInfo.Config.DCAConfig.WebsocketServer,
			dcaInfo.DataKit,
			ActionHandlerMap,
		); err != nil {
			l.Warnf("create websocket client failed: %s", err.Error())
		} else {
			l.Info("websocket client created")
		}
	}

	// update datakit remote
	dcaClient.UpdateDatakit(dcaInfo.DataKit, true)

	// replace with new config
	c.info = dcaInfo
}

var dkClient DataKitClient

func startDCA(p *serviceImpl) error {
	dkClient = &hostDatakitClient{}

	if err := dkClient.Init(); err != nil {
		return fmt.Errorf("init dkClient failed: %w", err)
	}

	httpapi.SetDCALogger(l)

	signals := make(chan os.Signal, datakit.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		dkClient.SyncDataKit()
		select {
		case sig := <-signals:
			l.Infof("get signal %v, wait & exit", sig)
			return nil
		case <-p.done:
			l.Info("stop dca")
			dcaClient.Exit()
			return nil
		case <-ticker.C:
		}
	}
}

func getDatakitStatsAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	res := ws.DCAResponse{}
	if err := dkClient.Request(http.MethodGet, "/stat", &res, datakit.URL, nil); err != nil {
		l.Warnf("get datakit stats failed: %s", err.Error())
		response.SetError()
		return
	}

	response.SetResponse(&res)
}

func reloadDatakitAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	res := ws.DCAResponse{}
	if err := dkClient.Request(http.MethodGet, "/reload", &res, datakit.URL, nil); err != nil {
		l.Warnf("reload datakit failed: %s", err.Error())
		response.SetError()
		return
	}

	response.SetResponse(&res)
}

func saveDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	res := ws.DCAResponse{}
	if err := dkClient.Request(http.MethodPost, "/saveConfig", &res, datakit.URL, data); err != nil {
		l.Warnf("save config failed: %s", err.Error())
		response.SetError()
		return
	}

	response.SetResponse(&res)
}

func deleteDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	res := ws.DCAResponse{}
	if err := dkClient.Request(http.MethodDelete, "/config", &res, datakit.URL, data); err != nil {
		l.Warnf("delete config failed: %s", err.Error())
		response.SetError()
		return
	}

	response.SetResponse(&res)
}

func getDatakitConfigAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	res := ws.DCAResponse{}
	if err := dkClient.Request(http.MethodGet, "/config", &res, datakit.URL, data); err != nil {
		l.Warnf("get config failed: %s", err.Error())
		response.SetError()
		return
	}

	response.SetResponse(&res)
}

// upgradeDatakitAction will restart datakit, and the service will be restarted.
func upgradeDatakitAction(client *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	version := data.Query.Get("version")
	force := data.Query.Get("force") != ""

	// not block here
	g.Go(func(ctx context.Context) error {
		if err := ui.upgrade(withVersion(version), withForce(force)); err != nil {
			l.Errorf("upgrade datakit error: %s", err.Error())
		} else {
			l.Info("upgrade datakit success")
		}

		return nil
	})

	response.SetSuccess("ok")
}

func stopDatakitAction(client *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	if err := ui.forceStopService(); err != nil {
		l.Errorf("force stop datakit failed: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 500, ErrorCode: "service.stop.failed", ErrorMsg: "Fail to stop datakit"})
		return
	}
	response.SetSuccess("ok")
}

func restartDatakitAction(_ *ws.Client, response *ws.DCAResponse, data *ws.ActionData, datakit *ws.DataKit) {
	g.Go(func(ctx context.Context) error {
		if err := ui.restartService(); err != nil {
			l.Errorf("restart datakit failed: %s", err.Error())
			return nil
		}

		return nil
	})
	response.SetSuccess("ok")
}

//nolint:gochecknoinits
func init() {
	for k, v := range httpapi.HostActionHandlerMap {
		ActionHandlerMap[k] = v
	}

	ActionHandlerMap[ws.GetDatakitStatsAction] = ws.GetActionHandler(ws.GetDatakitStatsAction, getDatakitStatsAction)
	ActionHandlerMap[ws.ReloadDatakitAction] = ws.GetActionHandler(ws.ReloadDatakitAction, reloadDatakitAction)
	ActionHandlerMap[ws.UpgradeDatakitAction] = ws.GetActionHandler(ws.UpgradeDatakitAction, upgradeDatakitAction)
	ActionHandlerMap[ws.StopDatakitAction] = ws.GetActionHandler(ws.StopDatakitAction, stopDatakitAction)
	ActionHandlerMap[ws.RestartDatakitAction] = ws.GetActionHandler(ws.RestartDatakitAction, restartDatakitAction)
	ActionHandlerMap[ws.SaveDatakitConfigAction] = ws.GetActionHandler(ws.SaveDatakitConfigAction, saveDatakitConfigAction)
	ActionHandlerMap[ws.DeleteDatakitConfigAction] = ws.GetActionHandler(ws.DeleteDatakitConfigAction, deleteDatakitConfigAction)
	ActionHandlerMap[ws.GetDatakitConfigAction] = ws.GetActionHandler(ws.GetDatakitConfigAction, getDatakitConfigAction)
}
