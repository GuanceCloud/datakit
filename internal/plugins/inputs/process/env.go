// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{
			FieldName: "OpenMetric",
			Type:      doc.Boolean,
			Default:   `false`,
			Desc:      "Enable process metric collecting",
			DescZh:    "开启进程指标采集",
		},
		{
			FieldName: "MatchedProcessNames",
			ENVName:   "PROCESS_NAME",
			Type:      doc.List,
			Example:   "`.*datakit.*,guance`",
			Desc:      "Whitelist of process",
			DescZh:    "进程名白名单",
		},
		{
			FieldName: "RunTime",
			ENVName:   "MIN_RUN_TIME",
			Type:      doc.TimeDuration,
			Default:   `10m`,
			Desc:      "Process minimal run time",
			DescZh:    "进程最短运行时间",
		},
		{
			FieldName: "ListenPorts",
			ENVName:   "ENABLE_LISTEN_PORTS",
			Type:      doc.Boolean,
			Default:   `false`,
			Desc:      "Enable listen ports tag",
			DescZh:    "启用监听端口标签",
		},
		{
			FieldName: "Tags",
		},

		{
			FieldName: "OnlyContainerProcesses",
			ENVName:   "ONLY_CONTAINER_PROCESSES",
			Type:      doc.Boolean,
			Default:   `false`,
			Desc:      "Only collect container process for metric and object",
			DescZh:    "只采集容器进程的指标和对象",
		},

		{
			FieldName: "MetricInterval",
			ENVName:   "METRIC_INTERVAL",
			Type:      doc.TimeDuration,
			Default:   `30s`,
			Desc:      "Collect interval on metric",
			DescZh:    "指标采集间隔",
		},

		{
			FieldName: "ObjectInterval",
			ENVName:   "object_interval",
			Type:      doc.TimeDuration,
			Default:   `300s`,
			Desc:      "Collect interval on object",
			DescZh:    "对象采集间隔",
		},
	}

	return doc.SetENVDoc("ENV_INPUT_HOST_PROCESSES_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_OPEN_METRIC : booler   // deprecated
//	ENV_INPUT_HOST_PROCESSES_OPEN_METRIC : booler
//	ENV_INPUT_HOST_PROCESSES_TAGS : "a=b,c=d"
//	ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
//	ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
//	ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS : booler
func (ipt *Input) ReadEnv(envs map[string]string) {
	// deprecated
	if open, ok := envs["ENV_INPUT_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			ipt.OpenMetric = b
		}
	}

	if open, ok := envs["ENV_INPUT_HOST_PROCESSES_OPEN_METRIC"]; ok {
		b, err := strconv.ParseBool(open)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_OPEN_METRIC to bool: %s, ignore", err)
		} else {
			ipt.OpenMetric = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_PROCESSES_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	//   ENV_INPUT_HOST_PROCESSES_PROCESS_NAME : []string
	//   ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME : datakit.Duration
	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_PROCESS_NAME"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add PROCESS_NAME from ENV: %v", arrays)
		ipt.MatchedProcessNames = append(ipt.MatchedProcessNames, arrays...)
	}

	if str, ok := envs["ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_MIN_RUN_TIME to time.Duration: %s, ignore", err)
		} else {
			ipt.RunTime.Duration = config.ProtectedInterval(minObjectInterval,
				maxObjectInterval,
				da)
		}
	}

	if port, ok := envs["ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS"]; ok {
		b, err := strconv.ParseBool(port)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_ENABLE_LISTEN_PORTS to bool: %s, ignore", err)
		} else {
			ipt.ListenPorts = b
		}
	}

	if x, ok := envs["ENV_INPUT_HOST_PROCESSES_ONLY_CONTAINER_PROCESSES"]; ok {
		b, err := strconv.ParseBool(x)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_ONLY_CONTAINER_PROCESSES to bool: %s, ignore", err)
		} else {
			ipt.OnlyContainerProcesses = b
		}
	}

	if x, ok := envs["ENV_INPUT_HOST_PROCESSES_METRIC_INTERVAL"]; ok {
		du, err := time.ParseDuration(x)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_METRIC_INTERVAL: %s, ignore", err)
		} else {
			ipt.MetricInterval = datakit.Duration{Duration: du}
		}
	}

	if x, ok := envs["ENV_INPUT_HOST_PROCESSES_OBJECT_INTERVAL"]; ok {
		du, err := time.ParseDuration(x)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOST_PROCESSES_OBJECT_INTERVAL: %s, ignore", err)
		} else {
			ipt.ObjectInterval = datakit.Duration{Duration: du}
		}
	}
}
