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

type PublishAppMessageRequestBody struct {
	//  message与message_structure二者选其一。  message, App消息发布。  message_structure, 使用消息结构体方式的App消息发布。  app推送的消息内容，当前支持的推送平台有HMS、APNS、APNS_SANDBOX。  HMS是为开发者提供的消息推送平台。  APNS和APNS_SANDBOX是用于推送iOS消息的服务平台。  HMS平台指定的消息内容不超过2K。  APNS和APNS_SANDBOX平台的消息内容不能超过4K。  推送平台的消息内容格式要求详情见application消息体格式。  华为透传消息  {   \"hps\": {     \"msg\": {       \"type\": 1,       \"body\": {         \"key\": \"value\"       }     }   } }  华为系统通知栏消息  {   \"hps\": {     \"msg\": {       \"type\": 3,       \"body\": {         \"content\": \"Push message content\",         \"title\": \"Push message content\"       },       \"action\": {         \"type\": 1,         \"param\": {           \"intent\": \"#Intent;compo=com.rvr/.Activity;S.W=U;end\"         }       }     },     \"ext\": {       \"biTag\": \"Trump\",       \"icon\": \"http://upload.w.org/00/150pxsvg.png\"     }   } }  苹果平台消息格式 {   \"aps\": {     \"alert\": \"hello world\"   } }
	Message *string `json:"message,omitempty"`
	// app推送的消息内容，当前支持的推送平台有HMS、APNS、APNS_SANDBOX。  HMS是为开发者提供的消息推送平台。  APNS和APNS_SANDBOX是用于推送iOS消息的服务平台。  HMS平台指定的消息内容不超过2K。  APNS和APNS_SANDBOX平台的消息内容不能超过4K。  推送平台的消息内容格式要求详情见application消息体格式。  华为透传消息  {   \"HMS\": {     \"hps\": {       \"msg\": {         \"type\": 1,         \"body\": {           \"key\": \"value\"         }       }     }   } }  华为系统通知栏消息  {   \"HMS\": {     \"hps\": {       \"msg\": {         \"type\": 3,         \"body\": {           \"content\": \"Push message content\",           \"title\": \"Push message content\"         },         \"action\": {           \"type\": 1,           \"param\": {             \"intent\": \"#Intent;compo=com.rvr/.Activity;S.W=U;end\"           }         }       },       \"ext\": {         \"biTag\": \"Trump\",         \"icon\": \"http://upload.w.org/00/150pxsvg.png\"       }     }   } }  苹果平台消息格式  {   \"APNS\": {     \"aps\": {       \"alert\": \"hello world\"     }   } }
	MessageStructure *string `json:"message_structure,omitempty"`
	// 消息发送的生存时间，是相对于发布时间的。  SMN系统将移动推送消息转交给推送平台前，会计算该消息在系统消耗的时间。只有消耗的时间小于time_to_live时，SMN才会将消息转交给推送平台，并将time_to_live减去消耗的时间传递给推送平台，否则消息废弃。  time _to_live的单位是s，变量默认值是3600s，即一小时。值为正整数且小于等于3600*24。
	TimeToLive *string `json:"time_to_live,omitempty"`
}

func (o PublishAppMessageRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublishAppMessageRequestBody struct{}"
	}

	return strings.Join([]string{"PublishAppMessageRequestBody", string(data)}, " ")
}
