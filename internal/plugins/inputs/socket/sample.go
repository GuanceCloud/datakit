// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

const (
	sample = `
[[inputs.socket]]
  ## Support TCP/UDP.
  ## If the quantity to be detected is too large, it is recommended to open more collectors
  dest_url = [
    "tcp://host:port",
    "udp://host:port",
  ]

  ## @param interval - number - optional - default: 30
  interval = "30s"

  ## @param interval - number - optional - default: 10
  tcp_timeout = "10s"

  ## @param interval - number - optional - default: 10
  udp_timeout = "10s"

  ## set false to disable election
  election = true

[inputs.socket.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

func (i *input) SampleConfig() string {
	return sample
}
