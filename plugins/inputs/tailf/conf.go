// +build !solaris

package tailf

const (
	pluginName = "tailf"

	tailfConfigSample = `
# [tailf]
# [[tailf.subscribes]]
#       ## ftdataway 地址
#       # ftdataway = "http://localhost:32758/v1/write/logging"
#
#       ## 文件路径名
#	filename = ""
#
#       ## 是否从文件首部读取
#	from_beginning = false
#
#       ## 是否是一个pipe
#	pipe = false
#
#	## 通知方式，默认是 inotify 即由操作系统进行变动通知
#       ## 可以设为 poll 改为轮询文件的方式
#	watch_method = "inotify"
#
#       ## 数据源名字，不可为空
#       source = "" 
`
)

type Subscribe struct {
	File          string `toml:"filename"`
	FormBeginning bool   `toml:"from_beginning"`
	Pipe          bool   `toml:"pipe"`
	WatchMethod   string `toml:"watch_method"`
	Measurement   string `toml:"source"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
