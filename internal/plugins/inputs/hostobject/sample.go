// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

const (
	sampleCfg = `
[inputs.hostobject]

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

##############################
# Disk related options
##############################
## Deprecated
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]

## We collect all devices prefixed with dev by default,If you want to collect additional devices, it's in extra_device add
# extra_device = []

## exclude some with dev prefix (We collect all devices prefixed with dev by default)
# exclude_device = ["/dev/loop0","/dev/loop1"]

# Physical devices only (e.g. hard disks, cd-rom drives, USB keys)
# and ignore all others (e.g. memory partitions such as /dev/shm)
only_physical_device = false

# Ignore the disk which space is zero
ignore_zero_bytes_disk = true

# Disable cloud provider information synchronization
disable_cloud_provider_sync = false

[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws/hwcloud/azure, probe automatically if not set
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
`
)
