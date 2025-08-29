// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package configwatcher

import "time"

const (
	inputName = "configwatcher"

	minInterval = time.Second
	maxInterval = time.Minute * 10

	sampleConfig = `
[[inputs.configwatcher]]
  ## Require. A name for this collection task for identification.
  source = "default"
  
  ## The root path to prepend to all monitored paths (e.g., for container host paths).
  ## Use "/rootfs" for the mount point in Kubernetes.
  mount_point = ""
  
  ## An array of files or directories to monitor for changes.
  ## Note: Wildcards (e.g., '/path/*.log') are not supported.
  paths = [
      # "/var/spool/cron/crontabs/",          # for entire directory
      # "/etc/hosts",                         # for specific file
  ]

  ## The interval at which to check for changes.
  interval = "5m"
  
  ## Whether to recursively monitor directories in the provided paths.
  recursive = true
  
  ## The maximum file size (in bytes) for which to compute content diffs, default is 256KiB.
  max_diff_size = 262144

  [inputs.configwatcher.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)
