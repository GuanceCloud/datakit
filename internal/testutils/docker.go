// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutils

import (
	"fmt"
	"log"
	"net"
	"os"
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
		Host: "",
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
