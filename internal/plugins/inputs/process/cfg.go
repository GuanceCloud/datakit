// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

const (
	inputName      = "host_processes"
	objectFeedName = inputName + "/O"
	category       = "host"

	sampleConfig = `
[[inputs.host_processes]]
  # Only collect these matched process' metrics. For process objects
  # these white list not applied. Process name support regexp.
  # process_name = [".*nginx.*", ".*mysql.*"]

  # Process minimal run time(default 10m)
  # If process running time less than the setting, we ignore it(both for metric and object)
  min_run_time = "10m"

  ## Enable process metric collecting
  open_metric = false

  ## Enable listen ports tag, default is false
  enable_listen_ports = false

  ## Enable open files field, default is false
  enable_open_files = false

  ## only collect container-based process(object and metric)
  only_container_processes = false

  # Extra tags
  [inputs.host_processes.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`
)
