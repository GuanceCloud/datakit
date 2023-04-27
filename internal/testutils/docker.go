// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

type RemoteInfo struct {
	// docker info
	Port string
	Host string
}

// RemoteAPIOK test if remote HTTP API ok.
func (i *RemoteInfo) RemoteAPIOK(port int,
	url string,
	args ...time.Duration,
) bool {
	return false // TODO
}

// PortOK test if remote container's port ok every second.
func (i *RemoteInfo) PortOK(port string, args ...time.Duration) bool {
	var (
		con net.Conn
		err error
	)

	addr := fmt.Sprintf("%s:%s", i.Host, port)

	if len(args) > 0 {
		iter := time.NewTicker(time.Second)
		defer iter.Stop()

		timeout := time.NewTicker(args[0])
		defer timeout.Stop()

		for {
			select {
			case <-timeout.C:
				return false

			case <-iter.C:
				log.Printf("check port %s...", addr)
				con, err = net.DialTimeout("tcp", addr, time.Second)
				if err == nil {
					goto end
				} else {
					log.Printf("check port: %s", err)
				}
			}
		}
	} else {
		for { // wait until ok
			log.Printf("check port %s...", addr)
			con, err = net.DialTimeout("tcp", addr, time.Second)
			if err == nil {
				goto end
			} else {
				log.Printf("check port: %s", err)
			}
			time.Sleep(time.Second)
		}
	}

end:
	defer con.Close() //nolint:errcheck
	return true
}

// TCPURL get TCP URL format.
func (i *RemoteInfo) TCPURL() string {
	return "tcp://" + net.JoinHostPort(i.Host, i.Port)
}

// GetRemote only return the IP of remote node.
func GetRemote() *RemoteInfo {
	ri := &RemoteInfo{
		Host: "0.0.0.0",
		Port: "2375",
	}

	if v := os.Getenv("REMOTE_HOST"); v != "" {
		ri.Host = v
	}

	if v := os.Getenv("DOCKER_PORT"); v != "" {
		ri.Port = v
	}

	return ri
}

var (
	maxPort    = 65535
	baseOffset = 10000
)

// RandPort return random port after offset baseOffset.
func RandPort(proto string) int {
	if v := os.Getenv("TESTING_BASE_PORT"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			baseOffset = int(i)
		}
	}

	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
		p := ((r.Int() % baseOffset) + baseOffset) % maxPort
		if !portInUse(proto, p) {
			return p
		}
	}
}

func portInUse(proto string, p int) bool {
	c, err := net.DialTimeout(proto, net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", p)), time.Second)
	if err != nil {
		return false
	}

	if c != nil {
		defer c.Close() //nolint:errcheck
	}

	return true
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
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
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
