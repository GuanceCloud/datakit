package config

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

func (c *Config) LoadEnvs() error {
	if c.IOConf == nil {
		c.IOConf = &dkio.IOConfig{}
	}

	for _, envkey := range []string{
		"ENV_MAX_CACHE_COUNT",
		"ENV_CACHE_DUMP_THRESHOLD",
		"ENV_MAX_DYNAMIC_CACHE_COUNT",
		"ENV_DYNAMIC_CACHE_DUMP_THRESHOLD",
	} {
		if v := datakit.GetEnv(envkey); v != "" {
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				l.Errorf("invalid env key value pair [%s:%s], ignored", envkey, v)
				continue
			}

			switch envkey {
			case "ENV_MAX_CACHE_COUNT":
				c.IOConf.MaxCacheCount = value
			case "ENV_CACHE_DUMP_THRESHOLD":
				c.IOConf.CacheDumpThreshold = value
			case "ENV_MAX_DYNAMIC_CACHE_COUNT":
				c.IOConf.MaxDynamicCacheCount = value
			case "ENV_DYNAMIC_CACHE_DUMP_THRESHOLD":
				c.IOConf.DynamicCacheDumpThreshold = value
			}
		}
	}

	if v := datakit.GetEnv("ENV_IPDB"); v != "" {
		switch v {
		case "iploc":
			c.Pipeline.IPdbType = v
		default:
			l.Warnf("unknown IPDB type: %s, ignored", v)
		}
	}

	if v := datakit.GetEnv("ENV_REQUEST_RATE_LIMIT"); v != "" {
		if x, err := strconv.ParseFloat(v, 64); err != nil {
			l.Warnf("invalid ENV_REQUEST_RATE_LIMIT, expect int or float, got %s, ignored", v)
		} else {
			c.HTTPAPI.RequestRateLimit = x
		}
	}

	if v := datakit.GetEnv("ENV_K8S_NODE_NAME"); v != "" {
		c.Hostname = v
		datakit.DatakitHostName = c.Hostname
	}

	if v := datakit.GetEnv("ENV_NAMESPACE"); v != "" {
		c.Namespace = v
	}

	if v := datakit.GetEnv("ENV_ENABLE_ELECTION"); v != "" {
		c.EnableElection = true
	}

	if v := datakit.GetEnv("ENV_GLOBAL_TAGS"); v != "" {
		c.GlobalTags = ParseGlobalTags(v)
	}

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

	// 多个 dataway 支持 ',' 分割
	if v := datakit.GetEnv("ENV_DATAWAY"); v != "" {
		if c.DataWayCfg == nil {
			c.DataWayCfg = &dataway.DataWayCfg{}
		}
		c.DataWayCfg.URLs = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_DATAWAY_TIMEOUT"); v != "" {
		if c.DataWayCfg == nil {
			c.DataWayCfg = &dataway.DataWayCfg{}
		}
		_, err := time.ParseDuration(v)
		if err != nil {
			l.Warnf("invalid ENV_DATAWAY_TIMEOUT: %s", v)
			c.DataWayCfg.HTTPTimeout = "30s"
		} else {
			c.DataWayCfg.HTTPTimeout = v
		}
	}

	if v := datakit.GetEnv("ENV_DATAWAY_ENABLE_HTTPTRACE"); v != "" {
		c.DataWayCfg.EnableHTTPTrace = true
	}

	if v := datakit.GetEnv("ENV_DATAWAY_HTTP_PROXY"); v != "" {
		c.DataWayCfg.HTTPProxy = v
		c.DataWayCfg.Proxy = true
	}

	if v := datakit.GetEnv("ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST"); v != "" {
		if c.DataWayCfg == nil {
			c.DataWayCfg = &dataway.DataWayCfg{}
		}
		value, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			if value <= 0 {
				l.Warnf("invalid ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST: %s", v)
			} else {
				c.DataWayCfg.MaxIdleConnsPerHost = int(value)
			}
		}
	}

	if v := datakit.GetEnv("ENV_HOSTNAME"); v != "" {
		c.Hostname = v
	}

	if v := datakit.GetEnv("ENV_NAME"); v != "" {
		c.Name = v
	}

	// HTTP server setting
	if v := datakit.GetEnv("ENV_HTTP_LISTEN"); v != "" {
		c.HTTPAPI.Listen = v
	}

	if v := datakit.GetEnv("ENV_HTTP_PUBLIC_APIS"); v != "" {
		c.HTTPAPI.PublicAPIs = strings.Split(v, ",")
	}

	// filters
	if v := datakit.GetEnv("ENV_IO_FILTERS"); v != "" {
		var x map[string][]string
		if err := json.Unmarshal([]byte(v), &x); err != nil {
			l.Warnf("json.Unmarshal: %s, ignored", err)
		} else {
			for k, arr := range x {
				for _, c := range arr {
					if parser.GetConds(c) == nil {
						l.Warnf("invalid filter condition on %s: %s, ignored", k, c)
					} else {
						l.Debugf("filter condition ok: %s", c)
					}
				}
			}

			c.IOConf.Filters = x
		}
	}

	// DCA settings
	if v := datakit.GetEnv("ENV_DCA_LISTEN"); v != "" {
		c.DCAConfig.Enable = true
		c.DCAConfig.Listen = v
	}

	if v := datakit.GetEnv("ENV_DCA_WHITE_LIST"); v != "" {
		c.DCAConfig.WhiteList = strings.Split(v, ",")
	}

	// RUM related
	if v := datakit.GetEnv("ENV_RUM_ORIGIN_IP_HEADER"); v != "" {
		c.HTTPAPI.RUMOriginIPHeader = v
	}

	if v := datakit.GetEnv("ENV_RUM_APP_ID_WHITE_LIST"); v != "" {
		c.HTTPAPI.RUMAppIDWhiteList = strings.Split(v, ",")
	}

	if v := datakit.GetEnv("ENV_DISABLE_404PAGE"); v != "" {
		c.HTTPAPI.Disable404Page = true
	}

	if v := datakit.GetEnv("ENV_ENABLE_PPROF"); v != "" {
		c.EnablePProf = true
	}

	if v := datakit.GetEnv("ENV_PPROF_LISTEN"); v != "" {
		c.PProfListen = v
	}

	if v := datakit.GetEnv("ENV_DISABLE_PROTECT_MODE"); v != "" {
		c.ProtectMode = false
	}

	if v := datakit.GetEnv("ENV_DEFAULT_ENABLED_INPUTS"); v != "" {
		c.DefaultEnabledInputs = strings.Split(v, ",")
	} else if v := datakit.GetEnv("ENV_ENABLE_INPUTS"); v != "" { // deprecated
		c.DefaultEnabledInputs = strings.Split(v, ",")
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
				}, // GitRepository
			}, // Repos
		} // GitRepost
	}

	if err := c.getSinkConfig(); err != nil {
		l.Fatalf("getSinkConfig failed: %v", err)
		return err
	}

	return nil
}
