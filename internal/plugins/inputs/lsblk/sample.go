// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package lsblk

const sampleCfg = `
[[inputs.lsblk]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  # exclude_device = ['/dev/sda1','/dev/sda2']

[inputs.lsblk.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
