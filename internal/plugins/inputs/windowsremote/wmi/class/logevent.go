// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package class

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

// Win32_NTLogEvent 表示 Windows 事件日志条目
// 该结构体映射自 WMI 的 Win32_NTLogEvent 类.
// nolint:stylecheck
type Win32_NTLogEvent struct {
	Category         uint16    // 事件分类代码（数字标识）.
	CategoryString   string    // 事件分类描述（可读形式）.
	ComputerName     string    // 生成事件的计算机名称.
	EventCode        uint16    // 事件 ID（如 1000 表示应用程序错误）.
	EventIdentifier  uint32    // 系统范围内唯一的事件标识符.
	EventType        uint8     // 事件类型（1=错误，2=警告，4=信息）.
	InsertionStrings []string  // 事件消息中的动态插入字符串.
	Logfile          string    // 所属日志类型（Application/System/Security）.
	Message          string    // 完整事件描述（包含动态参数）.
	RecordNumber     uint32    // 日志文件中的唯一记录编号.
	SourceName       string    // 事件来源程序/组件名称.
	TimeGenerated    time.Time // 事件生成时间（UTC 格式）.
	User             string    // 关联的用户账户（格式：域名\用户名）.

	//Type             string    // 事件类型描述（兼容旧字段）
	//Data             []byte    // 原始二进制数据（例如审计日志的二进制内容）
}

func eventType(t uint8) string {
	/*
		eventType
		1:Error
		2:Warning
		3:Information
		4:Security Audit Success
		5:Security Audit Failure
	*/
	switch t {
	case 1:
		return "error"
	case 2:
		return "warning"
	case 3:
		return "info"
	case 4:
		return "info"
	case 5:
		return "error"
	default:
		return "info"
	}
}

func (e Win32_NTLogEvent) String() string {
	return fmt.Sprintf(
		"[事件ID: 0x%X 事件类型：%s]\n"+
			"Category:%d CategoryString:%s\n"+
			"时间: \t\t%s\n"+
			"来源和唯一标识符: \t\t%s %d\n"+
			"日志类型: \t%s\n"+
			"计算机: \t%s\n"+
			"用户: \t\t%s\n"+
			"描述: \t\t%s",
		e.EventCode,
		eventType(e.EventType),
		e.Category,
		e.CategoryString,
		e.TimeGenerated.Local().Format("2006-01-02 15:04:05"),
		e.SourceName, e.EventIdentifier,
		e.Logfile,
		e.ComputerName,
		e.User,
		truncateString(e.Message, 100), // 截断长消息
	)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + " ..."
}

func (e Win32_NTLogEvent) ToPoint(tags map[string]string) *point.Point {
	opts := append(point.DefaultLoggingOptions(), point.WithTime(e.TimeGenerated))
	var kvs point.KVs
	for k, v := range tags {
		kvs = kvs.AddTag(k, v)
	}
	kvs = kvs.AddTag("host", strings.ToLower(e.ComputerName)).
		Add("message", e.Message).
		Add("status", eventType(e.EventType)).
		AddTag("user", e.User).
		AddTag("source", e.SourceName).
		AddTag("CategoryString", e.CategoryString).
		AddTag("service", e.Logfile).
		Add("event_code", e.EventCode).
		Add("event_identifier", e.EventIdentifier)

	return point.NewPoint("log_event", kvs, opts...)
}
