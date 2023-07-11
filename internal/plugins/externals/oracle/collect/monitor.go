// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collect

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Monitor struct {
	interval    string
	user        string
	password    string
	desc        string
	host        string
	port        string
	serviceName string
	tags        map[string]string
	election    bool

	db               *sqlx.DB
	intervalDuration time.Duration
	collectors       []dbMetricsCollector
	datakitPostURL   string
}

func NewMonitor() *Monitor {
	m := &Monitor{
		interval:    opt.Interval,
		user:        opt.Username,
		password:    opt.Password,
		desc:        opt.InstanceDesc,
		host:        opt.Host,
		port:        opt.Port,
		serviceName: opt.ServiceName,
		tags:        make(map[string]string),
		election:    opt.Election,
	}

	items := strings.Split(opt.Tags, ";")
	for _, item := range items {
		tagArr := strings.Split(item, "=")

		if len(tagArr) == 2 {
			tagKey := strings.Trim(tagArr[0], " ")
			tagVal := strings.Trim(tagArr[1], " ")
			if tagKey != "" {
				m.tags[tagKey] = tagVal
			}
		}
	}

	m.tags["oracle_service"] = m.serviceName
	m.tags["oracle_server"] = fmt.Sprintf("%s:%s", m.host, m.port)

	if m.interval != "" {
		du, err := time.ParseDuration(m.interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.interval, err.Error())
			m.intervalDuration = 10 * time.Minute
		} else {
			m.intervalDuration = du
		}
	}

	for {
		if err := m.ConnectDB(); err != nil {
			l.Errorf("oracle connect faild %v, retry each 3 seconds...", err)
			time.Sleep(time.Second * 3)
			continue
		}

		break
	}

	proMetric := newProcessMetrics(withMonitor(m))
	tsMetric := newTablespaceMetrics(withMonitor(m))
	sysMetric := newSystemMetrics(withMonitor(m))

	m.collectors = append(m.collectors, proMetric, tsMetric, sysMetric)
	l.Infof("collectors len = %d", len(m.collectors))

	var (
		ignoreGlobalHostTags = "ignore_global_host_tags=true"
		globalEnvTags        = "global_env_tags=true"
	)
	if opt.Election {
		m.datakitPostURL = fmt.Sprintf(
			"http://%s/v1/write/metric?input=oracle&%s&%s",
			net.JoinHostPort(opt.DatakitHTTPHost, strconv.Itoa(opt.DatakitHTTPPort)),
			ignoreGlobalHostTags, globalEnvTags)
	} else {
		m.datakitPostURL = fmt.Sprintf(
			"http://%s/v1/write/metric?input=oracle",
			net.JoinHostPort(opt.DatakitHTTPHost, strconv.Itoa(opt.DatakitHTTPPort)),
		)
	}
	l.Debugf("post to datakit URL: %s", m.datakitPostURL)

	return m
}

// ConnectDB establishes a connection to an Oracle instance and returns an open connection to the database.
func (m *Monitor) ConnectDB() error {
	db, err := sqlx.Open("godror",
		fmt.Sprintf("%s/%s@%s:%s/%s",
			m.user, m.password, m.host, m.port, m.serviceName))
	if err == nil {
		m.db = db
		return err
	}
	return nil
}

// CloseDB cleans up database resources used.
func (m *Monitor) CloseDB() {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			l.Warnf("failed to close oracle connection | server=[%s]: %s", m.host, err.Error())
		}
	}
}

func (m *Monitor) Run() {
	l.Info("start oracle...")

	tick := time.NewTicker(m.intervalDuration)
	defer tick.Stop()
	defer m.db.Close() //nolint:errcheck

	for {
		sa := newSafeArray()

		for idx := range m.collectors {
			pt, err := m.collectors[idx].collect()
			if err != nil {
				l.Errorf("Collect failed: %v", err)
			} else {
				line := pt.LineProto()
				sa.add(line)
			}
		}

		if sa.len() > 0 {
			if err := writeData(bytes.Join(sa.get(), []byte("\n")), m.datakitPostURL); err != nil {
				l.Errorf("writeData failed: %v", err)
			}
		}

		<-tick.C
	}
}
