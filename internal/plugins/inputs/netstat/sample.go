// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netstat

const sampleCfg = `
[[inputs.netstat]]
  ##(Optional) Collect interval, default is 10 seconds
  interval = '10s'

## The ports you want display
## Can add tags too
# [[inputs.netstat.addr_ports]]
  # ports = ["80","443"]

## Groups of ports and add different tags to facilitate statistics
# [[inputs.netstat.addr_ports]]
  # ports = ["80","443"]
# [inputs.netstat.addr_ports.tags]
  # service = "http"

# [[inputs.netstat.addr_ports]]
  # ports = ["9529"]
# [inputs.netstat.addr_ports.tags]
  # service = "datakit"
  # foo = "bar"

## Server may have multiple network cards
## Display only some network cards
## Can add tags too
# [[inputs.netstat.addr_ports]]
  # ports = ["1.1.1.1:80","2.2.2.2:80"]
  # ports_match is preferred if both ports and ports_match configured
  # ports_match = ["*:80","*:443"]

[inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
