module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v37.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/aeden/traceroute v0.0.0-20181124220833-147686d9cb0f
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.391
	github.com/aliyun/aliyun-log-go-sdk v0.1.5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/andygrunwald/go-jira v1.12.0
	github.com/antchfx/jsonquery v1.1.4
	github.com/apache/thrift v0.13.0
	github.com/aws/aws-sdk-go v1.31.12
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/containerd/containerd v1.4.0-beta.0
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/docker/docker v1.4.2-0.20180327123150-ed7b6428c133
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.6.3
	github.com/go-kit/kit v0.10.0
	github.com/go-ole/go-ole v1.2.4
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobwas/glob v0.2.3
	github.com/godror/godror v0.17.0
	github.com/gofrs/uuid v2.1.0+incompatible
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.1
	github.com/google/gopacket v1.1.17
	github.com/gorilla/websocket v1.4.0
	github.com/hpcloud/tail v1.0.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/influxdata/telegraf v1.15.2
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/jdcloud-api/jdcloud-sdk-go v1.43.0
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5 // indirect
	github.com/kardianos/service v1.0.0
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.4.0
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/mattn/go-zglob v0.0.3
	github.com/naoina/toml v0.1.1
	github.com/nickelser/parselogical v0.0.0-20171014195826-b07373e53c91
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/opencontainers/selinux v1.5.1 // indirect
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/parser v0.0.0-20200225075606-4cf2c2d4f51b
	github.com/pingcap/tidb v0.0.0-20200225134007-18ce601629fd
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/common v0.14.0
	github.com/prometheus/procfs v0.1.3
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
	github.com/siddontang/go-mysql v0.0.0-20200222075837-12e89848f047
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7
	github.com/tidwall/gjson v1.6.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/ucloud/ucloud-sdk-go v0.14.0
	github.com/vinllen/mgo v0.0.0-20190830033324-520f0e6e34b8
	github.com/xanzy/go-gitlab v0.31.0
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20200924074212-8d61bb78afeb
	gitlab.jiagouyun.com/cloudcare-tools/ftagent v1.0.2-0.20200421074654-24a7c53f8f54
	gitlab.jiagouyun.com/cloudcare-tools/kodo v0.0.0-20200928035315-f2d6f0279a19
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/sys v0.0.0-20200923182605-d9f96fdee20d
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.0.0-20200809012840-6f4f008689da // indirect
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.2 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
