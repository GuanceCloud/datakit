module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.16

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/MichaelMure/go-term-markdown v0.1.3
	github.com/aliyun/aliyun-oss-go-sdk v2.1.9+incompatible
	github.com/apache/thrift v0.13.0
	github.com/araddon/dateparse v0.0.0-20201001162425-8aadafed4dc4
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/cgroups v1.0.1
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.2
	github.com/go-ole/go-ole v1.2.4
	github.com/go-playground/validator/v10 v10.6.1 // indirect
	github.com/go-redis/redis/v8 v8.8.0
	github.com/go-redis/redismock/v8 v8.0.6
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/gobwas/glob v0.2.3
	github.com/godror/godror v0.17.0
	github.com/gofrs/flock v0.8.0
	github.com/golang/protobuf v1.5.2
	github.com/gomarkdown/markdown v0.0.0-20210208175418-bda154fe17d8
	github.com/gorilla/mux v1.8.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/influxdata/telegraf v1.15.2
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jessevdk/go-flags v1.4.0
	github.com/kardianos/service v1.0.0
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.8.0
	github.com/litao91/goldmark-mathjax v0.0.0-20210217064022-a43cf739a50f
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/mattn/go-zglob v0.0.3
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mssola/user_agent v0.5.2
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/procfs v0.6.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/sergi/go-diff v1.0.1-0.20180205163309-da645544ed44 // indirect
	github.com/shirou/gopsutil v3.20.12+incompatible
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.4
	github.com/tweekmonster/luser v0.0.0-20161003172636-3fa38070dbd7
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/ugorji/go/codec v1.2.6
	github.com/vjeantet/grok v1.0.0
	github.com/yuin/goldmark v1.3.2
	github.com/yuin/goldmark-highlighting v0.0.0-20200307114337-60d527fdb691
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20210801085853-2efd2cfc7023
	gitlab.jiagouyun.com/cloudcare-tools/kodo v0.0.0-20210603111111-890a3501d71c
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/term v0.0.0-20210422114643-f5beecf764ed
	golang.org/x/text v0.3.6
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/grpc v1.31.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.31.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.0.0-20190813020757-36bff7324fb7
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v12.0.0+incompatible
)

replace github.com/c-bata/go-prompt => github.com/coanor/go-prompt v0.2.6
