module gitlab.jiagouyun.com/cloudcare-tools/datakit

go 1.19

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.47.1 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/DataDog/gopsutil v1.2.1
	github.com/GuanceCloud/confd v0.1.101
	github.com/GuanceCloud/grok v1.1.5-0.20250416104424-34917bd63e69
	github.com/GuanceCloud/platypus v0.3.3-0.20250528074826-e3130ff5a05c
	github.com/GuanceCloud/timeout v1.9.1
	github.com/IBM/sarama v1.41.2
	github.com/aliyun/aliyun-oss-go-sdk v3.0.2+incompatible // indirect
	github.com/antchfx/xmlquery v1.3.18 // indirect
	github.com/apache/thrift v0.16.0
	github.com/araddon/dateparse v0.0.0-20201001162425-8aadafed4dc4
	github.com/bmatcuk/doublestar/v4 v4.7.0
	github.com/c-bata/go-prompt v0.2.5
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575
	github.com/containerd/cgroups/v3 v3.0.1
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/cortexproject/cortex v1.9.1-0.20210722081137-485474c9afb2
	github.com/dgraph-io/ristretto v0.2.0
	github.com/didip/tollbooth/v6 v6.1.2
	github.com/docker/docker v20.10.8+incompatible
	github.com/dustin/go-humanize v1.0.1
	github.com/elastic/go-lumber v0.1.1
	github.com/elazarl/goproxy v0.0.0-20231031074852-3ec07828be7a
	github.com/fatih/color v1.15.0
	github.com/gdamore/tcell/v2 v2.4.1-0.20210905002822-f057f0a857a1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-ole/go-ole v1.2.6
	github.com/go-ping/ping v1.1.0
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gobwas/ws v1.1.0
	github.com/godror/godror v0.17.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.4
	github.com/golang/snappy v0.0.4
	github.com/google/pprof v0.0.0-20230705174524-200ffdc848b8
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.0
	github.com/gosnmp/gosnmp v1.35.0
	github.com/grafana/loki v1.6.2-0.20210806161513-f5fd02966003
	github.com/hashicorp/go-retryablehttp v0.7.4 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20220302092344-a9ab5670611c
	github.com/influxdata/telegraf v1.16.3
	github.com/influxdata/toml v0.0.0-20190415235208-270119a8ce65
	github.com/ip2location/ip2location-go v8.3.0+incompatible // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/jessevdk/go-flags v1.5.0
	github.com/kardianos/service v1.2.1
	github.com/klauspost/compress v1.17.9
	github.com/mssola/user_agent v0.6.0 // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/ory/dockertest/v3 v3.9.1
	github.com/oschwald/geoip2-golang v1.9.0 // indirect
	github.com/pkg/sftp v1.11.0
	github.com/prometheus-operator/prometheus-operator/pkg/client v0.51.2
	github.com/prometheus/client_golang v1.16.0
	github.com/prometheus/client_model v0.4.0
	github.com/prometheus/common v0.44.0
	github.com/prometheus/procfs v0.11.0
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/pyroscope-io/pyroscope v0.36.0
	github.com/r3labs/diff/v3 v3.0.0
	github.com/rivo/tview v0.0.0-20220129131435-1f7581b67bd1
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/shirou/gopsutil/v3 v3.22.7
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cast v1.5.1
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.9.0
	github.com/tidwall/gjson v1.17.0 // indirect
	github.com/tinylib/msgp v1.1.6
	github.com/tweekmonster/luser v0.0.0-20161003172636-3fa38070dbd7
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/ugorji/go/codec v1.2.11
	github.com/vjeantet/grok v1.0.1
	github.com/whilp/git-urls v1.0.0
	go.mercari.io/go-dnscache v0.0.0-20220124075326-2701c2ab5df5
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.14.0
	golang.org/x/net v0.16.0
	golang.org/x/sys v0.13.0
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0
	google.golang.org/grpc v1.51.0
	gopkg.in/CodapeWild/dd-trace-go.v1 v1.35.18
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)

// indirect
require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/DataDog/datadog-go/v5 v5.1.1 // indirect
	github.com/DataDog/sketches-go v1.4.1 // indirect
	github.com/GuanceCloud/mdcheck v0.0.2-0.20250110081153-61d47275e386
	github.com/GuanceCloud/tracing-protos v0.0.0-20230619071516-54c8cff1b6b3
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210428141323-04723f9f07d7 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/antchfx/xpath v1.2.4 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/aws/aws-sdk-go v1.44.175
	github.com/aws/aws-sdk-go-v2 v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.8 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.27 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.21 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.18.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.0 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/c9s/goprocinfo v0.0.0-20210130143923-c95fcf8c64a8
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash v1.1.0
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/cli v20.10.17+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/eapache/go-resiliency v1.4.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0
	github.com/garyburd/redigo v1.6.4 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-pkgz/expirable-cache v0.0.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/goccy/go-json v0.10.3
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/status v1.0.3 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.6.0
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hashicorp/consul/api v1.18.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.7 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hashicorp/vault/api v1.8.2 // indirect
	github.com/hashicorp/vault/sdk v0.6.2 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/influxdata/line-protocol/v2 v2.2.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.4
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/naoina/go-stringutil v0.1.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/oschwald/maxminddb-golang v1.11.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.51.2 // indirect
	github.com/prometheus/node_exporter v1.0.0-rc.0.0.20200428091818-01054558c289 // indirect
	github.com/pyroscope-io/jfr-parser v0.5.2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/sercand/kuberesolver v2.4.0+incompatible // indirect
	github.com/sergi/go-diff v1.2.0
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20230807190133-6afddb37c1f0
	github.com/vishvananda/netns v0.0.4
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/weaveworks/common v0.0.0-20210419092856-009d1eebd624 // indirect
	github.com/weaveworks/promrus v1.2.0 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yuin/goldmark v1.5.4 // indirect
	github.com/yuin/goldmark-meta v1.1.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/etcd/api/v3 v3.5.6 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.6 // indirect
	go.etcd.io/etcd/client/v3 v3.5.6 // indirect
	go.mongodb.org/mongo-driver v1.10.2
	go.uber.org/multierr v1.9.0
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/mod v0.13.0 // indirect
	golang.org/x/oauth2 v0.5.0 // indirect
	golang.org/x/sync v0.4.0 // indirect
	golang.org/x/time v0.6.0
	golang.org/x/tools v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.34.2
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.25.0
	k8s.io/apiextensions-apiserver v0.25.0 // indirect
	k8s.io/apimachinery v0.25.0
	k8s.io/apiserver v0.25.0 // indirect
	k8s.io/client-go v0.25.0
	k8s.io/component-base v0.25.0 // indirect
	k8s.io/cri-api v0.20.6
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220803164354-a70c9af30aea // indirect
	k8s.io/metrics v0.20.5
	k8s.io/utils v0.0.0-20221012122500-cfd413dd9e85
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.24.1 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.6.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.25.0
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/DataDog/ebpf-manager v0.2.16
	github.com/GuanceCloud/kubernetes v0.0.0-20230801080916-ca299820872b
	github.com/GuanceCloud/zipstream v0.1.0 // indirect
	github.com/andrewkroh/sys v0.0.0-20151128191922-287798fe3e43
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/cilium/ebpf v0.11.0
	github.com/gin-contrib/size v0.0.0-20231230013409-e0f46cc9c1db
	github.com/google/gopacket v0.0.0-00010101000000-000000000000
	github.com/grafana/jfr-parser v0.0.1
	github.com/grafana/pyroscope/ebpf v0.2.1
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/ibmdb/go_ibm_db v0.4.4
	github.com/jackc/pgx/v5 v5.5.1
	github.com/mattn/go-oci8 v0.1.1
	github.com/microsoft/go-mssqldb v1.6.0
	github.com/netsampler/goflow2 v1.3.5
	github.com/oklog/ulid v1.3.1
	github.com/petermattis/goid v0.0.0-20240327183114-c42a807a84ba
	github.com/rosedblabs/wal v1.3.6
	github.com/sijms/go-ora/v2 v2.8.19
	github.com/spf13/cobra v1.9.1
	github.com/vmware/govmomi v0.38.0
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63
	k8s.io/kubelet v0.0.0-00010101000000-000000000000
)

require (
	github.com/GuanceCloud/cliutils v1.1.22-0.20250912100430-6d6393595d06
	github.com/VictoriaMetrics/easyproto v0.1.4 // indirect
	github.com/andybalholm/brotli v1.0.4
	github.com/avast/retry-go/v4 v4.1.0 // indirect
	github.com/avvmoto/buf-readerat v0.0.0-20171115124131-a17c8cb89270 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/pgzip v1.2.6
	github.com/libp2p/go-reuseport v0.3.0 // indirect
	github.com/outcaste-io/ristretto v0.2.1 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/valyala/fastjson v1.6.3
)

require (
	github.com/GuanceCloud/pipeline-go v1.0.9-0.20250819095325-01d2a81ed1c2
	github.com/hipages/php-fpm_exporter v1.2.1
	github.com/shopspring/decimal v1.2.0
)

require (
	github.com/speps/go-hashids v2.0.0+incompatible // indirect
	github.com/tomasen/fcgi_client v0.0.0-20180423082037-2bb3d819fd19 // indirect
)

replace (
	github.com/c-bata/go-prompt => github.com/coanor/go-prompt v0.2.6
	github.com/fsnotify/fsnotify => github.com/GuanceCloud/fsnotify v1.8.2
	github.com/google/gopacket => github.com/GuanceCloud/gopacket v0.0.1
	github.com/grafana/jfr-parser => github.com/GuanceCloud/jfr-parser v0.8.6
	github.com/influxdata/influxdb1-client => github.com/GuanceCloud/influxdb1-client v0.1.9
	github.com/iovisor/gobpf => github.com/DataDog/gobpf v0.0.0-20210322155958-9866ef4cd22c
	github.com/kardianos/service => github.com/GuanceCloud/service v1.2.4
	github.com/ory/dockertest/v3 v3.9.1 => github.com/GuanceCloud/dockertest/v3 v3.9.4
	github.com/prometheus/client_model => github.com/GuanceCloud/client_model v0.0.0-20230418154757-93bd4e878a5e
	github.com/prometheus/common => github.com/GuanceCloud/promcommon v0.0.0-20230828165048-8a8ac696e616
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20210720084720-59d02b5ef003
	github.com/pyroscope-io/pyroscope v0.36.0 => github.com/GuanceCloud/pyroscope v0.36.3
	github.com/sijms/go-ora/v2 => github.com/GuanceCloud/go-ora/v2 v2.8.20
	github.com/weaveworks/common => github.com/pyh4/common v0.0.0-20220923021349-874c8cc0db2c
	k8s.io/api => k8s.io/api v0.25.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.25.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.26.0-alpha.0
	k8s.io/apiserver => k8s.io/apiserver v0.25.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.25.0
	k8s.io/client-go => k8s.io/client-go v0.25.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.25.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.25.0
	k8s.io/code-generator => k8s.io/code-generator v0.25.1-rc.0
	k8s.io/component-base => k8s.io/component-base v0.25.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.25.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.25.0
	k8s.io/cri-api => k8s.io/cri-api v0.25.1-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.25.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.25.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.25.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.25.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.25.0
	k8s.io/kubectl => k8s.io/kubectl v0.25.0
	k8s.io/kubelet => k8s.io/kubelet v0.25.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.25.0
	k8s.io/metrics => k8s.io/metrics v0.25.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.25.3-rc.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.25.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.25.0
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.25.0
	k8s.io/sample-controller => k8s.io/sample-controller v0.25.0
)
