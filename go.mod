module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.13

require (
	github.com/Azure/azure-sdk-for-go v37.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.9.3
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/aeden/traceroute v0.0.0-20181124220833-147686d9cb0f // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.205
	github.com/aliyun/aliyun-log-go-sdk v0.1.5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.2+incompatible // indirect
	github.com/andygrunwald/go-jira v1.12.0
	github.com/antchfx/jsonquery v1.1.4
	github.com/apache/thrift v0.13.0
	github.com/aws/aws-sdk-go v1.27.0
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/containerd/containerd v1.4.0-beta.0
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/containerd/fifo v0.0.0-20200410184934-f15a3290365b // indirect
	github.com/containerd/ttrpc v1.0.1 // indirect
	github.com/containerd/typeurl v1.0.1
	github.com/cosiner/argv v0.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/ericchiang/k8s v1.2.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.6.2
	github.com/go-delve/delve v1.4.0 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-ole/go-ole v1.2.4
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobwas/glob v0.2.3
	github.com/godror/godror v0.17.0
	github.com/gofrs/uuid v2.1.0+incompatible
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/golang/snappy v0.0.1
	github.com/golangci/golangci-lint v1.27.0 // indirect
	github.com/google/gopacket v1.1.17 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/influxdata/influxdb v1.8.0 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d
	github.com/influxdata/telegraf v0.10.2-0.20200225003258-fc2486f24c26
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/jackc/pgx v3.6.0+incompatible
	github.com/kardianos/service v1.0.0
	github.com/lib/pq v1.4.0
	github.com/logrusorgru/gopb3any v0.0.0-20181002194712-b78f3858fa1f // indirect
	github.com/logrusorgru/lifo v0.0.0-20181002195007-26900045159d // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mehrdadrad/mylg v0.2.6 // indirect
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
	github.com/prometheus/common v0.9.1
	github.com/prometheus/node_exporter v0.18.1 // indirect
	github.com/prometheus/procfs v0.0.8
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
	github.com/siddontang/go-mysql v0.0.0-20200222075837-12e89848f047
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/stamblerre/gocode v1.0.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.123+incompatible
	github.com/tidwall/gjson v1.3.0
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/ucloud/ucloud-sdk-go v0.14.0
	github.com/vinllen/mgo v0.0.0-20190830033324-520f0e6e34b8
	github.com/xanzy/go-gitlab v0.31.0
	github.com/zcalusic/sysinfo v0.0.0-20200228145645-a159d7cc708b // indirect
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20200724094927-05a50c73056f
	gitlab.jiagouyun.com/cloudcare-tools/ftagent v1.0.2-0.20200421074654-24a7c53f8f54
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools/gopls v0.4.1 // indirect
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.24.0 // indirect
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools/v3 v3.0.2 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)
