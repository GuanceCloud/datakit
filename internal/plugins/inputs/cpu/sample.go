// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cpu

const sampleCfg = `
[[inputs.cpu]]
  ## Collect interval, default is 10 seconds. (optional)
  interval = '10s'

  ## Collect CPU usage per core, default is false. (optional)
  percpu = false

  ## Setting disable_temperature_collect to false will collect cpu temperature stats for linux. (deprecated)
  # disable_temperature_collect = false

  ## Enable to collect core temperature data.
  enable_temperature = true

  ## Enable gets average load information every five seconds.
  enable_load5s = true

[inputs.cpu.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
