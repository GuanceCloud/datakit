// +build linux

package systemd

import (
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "systemd"

	defaultMeasurement = "systemd"

	sampleCfg = `
[inputs.systemd]
	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
	# required
	interval = "10s"

	# [inputs.systemd.tags]
	# tags1 = "value1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Systemd{}
	})
}

type trie struct {
}

type Systemd struct {
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	conn     *dbus.Conn
	duration time.Duration
}

func (_ *Systemd) SampleConfig() string {
	return sampleCfg
}

func (_ *Systemd) Catalog() string {
	return inputName
}

func (s *Systemd) Run() {
	l = logger.SLogger(inputName)

	if s.loadcfg() {
		return
	}
	ticker := time.NewTicker(s.duration)
	defer ticker.Stop()

	l.Infof("systemd input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			s.conn.Close()
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := s.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (s *Systemd) loadcfg() bool {
	var err error

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		s.duration, err = time.ParseDuration(s.Interval)
		if err != nil || s.duration <= 0 {
			l.Errorf("invalid interval, err %s", err.Error())
			time.Sleep(time.Second)
			continue
		}

		s.conn, err = dbus.New()
		if err != nil {
			l.Errorf("connect systemd err: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return false
}

func (s *Systemd) getMetrics() ([]byte, error) {
	var loaded, active, service, socket, device, mount, automount,
		swap, target, path, timer, slice, scope int

	units, err := s.conn.ListUnits()
	if err != nil {
		return nil, err
	}

	for _, unitStatus := range units {
		nameBlocks := strings.Split(unitStatus.Name, ".")
		switch nameBlocks[len(nameBlocks)-1] {
		case "service":
			service++
		case "socket":
			socket++
		case "device":
			device++
		case "mount":
			mount++
		case "automount":
			automount++
		case "swap":
			swap++
		case "target":
			target++
		case "path":
			path++
		case "timer":
			timer++
		case "slice":
			slice++
		case "scope":
			scope++
		}

		if unitStatus.LoadState == "loaded" {
			loaded++
		}
		if unitStatus.ActiveState == "active" {
			active++
		}
	}
	fields := map[string]interface{}{
		"units_total":        len(units),
		"units_loaded_count": loaded,
		"units_active_count": active,
		"unit_service":       service,
		"unit_socket":        socket,
		"unit_device":        device,
		"unit_mount":         mount,
		"unit_automount":     automount,
		"unit_swap":          swap,
		"unit_target":        target,
		"unit_path":          path,
		"unit_timer":         timer,
		"unit_slice":         slice,
		"unit_scope":         scope,
	}
	return io.MakeMetric(defaultMeasurement, s.Tags, fields, time.Now())
}
