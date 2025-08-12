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

// Win32_OperatingSystem 表示Windows操作系统信息.
// nolint:stylecheck
type Win32_OperatingSystem struct {
	Caption                string    // 操作系统名称
	Version                string    // 系统版本号
	BuildNumber            string    // 构建号
	OSArchitecture         string    // 系统架构(32/64位)
	CSName                 string    // 计算机系统名称(主机名)
	InstallDate            time.Time // 安装日期
	LastBootUpTime         time.Time // 最后启动时间
	SerialNumber           string    // 产品序列号
	SystemDirectory        string    // 系统目录
	WindowsDirectory       string    // Windows目录
	CountryCode            string    // 国家代码
	Locale                 string    // 区域设置
	TotalVisibleMemorySize uint64    // 总物理内存(KB)
	FreePhysicalMemory     uint64    // 可用物理内存(KB)
	TotalVirtualMemorySize uint64    // 总虚拟内存(KB)
	FreeVirtualMemory      uint64    // 可用虚拟内存(KB)
	NumberOfProcesses      uint32    // 运行的进程数
	NumberOfUsers          uint32    // 登录用户数
}

// String 方法提供结构体的格式化输出.
func (os *Win32_OperatingSystem) String() string {
	var builder strings.Builder

	builder.WriteString("=== 操作系统信息 ===\n")
	builder.WriteString(fmt.Sprintf("操作系统名称: %s\n", os.Caption))
	builder.WriteString(fmt.Sprintf("版本: %s (Build %s)\n", os.Version, os.BuildNumber))
	builder.WriteString(fmt.Sprintf("系统架构: %s\n", os.OSArchitecture))
	builder.WriteString(fmt.Sprintf("主机名: %s\n", os.CSName))
	builder.WriteString(fmt.Sprintf("安装日期: %s\n", os.InstallDate.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("最后启动时间: %s\n", os.LastBootUpTime.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("产品序列号: %s\n", os.SerialNumber))
	builder.WriteString(fmt.Sprintf("系统目录: %s\n", os.SystemDirectory))
	builder.WriteString(fmt.Sprintf("Windows目录: %s\n", os.WindowsDirectory))
	builder.WriteString(fmt.Sprintf("区域设置: %s (国家代码: %s)\n", os.Locale, os.CountryCode))
	builder.WriteString("\n=== 内存信息 ===\n")
	builder.WriteString(fmt.Sprintf("总物理内存: %.2f GB\n", float64(os.TotalVisibleMemorySize)/1024/1024))
	builder.WriteString(fmt.Sprintf("可用物理内存: %.2f GB\n", float64(os.FreePhysicalMemory)/1024/1024))
	builder.WriteString(fmt.Sprintf("总虚拟内存: %.2f GB\n", float64(os.TotalVirtualMemorySize)/1024/1024))
	builder.WriteString(fmt.Sprintf("可用虚拟内存: %.2f GB\n", float64(os.FreeVirtualMemory)/1024/1024))
	builder.WriteString("\n=== 系统状态 ===\n")
	builder.WriteString(fmt.Sprintf("运行进程数: %d\n", os.NumberOfProcesses))
	builder.WriteString(fmt.Sprintf("登录用户数: %d", os.NumberOfUsers))

	return builder.String()
}

func (os Win32_OperatingSystem) ToMemPoint(host string) *point.Point {
	freePercent := float64(os.FreePhysicalMemory) / float64(os.TotalVisibleMemorySize)
	used := os.TotalVisibleMemorySize - os.FreePhysicalMemory
	usedPercent := float64(used) / float64(os.TotalVisibleMemorySize)
	opts := point.DefaultMetricOptions()
	var kvs point.KVs
	kvs = kvs.AddTag("host", host).
		Add("available", os.FreePhysicalMemory*1024).
		Add("available_percent", freePercent*100).
		Add("total", os.TotalVisibleMemorySize*1024).
		Add("used", used*1024).
		Add("used_percent", usedPercent*100)

	return point.NewPoint("mem", kvs, opts...)
}
