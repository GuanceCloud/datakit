// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package export

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type info []*inputs.ENVInfo

// example: map["env-dataway"]info
type content map[string]info

// example: map["doc/datakit-daemonset-deploy.md"]content
var nonInputDocs = map[string]content{
	"doc/datakit-daemonset-deploy.md": {
		"envCommon":   envCommon(),
		"envDataway":  envDataway(),
		"envLog":      envLog(),
		"envPprof":    envPprof(),
		"envElect":    envElect(),
		"envHTTPAPI":  envHTTPAPI(),
		"envConfd":    envConfd(),
		"envGit":      envGit(),
		"envSinker":   envSinker(),
		"envIO":       envIO(),
		"envDca":      envDca(),
		"envRefta":    envRefta(),
		"envRecorder": envRecorder(),
		"envOthers":   envOthers(),
		"remote_job":  envRemoteJob(),
		// "envPointPool": envPointPool(),
	},
}

func envCommon() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_DISABLE_PROTECT_MODE",
			Type:    doc.Boolean,
			Desc:    "Disable protect mode",
			DescZh:  "禁用「配置保护」模式",
		},
		{
			ENVName:  "ENV_DATAWAY",
			Type:     doc.URL,
			Example:  "`https://openway.<<<custom_key.brand_main_domain>>>.com?token=xxx`",
			Required: doc.Yes,
			Desc:     "Configure the DataWay address",
			DescZh:   "配置 DataWay 地址",
		},
		{
			ENVName: "ENV_DEFAULT_ENABLED_INPUTS",
			Type:    doc.List,
			Example: `cpu,mem,disk`,
			Desc:    "[The list of collectors](datakit-input-conf.md#default-enabled-inputs) is opened by default, divided by commas",
			DescZh:  "默认开启[采集器列表](datakit-input-conf.md#default-enabled-inputs)，以英文逗号分割，如 `cpu,mem,disk`",
		},
		{
			ENVName: "~~ENV_ENABLE_INPUTS~~",
			Type:    doc.List,
			Desc:    "Same as ENV_DEFAULT_ENABLED_INPUTS(Deprecated)",
			DescZh:  "同 ENV_DEFAULT_ENABLED_INPUTS，将废弃",
		},
		{
			ENVName: "ENV_GLOBAL_HOST_TAGS",
			Type:    doc.List,
			Example: `tag1=val,tag2=val2`,
			Desc:    "Global tag, multiple tags are divided by English commas. The old `ENV_GLOBAL_TAGS` will be discarded",
			DescZh:  "全局 tag，多个 tag 之间以英文逗号分割",
		},
		{
			ENVName: "ENV_PIPELINE_DEFAULT_PIPELINE",
			Type:    doc.Map,
			Example: "`{\"logging\":\"abc.p\",\"metric\":\"xyz.p\"}`",
			Desc: "Set the default Pipeline script for the specified data category. " +
				"This setting takes precedence when it conflicts with the remote setting.",
			DescZh: "为指定数据类别设置默认 Pipeline 脚本；与远程设置冲突时，此设置优先",
		},
		{
			ENVName: "ENV_PIPELINE_DISABLE_HTTP_REQUEST_FUNC",
			Type:    doc.Boolean,
			Desc:    "Disable Pipeline `http_request` function",
			DescZh:  "禁用 Pipeline `http_request` 函数",
		},
		{
			ENVName: "ENV_PIPELINE_HTTP_REQUEST_HOST_WHITELIST",
			Type:    doc.List,
			Desc:    "Set HOST whitelist for `http_request` function",
			DescZh:  "为 `http_request` 函数设置 HOST 白名单",
		},
		{
			ENVName: "ENV_PIPELINE_HTTP_REQUEST_CIDR_WHITELIST",
			Type:    doc.List,
			Desc:    "Set CIDR whitelist for `http_request` function",
			DescZh:  "为 `http_request` 函数设置 CIDR 白名单",
		},
		{
			ENVName: "ENV_PIPELINE_HTTP_REQUEST_DISABLE_INTERNAL_NET",
			Type:    doc.List,
			Desc:    "Disable `http_request` function to access internal network",
			DescZh:  "使 `http_request` 函数禁止访问内部网络",
		},
		{
			ENVName: "~~ENV_GLOBAL_TAGS~~",
			Type:    doc.List,
			Desc:    "Same as ENV_GLOBAL_HOST-TAGS(Deprecated)",
			DescZh:  "同 ENV_GLOBAL_HOST_TAGS，将废弃",
		},

		{
			ENVName: "ENV_K8S_CLUSTER_NODE_NAME",
			Type:    doc.String,
			Desc:    "If we got same node-name among multiple k8s cluster, we can add a prefix based on origin node-name via this ENV",
			DescZh:  "如果多个 k8s cluster 中存在同名的 node-name，可以通过该环境变量在原 node-name 上加一个前缀来区分",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envDataway() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_DATAWAY_DROP_EXPIRED_PACKAGE_AT",
			Type:    doc.TimeDuration,
			Example: "`24h`",
			Default: "`12h`",
			Desc:    "Set package expiration for packages within WAL",
			DescZh:  "设置 WAL 中数据包的超时时间",
		},

		{
			ENVName:  "ENV_DATAWAY",
			Type:     doc.URL,
			Example:  "`https://openway.<<<custom_key.brand_main_domain>>>?token=xxx`",
			Required: doc.Yes,
			Desc:     "Set DataWay address",
			DescZh:   "配置 DataWay 地址",
		},

		{
			ENVName: "ENV_DATAWAY_TIMEOUT",
			Type:    doc.TimeDuration,
			Default: `30s`,
			Desc:    "Set DataWay request timeout",
			DescZh:  "配置 DataWay 请求超时",
		},

		{
			ENVName: "ENV_DATAWAY_ENABLE_HTTPTRACE",
			Type:    doc.Boolean,
			Desc:    "Enable metrics on DataWay HTTP request",
			DescZh:  "开启 DataWay 请求时 HTTP 层面的指标暴露",
		},

		{
			ENVName: "ENV_DATAWAY_HTTP_PROXY",
			Type:    doc.URL,
			Desc:    "Set DataWay HTTP Proxy",
			DescZh:  "设置 DataWay HTTP 代理",
		},

		{
			ENVName: "ENV_DATAWAY_MAX_IDLE_CONNS",
			Type:    doc.Int,
			Desc:    "Set DataWay HTTP connection pool size [:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0)",
			DescZh:  "设置 DataWay HTTP 连接池大小 [:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0)",
		},

		{
			ENVName: "ENV_DATAWAY_IDLE_TIMEOUT",
			Type:    doc.TimeDuration,
			Default: `90s`,
			Desc:    "Set DataWay HTTP Keep-Alive timeout [:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0)",
			DescZh:  "设置 DataWay HTTP Keep-Alive 时长 [:octicons-tag-24: Version-1.7.0](changelog.md#cl-1.7.0)",
		},
		{
			ENVName: "ENV_DATAWAY_MAX_RETRY_COUNT",
			Type:    doc.Int,
			Default: `4`,
			Desc:    "Specify at most how many times the data sending operation will be performed when encounter failures [:octicons-tag-24: Version-1.18.0](changelog.md#cl-1.18.0)",
			DescZh:  "指定当把数据发送到中心时，最多可以发送的次数，最小值为 1（失败后不重试），最大值为 10 [:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)",
		},
		{
			ENVName: "ENV_DATAWAY_RETRY_DELAY",
			Type:    doc.TimeDuration,
			Default: `200ms`,
			Desc:    `The interval between two data sending retry, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h" [:octicons-tag-24: Version-1.18.0](changelog.md#cl-1.18.0)`,
			DescZh:  `数据发送失败时，两次重试之间的时间间隔 [:octicons-tag-24: Version-1.17.0](changelog.md#cl-1.17.0)`,
		},
		{
			ENVName: "ENV_DATAWAY_MAX_RAW_BODY_SIZE",
			Type:    doc.Int,
			Default: `10MB`,
			Desc:    "Set upload package size(before gzip)",
			DescZh:  "数据上传时单包（未压缩）大小",
		},

		{
			ENVName: "ENV_DATAWAY_CONTENT_ENCODING",
			Type:    doc.String,
			Desc:    "Set the encoding of the point data at upload time (optional list: 'v1' is the line protocol, 'v2' is Protobuf)",
			DescZh:  "设置上传时的 point 数据编码（可选列表：`v1` 即行协议，`v2` 即 Protobuf）",
		},

		{
			ENVName: "ENV_DATAWAY_TLS_INSECURE",
			Type:    doc.Boolean,
			Desc:    "Enable self-signed TLS certificate on Dataway [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
			DescZh:  "允许对应的 Dataway 上的证书是自签证书 [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
		},

		// NTP
		{
			ENVName: "ENV_DATAWAY_NTP_INTERVAL",
			Type:    doc.String,
			Desc:    "Set NTP sync interval [:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
			DescZh:  "设置 NTP 时间同步间隔 [:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
		},

		{
			ENVName: "ENV_DATAWAY_NTP_DIFF",
			Type:    doc.String,
			Desc:    "Set NTP sync difference [:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
			DescZh:  "设置 NTP 时间同步的误差 [:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
		},

		{
			ENVName: "ENV_DATAWAY_DISABLE_NTP",
			Type:    doc.Boolean,
			Desc:    "Disable NTP sync       [:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
			DescZh:  "设置 NTP 时间同步的误差[:octicons-tag-24: Version-1.75.0](changelog-2025.md#cl-1.75.0)",
		},

		// WAL
		{
			ENVName: "ENV_DATAWAY_WAL_CAPACITY",
			Type:    doc.Float,
			Desc:    "Set WAL disk cache capacity [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 WAL 占用的磁盘大小 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},

		{
			ENVName: "ENV_DATAWAY_WAL_WORKERS",
			Type:    doc.Int,
			Desc:    "Set WAL workers, default to limited CPU cores X 8 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 WAL worker 个数，默认为 CPU 配额核心数 X 8 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},

		{
			ENVName: "ENV_DATAWAY_WAL_MEM_CAPACITY",
			Type:    doc.Int,
			Desc:    "Set WAL memory queue length, default to limited CPU cores X 8 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 WAL 内存队列长度，默认为 CPU 配额核心数 X 8 [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},

		{
			ENVName: "ENV_DATAWAY_WAL_NO_DROP_CATEGORIES",
			Type:    doc.List,
			Desc:    "Set category list that do not drop data if WAL disk full [:octicons-tag-24: Version-1.71.0](changelog-2025.md#cl-1.71.0)",
			DescZh:  "当 WAL 磁盘限定容量写满时，设置不主动丢弃的数据类型列表 [:octicons-tag-24: Version-1.71.0](changelog-2025.md#cl-1.71.0)",
			Example: "`'L,T,N'`",
		},

		{
			ENVName: "ENV_DATAWAY_WAL_PATH",
			Type:    doc.String,
			Desc:    "Set WAL disk path, default path is *cache/dw-wal* under DataKit install path[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 WAL 磁盘目录，默认为 DataKit 安装目录下的 *cache/dw-wal* [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},

		{
			ENVName: "ENV_DATAWAY_WAL_FAIL_CACHE_CLEAN_INTERVAL",
			Type:    doc.TimeDuration,
			Desc:    "Set WAL fail-cache clean interval, default `30s`[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 WAL 失败队列的重试间隔，默认 `30s` [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envLog() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_GIN_LOG", Type: doc.String, Default: `*/var/log/datakit/gin.log*`, Desc: "If it is changed to `stdout`, the DataKit's own gin log will not be written to the file, but will be output by the terminal.", DescZh: "如果改成 `stdout`，DataKit 自身 gin 日志将不写文件，而是终端输出"},
		{ENVName: "ENV_LOG", Type: doc.String, Default: `*/var/log/datakit/log*`, Desc: "If it is changed to `stdout`, DataKit's own log will not be written to the file, but will be output by the terminal.", DescZh: "如果改成 `stdout`，DataKit 自身日志将不写文件，而是终端输出"},
		{ENVName: "ENV_LOG_LEVEL", Type: doc.String, Default: `info`, Desc: "Set DataKit's own log level, optional `info/debug`(case insensitive).", DescZh: "设置 DataKit 自身日志等级，可选 `info/debug`（不区分大小写）"},
		{ENVName: "ENV_DISABLE_LOG_COLOR", Type: doc.Boolean, Default: `-`, Desc: "Turn off log colors", DescZh: "关闭日志颜色"},
		{ENVName: "ENV_LOG_ROTATE_BACKUP", Type: doc.Int, Default: `5`, Desc: "The upper limit count for log files to be reserve.", DescZh: "设置最多保留日志分片的个数"},
		{ENVName: "ENV_LOG_ROTATE_SIZE_MB", Type: doc.Int, Default: `32`, Desc: "The threshold for automatic log rotating in MB, which automatically switches to a new file when the log file size reaches the threshold.", DescZh: "日志自动切割的阈值（单位：MB），当日志文件大小达到设置的值时，自动切换新的文件"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envPprof() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "~~ENV_ENABLE_PPROF~~", Type: doc.Boolean, Desc: "Whether to start port on for profiling(Deprecated: Default enabled)", DescZh: "是否开启 profiling 端口（已默认启用）"},
		{ENVName: "ENV_PPROF_LISTEN", Type: doc.String, Desc: "`pprof` service listening address", DescZh: "`pprof` 服务监听地址"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envElect() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_ENABLE_ELECTION",
			Type:    doc.Boolean,
			Default: "-",
			Desc:    "If you want to open the [election](election.md), it will not be opened by default. If you want to open it, you can give any non-empty string value to the environment variable.",
			DescZh:  "开启[选举](election.md)，默认不开启，如需开启，给该环境变量任意一个非空字符串值即可",
		},
		{
			ENVName: "ENV_NAMESPACE",
			Type:    doc.String,
			Default: "default",
			Desc:    "The namespace in which the DataKit resides, which defaults to null to indicate that it is namespace-insensitive and accepts any non-null string, such as `dk-namespace-example`. If the election is turned on, you can specify the workspace through this environment variable.",
			DescZh:  "DataKit 所在的命名空间，默认为空表示不区分命名空间，接收任意非空字符串，如 `dk-namespace-example`。如果开启了选举，可以通过此环境变量指定工作空间。",
		},
		{
			ENVName: "ENV_ENABLE_ELECTION_NAMESPACE_TAG",
			Type:    doc.Boolean,
			Default: "-",
			Desc:    "When this option is turned on, all election classes are collected with an extra tag of `election_namespace=<your-election-namespace>`, which may result in some timeline growth [:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7)",
			DescZh:  "开启该选项后，所有选举类的采集均会带上 `election_namespace=<your-election-namespace>` 的额外 tag，这可能会导致一些时间线的增长 [:octicons-tag-24: Version-1.4.7](changelog.md#cl-1.4.7)",
		},

		{
			ENVName: "ENV_GLOBAL_ELECTION_TAGS",
			Type:    doc.List,
			Example: "tag1=val,tag2=val2",
			Desc:    "Tags are elected globally, and multiple tags are divided by English commas. ENV_GLOBAL_ENV_TAGS will be discarded.",
			DescZh:  "全局选举 tag，多个 tag 之间以英文逗号分割。ENV_GLOBAL_ENV_TAGS 将被弃用",
		},

		{
			ENVName: "ENV_CLUSTER_NAME_K8S",
			Type:    doc.String,
			Default: "default",
			Desc:    "The cluster name in which the DataKit residers, if the cluster is not empty, a specified tag will be added to [global election tags](election.md#global-tags), the key is `cluster_name_k8s` and the value is the environment variable [:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)",
			DescZh:  "DataKit 所在的 cluster，如果非空，会在 [Global Election Tags](election.md#global-tags) 中添加一个指定 tag，key 是 `cluster_name_k8s`，value 是环境变量的值 [:octicons-tag-24: Version-1.5.8](changelog.md#cl-1.5.8)",
		},
		{
			ENVName: "ENV_ELECTION_NODE_WHITELIST",
			Type:    doc.List,
			Default: "[]",
			Desc:    "List of node names that are allowed to participate in elections [:octicons-tag-24: Version-1.35.0](changelog.md#cl-1.35.0)",
			DescZh:  "允许参加选举的节点名称列表 [:octicons-tag-24: Version-1.35.0](changelog.md#cl-1.35.0)",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envHTTPAPI() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_DISABLE_404PAGE",
			Type:    doc.Boolean,
			Default: "-",
			Desc:    "Disable the DataKit 404 page (commonly used when deploying DataKit RUM on the public network).",
			DescZh:  "禁用 DataKit 404 页面（公网部署 DataKit RUM 时常用）。",
		},
		{
			ENVName: "ENV_HTTP_LISTEN",
			Type:    doc.String,
			Default: "localhost:9529",
			Desc:    "The address can be modified so that the [DataKit interface](apis.md) can be called externally.",
			DescZh:  "可修改地址，使得外部可以调用 [DataKit 接口](apis.md)。",
		},
		{
			ENVName: "ENV_HTTP_LISTEN_SOCKET",
			Type:    doc.String,
			Example: "`/var/run/datakit/datakit.sock`",
			Desc:    "The address can be modified so that the [DataKit interface](apis.md) can be called externally.",
			DescZh:  "可修改地址，使得外部可以通过 unix socket 调用 [DataKit 接口](apis.md)。",
		},

		{
			ENVName: "ENV_HTTP_PUBLIC_APIS",
			Type:    doc.List,
			Desc:    "[API list](apis.md) that allow external access, separated by English commas between multiple APIs. When DataKit is deployed on the public network, it is used to disable some APIs.",
			DescZh:  "允许外部访问的 DataKit [API 列表](apis.md)，多个 API 之间以英文逗号分割。当 DataKit 部署在公网时，用来禁用部分 API。",
		},

		{
			ENVName: "ENV_HTTP_TIMEOUT",
			Type:    doc.TimeDuration,
			Default: "30s",
			Desc:    "Setting the 9529 HTTP API Server Timeout [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)",
			DescZh:  "设置 9529 HTTP API 服务端超时时间 [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)",
		},

		{
			ENVName: "ENV_HTTP_CLOSE_IDLE_CONNECTION",
			Type:    doc.Boolean,
			Default: "-",
			Desc:    "If turned on, the 9529 HTTP server actively closes idle connections(idle time equal to `ENV_HTTP_TIMEOUT`) [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)",
			DescZh:  "如果开启，则 9529 HTTP server 会主动关闭闲置连接（闲置时间等同于 `ENV_HTTP_TIMEOUT`） [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)",
		},

		{
			ENVName: "ENV_HTTP_ENABLE_TLS",
			Type:    doc.Boolean,
			Default: "-",
			Desc:    "Enable DataKit 9529 HTTPS [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
			DescZh:  "开启 DataKit 9529 HTTPS [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
		},

		{
			ENVName: "ENV_HTTP_TLS_CRT",
			Type:    doc.String,
			Default: "-",
			Desc:    "Set DataKit HTTP Server's TLS cert path [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
			DescZh:  "配置 DataKit HTTP Server 上的 TLS cert 路径 [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
		},

		{
			ENVName: "ENV_HTTP_TLS_KEY",
			Type:    doc.String,
			Default: "-",
			Desc:    "Set DataKit HTTP Server's TLS key path [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
			DescZh:  "配置 DataKit HTTP Server 上的 TLS key 路径 [:octicons-tag-24: Version-1.29.0](changelog.md#cl-1.29.0)",
		},

		{
			ENVName: "ENV_REQUEST_RATE_LIMIT",
			Type:    doc.Float,
			Default: "`100.0`",
			Desc:    "Limit 9529 [API requests per second](datakit-conf.md#set-http-api-limit).",
			DescZh:  "限制 9529 [API 每秒请求数](datakit-conf.md#set-http-api-limit)。",
		},

		{
			ENVName: "ENV_REQUEST_RATE_LIMIT_TTL",
			Type:    doc.TimeDuration,
			Default: "`60s`",
			Desc:    "Rate limit duration window",
			DescZh:  "9529 端口 HTTP API 限流窗口。",
		},

		{
			ENVName: "ENV_REQUEST_RATE_LIMIT_BURST",
			Type:    doc.Int,
			Default: "`500`",
			Desc:    "Rate limit burst",
			DescZh:  "9529 端口 HTTP API 限流允许的突发流量。如果设置成 0 则禁止突发请求数。",
		},

		{
			ENVName: "ENV_RUM_ORIGIN_IP_HEADER",
			Type:    doc.String,
			Default: "`X-Forwarded-For`",
			Desc:    "Set RUM HTTP request(`/v1/write/rum`) real IP forward header key.",
			DescZh:  "设置 RUM 请求中真实 IP forward 对应的 HTTP header key。DataKit 将从该 Header 上获取端上用户的真实 IP，否则拿到可能是网关 IP。",
		},

		{
			ENVName: "ENV_RUM_APP_ID_WHITE_LIST",
			Type:    doc.String,
			Example: "/appid-1,/appid-2",
			Desc:    "RUM app-id white list, split by `,`.",
			DescZh:  "RUM app-id 白名单列表，以 `,` 分割。",
		},

		{
			ENVName: "ENV_HTTP_ALLOWED_CORS_ORIGINS",
			Type:    doc.List,
			Default: "-",
			Example: "Origin,Access-Control-Allow-Origin,Access-Control-Allow-Methods",
			Desc:    "Setup CORS on DataKit HTTP APIs(split by `,`) [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "设置 DataKit API CORS 属性（英文逗号分割）[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envConfd() []*inputs.ENVInfo {
	// See also: https://github.com/kelseyhightower/confd/blob/master/docs/configuration-guide.md
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_CONFD_BACKEND", Type: doc.String, Example: "`etcdv3`", Desc: "The backend to use", DescZh: "要使用的后端"},
		{ENVName: "ENV_CONFD_BASIC_AUTH", Type: doc.Boolean, Default: "false", Desc: "Use Basic Auth to authenticate (used with `etcdv3`/consul)", DescZh: "使用 Basic Auth 进行身份验证（适用于 `etcdv3`/consul）"},
		{ENVName: "ENV_CONFD_CLIENT_CA_KEYS", Type: doc.String, Example: "`/opt/ca.crt`", Desc: "The client CA key file", DescZh: "客户端 CA 密钥文件"},
		{ENVName: "ENV_CONFD_CLIENT_CERT", Type: doc.String, Example: "`/opt/client.crt`", Desc: "The client cert file", DescZh: "客户端证书文件"},
		{ENVName: "ENV_CONFD_CLIENT_KEY", Type: doc.String, Example: "`/opt/client.key`", Desc: "The client key file", DescZh: "客户端密钥文件"},
		{
			ENVName: "ENV_CONFD_BACKEND_NODES",
			Type:    doc.JSON,
			Example: "`[\"http://aaa:2379\",\"1.2.3.4:2379\"]` (`Nacos must prefix http:// or https://`)",
			Desc:    "Backend source address",
			DescZh:  "后端源地址",
		},
		{ENVName: "ENV_CONFD_USERNAME", Type: doc.String, Desc: "The username to authenticate (used with `etcdv3/consul/nacos`)", DescZh: "身份验证的用户名（适用于 `etcdv3/consul/nacos`）"},
		{ENVName: "ENV_CONFD_PASSWORD", Type: doc.String, Desc: "The password to authenticate (used with `etcdv3/consul/nacos`)", DescZh: "身份验证的密码（适用于 `etcdv3/consul/nacos`）"},
		{ENVName: "ENV_CONFD_SCHEME", Type: doc.String, Example: "http/https", Desc: "The backend URI scheme", DescZh: "后端 URI 方案"},
		{ENVName: "ENV_CONFD_SEPARATOR", Type: doc.String, Default: "/", Desc: "The separator to replace '/' with when looking up keys in the backend, prefixed '/' will also be removed (used with rides)", DescZh: "在后端查找键时替换'/'的分隔符，前缀'/'也将被删除（适用于 redis）"},
		{ENVName: "ENV_CONFD_ACCESS_KEY", Type: doc.String, Desc: "Access Key Id (use with `nacos/aws`)", DescZh: "客户端身份 ID（适用于 `nacos/aws`）"},
		{ENVName: "ENV_CONFD_SECRET_KEY", Type: doc.String, Desc: "Secret Access Key (use with `nacos/aws`)", DescZh: "认证密钥（适用于 `nacos/aws`）"},
		{ENVName: "ENV_CONFD_CIRCLE_INTERVAL", Type: doc.Int, Default: "60", Desc: "Loop detection interval second (use with `nacos/aws`)", DescZh: "循环检测间隔秒数（适用于 `nacos/aws`）"},
		{ENVName: "ENV_CONFD_CONFD_NAMESPACE", Type: doc.String, Example: "`6aa36e0e-bd57-4483-9937-e7c0ccf59599`", Desc: "`confd` namespace id (use with `nacos`)", DescZh: "配置信息的空间 ID（适用于 `nacos`）"},
		{ENVName: "ENV_CONFD_PIPELINE_NAMESPACE", Type: doc.String, Example: "`d10757e6-aa0a-416f-9abf-e1e1e8423497`", Desc: "`pipeline` namespace id (use with `nacos`)", DescZh: "`pipeline` 的信息空间 ID（适用于 `nacos`）"},
		{ENVName: "ENV_CONFD_REGION", Type: doc.String, Example: "`cn-north-1`", Desc: "AWS Local Zone (use with aws)", DescZh: "AWS 服务区（适用于 aws）"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envGit() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_GIT_BRANCH", Type: doc.String, Example: "master", Desc: "Specifies the branch to pull. **If it is empty, it is the default.** And the default is the remotely specified main branch, which is usually `master`.", DescZh: "指定拉取的分支。**为空则是默认**，默认是远程指定的主分支，一般是 `master`"},
		{ENVName: "ENV_GIT_INTERVAL", Type: doc.TimeDuration, Example: "1m", Desc: "The interval of timed pull.", DescZh: "定时拉取的间隔"},
		{ENVName: "ENV_GIT_KEY_PATH", Type: doc.String, Example: "/Users/username/.ssh/id_rsa", Desc: "The full path of the local PrivateKey.", DescZh: "本地 PrivateKey 的全路径"},
		{ENVName: "ENV_GIT_KEY_PW", Type: doc.String, Example: "passwd", Desc: "Use password of local PrivateKey.", DescZh: "本地 PrivateKey 的使用密码"},
		{ENVName: "ENV_GIT_URL", Type: doc.URL, Example: "`http://username:password@github.com/username/repository.git`", Desc: "Manage the remote git repo address of the configuration file.", DescZh: "管理配置文件的远程 git repo 地址"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envSinker() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_SINKER_GLOBAL_CUSTOMER_KEYS", Type: doc.String, Desc: "Sinker Global Customer Key list, keys are split with `,`", DescZh: "指定 Sinker 分流的自定义字段列表，各个 Key 之间以英文逗号分割"},
		{ENVName: "ENV_DATAWAY_ENABLE_SINKER", Type: doc.Boolean, Default: "-", Desc: "Enable DataWay Sinker [:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)", DescZh: "开启 DataWay 发送数据时的 Sinker 功能。该功能需新版本 Dataway 才能生效 [:octicons-tag-24: Version-1.14.0](changelog.md#cl-1.14.0)"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envIO() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_IO_AUTO_TIMESTAMP_CORRECTION",
			Type:    doc.Boolean,
			Desc:    "Enable/disable point time correction",
			DescZh:  "开启/关闭 point 时间戳自动校准（默认开启）",
		},

		{
			ENVName: "ENV_IO_FILTERS",
			Type:    doc.JSON,
			Desc:    "Add [line protocol filter](datakit-filter.md)",
			DescZh:  "添加[行协议过滤器](datakit-filter.md)",
		},
		{
			ENVName: "ENV_IO_FLUSH_INTERVAL",
			Type:    doc.TimeDuration,
			Default: "10s",
			Desc:    "Set compact interval [:octicons-tag-24: Version-1.22.0](changelog.md#cl-1.22.0)",
			DescZh:  "设置 compact 执行间隔 [:octicons-tag-24: Version-1.22.0](changelog.md#cl-1.22.0)",
		},
		{
			ENVName: "ENV_IO_FEED_CHAN_SIZE",
			Type:    doc.Int,
			Default: "1",
			Desc:    "Set compact queue size [:octicons-tag-24: Version-1.22.0](changelog.md#cl-1.22.0)",
			DescZh:  "设置 compact 队列长度 [:octicons-tag-24: Version-1.22.0](changelog.md#cl-1.22.0)",
		},
		{
			ENVName: "ENV_IO_FLUSH_WORKERS",
			Type:    doc.Int,
			Desc:    "Set compact workers, default to limited CPU cores x 2 [:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)",
			DescZh:  "设置 compactor worker 数，默认为 CPU 配额核心数 x 2 [:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)",
		},

		{
			ENVName: "ENV_IO_MAX_CACHE_COUNT",
			Type:    doc.Int,
			Default: "1024",
			Desc:    "Compact buffer size",
			DescZh:  "Compact 缓存的点数",
		},

		{
			ENVName: "~~ENV_IO_ENABLE_CACHE~~",
			Type:    doc.Boolean,
			Default: "false",
			Desc:    "Whether to open the disk cache that failed to send. Removed in [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "是否开启发送失败的磁盘缓存。[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 版本已移除",
		},
		{
			ENVName: "~~ENV_IO_CACHE_ALL~~",
			Type:    doc.Boolean,
			Default: "false",
			Desc:    "Cache failed data points of all categories. Removed in [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "是否 cache 所有发送失败的数据。[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 版本已移除",
		},
		{
			ENVName: "~~ENV_IO_CACHE_MAX_SIZE_GB~~",
			Type:    doc.Int,
			Default: "10",
			Desc:    "Disk size of send failure cache (in GB). Removed in [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "发送失败缓存的磁盘大小（单位 GB）。[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 版本已移除",
		},
		{
			ENVName: "~~ENV_IO_CACHE_CLEAN_INTERVAL~~",
			Type:    doc.TimeDuration,
			Default: "5s",
			Desc:    "Periodically send failed tasks cached on disk. Removed in [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)",
			DescZh:  "定期发送缓存在磁盘内的失败任务。[:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0) 版本已移除",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envDca() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_DCA_WEBSOCKET_SERVER", Type: doc.URL, Default: "", Desc: "The server address that the the DataKit can connect to. Once `ENV_DCA_WEBSOCKET_SERVER` is set, the DCA function is enabled by default", DescZh: "DataKit 可以连接该地址，使得 DCA 可以进行管理该 DataKit，一旦开启 ENV_DCA_WEBSOCKET_SERVER 即默认启用 DCA 功能"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envRefta() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_REFER_TABLE_URL", Type: doc.String, Desc: "Set the data source URL", DescZh: "设置数据源 URL"},
		{ENVName: "ENV_REFER_TABLE_PULL_INTERVAL", Type: doc.String, Default: "5m", Desc: "Set the request interval for the data source URL", DescZh: "设置数据源 URL 的请求时间间隔"},
		{ENVName: "ENV_REFER_TABLE_USE_SQLITE", Type: doc.Boolean, Default: "false", Desc: "Set whether to use SQLite to save data", DescZh: "设置是否使用 SQLite 保存数据"},
		{ENVName: "ENV_REFER_TABLE_SQLITE_MEM_MODE", Type: doc.Boolean, Default: "false", Desc: "When using SQLite to save data, use SQLite memory mode/disk mode", DescZh: "当使用 SQLite 保存数据时，使用 SQLite 内存模式/磁盘模式"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envRecorder() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_ENABLE_RECORDER", Type: doc.Boolean, Default: "false", Desc: "To enable or disable recorder", DescZh: "设置是否开启数据录制"},
		{ENVName: "ENV_RECORDER_PATH", Type: doc.String, Default: "*[DataKit install path]/recorder*", Desc: "Set recorder data path", DescZh: "设置数据录制的存放目录"},
		{ENVName: "ENV_RECORDER_ENCODING", Type: doc.String, Default: "v2", Desc: "Set recorder format. v1 is lineprotocol, v2 is JSON", DescZh: "设置数据录制的存放格式，v1 为行协议格式，v2 为 JSON 格式"},
		{ENVName: "ENV_RECORDER_DURATION", Type: doc.TimeDuration, Default: "30m", Desc: "Set recorder duration(since DataKit start). After the duration, the recorder will stop to write data to file", DescZh: "设置数据录制时长（自 DataKit 启动以后），一旦超过该时长，则不再录制"},
		{ENVName: "ENV_RECORDER_INPUTS", Type: doc.List, Example: "cpu,mem,disk", Desc: "Set allowed input names for recording, split by comma", DescZh: "设置录制的采集器名称列表，以英文逗号分割"},
		{ENVName: "ENV_RECORDER_CATEGORIES", Type: doc.List, Example: "metric,logging,object", Desc: "Set allowed categories for recording, split by comma, full list of categories see [here](apis.md#category)", DescZh: "设置录制的数据分类列表，以英文逗号分割，完整的 Category 列表参见[这里](apis.md#category)"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envOthers() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{ENVName: "ENV_CLOUD_PROVIDER", Type: doc.String, Example: "`aliyun/aws/tencent/hwcloud/azure`", Desc: "Support filling in cloud suppliers during installation", DescZh: "支持安装阶段填写云厂商"},
		{ENVName: "ENV_HOSTNAME", Type: doc.String, Desc: "The default is the local host name, which can be specified at installation time, such as, `dk-your-hostname`", DescZh: "默认为本地主机名，可安装时指定，如， `dk-your-hostname`"},
		{ENVName: "ENV_IPDB", Type: doc.String, Desc: "Specify the IP repository type, currently only supports `iploc/geolite2`", DescZh: "指定 IP 信息库类型，目前只支持 `iploc/geolite2` 两种"},
		{ENVName: "ENV_ULIMIT", Type: doc.Int, Desc: "Specify the maximum number of open files for DataKit", DescZh: "指定 DataKit 最大的可打开文件数"},
		{ENVName: "ENV_PIPELINE_OFFLOAD_RECEIVER", Type: doc.String, Default: "`datakit-http`", Desc: "Set offload receiver", DescZh: "设置 Offload 目标接收器的类型"},
		{ENVName: "ENV_PIPELINE_OFFLOAD_ADDRESSES", Type: doc.List, Example: "`http://aaa:123,http://1.2.3.4:1234`", Desc: "Set offload addresses", DescZh: "设置 Offload 目标地址"},
		{ENVName: "ENV_PIPELINE_DISABLE_APPEND_RUN_INFO", Type: doc.Boolean, Default: "`false`", Desc: "Disable appending the Pipeline run info", DescZh: "禁用追加 Pipeline 运行信息"},
		{ENVName: "ENV_CRYPTO_AES_KEY", Type: doc.String, Example: "`0123456789abcdef`", Desc: "The crypto key(len 16)", DescZh: "AES 加解密的 key 长度是 16"},
		{ENVName: "ENV_CRYPTO_AES_KEY_FILE", Type: doc.String, Example: "`/usr/local/datakit/enc4mysql`", Desc: "File path for storing AES encryption and decryption key", DescZh: "AES 加解密的 key 存放的文件路径"},

		{ENVName: "ENV_LOGGING_MAX_OPEN_FILES", Type: doc.Int, Example: "`1000`", Desc: "Specify the maximum number of open files for logging collection, if the value is -1 then there is no limit, default 500", DescZh: "指定日志采集的最大文件个数，如果值是 -1 则没有限制，默认值 500"},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

// nolint:deadcode,unused
func envPointPool() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			ENVName: "~~ENV_ENABLE_POINT_POOL~~",
			Type:    doc.Boolean,
			Example: "`on`",
			Desc:    "Enable point pool [:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0) default enabled",
			DescZh:  "开启 point pool [:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0) 版本已默认开启",
		},
		{
			ENVName: "ENV_DISABLE_POINT_POOL",
			Type:    doc.Boolean,
			Example: "`on`",
			Desc:    "Disable point pool [:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0)",
			DescZh:  "禁用 point pool [:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0)",
		},
		{
			ENVName: "ENV_POINT_POOL_RESERVED_CAPACITY",
			Type:    doc.Int,
			Default: "4096",
			Desc:    "Specify the pool size that is immune to GC. if DataKit got too many data to collect, we can increase this reserved pool size(such as 40960) to decrease GC payload.",
			DescZh:  "指定 point pool 常驻大小，即不会被 GC 回收的 point 保有数，如果所在 DataKit 的采集量非常大，可以酌情将该数值调大一点",
		},
	}

	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}

func envRemoteJob() []*inputs.ENVInfo {
	infos := []*inputs.ENVInfo{
		{
			ENVName: "ENV_REMOTE_JOB_ENABLE",
			Type:    doc.Boolean,
			Default: "false",
			Example: "`true`",
			Desc:    "Enable remote job",
			DescZh:  "开启 remote job 功能",
		},
		{
			ENVName: "ENV_REMOTE_JOB_ENVS",
			Type:    doc.String,
			Default: "false",
			Example: "`true`",
			Desc:    "Mainly used to send generated files to OSS",
			DescZh:  "主要作用于将生成的文件发送到 OSS.",
		},
		{
			ENVName: "ENV_REMOTE_JOB_INTERVAL",
			Type:    doc.String,
			Default: "10s",
			Example: "10s",
			Desc:    "Regularly request the server to obtain tasks, with a default of 10 seconds",
			DescZh:  "定时请求服务端获取任务，默认 10 秒",
		},
	}
	for idx := range infos {
		infos[idx].DocType = doc.NonInput
	}

	return doc.SetENVDoc("", infos)
}
