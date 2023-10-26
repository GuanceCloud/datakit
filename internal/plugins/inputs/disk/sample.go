// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

const sampleCfg = `
[[inputs.disk]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  ## Physical devices only (e.g. hard disks, cd-rom drives, USB keys)
  ## and ignore all others (e.g. memory partitions such as /dev/shm)
  only_physical_device = false

  ## Deprecated
  # ignore_mount_points = ["/"]

  ## Deprecated
  # mount_points = ["/"]

  ## Deprecated
  # ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

  ## Deprecated
  # fs = ["ext2", "ext3", "ext4", "NTFS"]

  ## We collect all devices prefixed with dev by default,If you want to collect additional devices, it's in extra_device add
  # extra_device = ["/nfsdata"]

  ## exclude some with dev prefix (We collect all devices prefixed with dev by default)
  # exclude_device = ["/dev/loop0","/dev/loop1"]

  [inputs.disk.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
