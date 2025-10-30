// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 文件监控相关指标.
	createEventCounter *prometheus.CounterVec // 接收到的文件创建事件总数
	discardCounter     *prometheus.CounterVec // 基于白名单丢弃的日志总数
	openFilesGauge     *prometheus.GaugeVec   // 当前打开的文件数量
	rotateCounter      *prometheus.CounterVec // 文件轮转总数
	parseFailCounter   *prometheus.CounterVec // 解析失败的日志总数
	multilineCounter   *prometheus.CounterVec // 多行日志状态总数

	// 套接字日志相关指标.
	socketConnectCounter *prometheus.CounterVec // 套接字连接状态总数
	socketMessageCounter *prometheus.CounterVec // 套接字日志消息总数
	socketLengthSummary  *prometheus.SummaryVec // 套接字日志长度分布

	// 文本处理相关指标.
	decodeErrorCounter *prometheus.CounterVec // 文本解码错误总数
)

func setupMetrics() {
	// 文件创建事件指标
	createEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "receive_create_event_total",
			Help:      "Total number of received create events",
		},
		[]string{
			"source", // 数据源名称
			"type",   // 事件类型（file/directory）
		},
	)

	// 日志丢弃指标
	discardCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "discard_log_total",
			Help:      "Total number of discarded logs based on the whitelist",
		},
		[]string{
			"source",   // 数据源名称
			"filepath", // 文件路径
		},
	)

	// 打开文件数量指标
	openFilesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "open_files",
			Help:      "Total number of currently open files",
		},
		[]string{
			"source", // 数据源名称
			"max",    // 最大文件数限制
		},
	)

	// 文件轮转指标
	rotateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "file_rotate_total",
			Help:      "Total number of file rotations performed",
		},
		[]string{
			"source",   // 数据源名称
			"filepath", // 文件路径
		},
	)

	// 解析失败指标
	parseFailCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "parse_fail_total",
			Help:      "Total number of failed parse attempts",
		},
		[]string{
			"source",   // 数据源名称
			"filepath", // 文件路径
			"mode",     // 解析模式
		},
	)

	// 多行日志状态指标
	multilineCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "multiline_state_total",
			Help:      "Total number of multiline states encountered",
		},
		[]string{
			"source", // 数据源名称
			"state",  // 多行状态
		},
	)

	// 套接字连接状态指标
	socketConnectCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_connect_status_total",
			Help:      "Total number of socket connection status events",
		},
		[]string{
			"network", // 网络类型
			"status",  // 连接状态
		},
	)

	// 套接字日志消息数量指标
	socketMessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_feed_message_count_total",
			Help:      "Total number of messages fed to the socket",
		},
		[]string{
			"network", // 网络类型
		},
	)

	// 套接字日志长度分布指标
	socketLengthSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "socket_log_length",
			Help:      "Length distribution of logs for socket communication",
			Objectives: map[float64]float64{
				0.5:  0.05,  // 50% 分位数，误差 5%
				0.90: 0.01,  // 90% 分位数，误差 1%
				0.99: 0.001, // 99% 分位数，误差 0.1%
			},
		},
		[]string{
			"network", // 网络类型
		},
	)

	// 文本解码错误指标
	decodeErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "tailer",
			Name:      "text_decode_errors_total",
			Help:      "Total count of text decoding failures",
		},
		[]string{
			"source",             // 数据源名称
			"character_encoding", // 字符编码
			"error_type",         // 错误类型
		},
	)

	// 注册所有指标到 Prometheus
	metrics.MustRegister(
		createEventCounter,
		discardCounter,
		openFilesGauge,
		parseFailCounter,
		multilineCounter,
		rotateCounter,
		socketLengthSummary,
		socketMessageCounter,
		socketConnectCounter,
		decodeErrorCounter,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
