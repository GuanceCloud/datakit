package replication

const (
	pluginName = "replication"

	replicationConfigSample = `
# ## PostgreSQL Config
# ## Need version 9.4 or later
#
# ## 参数修改
# ## wal_level = 'logical';
# ## max_replication_slots = 5; #该值要大于1
#
# ## 创建有replication权限的用户，test_rep 为测试用户名
# ## CREATE ROLE test_rep LOGIN  ENCRYPTED PASSWORD 'xxxx' REPLICATION;
# ## GRANT CONNECT ON DATABASE test_database to test_rep;
#
# ## 修改白名单配置
# ## 在 pg_hba.conf 中增加配置: host replication test_rep all md5
# 
# ## 修改后需要reload才能生效
#
# [replication]
# [[replication.subscribes]]
#	## postgres host
#	host="10.100.64.106"
#	## postgres port
# 	port=25432
#	## postgres user (need replication privilege)
# 	user="rep_name"
#	## login password
# 	password="Sql123456"
#	## data from database
# 	database="testdb"
#	## data from table
# 	table="tb"
#	## replication slot name, only
# 	slotname="slot_for_datakit"
#	## exlcude the events of postgres, there are 3 events: "insert","update","delete"
#	events=['insert']
# 
#	## point measurement, default is table name
# 	measurement=""
#	## tags 
# 	tags=['colunm1']
#	## fields
# 	fields=['colunm0']
`
)

// Subscribe 订阅一个数据库中的表的wal，根据规则保存到es里相应的index，type中
type Subscribe struct {
	// Host 地址
	Host string `toml:"host"`
	// Port 端口
	Port uint16 `toml:"port"`
	// Database database
	Database string `toml:"database"`
	// User user
	User string `toml:"user"`
	// Password password
	Password string `toml:"password"`
	// table
	Table string `toml:"table"`
	// SlotName 逻辑复制槽
	SlotName string `toml:"slotname"`
	// events
	Events []string `toml:"events"`

	Measurement string   `toml:"measurement"`
	Tags        []string `toml:"tags"`
	Fields      []string `toml:"fields"`

	eventsOperation map[string]byte

	pointConfig pointConfig
}

type pointConfig struct {
	measurement string
	tags        map[string]byte
	fields      map[string]byte
}

type Config struct {
	Subscribes []Subscribe `toml:"subscribes"`
}

func configUpdatePoint(sub *Subscribe) {

	sub.pointConfig = pointConfig{
		measurement: func() string {
			if sub.Measurement != "" {
				return sub.Measurement
			} else {
				return sub.Table
			}
		}(),
		tags:   make(map[string]byte, len(sub.Tags)),
		fields: make(map[string]byte, len(sub.Fields)),
	}

	for _, value := range sub.Tags {
		sub.pointConfig.tags[value] = '0'
	}

	for _, value := range sub.Fields {
		sub.pointConfig.fields[value] = '0'
	}

	for _, value := range sub.Events {
		sub.eventsOperation[value] = '0'
	}

}
