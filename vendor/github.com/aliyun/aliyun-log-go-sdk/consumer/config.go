package consumerLibrary

type LogHubConfig struct {
	//:param Endpoint:
	//:param AccessKeyID:
	//:param AccessKeySecret:
	//:param Project:
	//:param Logstore:
	//:param ConsumerGroupName:
	//:param ConsumerName:
	//:param CursorPosition: This options is used for initialization, will be ignored once consumer group is created and each shard has beeen started to be consumed.
	//  Provide three options ：BEGIN_CURSOR,END_CURSOR,SPECIAL_TIMER_CURSOR,when you choose SPECIAL_TIMER_CURSOR, you have to set CursorStartTime parameter.
	//:param HeartbeatIntervalInSecond:
	// default 20, once a client doesn't report to server * heartbeat_interval * 3 interval,
	// server will consider it's offline and re-assign its task to another consumer.
	// don't set the heatbeat interval too small when the network badwidth or performance of consumtion is not so good.
	//:param DataFetchIntervalInMs: default 200(Millisecond), don't configure it too small (<100Millisecond)
	//:param MaxFetchLogGroupCount: default 1000, fetch size in each request, normally use default. maximum is 1000, could be lower. the lower the size the memory efficiency might be better.
	//:param CursorStartTime: Will be used when cursor_position when could be "begin", "end", "specific time format in time stamp", it's log receiving time. The unit of parameter is seconds.
	//:param InOrder:
	// 	default False, during consuption, when shard is splitted,
	// 	if need to consume the newly splitted shard after its parent shard (read-only) is finished consumption or not.
	// 	suggest keep it as False (don't care) until you have good reasion for it.
	//:param AllowLogLevel: default info,optional: debug,info,warn,error
	//:param LogFileName: Setting Log File Path，for example "/root/log/log_file.log",default
	//:param IsJsonType: Set whether the log output type is JSON，default false.
	//:param LogMaxSize: MaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
	//:param LogMaxBackups:
	// 	MaxBackups is the maximum number of old log files to retain.  The default
	// 	is to retain all old log files (though MaxAge may still cause them to get
	// 	deleted.)
	//:param LogCompass: Compress determines if the rotated log files should be compressed using gzip.

	Endpoint                  string
	AccessKeyID               string
	AccessKeySecret           string
	Project                   string
	Logstore                  string
	ConsumerGroupName         string
	ConsumerName              string
	CursorPosition            string
	HeartbeatIntervalInSecond int
	DataFetchIntervalInMs     int64
	MaxFetchLogGroupCount     int
	CursorStartTime           int64 // Unix time stamp; Units are seconds.
	InOrder                   bool
	AllowLogLevel             string
	LogFileName               string
	IsJsonType                bool
	LogMaxSize                int
	LogMaxBackups             int
	LogCompass                bool
	// SecurityToken        string
}

const (
	BEGIN_CURSOR            = "BEGIN_CURSOR"
	END_CURSOR              = "END_CURSOR"
	SPECIAL_TIMER_CURSOR    = "SPECIAL_TIMER_CURSOR"
	INITIALIZING            = "INITIALIZING"
	INITIALIZING_DONE       = "INITIALIZING_DONE"
	PULL_PROCESSING         = "PULL_PROCESSING"
	PULL_PROCESSING_DONE    = "PULL_PROCESSING_DONE"
	CONSUME_PROCESSING      = "CONSUME_PROCESSING"
	CONSUME_PROCESSING_DONE = "CONSUME_PROCESSING_DONE"
	SHUTDOWN_COMPLETE       = "SHUTDOWN_COMPLETE"
)

