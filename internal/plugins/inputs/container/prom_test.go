// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProm(t *T.T) {
	t.Run("basic", func(t *T.T) {
		txt := `
[[inputs.prom]]
  ## Exporter 地址
  urls = ["http://10.3.7.85:9527/v1/metric?metrics_api_key=apikey_5577006791947779410"]

  ## 采集器别名
 source = "kodo-prom"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  # metric_types = ["counter","gauge"]
  metric_types = []

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  # metric_name_filter = ["cpu"]

  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = ""

  ## 指标集名称
  # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  # 如果配置measurement_name, 则不进行指标名称的切割
  # 最终的指标集名称会添加上measurement_prefix前缀
  # measurement_name = "prom"

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## 过滤tags, 可配置多个tag
  # 匹配的tag将被忽略
  # tags_ignore = ["xxxx"]

  ## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  [[inputs.prom.measurements]]
    prefix = "kodo_api_"
    name = "kodo_api"

 [[inputs.prom.measurements]]
   prefix = "kodo_workers_"
   name = "kodo_workers"

 [[inputs.prom.measurements]]
   prefix = "kodo_workspace_"
   name = "kodo_workspace"

 [[inputs.prom.measurements]]
   prefix = "kodo_dql_"
   name = "kodo_dql"

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
		`
		r, err := newPromRunnerWithTomlConfig(txt)
		assert.NoError(t, err)
		t.Logf("source: %s", r[0].conf.Source)
	})
}
