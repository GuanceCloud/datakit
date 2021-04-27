package hostobject

const (
	InputName = "hostobject"
	InputCat  = "host"

	SampleConfig = `
[inputs.hostobject]

# ##(optional) collect interval, default is 5 miniutes
interval = '5m'

# ##(optional) 
#pipeline = ''

# ## Datakit does not collect network virtual interfaces under the linux system.
# ## Setting enable_net_virtual_interfaces to true will collect network virtual interfaces stats for linux.
# enable_net_virtual_interfaces = true

# ## Ignore mount points by filesystem type.
ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "autofs", "squashfs", "aufs"]

# ##(optional) custom tags
#[inputs.hostobject.tags]
#  key1 = "value1"
#  key2 = "value2"
`

	pipelineSample = ``
)
