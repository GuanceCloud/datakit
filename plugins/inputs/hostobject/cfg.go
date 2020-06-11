package hostobject

const (
	inputName = "host"

	sampleConfig = `
# ##(required) dataway的地址
#dataway_path='/v1/write/object'

# ##(optional) 默认使用host name
#name = ''

# ##(optional) 默认为Servers
#class = 'Servers'

# ## 采集间隔，默认3分钟
#interval = '3m'

# ##(optional) 自定义tags
#[tags]
# key1 = 'val1'
`
)
