/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type AddSubscriptionRequestBody struct {
	// 说明：http协议，接入点必须以“http://”开头。  https协议，接入点必须以“https://”开头。  email协议，接入点必须是邮件地址。  sms协议，接入点必须是一个电话号码。  functionstage协议，接入点必须是一个函数。  functiongraph协议，接入点必须是一个函数工作流。  dms协议，接入点必须是一个消息队列。  application协议，接入点必须是一个应用平台的设备终端。  callnotify协议，接入点必须是一个电话号码。
	Endpoint string `json:"endpoint"`
	// 不同协议对应不同的endpoint（接受消息的接入点）。 目前支持的协议包括：  “email”：邮件传输协议，endpoint为邮箱地址。  “default”  “sms”：短信传输协议，endpoint为手机号码。  “functionstage”：FunctionGraph（函数）传输协议，endpoint为一个函数。  “functiongraph”：FunctionGraph（工作流）传输协议，endpoint为由一组函数编排成的工作流。  “http”、“https”：HTTP/HTTPS传输协议，endpoint为URL。  “callnotify”：语音通知传输协议，endpoint为手机号码。
	Protocol string `json:"protocol"`
	// 备注。最大支持128字节，约42个中文，必须是UTF-8编码的字符串，否则无法正常显示中文。
	Remark string `json:"remark"`
}

func (o AddSubscriptionRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddSubscriptionRequestBody struct{}"
	}

	return strings.Join([]string{"AddSubscriptionRequestBody", string(data)}, " ")
}
