package dockerlog

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	docker "github.com/docker/docker/client"
	"github.com/gobwas/glob"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/docker"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  # endpoint = "unix:///var/run/docker.sock"

  ## When true, container logs are read from the beginning; otherwise
  ## reading begins at the end of the log.
  # from_beginning = false

  ## Timeout for Docker API calls.
  # timeout = "5s"

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  # container_name_include = []
  # container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  # docker_label_include = []
  # docker_label_exclude = []

  ## Set the source tag for the metrics to the container ID hostname, eg first 12 chars
  source_tag = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

const (
	defaultEndpoint = "unix:///var/run/docker.sock"

	// Maximum bytes of a log line before it will be split, size is mirroring
	// docker code:
	// https://github.com/moby/moby/blob/master/daemon/logger/copier.go#L21
	maxLineBytes = 16 * 1024
)

var (
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
	// ensure *DockerLogs implements telegraf.ServiceInput
	_ telegraf.ServiceInput = (*DockerLogs)(nil)
)

type DockerLogs struct {
	Endpoint              string            `toml:"endpoint"`
	FromBeginning         bool              `toml:"from_beginning"`
	Timeout               internal.Duration `toml:"timeout"`
	LabelInclude          []string          `toml:"docker_label_include"`
	LabelExclude          []string          `toml:"docker_label_exclude"`
	ContainerInclude      []string          `toml:"container_name_include"`
	ContainerExclude      []string          `toml:"container_name_exclude"`
	ContainerStateInclude []string          `toml:"container_state_include"`
	ContainerStateExclude []string          `toml:"container_state_exclude"`
	IncludeSourceTag      bool              `toml:"source_tag"`

	tlsint.ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
	opts            types.ContainerListOptions
	wg              sync.WaitGroup
	mu              sync.Mutex
	containerList   map[string]context.CancelFunc
}

func (d *DockerLogs) Description() string {
	return "Read logging output from the Docker engine"
}

func (d *DockerLogs) SampleConfig() string {
	return sampleConfig
}

func (d *DockerLogs) Init() error {
	var err error
	if d.Endpoint == "ENV" {
		d.client, err = d.newEnvClient()
		if err != nil {
			return err
		}
	} else {
		tlsConfig, err := d.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		d.client, err = d.newClient(d.Endpoint, tlsConfig)
		if err != nil {
			return err
		}
	}

	// Create filters
	err = d.createLabelFilters()
	if err != nil {
		return err
	}
	err = d.createContainerFilters()
	if err != nil {
		return err
	}
	err = d.createContainerStateFilters()
	if err != nil {
		return err
	}

	filterArgs := filters.NewArgs()
	for _, state := range containerStates {
		if d.stateFilter.Match(state) {
			filterArgs.Add("status", state)
		}
	}

	if filterArgs.Len() != 0 {
		d.opts = types.ContainerListOptions{
			Filters: filterArgs,
		}
	}

	return nil
}

func (d *DockerLogs) addToContainerList(containerID string, cancel context.CancelFunc) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.containerList[containerID] = cancel
	return nil
}

func (d *DockerLogs) removeFromContainerList(containerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.containerList, containerID)
	return nil
}

func (d *DockerLogs) containerInContainerList(containerID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.containerList[containerID]
	return ok
}

func (d *DockerLogs) cancelTails() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, cancel := range d.containerList {
		cancel()
	}
	return nil
}

func (d *DockerLogs) matchedContainerName(names []string) string {
	// Check if all container names are filtered; in practice I believe
	// this array is always of length 1.
	for _, name := range names {
		trimmedName := strings.TrimPrefix(name, "/")
		match := d.containerFilter.Match(trimmedName)
		if match {
			return trimmedName
		}
	}
	return ""
}

func (d *DockerLogs) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()
	acc.SetPrecision(time.Nanosecond)

	ctx, cancel := context.WithTimeout(ctx, d.Timeout.Duration)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, d.opts)
	if err != nil {
		return err
	}

	for _, container := range containers {
		if d.containerInContainerList(container.ID) {
			continue
		}

		containerName := d.matchedContainerName(container.Names)
		if containerName == "" {
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		d.addToContainerList(container.ID, cancel)

		// Start a new goroutine for every new container that has logs to collect
		d.wg.Add(1)
		go func(container types.Container) {
			defer d.wg.Done()
			defer d.removeFromContainerList(container.ID)

			err = d.tailContainerLogs(ctx, acc, container, containerName)
			if err != nil && err != context.Canceled {
				acc.AddError(err)
			}
		}(container)
	}
	return nil
}

func (d *DockerLogs) hasTTY(ctx context.Context, container types.Container) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout.Duration)
	defer cancel()
	c, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}
	return c.Config.Tty, nil
}

func (d *DockerLogs) tailContainerLogs(
	ctx context.Context,
	acc telegraf.Accumulator,
	container types.Container,
	containerName string,
) error {
	imageName, imageVersion := docker.ParseImage(container.Image)
	tags := map[string]string{
		"container_name":    containerName,
		"container_image":   imageName,
		"container_version": imageVersion,
	}

	if d.IncludeSourceTag {
		tags["source"] = hostnameFromID(container.ID)
	}

	// Add matching container labels as tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	hasTTY, err := d.hasTTY(ctx, container)
	if err != nil {
		return err
	}

	tail := "0"
	if d.FromBeginning {
		tail = "all"
	}

	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Details:    false,
		Follow:     true,
		Tail:       tail,
	}

	logReader, err := d.client.ContainerLogs(ctx, container.ID, logOptions)
	if err != nil {
		return err
	}

	// If the container is using a TTY, there is only a single stream
	// (stdout), and data is copied directly from the container output stream,
	// no extra multiplexing or headers.
	//
	// If the container is *not* using a TTY, streams for stdout and stderr are
	// multiplexed.
	if hasTTY {
		return tailStream(acc, tags, container.ID, logReader, "tty")
	} else {
		return tailMultiplexed(acc, tags, container.ID, logReader)
	}
}

func parseLine(line []byte) (time.Time, string, error) {
	parts := bytes.SplitN(line, []byte(" "), 2)

	switch len(parts) {
	case 1:
		parts = append(parts, []byte(""))
	}

	tsString := string(parts[0])

	// Keep any leading space, but remove whitespace from end of line.
	// This preserves space in, for example, stacktraces, while removing
	// annoying end of line characters and is similar to how other logging
	// plugins such as syslog behave.
	message := bytes.TrimRightFunc(parts[1], unicode.IsSpace)

	ts, err := time.Parse(time.RFC3339Nano, tsString)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("error parsing timestamp %q: %v", tsString, err)
	}

	return ts, string(message), nil
}

func tailStream(
	acc telegraf.Accumulator,
	baseTags map[string]string,
	containerID string,
	reader io.ReadCloser,
	stream string,
) error {
	defer reader.Close()

	tags := make(map[string]string, len(baseTags)+1)
	for k, v := range baseTags {
		tags[k] = v
	}
	tags["stream"] = stream

	r := bufio.NewReaderSize(reader, 64*1024)

	for {
		line, err := r.ReadBytes('\n')

		if len(line) != 0 {
			ts, message, err := parseLine(line)
			if err != nil {
				acc.AddError(err)
			} else {
				acc.AddFields("docker_log", map[string]interface{}{
					"container_id": containerID,
					"message":      message,
				}, tags, ts)
			}
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func tailMultiplexed(
	acc telegraf.Accumulator,
	tags map[string]string,
	containerID string,
	src io.ReadCloser,
) error {
	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, outReader, "stdout")
		if err != nil {
			acc.AddError(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := tailStream(acc, tags, containerID, errReader, "stderr")
		if err != nil {
			acc.AddError(err)
		}
	}()

	_, err := stdcopy.StdCopy(outWriter, errWriter, src)
	outWriter.Close()
	errWriter.Close()
	src.Close()
	wg.Wait()
	return err
}

// Start is a noop which is required for a *DockerLogs to implement
// the telegraf.ServiceInput interface
func (d *DockerLogs) Start(telegraf.Accumulator) error {
	return nil
}

func (d *DockerLogs) Stop() {
	d.cancelTails()
	d.wg.Wait()
}

// Following few functions have been inherited from telegraf docker input plugin
func (d *DockerLogs) createContainerFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}
	d.containerFilter = filter
	return nil
}

func (d *DockerLogs) createLabelFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return err
	}
	d.labelFilter = filter
	return nil
}

func (d *DockerLogs) createContainerStateFilters() error {
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	filter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return err
	}
	d.stateFilter = filter
	return nil
}

func init() {
	inputs.Add("docker_log", func() telegraf.Input {
		return &DockerLogs{
			Timeout:       internal.Duration{Duration: time.Second * 5},
			Endpoint:      defaultEndpoint,
			newEnvClient:  NewEnvClient,
			newClient:     NewClient,
			containerList: make(map[string]context.CancelFunc),
		}
	})
}

func hostnameFromID(id string) string {
	if len(id) > 12 {
		return id[0:12]
	}
	return id
}

// ======================= filter.go ===============
type Filter interface {
	Match(string) bool
}

// Compile takes a list of string filters and returns a Filter interface
// for matching a given string against the filter list. The filter list
// supports glob matching too, ie:
//
//   f, _ := Compile([]string{"cpu", "mem", "net*"})
//   f.Match("cpu")     // true
//   f.Match("network") // true
//   f.Match("memory")  // false
//
func Compile(filters []string) (Filter, error) {
	// return if there is nothing to compile
	if len(filters) == 0 {
		return nil, nil
	}

	// check if we can compile a non-glob filter
	noGlob := true
	for _, filter := range filters {
		if hasMeta(filter) {
			noGlob = false
			break
		}
	}

	switch {
	case noGlob:
		// return non-globbing filter if not needed.
		return compileFilterNoGlob(filters), nil
	case len(filters) == 1:
		return glob.Compile(filters[0])
	default:
		return glob.Compile("{" + strings.Join(filters, ",") + "}")
	}
}

// hasMeta reports whether path contains any magic glob characters.
func hasMeta(s string) bool {
	return strings.IndexAny(s, "*?[") >= 0
}

type filter struct {
	m map[string]struct{}
}

func (f *filter) Match(s string) bool {
	_, ok := f.m[s]
	return ok
}

type filtersingle struct {
	s string
}

func (f *filtersingle) Match(s string) bool {
	return f.s == s
}

func compileFilterNoGlob(filters []string) Filter {
	if len(filters) == 1 {
		return &filtersingle{s: filters[0]}
	}
	out := filter{m: make(map[string]struct{})}
	for _, filter := range filters {
		out.m[filter] = struct{}{}
	}
	return &out
}

type IncludeExcludeFilter struct {
	include Filter
	exclude Filter
}

func NewIncludeExcludeFilter(
	include []string,
	exclude []string,
) (Filter, error) {
	in, err := Compile(include)
	if err != nil {
		return nil, err
	}

	ex, err := Compile(exclude)
	if err != nil {
		return nil, err
	}

	return &IncludeExcludeFilter{in, ex}, nil
}

func (f *IncludeExcludeFilter) Match(s string) bool {
	if f.include != nil {
		if !f.include.Match(s) {
			return false
		}
	}

	if f.exclude != nil {
		if f.exclude.Match(s) {
			return false
		}
	}
	return true
}

// ================================= client.go ========

/*This file is inherited from telegraf docker input plugin*/
var (
	version        = "1.24"
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func NewEnvClient() (Client, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

func NewClient(host string, tlsConfig *tls.Config) (Client, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}
	client, err := docker.NewClientWithOpts(
		docker.WithHTTPHeaders(defaultHeaders),
		docker.WithHTTPClient(httpClient),
		docker.WithVersion(version),
		docker.WithHost(host))

	if err != nil {
		return nil, err
	}
	return &SocketClient{client}, nil
}

type SocketClient struct {
	client *docker.Client
}

func (c *SocketClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return c.client.ContainerList(ctx, options)
}

func (c *SocketClient) ContainerLogs(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return c.client.ContainerLogs(ctx, containerID, options)
}
func (c *SocketClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.client.ContainerInspect(ctx, containerID)
}

// ============================docker.go==========================
// Adapts some of the logic from the actual Docker library's image parsing
// routines:
// https://github.com/docker/distribution/blob/release/2.7/reference/normalize.go
func ParseImage(image string) (string, string) {
	domain := ""
	remainder := ""

	i := strings.IndexRune(image, '/')

	if i == -1 || (!strings.ContainsAny(image[:i], ".:") && image[:i] != "localhost") {
		remainder = image
	} else {
		domain, remainder = image[:i], image[i+1:]
	}

	imageName := ""
	imageVersion := "unknown"

	i = strings.LastIndex(remainder, ":")
	if i > -1 {
		imageVersion = remainder[i+1:]
		imageName = remainder[:i]
	} else {
		imageName = remainder
	}

	if domain != "" {
		imageName = domain + "/" + imageName
	}

	return imageName, imageVersion
}

// ============================ tls/config.go ========================

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	// Deprecated in 1.7; use TLS variables above
	SSLCA   string `toml:"ssl_ca"`
	SSLCert string `toml:"ssl_cert"`
	SSLKey  string `toml:"ssl_key"`
}

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TLSCert           string   `toml:"tls_cert"`
	TLSKey            string   `toml:"tls_key"`
	TLSAllowedCACerts []string `toml:"tls_allowed_cacerts"`
	TLSCipherSuites   []string `toml:"tls_cipher_suites"`
	TLSMinVersion     string   `toml:"tls_min_version"`
	TLSMaxVersion     string   `toml:"tls_max_version"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ClientConfig) TLSConfig() (*tls.Config, error) {
	// Support deprecated variable names
	if c.TLSCA == "" && c.SSLCA != "" {
		c.TLSCA = c.SSLCA
	}
	if c.TLSCert == "" && c.SSLCert != "" {
		c.TLSCert = c.SSLCert
	}
	if c.TLSKey == "" && c.SSLKey != "" {
		c.TLSKey = c.SSLKey
	}

	// TODO: return default tls.Config; plugins should not call if they don't
	// want TLS, this will require using another option to determine.  In the
	// case of an HTTP plugin, you could use `https`.  Other plugins may need
	// the dedicated option `TLSEnable`.
	if c.TLSCA == "" && c.TLSKey == "" && c.TLSCert == "" && !c.InsecureSkipVerify {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ServerConfig) TLSConfig() (*tls.Config, error) {
	if c.TLSCert == "" && c.TLSKey == "" && len(c.TLSAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(c.TLSAllowedCACerts) != 0 {
		pool, err := makeCertPool(c.TLSAllowedCACerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	if len(c.TLSCipherSuites) != 0 {
		cipherSuites, err := ParseCiphers(c.TLSCipherSuites)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse server cipher suites %s: %v", strings.Join(c.TLSCipherSuites, ","), err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}

	if c.TLSMaxVersion != "" {
		version, err := ParseTLSVersion(c.TLSMaxVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls max version %q: %v", c.TLSMaxVersion, err)
		}
		tlsConfig.MaxVersion = version
	}

	if c.TLSMinVersion != "" {
		version, err := ParseTLSVersion(c.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls min version %q: %v", c.TLSMinVersion, err)
		}
		tlsConfig.MinVersion = version
	}

	if tlsConfig.MinVersion != 0 && tlsConfig.MaxVersion != 0 && tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return nil, fmt.Errorf(
			"tls min version %q can't be greater than tls max version %q", tlsConfig.MinVersion, tlsConfig.MaxVersion)
	}

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()
	return nil
}
