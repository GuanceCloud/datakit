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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "OpenMetric", Type: doc.Boolean, Default: `false`, Desc: "Enable process metric collecting", DescZh: "开启进程指标采集"},
		{FieldName: "MatchedProcessNames", ENVName: "PROCESS_NAME", Type: doc.List, Example: "`\".*datakit.*\", \"guance\"`", Desc: "Whitelist of process", DescZh: "进程名白名单"},
		{FieldName: "RunTime", ENVName: "MIN_RUN_TIME", Type: doc.TimeDuration, Default: `10m`, Desc: "Process minimal run time", DescZh: "进程最短运行时间"},
		{FieldName: "ListenPorts", ENVName: "ENABLE_LISTEN_PORTS", Type: doc.Boolean, Default: `false`, Desc: "Enable listen ports tag", DescZh: "启用监听端口标签"},
		{FieldName: "Tags"},
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
}
