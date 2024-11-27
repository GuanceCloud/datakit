// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
)

type ActionHandler func(client *Client, datakit *ws.DataKit, ctx *gin.Context) (any, error)

var ActionHandlerMap map[string]ActionHandler

// doCommonAction is a handler for not-request-reply action.
// Only support UpdateDatakitStatus, UpdateDatakit action now.
func doCommonAction(client *Client, msg *ws.WebsocketMessage) {
	data, _ := msg.Data.(*ws.ActionData)
	if data == nil {
		data = &ws.ActionData{}
	}

	switch msg.Action {
	case ws.UpdateDatakitStatus:
		dkStatus := ws.StatusStopped
		if status := data.Query.Get("status"); status != "" {
			dkStatus = ws.DataKitStatus(status)
		}
		if err := datakitDB.UpdateStatus(client.DataKit, dkStatus); err != nil {
			l.Errorf("failed to update datakit status: %s", err.Error())
		}
	case ws.DeleteDatakit:
		if err := datakitDB.Delete(client.DataKit, true); err != nil {
			l.Errorf("failed to delete datakit: %s", err.Error())
		}
	case ws.UpdateDatakit:
		dk := &ws.DataKit{}
		if err := json.Unmarshal([]byte(data.Body), &dk); err != nil {
			l.Errorf("failed to unmarshal datakit: %s", err.Error())
		} else {
			if err := datakitDB.Update(dk); err != nil {
				l.Errorf("failed to update datakit: %s", err.Error())
			}
		}
		client.DataKit = dk

	default:
		l.Warnf("action %s not found for common handler", msg.Action)
	}
}

func upgradeDatakitAction(client *Client, datakit *ws.DataKit, ctx *gin.Context) (any, error) {
	response := preOperation(datakit, ws.StatusUpgrading)
	if !response.Success {
		return response, nil
	}
	return getActionHandler(ws.UpgradeDatakitAction)(client, datakit, ctx)
}

func isOperationAllowed(status ws.DataKitStatus) bool {
	return (status != ws.StatusOffline) && (status != ws.StatusRestarting) && (status != ws.StatusUpgrading)
}

func preOperation(datakit *ws.DataKit, status ws.DataKitStatus) (response *ws.DCAResponse) {
	response = &ws.DCAResponse{Success: true}
	if datakit == nil {
		response.SetError(&ws.ResponseError{Code: 400, ErrorMsg: "datakit is not available"})
		return
	}

	if dk, err := datakitDB.Find(datakit); err != nil {
		l.Errorf("failed to find datakit: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 500, ErrorMsg: "failed to find datakit"})
		return
	} else if !isOperationAllowed(dk.Status) {
		response.SetError(&ws.ResponseError{Code: 400, ErrorMsg: fmt.Sprintf("Operation is not allowed, current status is %s", dk.Status.String())})
		return
	}

	if err := datakitDB.UpdateStatus(datakit, status); err != nil {
		l.Errorf("failed to update datakit status: %s", err.Error())
		response.SetError(&ws.ResponseError{Code: 500, ErrorMsg: "failed to update datakit status"})
	}

	return
}

func restartDatakitAction(client *Client, datakit *ws.DataKit, ctx *gin.Context) (any, error) {
	response := preOperation(datakit, ws.StatusRestarting)
	if !response.Success {
		return response, nil
	}

	return getActionHandler(ws.RestartDatakitAction)(client, datakit, ctx)
}

func getDatakitStatsAction(client *Client, datakit *ws.DataKit, _ *gin.Context) (any, error) {
	msg := &ws.WebsocketMessage{
		Action: ws.GetDatakitStatsAction,
	}

	out := ws.WebsocketMessage{
		Data: &ws.DCAResponse{},
	}

	if err := client.request(msg, &out); err != nil {
		return nil, fmt.Errorf("failed to get datakit stats: %w", err)
	}

	return out.Data, nil
}

func getActionHandler(action string) ActionHandler {
	return func(client *Client, datakit *ws.DataKit, ctx *gin.Context) (any, error) {
		body, err := io.ReadAll(ctx.Request.Body)
		defer ctx.Request.Body.Close() //nolint:errcheck

		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		msg := &ws.WebsocketMessage{
			Action: action,
			Data: ActionData{
				Body:  string(body),
				Query: ctx.Request.URL.Query(),
			},
		}

		out := ws.WebsocketMessage{
			Data: &ws.DCAResponse{},
		}
		if err := client.request(msg, &out); err != nil {
			return nil, fmt.Errorf("failed to get datakit config: %w", err)
		}

		return out.Data, nil
	}
}

//nolint:gochecknoinits
func init() {
	// default handler
	ActionHandlerMap = map[string]ActionHandler{
		ws.GetDatakitStatsAction: getDatakitStatsAction,
		ws.UpgradeDatakitAction:  upgradeDatakitAction,
		ws.RestartDatakitAction:  restartDatakitAction,
	}
}
