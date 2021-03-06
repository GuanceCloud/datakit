// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	EnableElection = false

	// 对于 DataKit 自己主动采集的数据，如果数据中没有带上时间戳，那么可以用下面的这些全局
	// point-option。

	// 对于数据中带有时间戳（不管是不是主动采集）的，就不要用下面的这些 point-option 了，
	// 因为这些 option 中均未带上时间，建议采用原始数据中的时间戳，比如 tracing 类 HTTP
	// 外部主动打上来的数据。当然，如果原始数据中的时间戳不太重要（比如 prom 类），也可以
	// 使用这些全局 point-option.

	// 选举类 point-option，它们只会带上 global-env-tag(config.GlobalEnvTags).
	optMetricElection  = &PointOption{GlobalEnvTags: true, Category: datakit.Metric}
	optLoggingElection = &PointOption{GlobalEnvTags: true, Category: datakit.Logging}
	optObjectElection  = &PointOption{GlobalEnvTags: true, Category: datakit.Object}
	optNetworkElection = &PointOption{GlobalEnvTags: true, Category: datakit.Network}
	optProfileElection = &PointOption{GlobalEnvTags: true, Category: datakit.Profile}

	// 非选举类 point-option，它们只会带上 global-host-tag(config.GlobalHostTags).
	optLogging = &PointOption{Category: datakit.Logging}
	optMetric  = &PointOption{Category: datakit.Metric}
	optNetwork = &PointOption{Category: datakit.Network}
	optObject  = &PointOption{Category: datakit.Object}
	optProfile = &PointOption{Category: datakit.Profile}

	// TODO: 其它类数据（CO/S/E/R/T）可在此追加...
)

func MOptElection() *PointOption {
	if EnableElection {
		return optMetricElection
	} else {
		return optMetric
	}
}

func LOptElection() *PointOption {
	if EnableElection {
		return optLoggingElection
	} else {
		return optLogging
	}
}

func OOptElection() *PointOption {
	if EnableElection {
		return optObjectElection
	} else {
		return optObject
	}
}

func POptElection() *PointOption {
	if EnableElection {
		return optProfileElection
	} else {
		return optProfile
	}
}

func NOptElection() *PointOption {
	if EnableElection {
		return optNetworkElection
	} else {
		return optNetwork
	}
}

func LOpt() *PointOption { return optLogging }
func MOpt() *PointOption { return optMetric }
func NOpt() *PointOption { return optNetwork }
func OOpt() *PointOption { return optObject }
func POpt() *PointOption { return optProfile }
