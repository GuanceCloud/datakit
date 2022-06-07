// Package dnswatcher contains dns watcher control logic
package dnswatcher

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/compareutil"
)

//------------------------------------------------------------------------------

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

//------------------------------------------------------------------------------

func StartWatch() error {
	runDNSWatcher.Do(func() {
		l = logger.SLogger(packageName)
		g := datakit.G(packageName)

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
	watcherList = append(watcherList, one)
}

//------------------------------------------------------------------------------

func dnsWatcherMain(chkInterval time.Duration) error {
	l.Info("start")

	tick := time.NewTicker(chkInterval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil

		case <-tick.C:
			l.Debugf("triggered, watcherList = %#v", watcherList)
			if err := doRun(); err != nil {
				tip := fmt.Sprintf("[%s] failed: %v", packageName, err)
				l.Error(tip)
			}
		} // select
	} // for
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
		} // if changed
	} // for

	return nil
}

func checkDNSChanged(watcher IDNSWatcher) (bool, []string) {
	domain := watcher.GetDomain()
	oldIPs := watcher.GetIPs()

	netIPs, err := net.LookupIP(domain)
	if err != nil {
		l.Warnf("LookupIP failed: err = %v, domain = %s", err, domain)
		return false, nil
	}
	var newIPs []string
	for _, ip := range netIPs {
		newIPs = append(newIPs, ip.String())
	}

	return !compareutil.CompareListDisordered(oldIPs, newIPs), newIPs
}

func getCheckInterval(checkInterval string) time.Duration {
	du, err := time.ParseDuration(checkInterval)
	if err != nil {
		l.Warnf("parse dns interval failed: %v, default to %v", err, defaultCheckInterval)
		du = defaultCheckInterval
	}
	return du
}
