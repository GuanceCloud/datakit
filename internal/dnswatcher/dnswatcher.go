// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dnswatcher contains dns watcher control logic
package dnswatcher

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/compareutil"
)

const (
	packageName          = "dnswatcher"
	defaultCheckInterval = time.Minute
)

var (
	l             = logger.DefaultSLogger(packageName)
	runDNSWatcher sync.Once
	watcherList   []IDNSWatcher
	locker        sync.Mutex
)

type IDNSWatcher interface {
	GetDomain() string
	GetIPs() []string
	SetIPs([]string)
	Update() error
}

func StartWatch() error {
	runDNSWatcher.Do(func() {
		l = logger.SLogger(packageName)
		g := datakit.G("io_dnswatcher")

		// Uncomment this when you wanna get check interval from config file.
		// For now we write it permanently in the code.
		// du := getCheckInterval("1m")

		g.Go(func(ctx context.Context) error {
			return dnsWatcherMain(defaultCheckInterval)
		})
	})

	return nil
}

func AddWatcher(one IDNSWatcher) {
	locker.Lock()
	defer locker.Unlock()

	defer func() {
		dnsDomainCounter.Inc()
	}()

	watcherList = append(watcherList, one)
}

func dnsWatcherMain(chkInterval time.Duration) error {
	l.Info("start")

	tick := time.NewTicker(chkInterval)
	defer tick.Stop()

	for {
		l.Debugf("triggered, watcherList = %#v", watcherList)
		if err := doRun(); err != nil {
			l.Warnf("[%s] failed: %v, ignored", packageName, err)
		}

		watchRunCounter.WithLabelValues(chkInterval.String()).Inc()

		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil

		case <-tick.C:
		}
	}
}

func doRun() error {
	locker.Lock()
	defer locker.Unlock()

	for _, v := range watcherList {
		changed, newIPs := checkDNSChanged(v)

		if changed {
			if err := v.Update(); err != nil {
				l.Warnf("Update failed: err = %v, domain = %v", err, v.GetDomain())
				return err
			}

			v.SetIPs(newIPs)
		}
	}

	return nil
}

func checkDNSChanged(watcher IDNSWatcher) (bool, []string) {
	var (
		start   = time.Now()
		changed = false
		domain  = watcher.GetDomain()
		oldIPs  = watcher.GetIPs()
		err     error
	)

	defer func() {
		if changed {
			dnsUpdateCounter.WithLabelValues(domain).Inc()
		}

		if err == nil {
			watchLatency.WithLabelValues(domain, "ok").Observe(float64(time.Since(start)) / float64(time.Second))
		} else {
			watchLatency.WithLabelValues(domain, err.Error()).Observe(float64(time.Since(start)) / float64(time.Second))
		}
	}()

	netIPs, err := net.LookupIP(domain)
	if err != nil {
		l.Warnf("LookupIP failed: err = %v, domain = %s", err, domain)
		return false, nil
	}
	var newIPs []string
	for _, ip := range netIPs {
		newIPs = append(newIPs, ip.String())
	}

	changed = !compareutil.CompareListDisordered(oldIPs, newIPs)

	return changed, newIPs
}

func getCheckInterval(checkInterval string) time.Duration {
	du, err := time.ParseDuration(checkInterval)
	if err != nil {
		l.Warnf("parse dns interval failed: %v, default to %v", err, defaultCheckInterval)
		du = defaultCheckInterval
	}
	return du
}
