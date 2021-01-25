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

type PublishMessageRequestBody struct {
	// 消息标题，给邮箱订阅者发送邮件时作为邮件主题，长度不能超过512个字符。
	Subject string `json:"subject"`
	// 发送的消息。消息体必须是UTF-8编码的字符串，大小至多256KB。如果订阅者是手机号码，长度不超过490个字符，超出部分系统自动截断。短信内容不能包含“[]”或者“【】”符号。  对于移动推送订阅者推送消息，message消息必须符合移动推送平台的消息格式，消息格式请参见application消息体格式。否则移动app无法收到消息，目前支持的平台有HMS、APNS、APNS_SANDBOX。
	Message *string `json:"message,omitempty"`
	// Json格式的字符串。支持“email”、“sms”、 “http”、“https”、“dms”、“functiongraph”、“functionstage”、“HMS”、“APNS”以及“APNS_SANDBOX”。其中，“HMS”、“APNS”以及“APNS_SANDBOX”三种消息的格式请参见application消息体格式。必须设置默认的消息“default”，当匹配不到消息协议时，按“default”中的内容发送。  说明： 三种消息发送方式  message  message_structure  message_template_name  只需要设置其中一个，如果同时设置，生效的优先级为 message_structure > message_template_name > message。  华为透传消息  {   \"hps\": {     \"msg\": {       \"type\": 1,       \"body\": {         \"key\": \"value\"       }     }   } }  华为系统通知栏消息  {   \"hps\": {     \"msg\": {       \"type\": 3,       \"body\": {         \"content\": \"Push message content\",         \"title\": \"Push message content\"       },       \"action\": {         \"type\": 1,         \"param\": {           \"intent\": \"#Intent;compo=com.rvr/.Activity;S.W=U;end\"         }       }     },     \"ext\": {       \"biTag\": \"Trump\",       \"icon\": \"http://upload.w.org/00/150pxsvg.png\"     }   } }  苹果平台消息格式  {   \"aps\": {     \"alert\": \"hello world\"   } }
	MessageStructure *string `json:"message_structure,omitempty"`
	// 消息模板名称，可通过[查询消息模板列表](https://support.huaweicloud.com/api-smn/smn_api_53004.html)获取名称。  说明： 三种消息发送方式:  message  message_structure  message_template_name  只需要设置其中一个，如果同时设置，生效的优先级为 message_structure > message_template_name > message。
	MessageTemplateName *string `json:"message_template_name,omitempty"`
	// tag以及替换tag的参数组成的字典。消息模板中的标签对应的值。使用消息模板方式的消息发布必须携带该参数。字典中的key为消息模板中的参数名称，不超过21个字符。字典中的value为消息模板中的参数被替换后的值，不超过1KB。
	Tags map[string]string `json:"tags,omitempty"`
	// 指消息在SMN系统内部的最长存留时间。超过该存留时间，系统将不再发送该消息。单位是s，变量默认值是3600s，即一小时。值为正整数且小于等于3600*24。
	TimeToLive string `json:"time_to_live"`
}

func (o PublishMessageRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublishMessageRequestBody struct{}"
	}

	return strings.Join([]string{"PublishMessageRequestBody", string(data)}, " ")
}
