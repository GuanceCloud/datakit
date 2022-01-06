package hostobject

const (
	InputName = "hostobject"
	InputCat  = "host"

	SampleConfig = `
[inputs.hostobject]

#pipeline = '' # optional

## Datakit does not collect network virtual interfaces under the linux system.
## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

# Ignore the disk which space is zero
ignore_zero_bytes_disk = true

## Ignore mount points by filesystem type. Default ignored following FS types
# ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]


[inputs.hostobject.tags] # (optional) custom tags
# cloud_provider = "aliyun" # aliyun/tencent/aws
# some_tag = "some_value"
# more_tag = "some_other_value"
# ...
`
)
