package utils

var (
	XToken         = `X-Token`
	XAgentUID      = `X-Agent-Uid`
	XDatakitUUID   = `X-Datakit-UUID`
	XAgentIp       = `X-Agent-Ip`
	XVersion       = `X-Version`
	XDomainName    = `X-Domain-Name`
	XTraceId       = `X-Trace-Id`
	XRP            = `X-RP`
	XSource        = `X-Source`
	XCQRP          = `X-CQ-RP`
	XDBUUID        = `X-DB-UUID`
	XWorkspaceUUID = `X-Workspace-UUID`
	XTableName     = `X-Table-Name`

	BuiltInClass   = `sys_class`
	UnBuiltInClass = `personal_class`

	InfluxPermReadOnly  = `ro`
	InfluxPermReadWrite = `wr`
	InfluxPermAdmin     = `admin`

	InfluxPrecisonDefault = `s`
	InfluxInstanceMaxDB   = 500
	InfluxDefaultCQRP     = `autogen`
	InfluxdbBizPreifx     = `biz_`
	InfluxCQPrefix        = `cq_`
	ObjectClassPrefix     = `objc_`

	//default string aggr
	InfluxCQDefaultAggr = `last`
	//ck
	ClusterNameConst = `default`

	//对象
	EsDocId = `__docid`

	KeepObjectName   = `name`
	KeepObjectUUID   = `__uuid`
	KeepObjectClass  = `class`
	EsLastUpdateTime = `last_update_time`

	EsKeepLogSource  = `source`
	EsKeepCreateTime = `create_time`

	//EsMeta           = `meta`
	//EsKeepTimeStamp  = `timestampUs`
	EsKeepTimeStampM = `date`

	EsKeepEventId   = `__eventId`
	EsKeepStatus    = `__status`
	EsKeepTitle     = `__title`
	EsKeepEventType = `__type`

	RUMType = `source`

	MaxRetryTimes = int64(3600 * 24)

	BillState     = `bill_state`
	VerType       = `version_type`
	EsKeepTracing = `tracing`

	WsFreeState  = `free`
	WsNormal     = `normal`
	WsOwningFee  = `arrearage`
	WsExpireTime = `expired`

	Times = `time`
)
