// +build !solaris

package tailf

const (
	pluginName = "tailf"

	tailfConfigSample = `
# [tailf]
# [[tailf.subscribes]]
#
#	file = ""
#
#	from_beginning = false
#
#	pipe = false
#
#	## Method used to watch for file updates.  Can be either "inotify" or "poll".
#	watch_method = "inotify"
#
#       ## measurement，不可重复
#       measurement = "" 
`
)

type Subscribe struct {
	File          string `toml:"file"`
	FormBeginning bool   `toml:"from_beginning"`
	Pipe          bool   `toml:"pipe"`
	WatchMethod   string `toml:"watch_method"`
	Measurement   string `toml:"measurement"`
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
