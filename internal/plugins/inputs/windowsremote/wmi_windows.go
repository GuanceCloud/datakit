// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package windowsremote

import (
	"github.com/GuanceCloud/cliutils/point"
	"sync"
	"time"
)

type Wmi struct {
	cfg *WmiConfig

	serversLock sync.RWMutex
	servers     map[string]*winServer
}

func (w *Wmi) Name() string {
	return "wmi"
}

func newWmi(cfg *WmiConfig) *Wmi {
	w := &Wmi{cfg: cfg, servers: make(map[string]*winServer)}
	// do something...
	// wmi.Log = l
	return w
}

// Close  关闭连接。
func (w *Wmi) Close() {
	for _, ser := range w.servers {
		_ = ser.conn.Close()
	}
}

func (w *Wmi) CollectMetric(ip string, timestamp int64) []*point.Point {
	l.Debugf("collect metric ip:%s timestamp:%d", ip, timestamp)
	if w.servers[ip] == nil {
		s := newServer(ip, w.cfg.Auth.Username, w.cfg.Auth.Password)
		if s == nil {
			l.Errorf("NewServer ip:%s error", ip)
			return []*point.Point{}
		}
		s.tags = w.cfg.extraTags
		w.serversLock.Lock()
		w.servers[ip] = s
		w.serversLock.Unlock()
	}
	// 对齐时间 timestamp 纳秒单位。
	start := time.Unix(0, timestamp)
	return w.servers[ip].collectMetric(start)
}

func (w *Wmi) CollectObject(ip string) []*point.Point {
	l.Debugf("wmi CollectObject ip:%s", ip)
	if w.servers[ip] == nil {
		s := newServer(ip, w.cfg.Auth.Username, w.cfg.Auth.Password)
		if s == nil {
			l.Errorf("NewServer ip:%s error", ip)
			return []*point.Point{}
		}
		s.tags = w.cfg.extraTags
		w.serversLock.Lock()
		w.servers[ip] = s
		w.serversLock.Unlock()
	}
	if w.servers[ip] != nil {
		w.servers[ip].beginCollectObject()
		return w.servers[ip].toObjectPoints()
	}
	return []*point.Point{}
}

func (w *Wmi) CollectLogging(ip string) []*point.Point {
	if !w.cfg.LogEnable {
		return []*point.Point{}
	}
	l.Debugf("wmi CollectLogging ip:%s", ip)
	if w.servers[ip] != nil {
		return w.servers[ip].collectLog()
	}
	return []*point.Point{}
}
