// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/pipeline-go/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
)

func (c *Config) loadConfdEnvs() {
	// k8s ENV confd
	if backend := datakit.GetEnv("ENV_CONFD_BACKEND"); backend != "" {
		authToken := datakit.GetEnv("ENV_CONFD_AUTH_TOKEN")
		authType := datakit.GetEnv("ENV_CONFD_AUTH_TYPE")
		basicAuthBool := datakit.GetEnv("ENV_CONFD_BASIC_AUTH")
		clientCaKeys := datakit.GetEnv("ENV_CONFD_CLIENT_CA_KEYS")
		clientCert := datakit.GetEnv("ENV_CONFD_CLIENT_CERT")
		clientKey := datakit.GetEnv("ENV_CONFD_CLIENT_KEY")
		clientInsecureBool := datakit.GetEnv("ENV_CONFD_CLIENT_INSECURE")
		backendNodesArry := datakit.GetEnv("ENV_CONFD_BACKEND_NODES")
		password := datakit.GetEnv("ENV_CONFD_PASSWORD")
		scheme := datakit.GetEnv("ENV_CONFD_SCHEME")
		table := datakit.GetEnv("ENV_CONFD_TABLE")
		separator := datakit.GetEnv("ENV_CONFD_SEPARATOR")
		username := datakit.GetEnv("ENV_CONFD_USERNAME")
		appID := datakit.GetEnv("ENV_CONFD_APP_ID")
		userID := datakit.GetEnv("ENV_CONFD_USER_ID")
		roleID := datakit.GetEnv("ENV_CONFD_ROLE_ID")
		secretID := datakit.GetEnv("ENV_CONFD_SECRET_ID")
		filter := datakit.GetEnv("ENV_CONFD_FILTER")
		path := datakit.GetEnv("ENV_CONFD_PATH")
		role := datakit.GetEnv("ENV_CONFD_ROLE")
		accessKey := datakit.GetEnv("ENV_CONFD_ACCESS_KEY")
		secretKey := datakit.GetEnv("ENV_CONFD_SECRET_KEY")
		circleIntervalInt := datakit.GetEnv("ENV_CONFD_CIRCLE_INTERVAL")
		confdNamespace := datakit.GetEnv("ENV_CONFD_CONFD_NAMESPACE")
		pipelineNamespace := datakit.GetEnv("ENV_CONFD_PIPELINE_NAMESPACE")
		region := datakit.GetEnv("ENV_CONFD_REGION")

		// some data types need to be converted
		var backendNodes []string
		err := json.Unmarshal([]byte(backendNodesArry), &backendNodes)
		if err != nil {
			l.Warnf("parse ENV_CONFD_BACKEND_NODES: %s, ignore", err)
			backendNodes = make([]string, 0)
		}
		basicAuth := false
		if basicAuthBool == "true" {
			basicAuth = true
		}
		clientInsecure := false
		if clientInsecureBool == "true" {
			clientInsecure = true
		}
		circleInterval := 60
		if interval, err := strconv.Atoi(circleIntervalInt); err == nil {
			circleInterval = interval
		} else {
			l.Warnf("parse ENV_CONFD_CIRCLE_INTERVAL: %s, ignore", err)
		}

		c.Confds = append(c.Confds, &ConfdCfg{
			Enable:            true,
			Backend:           backend,
			AuthToken:         authToken,
			AuthType:          authType,
			BasicAuth:         basicAuth,
			ClientCaKeys:      clientCaKeys,
			ClientCert:        clientCert,
			ClientKey:         clientKey,
			ClientInsecure:    clientInsecure,
			BackendNodes:      append(backendNodes[0:0], backendNodes...),
			Password:          password,
			Scheme:            scheme,
			Table:             table,
			Separator:         separator,
			Username:          username,
			AppID:             appID,
			UserID:            userID,
			RoleID:            roleID,
			SecretID:          secretID,
			Filter:            filter,
			Path:              path,
			Role:              role,
			AccessKey:         accessKey,
			SecretKey:         secretKey,
			CircleInterval:    circleInterval,
			ConfdNamespace:    confdNamespace,
			PipelineNamespace: pipelineNamespace,
			Region:            region,
		})
	}

	if v := datakit.GetEnv("ENV_GIT_URL"); v != "" {
		interval := datakit.GetEnv("ENV_GIT_INTERVAL")
		keyPath := datakit.GetEnv("ENV_GIT_KEY_PATH")
		keyPasswd := datakit.GetEnv("ENV_GIT_KEY_PW")
		branch := datakit.GetEnv("ENV_GIT_BRANCH")

		c.GitRepos = &GitRepost{
			PullInterval: interval,
			Repos: []*GitRepository{
				{
					Enable:                true,
					URL:                   v,
					SSHPrivateKeyPath:     keyPath,
					SSHPrivateKeyPassword: keyPasswd,
					Branch:                branch,
				},
			},
		}
	}
}

func (c *Config) loadPprofEnvs() {
	if v := datakit.GetEnv("ENV_ENABLE_DEBUG_FIELDS"); v != "" {
		b, _ := strconv.ParseBool(v)
		c.EnableDebugFields = b
	}

	if v := datakit.GetEnv("ENV_ENABLE_PPROF"); v != "" {
		c.EnablePProf = true
	}

	if v := datakit.GetEnv("ENV_PPROF_LISTEN"); v != "" {
		c.PProfListen = v
	}
}

func (c *Config) loadDCAEnvs() {
	// DCA settings
	if v := datakit.GetEnv("ENV_DCA_WEBSOCKET_SERVER"); v != "" {
		c.DCAConfig.Enable = true
		c.DCAConfig.WebsocketServer = v
	}
}

func (c *Config) loadLogEnvs() {
	// set logging
	if v := datakit.GetEnv("ENV_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}

	if v := datakit.GetEnv("ENV_LOG"); v != "" {
		c.Logging.Log = v
	}

	if v := datakit.GetEnv("ENV_GIN_LOG"); v != "" {
		c.Logging.GinLog = v
	}

	if v := datakit.GetEnv("ENV_DISABLE_LOG_COLOR"); v != "" {
		c.Logging.DisableColor = true
	}

	if v := datakit.GetEnv("ENV_LOG_ROTATE_BACKUP"); v != "" {
		count, err := strconv.Atoi(v)
		if err != nil || count <= 0 {
			l.Warnf("invalid value for ENV_LOG_ROTATE_BACKUP, need positive number but got [%s], use default value instead", v)
			count = logger.MaxBackups
		}
		c.Logging.RotateBackups = count
	}

	if v := datakit.GetEnv("ENV_LOG_ROTATE_SIZE_MB"); v != "" {
		size, err := strconv.Atoi(v)
		if err != nil || size <= 0 {
			l.Warnf("invalid value for ENV_LOG_ROTATE_SIZE_MB, need positive number but got [%s], use default value instead", v)
			size = logger.MaxSize
		}
		c.Logging.Rotate = size
	}
}

func (c *Config) loadPipelineEnvs() {
	if v := datakit.GetEnv("ENV_IPDB"); v != "" {
		switch v {
		case "iploc":
			c.Pipeline.IPdbType = v
		default:
			l.Warnf("unknown IPDB type: %s, ignored", v)
		}
	}

	if v := datakit.GetEnv("ENV_REFER_TABLE_URL"); v != "" {
		c.Pipeline.ReferTableURL = v
	}

	if v := datakit.GetEnv("ENV_REFER_TABLE_PULL_INTERVAL"); v != "" {
		c.Pipeline.ReferTablePullInterval = v
	}

	if v := datakit.GetEnv("ENV_REFER_TABLE_USE_SQLITE"); v != "" {
		c.Pipeline.UseSQLite = true
	}

	if v := datakit.GetEnv("ENV_REFER_TABLE_SQLITE_MEM_MODE"); v != "" {
		c.Pipeline.SQLiteMemMode = true
	}

	if v := datakit.GetEnv("ENV_PIPELINE_OFFLOAD_RECEIVER"); v != "" {
		if c.Pipeline.Offload == nil {
			c.Pipeline.Offload = &offload.OffloadConfig{
				Receiver: v,
			}
		} else {
			c.Pipeline.Offload.Receiver = v
		}
	}

	if v := datakit.GetEnv("ENV_PIPELINE_OFFLOAD_ADDRESSES"); v != "" {
		if c.Pipeline.Offload == nil {
			c.Pipeline.Offload = &offload.OffloadConfig{
				Receiver:  offload.DKRcv,
				Addresses: strings.Split(v, ","),
			}
		} else {
			if c.Pipeline.Offload.Receiver == "" {
				c.Pipeline.Offload.Receiver = offload.DKRcv
			}
			c.Pipeline.Offload.Addresses = strings.Split(v, ",")
		}
	}

	if v := datakit.GetEnv("ENV_PIPELINE_DEFAULT_PIPELINE"); v != "" {
		var result map[string]string
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			l.Errorf("unmarshal `ENV_PIPELINE_DEFAULT_PIPELINE` failed: %s",
				err.Error())
		} else {
			c.Pipeline.DefaultPipeline = result
		}
	}

	if v := datakit.GetEnv("ENV_PIPELINE_DISABLE_HTTP_REQUEST_FUNC"); v != "" {
		var err error
		c.Pipeline.DisableHTTPRequestFunc, err = strconv.ParseBool(v)
		if err != nil {
			l.Errorf("parse `ENV_PIPELINE_DISABLE_HTTP_REQUEST_FUNC` failed: %s", err)
		}
	}

	if v := datakit.GetEnv("ENV_PIPELINE_HTTP_REQUEST_HOST_WHITELIST"); v != "" {
		var r []string
		if err := json.Unmarshal([]byte(v), &r); err != nil {
			l.Errorf("parse `ENV_PIPELINE_HTTP_REQUEST_HOST_WHITELIST` failed: %s", err)
		}
		c.Pipeline.HTTPRequestHostWhitelist = r
	}

	if v := datakit.GetEnv("ENV_PIPELINE_HTTP_REQUEST_CIDR_WHITELIST"); v != "" {
		var r []string
		if err := json.Unmarshal([]byte(v), &r); err != nil {
			l.Errorf("parse `ENV_PIPELINE_HTTP_REQUEST_CIDR_WHITELIST` failed: %s", err)
		}
		c.Pipeline.HTTPRequestCIDRWhitelist = r
	}

	if v := datakit.GetEnv("ENV_PIPELINE_HTTP_REQUEST_DISABLE_INTERNAL_NET"); v != "" {
		var err error
		c.Pipeline.HTTPRequestDisableInternalNet, err = strconv.ParseBool(v)
		if err != nil {
			l.Errorf("parse `ENV_PIPELINE_HTTP_REQUEST_DISABLE_INTERNAL_NET` failed: %s", err)
		}
	}
}

func (c *Config) loadPointPoolEnvs() {
	if v := datakit.GetEnv("ENV_DISABLE_POINT_POOL"); v != "" {
		l.Warn("point pool disabled, this may cost too many memory")
		c.PointPool.Enable = false
	}

	if v := datakit.GetEnv("ENV_POINT_POOL_RESERVED_CAPACITY"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.PointPool.ReservedCapacity = i
		} else {
			l.Warnf("invalid ENV_POINT_POOL_RESERVED_CAPACITY: %s, use default %d", v, c.PointPool.ReservedCapacity)
		}
	}
}

func (c *Config) loadDatawayEnvs() {
	if v := datakit.GetEnv("ENV_DATAWAY_TLS_INSECURE"); v != "" {
		// NOTE: do not checking encoding here, invalid encoding will reset to line-protocol
		l.Info("ENV_DATAWAY_TLS_INSECURE set to true")
		c.Dataway.InsecureSkipVerify = true
	}

	if v := datakit.GetEnv("ENV_DATAWAY_CONTENT_ENCODING"); v != "" {
		// NOTE: do not checking encoding here, invalid encoding will reset to line-protocol
		l.Infof("ENV_DATAWAY_CONTENT_ENCODING set to %q", v)
		c.Dataway.ContentEncoding = v
	}

	if v := datakit.GetEnv("ENV_DATAWAY_DISABLE_GZIP"); v != "" {
		// NOTE: list the entry here only for test.
		// Do NOT enable this ENV, kodo only accept gzip /v1/write/ payload
		l.Info("ENV_DATAWAY_GZIP disabled")
		c.Dataway.GZip = false
	}

	if v := datakit.GetEnv("ENV_DATAWAY_MAX_RAW_BODY_SIZE"); v != "" {
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Warnf("invalid ENV_DATAWAY_MAX_RAW_BODY_SIZE: %s", v)
		} else {
			l.Info("ENV_DATAWAY_MAX_RAW_BODY_SIZE set to %q", v)

			if value < dataway.MinimalRawBodySize && c.ProtectMode {
				l.Info("ENV_DATAWAY_MAX_RAW_BODY_SIZE(%q) too small, not applied", v)
				c.Dataway.MaxRawBodySize = dataway.MinimalRawBodySize
			} else {
				c.Dataway.MaxRawBodySize = int(value)
			}
		}
	}

	// 多个 dataway 支持 ',' 分割
	// c.Dataway should not nil
	if v := datakit.GetEnv("ENV_DATAWAY"); v != "" {
		c.Dataway.URLs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_DATAWAY_TIMEOUT"); v != "" {
		du, err := time.ParseDuration(v)
		if err != nil {
			l.Warnf("invalid ENV_DATAWAY_TIMEOUT: %s", v)
			c.Dataway.HTTPTimeout = time.Second * 30
		} else {
			c.Dataway.HTTPTimeout = du
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_ENABLE_HTTPTRACE"); v != "" {
		c.Dataway.EnableHTTPTrace = true
	}

	if v := datakit.GetEnv("ENV_DATAWAY_HTTP_PROXY"); v != "" {
		c.Dataway.HTTPProxy = v
	}

	if v := datakit.GetEnv("ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST"); v != "" {
		value, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			if value <= 0 {
				l.Warnf("invalid ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST(%s): %s, ignored", v, err)
			} else {
				c.Dataway.MaxIdleConnsPerHost = int(value)
			}
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_MAX_IDLE_CONNS"); v != "" {
		value, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			if value <= 0 {
				l.Warnf("invalid ENV_DATAWAY_MAX_IDLE_CONNS(%q): %s, ignored", v)
			} else {
				c.Dataway.MaxIdleConns = int(value)
			}
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_IDLE_TIMEOUT"); v != "" {
		du, err := time.ParseDuration(v)
		if err == nil {
			c.Dataway.IdleTimeout = du
		} else {
			l.Warnf("invalid ENV_DATAWAY_IDLE_TIMEOUT(%q): %s, ignored", v, err)
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_MAX_RETRY_COUNT"); v != "" {
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Warnf("invalid ENV_DATAWAY_MAX_RETRY_COUNT: %q, must be an integer number", v)
		} else {
			if value < 1 {
				l.Warnf("invalid ENV_DATAWAY_MAX_RETRY_COUNT: %q, must be greater then 0", v)
			} else {
				if value > 10 {
					value = 10
				}
				c.Dataway.MaxRetryCount = int(value)
			}
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_RETRY_DELAY"); v != "" {
		value, err := time.ParseDuration(v)
		if err != nil {
			l.Warnf("invalid ENV_DATAWAY_RETRY_DELAY: %q, must be an valid golang duration", v)
		} else {
			if value < 0 {
				l.Warnf("invalid ENV_DATAWAY_RETRY_DELAY: %q, must not be a negative duration, ignored", v)
			} else {
				c.Dataway.RetryDelay = value
			}
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_ENABLE_SINKER"); v != "" {
		c.Dataway.EnableSinker = true
		l.Infof("enable sinker on dataway")
	}

	if v := datakit.GetEnv("ENV_SINKER_GLOBAL_CUSTOMER_KEYS"); v != "" {
		c.Dataway.GlobalCustomerKeys = dataway.ParseGlobalCustomerKeys(v)
		l.Infof("set global custom keys to %v", c.Dataway.GlobalCustomerKeys)
	}

	if c.Dataway.NTP != nil {
		if v := datakit.GetEnv("ENV_DATAWAY_NTP_INTERVAL"); v != "" {
			if du, err := time.ParseDuration(v); err == nil {
				c.Dataway.NTP.Interval = du
			} else {
				l.Warnf("invalid ENV_DATAWAY_NTP_INTERVAL: %q: %s, ignored", v, err.Error())
			}
		}

		if v := datakit.GetEnv("ENV_DATAWAY_NTP_DIFF"); v != "" {
			if du, err := time.ParseDuration(v); err == nil {
				c.Dataway.NTP.SyncOnDiff = du
			} else {
				l.Warnf("invalid ENV_DATAWAY_NTP_DIFF: %q: %s, ignored", v, err.Error())
			}
		}
	}

	// WAL
	if c.Dataway.WAL != nil {
		if v := datakit.GetEnv("ENV_DATAWAY_WAL_CAPACITY"); v != "" {
			if x, err := strconv.ParseFloat(v, 64); err != nil {
				l.Warnf("invalid ENV_DATAWAY_WAL_CAPACITY, expect int or float, got %s, ignored", v)
			} else {
				c.Dataway.WAL.MaxCapacityGB = x
			}
		}

		if v := datakit.GetEnv("ENV_DATAWAY_WAL_WORKERS"); v != "" {
			if x, err := strconv.ParseInt(v, 10, 64); err != nil {
				l.Warnf("invalid ENV_DATAWAY_WAL_WORKERS, expect int, got %s, ignored", v)
			} else {
				c.Dataway.WAL.Workers = int(x)
			}
		}

		if v := datakit.GetEnv("ENV_DATAWAY_WAL_MEM_CAPACITY"); v != "" {
			if x, err := strconv.ParseInt(v, 10, 64); err != nil {
				l.Warnf("invalid ENV_DATAWAY_WAL_MEM_CAPACITY, expect int, got %s, ignored", v)
			} else {
				c.Dataway.WAL.MemCap = int(x)
			}
		}

		if v := datakit.GetEnv("ENV_DATAWAY_WAL_PATH"); v != "" {
			c.Dataway.WAL.Path = v
		}

		if v := datakit.GetEnv("ENV_DATAWAY_WAL_FAIL_CACHE_CLEAN_INTERVAL"); v != "" {
			if x, err := time.ParseDuration(v); err != nil {
				l.Warnf("invalid ENV_DATAWAY_WAL_FAIL_CACHE_CLEAN_INTERVAL, expect duration, got %s, ignored", v)
			} else {
				c.Dataway.WAL.FailCacheCleanInterval = x
			}
		}

		if v := datakit.GetEnv("ENV_DATAWAY_WAL_NO_DROP_CATEGORIES"); v != "" {
			c.Dataway.WAL.NoDropCategories = strings.Split(v, ",")
		}
	} else {
		l.Errorf("WAL not set, should not been here")
	}
}

func (c *Config) loadRemoteJobEnvs() {
	if s := datakit.GetEnv("ENV_REMOTE_JOB_ENABLE"); s == "" {
		return
	}
	if c.RemoteJob == nil {
		c.RemoteJob = &io.RemoteJob{}
	}
	c.RemoteJob.Enable = true
	if v := datakit.GetEnv("ENV_REMOTE_JOB_ENVS"); v != "" {
		c.RemoteJob.ENVs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_REMOTE_JOB_JAVA_HOME"); v != "" {
		c.RemoteJob.JavaHome = v
	}
	if v := datakit.GetEnv("ENV_REMOTE_JOB_INTERVAL"); v != "" {
		c.RemoteJob.Interval = v
	}
}

func (c *Config) loadElectionEnvs() {
	if v := datakit.GetEnv("ENV_ENABLE_ELECTION"); v == "" {
		return
	}

	c.Election.Enable = true

	if s := datakit.GetEnv("ENV_ELECTION_NODE_WHITELIST"); s != "" {
		err := json.Unmarshal([]byte(s), &c.Election.NodeWhitelist)
		if err != nil {
			c.Election.NodeWhitelist = strings.Split(s, ",")
		}
	}

	// default election namespace is `default`
	if v := datakit.GetEnv("ENV_NAMESPACE"); v != "" {
		c.Election.Namespace = v
	}

	if v := datakit.GetEnv("ENV_ENABLE_ELECTION_NAMESPACE_TAG"); v != "" {
		// add to global-env-tags
		c.Election.EnableNamespaceTag = true
		c.Election.Tags["election_namespace"] = c.Election.Namespace
	}

	for _, x := range []string{
		"ENV_GLOBAL_ELECTION_TAGS",
		"ENV_GLOBAL_ENV_TAGS", // Deprecated
	} {
		if v := datakit.GetEnv(x); v != "" {
			for k, v := range ParseGlobalTags(v) {
				c.Election.Tags[k] = v
			}
			break
		}
	}
}

func (c *Config) loadRecorderEnvs() {
	if v := datakit.GetEnv("ENV_ENABLE_RECORDER"); v == "" {
		return
	}

	c.Recorder.Enabled = true

	if v := datakit.GetEnv("ENV_RECORDER_PATH"); v != "" {
		c.Recorder.Path = v
	}

	if v := datakit.GetEnv("ENV_RECORDER_ENCODING"); v != "" {
		c.Recorder.Encoding = v
	}

	if v := datakit.GetEnv("ENV_RECORDER_DURATION"); v != "" {
		du, err := time.ParseDuration(v)
		if err == nil {
			c.Recorder.Duration = du
		}
	}

	if v := datakit.GetEnv("ENV_RECORDER_INPUTS"); v != "" {
		c.Recorder.Inputs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_RECORDER_CATEGORIES"); v != "" {
		c.Recorder.Categories = strings.Split(v, ",")
	}
}

func (c *Config) loadIOEnvs() {
	if v := datakit.GetEnv("ENV_IO_MAX_CACHE_COUNT"); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Warnf("invalid env key ENV_IO_MAX_CACHE_COUNT, value %s, ignored", v)
		} else {
			if val < 1000 {
				l.Warnf("reset cache count from %d to %d", val, 1000)
			} else {
				l.Infof("set cache count to %d", val)
				c.IO.MaxCacheCount = int(val)
			}
		}
	}

	if v := datakit.GetEnv("ENV_IO_FLUSH_WORKERS"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Warnf("invalid env key ENV_IO_FLUSH_WORKERS, value %s, err: %s ignored", v, err)
		} else {
			l.Infof("set ENV_IO_FLUSH_WORKERS to %d", n)
			c.IO.CompactWorkers = int(n)
		}
	}

	if v := datakit.GetEnv("ENV_IO_FLUSH_INTERVAL"); v != "" {
		du, err := time.ParseDuration(v)
		if err != nil {
			l.Warnf("invalid env key ENV_IO_FLUSH_INTERVAL, value %s, err: %s ignored", v, err)
		} else {
			l.Infof("set ENV_IO_FLUSH_INTERVAL to %s", du)
			c.IO.CompactInterval = du
		}
	}

	if v := datakit.GetEnv("ENV_IO_FEED_CHAN_SIZE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Warnf("invalid env key ENV_IO_FEED_CHAN_SIZE, value %s, err: %s ignored", v, err)
		} else {
			l.Infof("set ENV_IO_FEED_CHAN_SIZE to %d", n)
			c.IO.FeedChanSize = int(n)
		}
	}

	// filters
	if v := datakit.GetEnv("ENV_IO_FILTERS"); v != "" {
		var x map[string]filter.FilterConditions

		if err := json.Unmarshal([]byte(v), &x); err != nil {
			l.Warnf("json.Unmarshal: %s, ignored", err)
		} else {
			for k, arr := range x {
				for _, c := range arr {
					arr, err := fp.GetConds(c)
					if err != nil {
						l.Warnf("parse filter condition failed %q: %q, ignored", k, c)
					}

					if len(arr) == 0 {
						l.Warnf("empty filter conditions %q", c)
					} else {
						l.Infof("filter condition ok %q", c)
					}
				}
			}

			c.IO.Filters = x
		}
	}
}

func (c *Config) loadHTTPAPIEnvs() {
	if v := datakit.GetEnv("ENV_REQUEST_RATE_LIMIT"); v != "" {
		if x, err := strconv.ParseFloat(v, 64); err != nil {
			l.Warnf("invalid ENV_REQUEST_RATE_LIMIT, expect int or float, got %s, ignored", v)
		} else {
			c.HTTPAPI.RequestRateLimit = x
		}
	}

	// HTTP server setting
	if v := datakit.GetEnv("ENV_HTTP_LISTEN"); v != "" {
		c.HTTPAPI.Listen = v
	}

	if v := datakit.GetEnv("ENV_HTTP_LISTEN_SOCKET"); v != "" {
		c.HTTPAPI.ListenSocket = v
	}

	if v := datakit.GetEnv("ENV_HTTP_TIMEOUT"); v != "" {
		c.HTTPAPI.Timeout = v
	}

	if v := datakit.GetEnv("ENV_HTTP_CLOSE_IDLE_CONNECTION"); v != "" {
		c.HTTPAPI.CloseIdleConnection = true
	}

	if v := datakit.GetEnv("ENV_HTTP_PUBLIC_APIS"); v != "" {
		c.HTTPAPI.PublicAPIs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_HTTP_ALLOWED_CORS_ORIGINS"); v != "" {
		c.HTTPAPI.AllowedCORSOrigins = strings.Split(v, ",")
	}

	// RUM related
	if v := datakit.GetEnv("ENV_RUM_ORIGIN_IP_HEADER"); v != "" {
		c.HTTPAPI.RUMOriginIPHeader = v
	}

	if v := datakit.GetEnv("ENV_RUM_APP_ID_WHITE_LIST"); v != "" {
		c.HTTPAPI.RUMAppIDWhiteList = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_ALLOWED_CORS_ORIGINS"); v != "" {
		c.HTTPAPI.AllowedCORSOrigins = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_DISABLE_404PAGE"); v != "" {
		c.HTTPAPI.Disable404Page = true
	}

	if v := datakit.GetEnv("ENV_HTTP_ENABLE_TLS"); v != "" {
		if v := datakit.GetEnv("ENV_HTTP_TLS_CRT"); v != "" {
			c.HTTPAPI.TLSConf.Cert = v
		}

		if v := datakit.GetEnv("ENV_HTTP_TLS_KEY"); v != "" {
			c.HTTPAPI.TLSConf.PrivKey = v
		}
	}
}

func (c *Config) setNodenameAsHostname() {
	for _, x := range []string{
		"ENV_K8S_NODE_NAME",
		"NODE_NAME", // Deprecated
	} {
		if v := datakit.GetEnv(x); v != "" {
			c.Hostname = v
			datakit.DatakitHostName = c.Hostname
			break
		}
	}

	if v := datakit.GetEnv("ENV_K8S_CLUSTER_NODE_NAME"); v != "" {
		l.Infof("ENV_K8S_CLUSTER_NODE_NAME set to %s", v)
		c.Hostname = v
		datakit.DatakitHostName = c.Hostname
	}
}

//nolint:funlen
func (c *Config) LoadEnvs() error {
	if c.IO == nil {
		c.IO = &io.IOConf{}
	}

	// Save inputs .conf form env to disk.
	if v := datakit.GetEnv("ENV_DATAKIT_INPUTS"); v != "" {
		p := filepath.Join(datakit.ConfdDir, "ENV_DATAKIT_INPUTS.conf")
		if err := os.WriteFile(p, []byte(v), datakit.ConfPerm); err != nil {
			l.Errorf("error creating %s: %s", p, err)
			return err
		}
	}

	// first load protect mode settings, other settings depends on this flag.
	if v := datakit.GetEnv("ENV_DISABLE_PROTECT_MODE"); v != "" {
		c.ProtectMode = false
	}

	c.loadIOEnvs()
	c.loadRecorderEnvs()
	c.loadPipelineEnvs()
	c.loadHTTPAPIEnvs()
	c.loadElectionEnvs()

	for _, x := range []string{
		"ENV_GLOBAL_HOST_TAGS",
		"ENV_GLOBAL_TAGS", // Deprecated
	} {
		if v := datakit.GetEnv(x); v != "" {
			for k, v := range ParseGlobalTags(v) {
				c.GlobalHostTags[k] = v
			}
			break
		}
	}

	c.loadLogEnvs()
	c.loadDatawayEnvs()
	c.loadPointPoolEnvs()

	c.setNodenameAsHostname()
	c.loadRemoteJobEnvs()

	// Don't Add to ElectionTags.
	if v := datakit.GetEnv("ENV_CLUSTER_NAME_K8S"); v != "" {
		c.GlobalHostTags["cluster_name_k8s"] = v
	}

	// misc
	if v := datakit.GetEnv("ENV_HOSTNAME"); v != "" {
		c.Hostname = v
	}

	if v := datakit.GetEnv("ENV_NAME"); v != "" {
		c.Name = v
	}

	c.loadDCAEnvs()
	c.loadPprofEnvs()

	for _, x := range []string{
		"ENV_DEFAULT_ENABLED_INPUTS",
		"ENV_ENABLE_INPUTS", // Deprecated
	} {
		if v := datakit.GetEnv(x); v != "" {
			if v == "-" {
				l.Warnf("no default inputs enabled!")
			} else {
				c.DefaultEnabledInputs = strings.Split(v, ",")
			}
			break
		}
	}

	if v := datakit.GetEnv("ENV_ULIMIT"); v != "" {
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			l.Warnf("invalid ulimit input through ENV_ULIMIT: %v", err)
		} else {
			c.Ulimit = u
		}
	}

	if v := datakit.GetEnv("ENV_CRYPTO_AES_KEY"); v != "" {
		if c.Crypto == nil {
			c.Crypto = &configCrpto{}
		}
		c.Crypto.AESKey = v
	}

	if v := datakit.GetEnv("ENV_CRYPTO_AES_KEY_FILE"); v != "" {
		if c.Crypto == nil {
			c.Crypto = &configCrpto{}
		}
		c.Crypto.AESKeyFile = v
	}

	c.loadConfdEnvs()

	return nil
}
