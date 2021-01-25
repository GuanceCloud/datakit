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

type Statement struct {
	// Statement语句的ID。 Statement语句ID必须是唯一的，例如statement01、statement02。
	Sid string `json:"Sid"`
	// Statement语句的效果。“Allow”或者“Deny”。
	Effect string `json:"Effect"`
	// Statement语句作用的对象。 目前支持“CSP”和“Service”两类对象。  “CSP”对象指的是其他用户，可以作用于多个用户。  “Service”对象指的是云服务，可以作用于多个云服务。  Principal元素和NotPrincipal元素两者任选其一。选定后， “CSP”对象填写内容的格式为格式为urn:csp:iam::domainId:root，其中domainId为其他用户的“账号ID”。  “Service”对象填写内容的格式为小写的云服务名称缩写。
	Principal *string `json:"Principal,omitempty"`
	// NotPrincipal：Statement语句排除作用的对象。  目前支持“CSP”和“Service”两类对象。  “CSP”对象指的是其他用户，可以作用于多个用户。  “Service”对象指的是云服务，可以作用于多个云服务。  Principal元素和NotPrincipal元素两者任选其一。选定后， “CSP”对象填写内容的格式为格式为urn:csp:iam::domainId:root，其中domainId为其他用户的“账号ID”。  “Service”对象填写内容的格式为小写的云服务名称缩写。
	NotPrincipal *string `json:"NotPrincipal,omitempty"`
	// Statement语句作用的操作。  允许使用通配符来表示一类操作，例如：SMN:Update*、SMN:Delete*。如果只填写“*”，表示Statement语句作用的操作为该资源支持的所有操作。  Action元素和NotAction元素两者任选其一。  目前支持的操作有：  SMN:UpdateTopic SMN:DeleteTopic SMN:QueryTopicDetail SMN:ListTopicAttributes SMN:UpdateTopicAttribute SMN:DeleteTopicAttributes SMN:DeleteTopicAttributeByName SMN:ListSubscriptionsByTopic SMN:Subscribe SMN:Unsubscribe SMN:Publish
	Action *string `json:"Action,omitempty"`
	// Statement语句排除作用的操作。  允许使用通配符来表示一类操作，例如：SMN:Update*、SMN:Delete*。如果只填写“*”，表示Statement语句作用的操作为该资源支持的所有操作。  Action元素和NotAction元素两者任选其一。  目前支持的操作有：  SMN:UpdateTopic  SMN:DeleteTopic  SMN:QueryTopicDetail  SMN:ListTopicAttributes  SMN:UpdateTopicAttribute  SMN:DeleteTopicAttributes  SMN:DeleteTopicAttributeByName  SMN:ListSubscriptionsByTopic  SMN:Subscribe  SMN:Unsubscribe  SMN:Publish
	NotAction *string `json:"NotAction,omitempty"`
	// Statement语句作用的主题。  Resource和NotResource两者任选其一。选定后，填写内容为主题URN。
	Resource *string `json:"Resource,omitempty"`
	// Statement语句排除作用的主题。  Resource和NotResource两者任选其一。选定后，填写内容为主题URN。
	NotResource *string `json:"NotResource,omitempty"`
}

func (o Statement) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Statement struct{}"
	}

	return strings.Join([]string{"Statement", string(data)}, " ")
}
