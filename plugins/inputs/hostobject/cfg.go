package hostobject

const (
	inputName = "hostobject"

	sampleConfig = `
#[inputs.hostobject]
# ##(optional) 默认使用host name
#name = ''

# ##(optional) 默认为Servers
#class = 'Servers'

# ## 采集间隔，默认3分钟
#interval = '3m'

# ##(optional) 自定义tags
#[inputs.hostobject.tags]
# key1 = 'val1'
`
)
