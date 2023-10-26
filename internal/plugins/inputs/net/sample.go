// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

const sampleCfg = `
[[inputs.net]]
  ## (optional) collect interval, default is 10 seconds
  interval = '10s'

  ## By default, gathers stats from any up interface, but Linux does not contain virtual interfaces.
  ## Setting interfaces using regular expressions will collect these expected interfaces.
  # interfaces = ['''eth[\w-]+''', '''lo''', ]

  ## Datakit does not collect network virtual interfaces under the linux system.
  ## Setting enable_virtual_interfaces to true will collect virtual interfaces stats for linux.
  # enable_virtual_interfaces = true

  ## On linux systems also collects protocol stats.
  ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
  # ignore_protocol_stats = false

[inputs.net.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
