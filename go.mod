module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.18

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/DataDog/ebpf v0.0.0-20210419131141-ea64821c9793
	github.com/MichaelMure/go-term-markdown v0.1.3
	github.com/aliyun/aliyun-oss-go-sdk v2.2.5+incompatible
	github.com/antchfx/xmlquery v1.3.11
	github.com/apache/thrift v0.13.0
	github.com/araddon/dateparse v0.0.0-20201001162425-8aadafed4dc4
	github.com/c-bata/go-prompt v0.2.5
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/containerd/cgroups v1.0.1
	github.com/containerd/containerd v1.5.5
	github.com/containerd/typeurl v1.0.2
	github.com/denisenkom/go-mssqldb v0.12.2
	github.com/dgraph-io/ristretto v0.1.0
	github.com/didip/tollbooth/v6 v6.1.2
	github.com/docker/docker v20.10.8+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/go-lumber v0.1.1
	github.com/elazarl/goproxy v0.0.0-20210801061803-8e322dfb79c4
	github.com/fatih/color v1.12.0
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/gdamore/tcell/v2 v2.4.1-0.20210905002822-f057f0a857a1
	github.com/gin-gonic/gin v1.7.4
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-ole/go-ole v1.2.5
	github.com/go-ping/ping v1.1.0
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gobwas/ws v1.1.0
	github.com/godror/godror v0.17.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.4
	github.com/gomarkdown/markdown v0.0.0-20210208175418-bda154fe17d8
	github.com/google/gopacket v1.1.19
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/grafana/loki v1.6.2-0.20210806161513-f5fd02966003
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/influxdata/influxdb1-client v0.0.0-20220302092344-a9ab5670611c
	github.com/influxdata/telegraf v1.16.3
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/ip2location/ip2location-go v8.3.0+incompatible
	github.com/jessevdk/go-flags v1.5.0
	github.com/kardianos/service v1.2.1
	github.com/lib/pq v1.10.2
	github.com/mssola/user_agent v0.5.2
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/oschwald/geoip2-golang v1.7.0
	github.com/pborman/ansi v1.0.0
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.37.0
	github.com/prometheus/procfs v0.7.3
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/rivo/tview v0.0.0-20220129131435-1f7581b67bd1
	github.com/shirou/gopsutil v3.21.8+incompatible
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.0
	github.com/tidwall/gjson v1.14.3
	github.com/tinylib/msgp v1.1.6
	github.com/tweekmonster/luser v0.0.0-20161003172636-3fa38070dbd7
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/ubwbu/grok v1.0.6
	github.com/ugorji/go/codec v1.2.6
	github.com/vjeantet/grok v1.0.0
	github.com/whilp/git-urls v1.0.0
	gitlab.jiagouyun.com/cloudcare-tools/cliutils v0.0.0-20220928075050-6a94ccf938f8
	go.etcd.io/bbolt v1.3.6
	go.mercari.io/go-dnscache v0.0.0-20220124075326-2701c2ab5df5
	go.opentelemetry.io/otel v1.4.1
	go.opentelemetry.io/otel/exporters/jaeger v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.27.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.4.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.4.1
	go.opentelemetry.io/otel/metric v0.27.0
	go.opentelemetry.io/otel/sdk v1.4.1
	go.opentelemetry.io/otel/sdk/metric v0.27.0
	go.opentelemetry.io/otel/trace v1.4.1
	go.opentelemetry.io/proto/otlp v0.12.0
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.1.0
	golang.org/x/net v0.1.0
	golang.org/x/sys v0.1.0
	golang.org/x/term v0.1.0
	golang.org/x/text v0.4.0
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/CodapeWild/dd-trace-go.v1 v1.35.18
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.25.0
	k8s.io/apimachinery v0.25.0
	k8s.io/client-go v0.25.0
	k8s.io/cri-api v0.20.6
	k8s.io/metrics v0.20.5
)

require (
	github.com/DataDog/gopsutil v1.1.0
	github.com/GuanceCloud/confd v0.0.0-20221011061652-936375adb49b
	github.com/Shopify/sarama v1.29.1
	github.com/cortexproject/cortex v1.9.1-0.20210722081137-485474c9afb2
	github.com/gin-contrib/timeout v0.0.3
	github.com/golang/protobuf v1.5.2
	github.com/klauspost/compress v1.15.9
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.60.1
	github.com/r3labs/diff/v3 v3.0.0
	github.com/schollz/progressbar/v3 v3.9.0
	github.com/tidwall/wal v1.1.7
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/armon/go-metrics v0.3.10 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.107 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/garyburd/redigo v1.6.4 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/status v1.0.3 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/consul/api v1.15.2 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.9.7 // indirect
	github.com/hashicorp/vault/api v1.8.0 // indirect
	github.com/hashicorp/vault/sdk v0.6.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/influxdata/line-protocol/v2 v2.2.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.60.1 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/node_exporter v1.0.0-rc.0.0.20200428091818-01054558c289 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/sercand/kuberesolver v2.4.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/weaveworks/common v0.0.0-20210419092856-009d1eebd624 // indirect
	github.com/weaveworks/promrus v1.2.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.etcd.io/etcd/api/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.5 // indirect
	go.etcd.io/etcd/client/v3 v3.5.5 // indirect
	golang.org/x/exp v0.0.0-20221026004748-78e5e7837ae6 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	k8s.io/apiextensions-apiserver v0.25.0 // indirect
	k8s.io/kube-openapi v0.0.0-20220803164354-a70c9af30aea // indirect
	sigs.k8s.io/controller-runtime v0.13.0 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
)

require (
	github.com/DataDog/sketches-go v1.4.1 // indirect
	github.com/MichaelMure/go-term-text v0.2.7 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.8.18 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/alecthomas/chroma v0.7.1 // indirect
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/antchfx/xpath v1.2.1 // indirect
	github.com/avast/retry-go v2.7.0+incompatible // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chromedp/cdproto v0.0.0-20220914223734-4ab9dc957c3e // indirect
	github.com/chromedp/chromedp v0.8.5 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/containerd/fifo v1.0.0 // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/dlclark/regexp2 v1.1.6 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/eliukblau/pixterm/pkg/ansimage v0.0.0-20191210081756-9fb6cf8c2f75 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/florianl/go-tc v0.2.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-pkgz/expirable-cache v0.0.3 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.6.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v0.0.0-20200817173448-b6b71def0850 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kyokomi/emoji v2.1.0+incompatible // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/mdlayher/netlink v1.4.1 // indirect
	github.com/mdlayher/socket v0.0.0-20210307095302-262dc9984e00 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mountinfo v0.4.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.1 // indirect
	github.com/opencontainers/selinux v1.8.2 // indirect
	github.com/oschwald/maxminddb-golang v1.9.0 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.3.1 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20210508154835-66ddd91f7ddd // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	go.mongodb.org/mongo-driver v1.10.2
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.4.1 // indirect
	go.opentelemetry.io/otel/internal/metric v0.27.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/image v0.0.0-20191206065243-da761ea9ff43 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/utils v0.0.0-20221012122500-cfd413dd9e85 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace github.com/c-bata/go-prompt => github.com/coanor/go-prompt v0.2.6

replace github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20210720084720-59d02b5ef003

replace github.com/weaveworks/common => github.com/pyh4/common v0.0.0-20220923021349-874c8cc0db2c

// added for ddtrace
replace (
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20210322155958-9866ef4cd22c
	github.com/kardianos/service => github.com/GuanceCloud/service v1.2.1-rc3
// k8s.io/api => k8s.io/api v0.20.5
// k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.5
// k8s.io/apimachinery => k8s.io/apimachinery v0.20.5
// k8s.io/apiserver => k8s.io/apiserver v0.20.5
// k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.5
// k8s.io/client-go => k8s.io/client-go v0.20.5
// k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.5
// k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
// k8s.io/code-generator => k8s.io/code-generator v0.20.5
// k8s.io/component-base => k8s.io/component-base v0.20.5
// k8s.io/component-helpers => k8s.io/component-helpers v0.20.5
// k8s.io/controller-manager => k8s.io/controller-manager v0.20.5
// k8s.io/cri-api => k8s.io/cri-api v0.20.5
// k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.5
// k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.5
// k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.5
// k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.5
// k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.5
// k8s.io/kubectl => k8s.io/kubectl v0.20.5
// k8s.io/kubelet => k8s.io/kubelet v0.20.5
// k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.5
// k8s.io/metrics => k8s.io/metrics v0.20.5
// k8s.io/mount-utils => k8s.io/mount-utils v0.20.3-rc.0
// k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.5
// k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.5
// k8s.io/sample-controller => k8s.io/sample-controller v0.20.5
)
