// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package usagetrace implements datakit's counting according to CPU cores.
package usagetrace

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"golang.org/x/exp/slices"
)

var (
	l = logger.DefaultSLogger("usage_trace")
	g = datakit.G("usage_trace")

	traceOptionCh = make(chan []UsageTraceOption, runtime.NumCPU())
)

type refresher interface {
	UsageTrace(body []byte) error
}

type nopRefresher struct{}

func (r *nopRefresher) UsageTrace([]byte) error {
	return nil
}

type usageTrace struct {
	Host      string `json:"hostname"`
	RuntimeID string `json:"runtime_id"`
	Token     string `json:"token"`
	IP        string `json:"ip"`
	PodName   string `json:"pod_name,omitempty"`
	RunMode   string `json:"run_mode"`

	DatakitVersion string `json:"datakit_version"`
	Arch           string `json:"arch"`
	OS             string `json:"os"`

	CPUCores   int `json:"cpu_cores"`
	CPULimites int `json:"cpu_limites,omitempty"`
	UsageCores int `json:"usage_cores"`

	startTime int64
	Uptime    int64 `json:"uptime"`

	RunInContainer bool `json:"run_in_container"`

	ServerListens []string `json:"server_listens,omitempty"`
	Inputs        []string `json:"inputs,omitempty"`

	DCAServer      string `json:"dca_server,omitempty"`
	UpgraderServer string `json:"upgrader_server,omitempty"`

	reservedInputs []string

	refreshInterval time.Duration

	refresher refresher
	exit      <-chan any
}

type UsageTraceOption func(*usageTrace)

func doClearServerListens() UsageTraceOption {
	return func(ut *usageTrace) {
		if len(ut.ServerListens) > 0 { // clear
			l.Infof("clear %d listen servers: %+#v", len(ut.ServerListens), ut.ServerListens)
			ut.ServerListens = ut.ServerListens[:0]
		}
	}
}

func doClearInputNames() UsageTraceOption {
	return func(ut *usageTrace) {
		if len(ut.Inputs) > 0 { // clear
			l.Infof("clear %d inputs: %+#v", len(ut.Inputs), ut.Inputs)
			ut.Inputs = ut.Inputs[:0]
		}
	}
}

func checkLoopbackServerListen(urlStr string) (bool, error) {
	host := urlStr

	// check if unix domain socket
	if _, err := os.Stat(urlStr); err == nil {
		l.Infof("local file(%q) as URL string", urlStr)
		return true, nil
	}

	if strings.Contains(urlStr, "://") { // this is a url like udp://localhost:1234
		if u, err := url.Parse(urlStr); err != nil {
			l.Debugf("url.Parse: %s, ignored", err.Error())
			host = urlStr
		} else {
			l.Debugf("url: %+#v", u)
			host = u.Host
		}
	}

	addrport, err := netip.ParseAddrPort(host)
	if err != nil {
		l.Warnf("invalid addr-port(%q): %s", host, err.Error())

		// For host like ':1234', the error is 'no IP', but this is a valid listen address.
		arr := strings.Split(host, ":")
		if len(arr) == 2 {
			switch arr[0] {
			case "":
				return false, nil

			default:
				if isDomainLoopback(arr[0]) {
					l.Infof("domain %q is loopback", arr[0])
					return true, nil
				} else {
					return false, err
				}
			}
		}

		// on any other errors, not a loopback address.
		return false, err
	}

	l.Debugf("addrport: %s, addr.Addr(): %s", addrport, addrport.Addr())
	return addrport.IsValid() && addrport.Addr().IsLoopback(), nil
}

func isDomainLoopback(domain string) bool {
	ips, err := net.LookupIP(domain)
	if err == nil {
		for _, ip := range ips {
			if !ip.IsLoopback() {
				return false
			}
		}
	} else {
		l.Warnf("Lookup (%q): %s", domain, err.Error())
		return false
	}

	return true
}

func WithUpgraderServer(s string) UsageTraceOption {
	return func(ut *usageTrace) {
		l.Infof("setup upgrader server to %q", s)
		ut.UpgraderServer = s
	}
}

func WithServerListens(listens ...string) UsageTraceOption {
	return func(ut *usageTrace) {
		for _, urlStr := range listens {
			if ok, err := checkLoopbackServerListen(urlStr); err != nil {
				l.Warnf("checkLoopbackServerListen: %s, ignored", err.Error())
				continue
			} else if !ok { // not loopback listen
				l.Infof("add server listen %q", urlStr)
				ut.ServerListens = append(ut.ServerListens, urlStr)
			}
		}
	}
}

func WithReservedInputs(ri ...string) UsageTraceOption {
	return func(ut *usageTrace) {
		l.Infof("add reserved inputs %+#v", ri)
		ut.reservedInputs = append(ut.reservedInputs, ri...)
	}
}

func WithInputNames(names ...string) UsageTraceOption {
	return func(ut *usageTrace) {
		for _, name := range names {
			if !slices.Contains(ut.reservedInputs, name) { // ignore non-reseved inputs
				return
			}

			if !slices.Contains(ut.Inputs, name) {
				l.Infof("add reserved input %q", name)
				ut.Inputs = append(ut.Inputs, name)
			}
		}
	}
}

func WithRefreshDuration(du time.Duration) UsageTraceOption {
	return func(ut *usageTrace) {
		l.Infof("setup refresh interval %s", du)
		ut.refreshInterval = du
	}
}

func WithCPULimits(limits float64) UsageTraceOption {
	return func(ut *usageTrace) {
		if limits < 1.0 {
			ut.CPULimites = 1
			l.Infof("set CPULimites from %f to %d", limits, ut.CPULimites)
		} else {
			ut.CPULimites = int(limits)
			l.Infof("set CPULimites to %d", ut.CPULimites)
		}
	}
}

func WithRefresher(r refresher) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.refresher = r
	}
}

func WithDatakitHostname(hostname string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.Host = hostname
		l.Infof("set host to %q", ut.Host)
	}
}

func WithDatakitRuntimeID(id string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.RuntimeID = id
		l.Infof("set runtimeID to %q", ut.RuntimeID)
	}
}

func WithDatakitPodname(podname string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.PodName = podname
		l.Infof("set pod name to %q", ut.PodName)
	}
}

func WithRunInContainer(on bool) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.RunInContainer = on
		l.Infof("is run in docker? %v", ut.RunInContainer)
	}
}

func WithWorkspaceToken(token string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.Token = token
	}
}

func WithMainIP(ip string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.IP = ip
		l.Infof("set main IP to %q", ut.IP)
	}
}

func WithDCAAPIServer(ipPort string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.DCAServer = ipPort
		l.Infof("set DCA server to %q", ut.DCAServer)
	}
}

func WithDatakitVersion(version string) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.DatakitVersion = version
		l.Infof("set Datakit version to %q", ut.DatakitVersion)
	}
}

func WithDatakitStartTime(ts int64) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.startTime = ts
	}
}

func WithExitChan(ch <-chan any) UsageTraceOption {
	return func(ut *usageTrace) {
		ut.exit = ch
	}
}

func (ut *usageTrace) refreshInfo() {
	ut.Uptime = int64(time.Since(time.Unix(ut.startTime, 0)) / time.Second)

	// go some reserved inputs in use, or enabled multiple servers
	if len(ut.Inputs) > 0 || len(ut.ServerListens) > 0 {
		if ut.CPULimites > 0 {
			ut.RunMode = "gateway" // It seems that the Datakit running as a gateway server.
			ut.UsageCores = ut.CPULimites
		}
	}
}

func (ut *usageTrace) doRefresh() error {
	ut.refreshInfo()

	j, err := json.Marshal(ut)
	if err != nil {
		return fmt.Errorf("refresh: %w", err)
	}

	l.Debugf("usage trace: %s", string(j))

	if err := ut.refresher.UsageTrace(j); err != nil {
		return fmt.Errorf("UsageTrace: %w", err)
	}

	return nil
}

// UpdateTraceOptions used to set dynamic infos on datakit usage tracing fields.
//
// Note: this may block the caller if usage tracer busy.
func UpdateTraceOptions(opts ...UsageTraceOption) {
	l.Infof("update usage tracer with %d options", len(opts))
	if len(opts) > 0 {
		traceOptionCh <- opts
	}
}

func ClearServerListens() {
	UpdateTraceOptions(doClearServerListens())
}

func ClearInputNames() {
	UpdateTraceOptions(doClearInputNames())
}

func (ut *usageTrace) loop() error {
	tick := time.NewTicker(ut.refreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if err := ut.doRefresh(); err != nil {
				l.Warnf("usage trace refresh: %s, ignored", err.Error())
			}
		case <-ut.exit:
			return nil
		case opts := <-traceOptionCh:
			ut.applyOptions(opts...)
		}
	}
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (ut *usageTrace) applyOptions(opts ...UsageTraceOption) {
	for _, opt := range opts {
		if opt != nil {
			l.Infof("apply option %s", getFunctionName(opt))
			opt(ut)
		} else {
			l.Warnf("ignore nil option")
		}
	}
}

func doStart(opts ...UsageTraceOption) error {
	ut := &usageTrace{
		CPUCores:        runtime.NumCPU(),
		refresher:       &nopRefresher{},
		UsageCores:      1,
		refreshInterval: time.Minute,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
	}

	ut.applyOptions(opts...)

	return ut.loop()
}

func Start(opts ...UsageTraceOption) {
	l = logger.SLogger("usage_trace")

	g.Go(func(_ context.Context) error {
		return doStart(opts...)
	})
}
