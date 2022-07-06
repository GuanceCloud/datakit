// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	// 对于 DataKit 自己主动采集的数据，如果数据中没有带上时间戳，那么可以用下面的这些全局
	// point-option。

	// 对于数据中带有时间戳（不管是不是主动采集）的，就不要用下面的这些 point-option 了，
	// 因为这些 option 中均未带上时间，建议采用原始数据中的时间戳，比如 tracing 类 HTTP
	// 外部主动打上来的数据。当然，如果原始数据中的时间戳不太重要（比如 prom 类），也可以
	// 使用这些全局 point-option.

	// 选举类 point-option，它们只会带上 global-env-tag(config.GlobalEnvTags).
	OptElectionMetric  = &io.PointOption{GlobalEnvTags: true, Category: datakit.Metric}
	OptElectionLogging = &io.PointOption{GlobalEnvTags: true, Category: datakit.Logging}
	OptElectionObject  = &io.PointOption{GlobalEnvTags: true, Category: datakit.Object}

	// 非选举类 point-option，它们只会带上 global-host-tag(config.GlobalHostTags).
	OptMetric  = &io.PointOption{Category: datakit.Metric}
	OptLogging = &io.PointOption{Category: datakit.Logging}
	OptObject  = &io.PointOption{Category: datakit.Object}
	OptNetwork = &io.PointOption{Category: datakit.Network}
	OptProfile = &io.PointOption{Category: datakit.Profile}
)
