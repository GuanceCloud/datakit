module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.16

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/BurntSushi/toml v0.3.1
	github.com/DataDog/datadog-agent v0.0.0-20210913210003-1f013d7aad22
	github.com/MichaelMure/go-term-markdown v0.1.3
	github.com/aliyun/aliyun-oss-go-sdk v2.1.10+incompatible
	github.com/apache/thrift v0.13.0
	github.com/araddon/dateparse v0.0.0-20201001162425-8aadafed4dc4
	github.com/c-bata/go-prompt v0.2.5
	github.com/chromedp/cdproto v0.0.0-20210910012206-68626162910d // indirect
	github.com/chromedp/chromedp v0.7.4 // indirect
	github.com/containerd/cgroups v1.0.1
	github.com/denisenkom/go-mssqldb v0.10.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4
	github.com/fatih/color v1.12.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-gonic/gin v1.7.2
	github.com/go-ole/go-ole v1.2.5
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
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/influxdata/telegraf v1.15.2
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jessevdk/go-flags v1.4.0
	github.com/kardianos/service v1.0.0
	github.com/lib/pq v1.10.0
	github.com/litao91/goldmark-mathjax v0.0.0-20210217064022-a43cf739a50f
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/mssola/user_agent v0.5.2
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pborman/ansi v1.0.0
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/procfs v0.6.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/shirou/gopsutil v3.21.7+incompatible
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/spf13/cast v1.3.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.4
	github.com/tinylib/msgp v1.1.6
	github.com/tweekmonster/luser v0.0.0-20161003172636-3fa38070dbd7
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/ugorji/go/codec v1.2.6
	github.com/vjeantet/grok v1.0.0
	github.com/vmihailenco/msgpack/v4 v4.3.12
	github.com/yuin/goldmark v1.3.5
	github.com/yuin/goldmark-highlighting v0.0.0-20200307114337-60d527fdb691
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20210819175148-5d0565ecd466
	gitlab.jiagouyun.com/cloudcare-tools/kodo v0.0.0-20210818074822-5341861c9cbf
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20210907225631-ff17edfbf26d
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
	golang.org/x/term v0.0.0-20210422114643-f5beecf764ed
	golang.org/x/text v0.3.7
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83 // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.33.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
)

replace github.com/c-bata/go-prompt => github.com/coanor/go-prompt v0.2.6

// added for ddtrace
replace (
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20210322155958-9866ef4cd22c
	k8s.io/api => k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.5
	k8s.io/apiserver => k8s.io/apiserver v0.20.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.5
	k8s.io/client-go => k8s.io/client-go v0.20.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
	k8s.io/code-generator => k8s.io/code-generator v0.20.5
	k8s.io/component-base => k8s.io/component-base v0.20.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.5
	k8s.io/cri-api => k8s.io/cri-api v0.20.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.5
	k8s.io/kubectl => k8s.io/kubectl v0.20.5
	k8s.io/kubelet => k8s.io/kubelet v0.20.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.5
	k8s.io/metrics => k8s.io/metrics v0.20.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.3-rc.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.5
)
