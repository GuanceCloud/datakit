module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v37.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/date v0.2.0
	github.com/Azure/go-autorest/autorest/to v0.3.0
	github.com/Azure/go-autorest/autorest/validation v0.2.0
	github.com/Azure/go-autorest/tracing v0.5.0
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5
	github.com/Microsoft/hcsshim v0.8.9
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d
	github.com/aeden/traceroute v0.0.0-20181124220833-147686d9cb0f
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.391
	github.com/aliyun/aliyun-log-go-sdk v0.1.5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/andygrunwald/go-jira v1.12.0
	github.com/antchfx/jsonquery v1.1.4
	github.com/apache/thrift v0.13.0
	github.com/aws/aws-sdk-go v1.27.0
	github.com/basgys/goxml2json v1.1.0
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/caio/go-tdigest v2.3.0+incompatible
	github.com/containerd/cgroups v0.0.0-20190919134610-bf292b21730f
	github.com/containerd/containerd v1.4.0-beta.0
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b
	github.com/containerd/ttrpc v1.0.1
	github.com/containerd/typeurl v1.0.1
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548
	github.com/denisenkom/go-mssqldb v0.0.0-20200206145737-bbfc9a55622e
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/dustin/go-humanize v1.0.0
	github.com/ericchiang/k8s v1.2.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.6.2
	github.com/go-kit/kit v0.10.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-ole/go-ole v1.2.4
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-stack/stack v1.8.0
	github.com/gobwas/glob v0.2.3
	github.com/godbus/dbus/v5 v5.0.3
	github.com/godror/godror v0.17.0
	github.com/gofrs/uuid v2.1.0+incompatible
	github.com/gogo/googleapis v1.4.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.1
	github.com/google/pprof v0.0.0-20191218002539-d4f498aebedc // indirect
	github.com/google/uuid v1.1.1
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/hpcloud/tail v1.0.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20200515024757-02f0bf5dbca3
	github.com/influxdata/telegraf v0.10.2-0.20200225003258-fc2486f24c26
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/jmespath/go-jmespath v0.3.0
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/junhsieh/goexamples v0.0.0-20190721045834-1c67ae74caa6
	github.com/kardianos/service v1.0.0
	github.com/klauspost/compress v1.9.5
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.4.0
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/mattn/go-zglob v0.0.3
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mibk/dupl v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/naoina/go-stringutil v0.1.0
	github.com/naoina/toml v0.1.1
	github.com/nickelser/parselogical v0.0.0-20171014195826-b07373e53c91
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1
	github.com/opencontainers/runtime-spec v0.1.2-0.20190507144316-5b71a03e2700
	github.com/opencontainers/selinux v1.5.1
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pierrec/lz4 v2.2.6+incompatible
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/log v0.0.0-20200117041106-d28c14d3b1cd
	github.com/pingcap/parser v0.0.0-20200225075606-4cf2c2d4f51b
	github.com/pingcap/tidb v0.0.0-20200225134007-18ce601629fd
	github.com/pingcap/tipb v0.0.0-20200212061130-c4d518eb1d60
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/prometheus/procfs v0.0.8
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/remyoudompheng/bigfft v0.0.0-20190728182440-6a916e37a237
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/securego/gosec v0.0.0-20200401082031-e946c8c39989 // indirect
	github.com/shirou/gopsutil v2.20.1+incompatible
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
	github.com/siddontang/go-mysql v0.0.0-20200222075837-12e89848f047
	github.com/sirupsen/logrus v1.5.0
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7
	github.com/tidwall/gjson v1.3.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/tommy-muehle/go-mnd v1.3.0 // indirect
	github.com/trivago/tgo v1.0.1
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/ucloud/ucloud-sdk-go v0.14.0
	github.com/vinllen/mgo v0.0.0-20190830033324-520f0e6e34b8
	github.com/wavefronthq/wavefront-sdk-go v0.9.2
	github.com/xanzy/go-gitlab v0.31.0
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/xdg/stringprep v1.0.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20200818074621-1fc80b8dd3a6
	gitlab.jiagouyun.com/cloudcare-tools/ftagent v1.0.2-0.20200421074654-24a7c53f8f54
	go.mongodb.org/mongo-driver v1.3.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/exp v0.0.0-20191227195350-da58074b4299 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200808120158-1030fc2bf1d9
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.0.0-20200809012840-6f4f008689da // indirect
	google.golang.org/appengine v1.6.5
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/ini.v1 v1.57.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools/v3 v3.0.2 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
