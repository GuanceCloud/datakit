// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package w32time

const sampleCfg = `
[[inputs.w32time]]
  ## Collect interval, default is 60 seconds.
  interval = "60s"

[inputs.w32time.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
