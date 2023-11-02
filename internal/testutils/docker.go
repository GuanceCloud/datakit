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
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
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

// RandPortUDP return random UDP port after offset baseOffset.
func RandPortUDP() (*net.UDPConn, int, error) {
	if v := os.Getenv("TESTING_BASE_PORT"); v != "" {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			baseOffset = int(i)
		}
	}

	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
		p := ((r.Int() % baseOffset) + baseOffset) % maxPort
		if conn, err := udpPortInUse(p); err == nil {
			return conn, p, nil
		}
	}
}

func udpPortInUse(port int) (*net.UDPConn, error) {
	s, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", s)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// ExternalIP returns running host's external IP address.
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

// RetryTestRun retries function under specificed conditions.
//
//nolint:lll
func RetryTestRun(f func() error) error {
	retryCount := 0
	var errMsgs []string

	for {
		retryCount++
		if retryCount > 3 {
			return fmt.Errorf("exceeded retry count: %v", errMsgs)
		}

		if err := f(); err != nil {
			fmt.Printf("RetryTestRun, err = %v\n", err)
			switch {
			case strings.Contains(err.Error(), "already"):
				// API error (500): driver failed programming external connectivity on endpoint memcached (7bdcaf6b4a5dba4fa54c118e455a9f0220f9d3514e682f0dfdb92fddebc6823f): Error starting userland proxy: listen tcp4 0.0.0.0:10828: bind: address already in use
				// API error (500): driver failed programming external connectivity on endpoint java (7a26eeed3d3eefb86e7f043661f55e19f80ed5ed60a8d27f4663dc0ff87b404f): Bind for 0.0.0.0:8080 failed: port is already allocated
				fallthrough
			case strings.Contains(err.Error(), "timeout"):
				errMsgs = append(errMsgs, err.Error()) // not return, retry.
			default:
				return err // other conditions return error immediately.
			}
		} else {
			return nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

const envExcludeIntegrationTesting = "UT_EXCLUDE_INTEGRATION_TESTING"

// CheckIntegrationTestingRunning returns whether performing the integration testing.
// Default is running.
func CheckIntegrationTestingRunning() bool {
	if val := os.Getenv(envExcludeIntegrationTesting); len(val) > 0 {
		lower := strings.ToLower(val)
		if lower == "on" {
			return false
		}
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////

const testingPrefix = "testing-"

// GetUniqueContainerName returns unique combined container name.
func GetUniqueContainerName(name string) string {
	rawName := dockertest.GetRawName(name)

	nanoStr := fmt.Sprintf("-%d", time.Now().Nanosecond())
	randStr := cliutils.CreateRandomString(5)

	return testingPrefix + rawName + nanoStr + randStr
}

// GetTestingPrefix returns testing prefix name.
func GetTestingPrefix(name string) string {
	return testingPrefix + name
}

////////////////////////////////////////////////////////////////////////////////

// PurgeRemoteByName purges containers with specified testing names.
func PurgeRemoteByName(name string) error {
	r := GetRemote()
	dockerTCP := r.TCPURL()

	log.Printf("get remote: %+#v, TCP: %s", r, dockerTCP)

	p, err := GetPool(dockerTCP)
	if err != nil {
		if r.Host != "0.0.0.0" {
			return err
		} else if p, err = GetPool(""); err != nil {
			return err
		}
	}

	containerName := GetTestingPrefix(name)
	containers, err := p.Client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"name": {containerName},
		},
	})
	if err != nil {
		return fmt.Errorf("error while listing containers with name %s", containerName)
	}

	if len(containers) == 0 {
		return nil
	}

	var errs []error
	for k := range containers {
		err = p.Client.RemoveContainer(docker.RemoveContainerOptions{
			ID:            containers[k].ID,
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		for _, v := range errs {
			log.Printf("RemoveContainer failed: %v", v)
		}
		return fmt.Errorf("error while removing container with name %s: %s", containerName, errs[0].Error()) //nolint:errorlint
	}

	return nil
}

// GetPool returns dockertest Pool connection.
func GetPool(endpoint string) (*dockertest.Pool, error) {
	p, err := dockertest.NewPool(endpoint)
	if err != nil {
		return nil, err
	}
	err = p.Client.Ping()
	if err != nil {
		log.Printf("Could not connect to Docker: %v", err)
		return nil, err
	}
	return p, nil
}

////////////////////////////////////////////////////////////////////////////////

//nolint:stylecheck
const (
	SERVICE_NAME     = "service_name"
	PSPAN_ANNOTATION = "pspan_annotation"
	RUNTIME_ID       = "runtime_id"
)

////////////////////////////////////////////////////////////////////////////////

const (
	containerName   = "oraemon-packet-container"
	volumeName      = "oraemon-packet"
	volumeMountPath = "/tmp/myapp"
)

// RunOraemon runs datakit-oraemon image which contains necessary files.
// ATTENTION: Inputs' integration testings used RunOraemon should NOT be paralleled with each other!
//
//	But they can paralleled inside themselves.
func RunOraemon(endpoint string) (p *dockertest.Pool, resource *dockertest.Resource, mounts string, err error) {
	p, err = GetPool(endpoint)
	if err != nil {
		return nil, nil, "", err
	}

	err = p.Client.RemoveVolumeWithOptions(docker.RemoveVolumeOptions{
		Name: volumeName,
	})
	if err != nil {
		switch err.Error() {
		case "no such volume", "volume in use and cannot be removed":
			// Ignore.
		default:
			return nil, nil, "", err
		}
	}

	_, err = p.Client.CreateVolume(docker.CreateVolumeOptions{
		Name: volumeName,
	})
	if err != nil {
		return nil, nil, "", err
	}

	mounts = volumeName + ":" + volumeMountPath

	resource, err = p.RunWithOptions(
		&dockertest.RunOptions{
			Name:       containerName,
			Repository: "pubrepo.guance.com/image-repo-for-testing/datakit-oraemon",
			Tag:        "v1",
			Mounts:     []string{mounts},
		},

		func(c *docker.HostConfig) {
			c.RestartPolicy = docker.RestartPolicy{Name: "no"}
			c.AutoRemove = true
		},
	)
	if err != nil {
		return nil, nil, "", err
	}

	return p, resource, mounts, nil
}

func RemoveOraemon(p *dockertest.Pool, res *dockertest.Resource) error {
	if err := p.Purge(res); err != nil {
		log.Println("p.Purge failed:", err.Error())
	}
	return p.Client.RemoveVolumeWithOptions(docker.RemoveVolumeOptions{
		Name: volumeName,
	})
}

////////////////////////////////////////////////////////////////////////////////
