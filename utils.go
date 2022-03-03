package datakit

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
)

func TrimSuffixAll(s, sfx string) string {
	var x string
	for {
		x = strings.TrimSuffix(s, sfx)
		if x == s {
			break
		}
		s = x
	}

	return x
}

func MonitProc(proc *os.Process, name string, stopCh *cliutils.Sem) error {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	if proc == nil {
		return fmt.Errorf("invalid proc %s", name)
	}

	for {
		select {
		case <-tick.C:
			p, err := os.FindProcess(proc.Pid)
			if err != nil {
				continue
			}

			switch runtime.GOOS {
			case OSWindows:

			default:
				if err := p.Signal(syscall.Signal(0)); err != nil {
					return err
				}
			}

		case <-Exit.Wait():
			return doKill(proc, name)

		case <-stopCh.Wait():
			return doKill(proc, name)
		}
	}
}

func doKill(proc *os.Process, name string) error {
	if err := proc.Kill(); err != nil { // XXX: should we wait here?
		return err
	}
	sts, err := proc.Wait()
	if err != nil {
		return err
	}
	l.Infof("proc wait, proc name: %ss exit code: %v", name, sts.ExitCode())
	return nil
}

func RndTicker(s string) (*time.Ticker, error) {
	du, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	if du <= 0 {
		return nil, fmt.Errorf("duration should larger than 0")
	}

	now := time.Now().UnixNano()
	rnd := now % int64(du)
	time.Sleep(time.Duration(rnd))
	return time.NewTicker(du), nil
}

func RawTicker(s string) (*time.Ticker, error) {
	du, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	if du <= 0 {
		return nil, fmt.Errorf("duration should larger than 0")
	}

	return time.NewTicker(du), nil
}

// SleepContext sleeps until the context is closed or the duration is reached.
func SleepContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	t := time.NewTimer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

// Duration just wraps time.Duration.
type Duration struct {
	Duration time.Duration
}

// UnmarshalTOML parses the duration from the TOML config file.
func (d *Duration) UnmarshalTOML(b []byte) error {
	b = bytes.Trim(b, "'")

	// see if we can directly convert it
	if du, err := time.ParseDuration(string(b)); err == nil {
		d.Duration = du
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err == nil {
			return nil
		}
	}

	// First try parsing as integer seconds
	if sI, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}
	// Second try parsing as float seconds
	if sF, err := strconv.ParseFloat(string(b), 64); err == nil {
		d.Duration = time.Second * time.Duration(sF)
	} else {
		return err
	}

	return nil
}

func (d *Duration) UnitString(unit time.Duration) string {
	ts := fmt.Sprintf("%d", d.Duration/unit)
	switch unit {
	case time.Second:
		return ts + "s"
	case time.Millisecond:
		return ts + "ms"
	case time.Microsecond:
		return ts + "mics"
	case time.Minute:
		return ts + "m"
	case time.Hour:
		return ts + "h"
	case time.Nanosecond:
		return ts + "ns"
	default:
		return ts + "unknow"
	}
}

// Size just wraps an int64.
type Size struct {
	Size int64
}

func (s *Size) UnmarshalTOML(b []byte) error {
	var err error
	b = bytes.Trim(b, `'`)

	val, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	s.Size = val
	return nil
}

func NumberFormat(str string) string {
	// 1,234.0
	arr := strings.Split(str, ".")
	if len(arr) == 0 {
		return str
	}
	part1 := arr[0]

	ps := strings.Split(part1, ",")
	if len(ps) == 0 {
		return str
	}

	n := strings.Join(ps, "")

	if len(arr) > 1 {
		n += "." + arr[1]
	}

	return n
}

func GZipStr(str string) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := io.WriteString(zw, str); err != nil {
		return nil, err
	}

	if err := zw.Flush(); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return z.Bytes(), nil
}

func GZip(data []byte) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}

	if err := zw.Flush(); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return z.Bytes(), nil
}

var dnsdests = []string{
	`114.114.114.114:80`,
	`8.8.8.8:80`,
}

func LocalIP() (string, error) {
	for _, dest := range dnsdests {
		conn, err := net.DialTimeout("udp", dest, time.Second)
		if err == nil {
			defer conn.Close() //nolint:errcheck
			localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
			if !ok {
				return "", fmt.Errorf("expect net.UDPAddr")
			}

			return localAddr.IP.String(), nil
		}
	}

	return GetFirstGlobalUnicastIP()
}

func GetFirstGlobalUnicastIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				// pass
			}

			switch {
			case ip.IsGlobalUnicast():
				return ip.String(), nil
			default:
				// pass
			}
		}
	}

	return "", fmt.Errorf("no IP found")
}

func TomlMarshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func CheckExcluded(item string, blacklist, whitelist []string) bool {
	for _, v := range blacklist {
		if v == item {
			return true
		}
	}

	if len(whitelist) > 0 {
		exclude := true
		for _, v := range whitelist {
			if v == item {
				exclude = false
				break
			}
		}
		return exclude
	}

	return false
}

func TimestampMsToTime(ms int64) time.Time {
	return time.Unix(0, ms*1000000)
}

func GetEnv(env string) string {
	if v, ok := os.LookupEnv(env); ok {
		if v != "" {
			return v
		}
	}
	return ""
}
