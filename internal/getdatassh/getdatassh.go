// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package getdatassh for ssh get data.
package getdatassh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"golang.org/x/crypto/ssh"
)

const packageName = "getdatassh"

var (
	l = logger.DefaultSLogger(packageName)
	g = datakit.G(packageName)
)

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

// GetDataSSH use goroutine in goroutine, parallel get gpu-smi data through ssh.
// need not context timeout because "golang.org/x/crypto/ssh" have it.
func GetDataSSH(servers *SSHServers, timeout time.Duration) ([]*SSHData, error) {
	if len(servers.RemoteUsers) == 0 {
		l.Errorf("SSHServers RemoteUsers is null.")
		return nil, errors.New("RemoteUsers is null")
	}
	if len(servers.RemoteCommand) == 0 {
		l.Errorf("SSHServers RemoteCommand is null.")
		return nil, errors.New("SSHServers RemoteCommand is null")
	}
	if len(servers.RemotePasswords) == 0 && len(servers.RemoteRsaPaths) == 0 {
		l.Errorf("SSHServers RemotePasswords RemoteRsaPaths all be null.")
		return nil, errors.New("SSHServers RemotePasswords RemoteRsaPaths all be null")
	}

	// walk all remote servers
	sshData := make([]*SSHData, 0)
	ch := make(chan *SSHData, 1)
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
		func(index int, addr, command, username, password, rsa string, timeout time.Duration) {
			g.Go(func(ctx context.Context) error {
				defer wg.Done()
				// get data from ipmiServer（[]byte）
				data, err := getData(addr, command, username, password, rsa, timeout)
				if err != nil {
					l.Errorf("get SSH data : %s .", servers.RemoteAddrs[index], err)
					return err
				} else {
					ch <- &SSHData{
						Server: servers.RemoteAddrs[index],
						Data:   data,
					}
				}
				return nil
			})
		}(i, addr, command, username, password, rsa, timeout)
	}

	wg2 := sync.WaitGroup{}
	wg2.Add(1)
	g.Go(func(ctx context.Context) error {
		for v := range ch {
			sshData = append(sshData, v)
		}
		wg2.Done()
		return nil
	})

	wg.Wait()
	// all finish
	close(ch)
	wg2.Wait()

	return sshData, nil
}

// get data from ssh server, need not context timeout because "golang.org/x/crypto/ssh" have it.
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
		key, err := os.ReadFile(rsa)
		if err != nil {
			return nil, fmt.Errorf("unable to read rsa public key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse rsa public key: %w", err)
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

	client, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %s error %w", addr, err)
	}
	// nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh new session error %w", err)
	}
	// nolint:errcheck
	defer session.Close()

	// get data
	return session.Output(command)
}
