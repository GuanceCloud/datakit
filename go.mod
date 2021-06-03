module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.13

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/Azure/azure-sdk-for-go v37.1.0+incompatible
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest/autorest v0.11.1
	github.com/Azure/go-autorest/autorest/adal v0.9.5 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/BurntSushi/toml v0.3.1
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/MichaelMure/go-term-markdown v0.1.3
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Microsoft/hcsshim/test v0.0.0-20201218223536-d3e5debf77da // indirect
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/aeden/traceroute v0.0.0-20181124220833-147686d9cb0f
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.391
	github.com/aliyun/aliyun-log-go-sdk v0.1.5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.8+incompatible
	github.com/andygrunwald/go-jira v1.13.0
	github.com/antchfx/jsonquery v1.1.4
	github.com/apache/thrift v0.13.0
	github.com/araddon/dateparse v0.0.0-20201001162425-8aadafed4dc4
	github.com/aws/aws-sdk-go v1.35.20
	github.com/blang/semver v3.5.1+incompatible
	github.com/blang/semver/v4 v4.0.0
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/cgroups v0.0.0-20201119153540-4cbc285b3327 // indirect
	github.com/containerd/containerd v1.4.1
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/elazarl/goproxy v0.0.0-20180725130230-947c36da3153
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.6.3
	github.com/go-kit/kit v0.10.0
	github.com/go-ole/go-ole v1.2.4
	github.com/go-redis/redis/v8 v8.8.0
	github.com/go-redis/redismock/v8 v8.0.6
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/gobwas/glob v0.2.3
	github.com/godror/godror v0.17.0
	github.com/gofrs/flock v0.8.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/gomarkdown/markdown v0.0.0-20210208175418-bda154fe17d8
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gopacket v1.1.17
	github.com/google/pprof v0.0.0-20200229191704-1ebb73c60ed3 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hpcloud/tail v1.0.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.0.30-rc.1
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	//github.com/influxdata/telegraf v1.16.1
	github.com/influxdata/telegraf v1.15.2
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/jdcloud-api/jdcloud-sdk-go v1.43.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5 // indirect
	github.com/kardianos/service v1.0.0
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/litao91/goldmark-mathjax v0.0.0-20210217064022-a43cf739a50f
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/mattn/go-zglob v0.0.3
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mssola/user_agent v0.5.2
	github.com/naoina/toml v0.1.1
	github.com/nickelser/parselogical v0.0.0-20171014195826-b07373e53c91
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v0.1.1 // indirect
	github.com/opencontainers/selinux v1.5.1 // indirect
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/parser v0.0.0-20200225075606-4cf2c2d4f51b
	github.com/pingcap/tidb v0.0.0-20200225134007-18ce601629fd
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.25.0
	github.com/prometheus/procfs v0.1.3
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/shirou/gopsutil v3.20.12+incompatible
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
	github.com/siddontang/go-mysql v0.0.0-20200222075837-12e89848f047
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible
	github.com/tencentyun/cos-go-sdk-v5 v0.7.7
	github.com/tidwall/gjson v1.7.4
	github.com/trivago/tgo v1.0.7 // indirect
	github.com/tweekmonster/luser v0.0.0-20161003172636-3fa38070dbd7
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/ucloud/ucloud-sdk-go v0.14.0
	github.com/ugorji/go/codec v1.2.4
	github.com/urfave/negroni v1.0.0 // indirect
	github.com/vinllen/mgo v0.0.0-20190830033324-520f0e6e34b8
	github.com/vjeantet/grok v1.0.0
	github.com/yuin/goldmark v1.3.2
	github.com/yuin/goldmark-highlighting v0.0.0-20200307114337-60d527fdb691
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20210528040150-d44a55a4a70a
	gitlab.jiagouyun.com/cloudcare-tools/kodo v0.0.0-20210602132627-8797c1bb76ba
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210525143221-35b2ab0089ea
	golang.org/x/term v0.0.0-20210422114643-f5beecf764ed // indirect
	golang.org/x/text v0.3.4
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.28.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/v3 v3.0.2 // indirect
	k8s.io/api v0.0.0-20190813020757-36bff7324fb7
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v12.0.0+incompatible
)

replace github.com/koding/websocketproxy v0.0.0-20181220232114-7ed82d81a28c => github.com/1157987916/websocketproxy v0.0.0-20201229082103-cfa96d57158c

replace github.com/c-bata/go-prompt => github.com/coanor/go-prompt v0.2.6
