package constants

const (
	// pipeline关键字段.
	FieldTime       = "time"
	FieldMessage    = "message"
	FieldStatus     = "status"
	PlLoggingSource = "source"

	DefaultStatus = "info"
	NSDefault     = "default" // 内置 pl script， 优先级最低
	NSGitRepo     = "gitrepo" // git 管理的 pl script
	NSConfd       = "confd"   // confd 管理的 pl script
	NSRemote      = "remote"  // remote pl script，优先级最高
)
