// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

const sampleCfg = `
[[inputs.hostdir]]
  interval = "10s"

  # directory to collect
  # Windows example: C:\\Users
  # UNIX-like example: /usr/local/
  dir = "" # required

  # optional, i.e., "*.exe", "*.so"
  exclude_patterns = []

[inputs.hostdir.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
