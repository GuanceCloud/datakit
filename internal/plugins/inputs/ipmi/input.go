// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ipmi collects host ipmi metrics.
package ipmi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// Run Start the process of timing acquisition.
func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	ipt.getBytesFromIPMIs() // Not test model

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	dataCache := make([]dataStruct, 0)
	warnLazyTimes := 3

	for {
		// Avoid false alarms when the input wins the election.
		if !ipt.pause {
			if warnLazyTimes > 0 {
				warnLazyTimes--
			} else {
				l.Infof("is leader, check server drop warning")
				ipt.handleServerDropWarn()
			}
		} else {
			warnLazyTimes = 3
		}

		select {
		case d := <-ipt.ipmiDataCh:
			dataCache = append(dataCache, d)
		case <-tick.C:
			if len(dataCache) > 0 || len(ipt.collectCache) > 0 {
				start := time.Now()

				if err := ipt.handleData(dataCache); err != nil {
					l.Errorf("Collect: %s", err)
					ipt.Feeder.FeedLastError(inputName, err.Error())
				}
				// If there is data in the cache, submit it.
				if len(ipt.collectCache) > 0 {
					err := ipt.Feeder.Feed(inputName, point.Metric, ipt.collectCache,
						&io.Option{CollectCost: time.Since(start)})
					if err != nil {
						l.Errorf("Feed: %V", err)
						ipt.Feeder.FeedLastError(inputName, err.Error())
					}
				}
				dataCache = make([]dataStruct, 0)
				ipt.collectCache = make([]*point.Point, 0)
			}
		case ipt.pause = <-ipt.pauseCh:
		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return
		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
		}
	}
}

func (ipt *Input) checkConfig() error {
	if ipt.BinPath == "" {
		return fmt.Errorf("BinPath is null")
	}

	if len(ipt.IpmiServers) == 0 {
		return fmt.Errorf("IpmiServers is null")
	}

	if len(ipt.HexKeys) == 0 && (len(ipt.IpmiUsers) == 0 || len(ipt.IpmiPasswords) == 0) {
		l.Errorf("No HexKeys && No IpmiUsers No IpmiPasswords")
		return fmt.Errorf("no HexKeys and No IpmiUsers No IpmiPasswords")
	}

	// Check if have same ip.
	arr := make([]string, len(ipt.IpmiServers))
	copy(arr, ipt.IpmiServers)
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] == arr[i+1] {
			return fmt.Errorf("IpmiServers have same ip: %v", arr[i])
		}
	}
	return nil
}

func (ipt *Input) getBytesFromIPMIs() {
	g := datakit.G("ipmi")
	// Traversal get bytes data from all ipmi servers.
	for i := 0; i < len(ipt.IpmiServers); i++ {
		// only add a recorder. (server, isActive, isOnlyAdd)
		ipt.serverOnlineInfo(ipt.IpmiServers[i], false, true)
		func(index int) {
			g.Go(func(ctx context.Context) error {
				ipt.getBytesFromIPMI(index)
				return nil
			})
		}(i)
	}
}

func (ipt *Input) getBytesFromIPMI(index int) {
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		if !ipt.pause {
			l.Infof("is leader, ipmi gathering...")
			data, err := ipt.getBytes(index)
			if err != nil {
				l.Errorf("getBytes error: %v", err)
			} else {
				ipt.ipmiDataCh <- *data
			}
		} else {
			l.Infof("not leader, skip gather")
		}

		select {
		case <-tick.C:
		case ipt.pause = <-ipt.pauseCh:
		case <-datakit.Exit.Wait():
			l.Infof("memory input exit")
			return
		case <-ipt.semStop.Wait():
			l.Infof("memory input return")
			return
		}
	}
}

func (ipt *Input) getBytes(index int) (*dataStruct, error) {
	if err := ipt.checkConfig(); err != nil {
		return nil, err
	}

	// Get exec parameters.
	opts, metricVersion, err := ipt.getParameters(index)
	if err != nil {
		l.Errorf("getParameters :%v", err)
		return nil, err
	}

	l.Info("before get bytes, server: ", ipt.IpmiServers[index])
	start := time.Now()

	//nolint:gosec
	c := exec.Command(ipt.BinPath, opts...)

	// exec.Command ENV
	if len(ipt.Envs) != 0 {
		// In windows here will broken old PATH.
		c.Env = ipt.Envs
	}

	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		l.Errorf("c.Start(): %v, %v", err, b.String())
		return nil, err
	}
	err = c.Wait()
	if err != nil {
		l.Errorf("c.Wait(): %s, %v, %v", ipt.IpmiServers[index], err, b.String())
		l.Info("env PATH: ", os.Getenv("PATH"))
		l.Info("env LD_LIBRARY_PATH: ", os.Getenv("LD_LIBRARY_PATH"))
		return nil, err
	}

	bytes := b.Bytes()
	l.Infof("get bytes len: %v. consuming: %v", len(bytes), time.Since(start))

	return &dataStruct{
		server:        ipt.IpmiServers[index],
		metricVersion: metricVersion,
		data:          bytes,
	}, err
}

// Get parameter that exec need.
func (ipt *Input) getParameters(i int) (opts []string, metricVersion int, err error) {
	var (
		ipmiInterface string
		ipmiUser      string
		ipmiPassword  string
		hexKey        string
	)

	// Add ipmiInterface ipmiServer parameter.
	if len(ipt.IpmiInterfaces) < i+1 {
		ipmiInterface = ipt.IpmiInterfaces[0]
	} else {
		ipmiInterface = ipt.IpmiInterfaces[i]
	}
	opts = []string{"-I", ipmiInterface, "-H", ipt.IpmiServers[i]}

	if len(ipt.HexKeys) > 0 {
		// Add hexkey parameter.
		if len(ipt.HexKeys) < i+1 {
			hexKey = ipt.HexKeys[0]
		} else {
			hexKey = ipt.HexKeys[i]
		}
		opts = append(opts, "-y", hexKey)
	}

	// Add ipmiUser ipmiPassword parameter.
	if len(ipt.IpmiUsers) == 0 || len(ipt.IpmiPasswords) == 0 {
		l.Error("have no hexKey && ipmiUser && ipmiPassword")
		err = errors.New("have no hexKey && ipmiUser && ipmiPassword")
		return
	}
	if len(ipt.IpmiUsers) < i+1 {
		ipmiUser = ipt.IpmiUsers[0]
	} else {
		ipmiUser = ipt.IpmiUsers[i]
	}
	if len(ipt.IpmiPasswords) < i+1 {
		ipmiPassword = ipt.IpmiPasswords[0]
	} else {
		ipmiPassword = ipt.IpmiPasswords[i]
	}
	opts = append(opts, "-U", ipmiUser, "-P", ipmiPassword)

	if len(ipt.MetricVersions) < i+1 {
		metricVersion = ipt.MetricVersions[0]
	} else {
		metricVersion = ipt.MetricVersions[i]
	}
	if metricVersion == 2 {
		opts = append(opts, "sdr", "elist")
	} else {
		opts = append(opts, "sdr")
	}

	return opts, metricVersion, err
}

// Terminate Stop.
func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// Catalog Catalog.
func (*Input) Catalog() string {
	return inputName
}

// SampleConfig : conf File samples, reflected in the document.
func (*Input) SampleConfig() string {
	return sampleCfg
}

// AvailableArchs : OS support, reflected in the document.
func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

// SampleMeasurement Sample measurement results, reflected in the document.
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&ipmiMeasurement{},
	}
}

// Add or update server info.
func (ipt *Input) serverOnlineInfo(server string, isActive bool, isOnlyAdd bool) {
	for i := 0; i < len(ipt.servers); i++ {
		if ipt.servers[i].server == server {
			if isActive && !isOnlyAdd {
				// Here is multi thread but different recorder，so need not mutex.
				ipt.servers[i].isWarned = false                        // Set never warning
				ipt.servers[i].activeTimestamp = time.Now().UnixNano() // Reset timestamp
			}
			return
		}
	}
	// Here is single thread，so need not mutex.
	// New, append, server not in ipt.servers.
	ipt.servers = append(ipt.servers, ipmiServer{server, time.Now().UnixNano(), false})
}

func (ipt *Input) handleServerDropWarn() {
	for i := 0; i < len(ipt.servers); i++ {
		if time.Now().UnixNano() > ipt.servers[i].activeTimestamp+ipt.DropWarningDelay.Duration.Nanoseconds() &&
			!ipt.servers[i].isWarned {
			// This server time stamp out of delay time && no alarm has been given.
			l.Warnf("before lost a server: %s", ipt.servers[i].server)

			// Send warning.
			tags := map[string]string{
				"host": ipt.servers[i].server,
			}
			fields := map[string]interface{}{
				"warning": 1,
			}
			ipt.collectCache = append(ipt.collectCache, ipt.newPoint(tags, fields))

			l.Warnf("lost a server: %s", ipt.servers[i].server)

			// Set warned state. Do not warn and warn.
			ipt.servers[i].isWarned = true
		}
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

// Collect get data convert to metrics. In this input no real useful.
func (ipt *Input) Collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	dataCache := make([]dataStruct, 0)

	for i := 0; i < len(ipt.IpmiServers); i++ {
		data, err := ipt.getBytes(i)
		if err != nil {
			l.Errorf("getBytes error: %v", err)
			return err
		} else {
			dataCache = append(dataCache, *data)
		}
	}

	if len(dataCache) > 0 {
		_ = ipt.handleData(dataCache)
	}

	return nil
}

func newDefaultInput() *Input {
	ipt := &Input{
		platform:       runtime.GOOS,
		IpmiInterfaces: []string{"lan"},
		Interval:       datakit.Duration{Duration: time.Second * 10},
		semStop:        cliutils.NewSem(),

		Tags:             make(map[string]string),
		Timeout:          datakit.Duration{Duration: time.Second * 5},
		MetricVersions:   []int{1},
		DropWarningDelay: datakit.Duration{Duration: time.Second * 300},
		servers:          make([]ipmiServer, 0, 1),
		pauseCh:          make(chan bool, inputs.ElectionPauseChannelLength),
		Election:         true,
		ipmiDataCh:       make(chan dataStruct, 1),
		collectCache:     make([]*point.Point, 0),
		Feeder:           io.DefaultFeeder(),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
