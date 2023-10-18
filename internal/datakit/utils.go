// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/klauspost/compress/zstd"
	pr "github.com/shirou/gopsutil/v3/process"
	"golang.org/x/crypto/ssh"

	"github.com/GuanceCloud/cliutils"
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

// UnmarshalText parses the duration from the TOML config file.
func (d *Duration) UnmarshalText(b []byte) error {
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
		return ts + "unknown"
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

func Zstdzip2(data []byte) ([]byte, error) {
	enc, _ := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(runtime.NumCPU()), zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	return enc.EncodeAll(data, make([]byte, 0, len(data))), nil
}

func Zstdzip(data []byte) ([]byte, error) {
	out := bytes.NewBuffer(nil)
	in := bytes.NewBuffer(data)

	enc, err := zstd.NewWriter(out, zstd.WithEncoderConcurrency(runtime.NumCPU()), zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(enc, in)
	if err != nil {
		enc.Close() //nolint: errcheck,gosec
		return nil, err
	}

	enc.Close() //nolint: errcheck,gosec
	return out.Bytes(), nil
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

func OpenFiles() int {
	pid := os.Getpid()
	p, err := pr.NewProcess(int32(pid))
	if err != nil {
		return -1
	}

	if fs, err := p.OpenFiles(); err != nil {
		return -1
	} else {
		return len(fs)
	}
}

// WaitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
func WaitTimeout(c *exec.Cmd, timeout time.Duration) error {
	var kill *time.Timer
	term := time.AfterFunc(timeout, func() {
		err := c.Process.Signal(syscall.SIGTERM)
		if err != nil {
			l.Infof("E! [agent] Error terminating process: %s", err)
			return
		}

		kill = time.AfterFunc(timeout+1, func() { // 这个地方 原本是定死的5秒,应该比exec.Command()的timeout长一点
			err := c.Process.Kill()
			if err != nil {
				l.Infof("E! [agent] Error killing process: %s", err)
				return
			}
		})
	})

	err := c.Wait()

	// Shutdown all timers
	if kill != nil {
		kill.Stop()
	}

	// If the process exited without error treat it as success.  This allows a
	// process to do a clean shutdown on signal.
	if err == nil {
		return nil
	}

	// If SIGTERM was sent then treat any process error as a timeout.
	if !term.Stop() {
		return errors.New("command timed out")
	}

	// Otherwise there was an error unrelated to termination.
	return err
}

// SSH remote get data
/*
    // use like this way

 	// use goroutine, send data through dataCh, with timeout, with tag
	dataCh := make(chan datakit.SSHData, 1)
	l = logger.SLogger("gpu_smi")
	g := datakit.G("gpu_smi")

	g.Go(func(ctx context.Context) error {
		return datakit.SSHGetData(dataCh, &ipt.SSHServers, ipt.Timeout.Duration)
	})

	// to receive data through dataCh
	ipt.handleDatas(dataCh)
*/

type SSHData struct {
	Server string
	Data   []byte
}
type SSHServers struct {
	RemoteAddrs     []string `toml:"remote_addrs"`     // remote server addr:port list
	RemoteUsers     []string `toml:"remote_users"`     // remote server username list
	RemotePasswords []string `toml:"remote_passwords"` // remote server password list
	RemoteRsaPaths  []string `toml:"remote_rsa_paths"` // rsa path for remote server list
	RemoteCommand   string   `toml:"remote_command"`   // remote command
}

// SSHGetData use goroutine in goroutine, parallel get gpu-smi data through ssh.
// need not context timeout because "golang.org/x/crypto/ssh" have.
func SSHGetData(ch chan SSHData, servers *SSHServers, timeout time.Duration) error {
	if len(servers.RemoteUsers) == 0 {
		l.Errorf("SSHServers RemoteUsers is null.")
		return errors.New("RemoteUsers is null")
	}
	if len(servers.RemoteCommand) == 0 {
		l.Errorf("SSHServers RemoteCommand is null.")
		return errors.New("SSHServers RemoteCommand is null")
	}
	if len(servers.RemotePasswords) == 0 && len(servers.RemoteRsaPaths) == 0 {
		l.Errorf("SSHServers RemotePasswords RemoteRsaPaths all be null.")
		return errors.New("SSHServers RemotePasswords RemoteRsaPaths all be null")
	}

	// walk all remote servers
	var wg sync.WaitGroup

	for i := 0; i < len(servers.RemoteAddrs); i++ {
		var (
			addr     string
			command  string
			username string
			password string
			rsa      string
		)
		addr = servers.RemoteAddrs[i]
		command = servers.RemoteCommand

		if i >= len(servers.RemoteUsers) {
			// RemoteUsers short than RemoteAddrs
			username = servers.RemoteUsers[0]
		} else {
			username = servers.RemoteUsers[i]
		}

		if len(servers.RemoteRsaPaths) > 0 {
			// rsa first than password
			// use rsa public key
			if i >= len(servers.RemoteRsaPaths) {
				// RemoteRsaPaths short than RemoteAddrs
				rsa = servers.RemoteRsaPaths[0]
			} else {
				rsa = servers.RemoteRsaPaths[i]
			}
		} else {
			// use password
			if i >= len(servers.RemotePasswords) {
				// RemotePasswords short than RemoteAddrs
				password = servers.RemotePasswords[0]
			} else {
				password = servers.RemotePasswords[i]
			}
		}

		// check addr:port
		_, err := netip.ParseAddrPort(servers.RemoteAddrs[i])
		if err != nil {
			l.Errorf("SSHServers ParseAddrPort : ", servers.RemoteAddrs[i])
			continue
		}

		// walk and do get data
		wg.Add(1)
		go func(ch chan SSHData, index int, addr, command, username, password, rsa string, timeout time.Duration) {
			defer wg.Done()

			// get data from ipmiServer（[]byte）
			data, err := getData(addr, command, username, password, rsa, timeout)
			if err != nil {
				l.Errorf("get SSH data : %s .", servers.RemoteAddrs[index], err)
			} else {
				ch <- SSHData{
					Server: servers.RemoteAddrs[index],
					Data:   data,
				}
			}
		}(ch, i, addr, command, username, password, rsa, timeout)
	}

	wg.Wait()
	// all finish
	close(ch)

	return nil
}

// get data from shh server, need not context timeout because "golang.org/x/crypto/ssh" have.
func getData(addr, command, username, password, rsa string, timeout time.Duration) ([]byte, error) {
	var config ssh.ClientConfig

	if rsa == "" {
		// use password
		config = ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: timeout,
		}
	} else {
		// use rsa public key
		// nolint:gosec
		key, err := ioutil.ReadFile(rsa)
		if err != nil {
			err = fmt.Errorf("unable to read rsa public key: %w", err)
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			err = fmt.Errorf("unable to parse rsa public key: %w", err)
			return nil, err
		}
		config = ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			Timeout: timeout,
		}
	}

	// creat client
	client, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %s error %w", addr, err)
	}
	// nolint:errcheck
	defer client.Close()

	// creat session
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh new session error %w", err)
	}
	// nolint:errcheck
	defer session.Close()

	// get data
	return session.Output(command)
}

// RebuildFolder rebuild folder. if exists, remove and rebuild.
func RebuildFolder(path string, perm fs.FileMode) error {
	isExists, _, err := IsPathExists(path)
	if err != nil {
		return err
	}

	if isExists {
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("remove %v: %w", path, err)
		}
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("create %s failed: %w", path, err)
	}

	return nil
}

// IsPathExists Check if the given path exists and if is folder.
func IsPathExists(path string) (isExists, isDir bool, err error) {
	s, err := os.Stat(path)
	if err == nil {
		return true, s.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, false, nil
	}

	return false, false, fmt.Errorf("cheack %v: %w", path, err)
}

// SaveStringToFile Save string to file. If exists, remove and rebuild.
func SaveStringToFile(path string, value string) error {
	// Create a file
	// #nosec
	f, err := os.Create(path)
	if err != nil {
		l.Errorf("os.Create(%v): %v", path, err)
		return err
	}
	// nolint:errcheck,gosec
	defer f.Close()

	n, err := io.WriteString(f, value)
	if err != nil {
		l.Errorf("os.WriteString(%v): %v", path, err)
		return err
	}
	l.Info("os.WriteString(%v) success: %d bytes", path, n)

	return nil
}
