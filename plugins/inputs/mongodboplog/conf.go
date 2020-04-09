package mongodboplog

const (
	pluginName = "mongodb_oplog"

	mongodboplogConfigSample = `
# ## MongoDB oplog Config
# ## Need version 3.6 or later
# ## 
# ## 启用 MongoDB Oplog
# ## ./mongodb --replSet mongo-set
# ## 
# ## 开启 replication oplog 功能后，可以在 MongoDB 交互中使用 "use local" "db.oplog.rs.find()" 查看
# ## 配置详情查看 MongoDB 官方文档： https://docs.mongodb.com/manual/tutorial/deploy-replica-set/#procedure
# ## MongoDB 连接时如果 user 或 password 中含有 '@' 需要使用百分比编码替换
#
# ## MongoDB 数据一般包含多层嵌套，需要指定最终元素的路径，以'/'分隔，如：
# ## { 
# ##	name: "tony", 
# ##	age: 22, 
# ##	address: {
# ##		school: "shanghai",
# ##		home: "beijing"
# ##	}
# ## }
# ## 如果想要获取 name 字段的数据，需要配置的路径为 '/name'
# ## 如果想要获取 address home 字段的数据，需要配置的路径为 '/address/home'
#
# [mongodb_oplog]
# [[mongodb_oplog.subscribes]]
#       ## MongoDB URL: mongodb://user:password@host:port/database
#       mongodb_url="mongodb://127.0.0.1:27017"
#	## MongoDB database
#       database="testdb"
#	## MongoDB collection
#       collection="testdb"
#
#       measurement="test"
#	## 配置 tags 数据所在路径，可以为空
# 	tags=[
#		"/path",
#		"/a/b/c/e"
#	]
#	## 配置 fields 数据所在路径，并且指定数据类型，不可以为空
#	## 支持 int float bool string 四种类型	
# 	[mongodb_oplog.subscribes.fields]
#		"/a/c/d" = "int"
#		"/a/c/e" = "float"
`
)

type Subscribe struct {
	MongodbURL string `toml:"mongodb_url"`
	Database   string `toml:"database"`
	Collection string `toml:"collection"`

	Measurement string   `toml:"measurement"`
	Tags        []string `toml:"tags"`
	Fields      map[string]string
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}
