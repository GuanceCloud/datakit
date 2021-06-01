package inputs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// --------------------------------------------------------------------
// --------------------------- jolokia agent --------------------------
const (
	defaultInterval   = "60s"
	MaxGatherInterval = 30 * time.Minute
	MinGatherInterval = 1 * time.Second
)

type JolokiaAgent struct {
	DefaultFieldPrefix    string
	DefaultFieldSeparator string
	DefaultTagPrefix      string

	URLs            []string `toml:"urls"`
	Username        string
	Password        string
	ResponseTimeout time.Duration `toml:"response_timeout"`
	Interval        string

	tls.ClientConfig

	Metrics  []MetricConfig `toml:"metric"`
	gatherer *Gatherer
	clients  []*Client

	collectCache []Measurement
	PluginName   string `toml:"-"`
	l            *logger.Logger

	Tags  map[string]string `toml:"-"`
	Types map[string]string `toml:"-"`
}

func (j *JolokiaAgent) Collect() {
	j.l = logger.DefaultSLogger(j.PluginName)
	j.l.Infof("%s input started...", j.PluginName)
	j = j.Adaptor()

	duration, err := time.ParseDuration(j.Interval)
	if err != nil {
		j.l.Error(err)
		return
	}

	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := j.Gather(); err != nil {
				io.FeedLastError(j.PluginName, err.Error())
				j.l.Error(err)
			} else {
				FeedMeasurement(j.PluginName, datakit.Metric, j.collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})
				j.collectCache = j.collectCache[:0] // NOTE: do not forget to clean cache
			}

		case <-datakit.Exit.Wait():
			j.l.Infof("input %s exit", j.PluginName)
			return
		}
	}
}

func (ja *JolokiaAgent) Gather() error {
	if ja.gatherer == nil {
		ja.gatherer = NewGatherer(ja.createMetrics())
	}

	// Initialize clients once
	if ja.clients == nil {
		ja.clients = make([]*Client, 0, len(ja.URLs))
		for _, url := range ja.URLs {
			client, err := ja.createClient(url)
			if err != nil {
				return err
			}
			ja.clients = append(ja.clients, client)
		}
	}

	var wg sync.WaitGroup
	for _, client := range ja.clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()

			err := ja.gatherer.Gather(client, ja)
			if err != nil {
				ja.l.Errorf("unable to gather metrics for %s: %v", client.URL, err)
			}
		}(client)
	}

	wg.Wait()

	return nil
}

func (ja *JolokiaAgent) createMetrics() []Metric {
	var metrics []Metric

	for _, config := range ja.Metrics {
		metrics = append(metrics, NewMetric(config,
			ja.DefaultFieldPrefix, ja.DefaultFieldSeparator, ja.DefaultTagPrefix))
	}

	return metrics
}

func (ja *JolokiaAgent) createClient(url string) (*Client, error) {
	return NewClient(url, &ClientConfig{
		Username:        ja.Username,
		Password:        ja.Password,
		ResponseTimeout: ja.ResponseTimeout,
		ClientConfig:    ja.ClientConfig,
	})
}

func (j *JolokiaAgent) Adaptor() *JolokiaAgent {
	for i, m := range j.Metrics {
		var t string
		if m.FieldPrefix != nil {
			t = strings.ReplaceAll(*m.FieldPrefix, "#", "$")
			m.FieldPrefix = &t
		}

		if m.FieldSeparator != nil {
			t = strings.ReplaceAll(*m.FieldSeparator, "#", "$")
			m.FieldSeparator = &t
		}

		if m.FieldName != nil {
			t = strings.ReplaceAll(*m.FieldName, "#", "$")
			m.FieldName = &t
		}

		j.Metrics[i] = m
	}
	return j
}

// ----------------------- gatherer --------------------------------

const defaultFieldName = "value"

type JolokiaMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (j *JolokiaMeasurement) LineProto() (*io.Point, error) {
	data, err := io.MakePoint(j.name, j.tags, j.fields, j.ts)
	return data, err
}

func (j *JolokiaMeasurement) Info() *MeasurementInfo {
	return &MeasurementInfo{}
}

type Gatherer struct {
	metrics  []Metric
	requests []ReadRequest
}

func NewGatherer(metrics []Metric) *Gatherer {
	return &Gatherer{
		metrics:  metrics,
		requests: makeReadRequests(metrics),
	}
}

// Gather adds points to an accumulator from responses returned
// by a Jolokia agent.
func (g *Gatherer) Gather(client *Client, ja *JolokiaAgent) error {
	tags := make(map[string]string)
	if client.config.ProxyConfig != nil {
		tags["jolokia_proxy_url"] = client.URL
	} else {
		tags["jolokia_agent_url"] = client.URL
	}

	requests := makeReadRequests(g.metrics)
	responses, err := client.read(requests)
	if err != nil {
		return err
	}

	g.gatherResponses(responses, mergeTags(tags, ja.Tags), ja)

	return nil
}

// gatherResponses adds points to an accumulator from the ReadResponse objects
// returned by a Jolokia agent.
func (g *Gatherer) gatherResponses(responses []ReadResponse, tags map[string]string, ja *JolokiaAgent) error {
	series := make(map[string][]point)

	for _, metric := range g.metrics {
		points, ok := series[metric.Name]
		if !ok {
			points = make([]point, 0)
		}

		responsePoints, responseErrors := g.generatePoints(metric, responses)
		points = append(points, responsePoints...)
		if len(responseErrors) != 0 {
			return responseErrors[0]
		}

		series[metric.Name] = points
	}

	for measurement, points := range series {
		for _, point := range compactPoints(points) {
			if len(point.Fields) == 0 {
				continue
			}

			field := make(map[string]interface{})
			for k, v := range point.Fields {
				if v == nil {
					continue
				}

				if jn, ok := v.(json.Number); ok {
					field[k] = convertJsonNumber(k, jn, ja.Types)
				} else {
					field[k] = v
				}
			}

			if len(field) == 0 {
				continue
			}

			ja.collectCache = append(ja.collectCache, &JolokiaMeasurement{
				measurement,
				mergeTags(point.Tags, tags),
				field,
				time.Now(),
			})
		}
	}
	return nil
}

// generatePoints creates points for the supplied metric from the ReadResponse
// objects returned by the Jolokia client.
func (g *Gatherer) generatePoints(metric Metric, responses []ReadResponse) ([]point, []error) {
	points := make([]point, 0)
	errors := make([]error, 0)

	for _, response := range responses {
		switch response.Status {
		case 200:
			break
		case 404:
			continue
		default:
			errors = append(errors, fmt.Errorf("unexpected status in response from target %s (%q): %d",
				response.RequestTarget, response.RequestMbean, response.Status))
			continue
		}

		if !metricMatchesResponse(metric, response) {
			continue
		}

		pb := newPointBuilder(metric, response.RequestAttributes, response.RequestPath)
		for _, point := range pb.Build(metric.Mbean, response.Value) {
			if response.RequestTarget != "" {
				point.Tags["jolokia_agent_url"] = response.RequestTarget
			}

			points = append(points, point)
		}
	}

	return points, errors
}

// mergeTags combines two tag sets into a single tag set.
func mergeTags(metricTags, outerTags map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range outerTags {
		tags[k] = v
	}
	for k, v := range metricTags {
		tags[k] = v
	}

	return tags
}

// metricMatchesResponse returns true when the name, attributes, and path
// of a Metric match the corresponding elements in a ReadResponse object
// returned by a Jolokia agent.
func metricMatchesResponse(metric Metric, response ReadResponse) bool {
	if !metric.MatchObjectName(response.RequestMbean) {
		return false
	}

	if len(metric.Paths) == 0 {
		return len(response.RequestAttributes) == 0
	}

	for _, attribute := range response.RequestAttributes {
		if metric.MatchAttributeAndPath(attribute, response.RequestPath) {
			return true
		}
	}

	return false
}

// compactPoints attempts to remove points by compacting points
// with matching tag sets. When a match is found, the fields from
// one point are moved to another, and the empty point is removed.
func compactPoints(points []point) []point {
	compactedPoints := make([]point, 0)

	for _, sourcePoint := range points {
		keepPoint := true

		for _, compactPoint := range compactedPoints {
			if !tagSetsMatch(sourcePoint.Tags, compactPoint.Tags) {
				continue
			}

			keepPoint = false
			for key, val := range sourcePoint.Fields {
				compactPoint.Fields[key] = val
			}
		}

		if keepPoint {
			compactedPoints = append(compactedPoints, sourcePoint)
		}
	}

	return compactedPoints
}

// tagSetsMatch returns true if two maps are equivalent.
func tagSetsMatch(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for ak, av := range a {
		bv, ok := b[ak]
		if !ok {
			return false
		}
		if av != bv {
			return false
		}
	}

	return true
}

// makeReadRequests creates ReadRequest objects from metrics definitions.
func makeReadRequests(metrics []Metric) []ReadRequest {
	var requests []ReadRequest
	for _, metric := range metrics {
		if len(metric.Paths) == 0 {
			requests = append(requests, ReadRequest{
				Mbean:      metric.Mbean,
				Attributes: []string{},
			})
		} else {
			attributes := make(map[string][]string)

			for _, path := range metric.Paths {
				segments := strings.Split(path, "/")
				attribute := segments[0]

				if _, ok := attributes[attribute]; !ok {
					attributes[attribute] = make([]string, 0)
				}

				if len(segments) > 1 {
					paths := attributes[attribute]
					attributes[attribute] = append(paths, strings.Join(segments[1:], "/"))
				}
			}
			rootAttributes := findRequestAttributesWithoutPaths(attributes)
			if len(rootAttributes) > 0 {
				requests = append(requests, ReadRequest{
					Mbean:      metric.Mbean,
					Attributes: rootAttributes,
				})
			}

			for _, deepAttribute := range findRequestAttributesWithPaths(attributes) {
				for _, path := range attributes[deepAttribute] {
					requests = append(requests, ReadRequest{
						Mbean:      metric.Mbean,
						Attributes: []string{deepAttribute},
						Path:       path,
					})
				}
			}
		}
	}
	return requests
}

func findRequestAttributesWithoutPaths(attributes map[string][]string) []string {
	results := make([]string, 0)
	for attr, paths := range attributes {
		if len(paths) == 0 {
			results = append(results, attr)
		}
	}

	sort.Strings(results)
	return results
}

func findRequestAttributesWithPaths(attributes map[string][]string) []string {
	results := make([]string, 0)
	for attr, paths := range attributes {
		if len(paths) != 0 {
			results = append(results, attr)
		}
	}

	sort.Strings(results)
	return results
}

func convertJsonNumber(fieldName string, jn json.Number, fieldTyp map[string]string) interface{} {
	if fieldTyp != nil {
		return convertSpecifyType(fieldName, jn, fieldTyp)
	}

	if intVal, err := jn.Int64(); err == nil {
		return intVal
	}

	if floatVal, err := jn.Float64(); err == nil {
		return floatVal
	}

	return jn.String()
}

func convertSpecifyType(fieldName string, jn json.Number, fieldTyp map[string]string) interface{} {
	val, _ := jn.Float64()
	if typ, ok := fieldTyp[fieldName]; ok && typ == "int" {
		return int64(val)
	}
	return val
}

// ----------------------------------- metric ---------------------------------------
// A MetricConfig represents a TOML form of
// a Metric with some optional fields.
type MetricConfig struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      *string
	FieldPrefix    *string
	FieldSeparator *string
	TagPrefix      *string
	TagKeys        []string
}

// A Metric represents a specification for a
// Jolokia read request, and the transformations
// to apply to points generated from the responses.
type Metric struct {
	Name           string
	Mbean          string
	Paths          []string
	FieldName      string
	FieldPrefix    string
	FieldSeparator string
	TagPrefix      string
	TagKeys        []string

	mbeanDomain     string
	mbeanProperties []string
}

func NewMetric(config MetricConfig, defaultFieldPrefix, defaultFieldSeparator, defaultTagPrefix string) Metric {
	metric := Metric{
		Name:    config.Name,
		Mbean:   config.Mbean,
		Paths:   config.Paths,
		TagKeys: config.TagKeys,
	}

	if config.FieldName != nil {
		metric.FieldName = *config.FieldName
	}

	if config.FieldPrefix == nil {
		metric.FieldPrefix = defaultFieldPrefix
	} else {
		metric.FieldPrefix = *config.FieldPrefix
	}

	if config.FieldSeparator == nil {
		metric.FieldSeparator = defaultFieldSeparator
	} else {
		metric.FieldSeparator = *config.FieldSeparator
	}

	if config.TagPrefix == nil {
		metric.TagPrefix = defaultTagPrefix
	} else {
		metric.TagPrefix = *config.TagPrefix
	}

	mbeanDomain, mbeanProperties := parseMbeanObjectName(config.Mbean)
	metric.mbeanDomain = mbeanDomain
	metric.mbeanProperties = mbeanProperties

	return metric
}

func (m Metric) MatchObjectName(name string) bool {
	if name == m.Mbean {
		return true
	}

	mbeanDomain, mbeanProperties := parseMbeanObjectName(name)
	if mbeanDomain != m.mbeanDomain {
		return false
	}

	if len(mbeanProperties) != len(m.mbeanProperties) {
		return false
	}

NEXT_PROPERTY:
	for _, mbeanProperty := range m.mbeanProperties {
		for i := range mbeanProperties {
			if mbeanProperties[i] == mbeanProperty {
				continue NEXT_PROPERTY
			}
		}

		return false
	}

	return true
}

func (m Metric) MatchAttributeAndPath(attribute, innerPath string) bool {
	path := attribute
	if innerPath != "" {
		path = path + "/" + innerPath
	}

	for i := range m.Paths {
		if path == m.Paths[i] {
			return true
		}
	}

	return false
}

func parseMbeanObjectName(name string) (string, []string) {
	index := strings.Index(name, ":")
	if index == -1 {
		return name, []string{}
	}

	domain := name[:index]

	if index+1 > len(name) {
		return domain, []string{}
	}

	return domain, strings.Split(name[index+1:], ",")
}

// ----------------------------------------------------------------------------------
// ------------------------------------- point builder ------------------------------

type point struct {
	Tags   map[string]string
	Fields map[string]interface{}
}

type pointBuilder struct {
	metric           Metric
	objectAttributes []string
	objectPath       string
	substitutions    []string
}

func newPointBuilder(metric Metric, attributes []string, path string) *pointBuilder {
	return &pointBuilder{
		metric:           metric,
		objectAttributes: attributes,
		objectPath:       path,
		substitutions:    makeSubstitutionList(metric.Mbean),
	}
}

// Build generates a point for a given mbean name/pattern and value object.
func (pb *pointBuilder) Build(mbean string, value interface{}) []point {
	hasPattern := strings.Contains(mbean, "*")
	if !hasPattern {
		value = map[string]interface{}{mbean: value}
	}

	valueMap, ok := value.(map[string]interface{})
	if !ok { // FIXME: log it and move on.
		panic(fmt.Sprintf("There should be a map here for %s!\n", mbean))
	}

	points := make([]point, 0)
	for mbean, value := range valueMap {
		points = append(points, point{
			Tags:   pb.extractTags(mbean),
			Fields: pb.extractFields(mbean, value),
		})
	}

	return compactPoints(points)
}

// extractTags generates the map of tags for a given mbean name/pattern.
func (pb *pointBuilder) extractTags(mbean string) map[string]string {
	propertyMap := makePropertyMap(mbean, pb.metric.Mbean)
	tagMap := make(map[string]string)

	for key, value := range propertyMap {
		if pb.includeTag(key) {
			tagName := pb.formatTagName(key)
			tagMap[tagName] = value
		}
	}

	return tagMap
}

func (pb *pointBuilder) includeTag(tagName string) bool {
	for _, t := range pb.metric.TagKeys {
		if tagName == t {
			return true
		}
	}

	return false
}

func (pb *pointBuilder) formatTagName(tagName string) string {
	if tagName == "" {
		return ""
	}

	if tagPrefix := pb.metric.TagPrefix; tagPrefix != "" {
		return tagPrefix + tagName
	}

	return tagName
}

// extractFields generates the map of fields for a given mbean name
// and value object.
func (pb *pointBuilder) extractFields(mbean string, value interface{}) map[string]interface{} {
	fieldMap := make(map[string]interface{})
	valueMap, ok := value.(map[string]interface{})

	if ok {
		// complex value
		if len(pb.objectAttributes) == 0 {
			// if there were no attributes requested,
			// then the keys are attributes
			pb.fillFields("", valueMap, fieldMap)
		} else if len(pb.objectAttributes) == 1 {
			// if there was a single attribute requested,
			// then the keys are the attribute's properties
			fieldName := pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
			pb.fillFields(fieldName, valueMap, fieldMap)
		} else {
			// if there were multiple attributes requested,
			// then the keys are the attribute names
			for _, attribute := range pb.objectAttributes {
				fieldName := pb.formatFieldName(attribute, pb.objectPath)
				pb.fillFields(fieldName, valueMap[attribute], fieldMap)
			}
		}
	} else {
		// scalar value
		var fieldName string
		if len(pb.objectAttributes) == 0 {
			fieldName = pb.formatFieldName(defaultFieldName, pb.objectPath)
		} else {
			fieldName = pb.formatFieldName(pb.objectAttributes[0], pb.objectPath)
		}

		pb.fillFields(fieldName, value, fieldMap)
	}

	if len(pb.substitutions) > 1 {
		pb.applySubstitutions(mbean, fieldMap)
	}

	return fieldMap
}

// formatFieldName generates a field name from the supplied attribute and
// path. The return value has the configured FieldPrefix and FieldSuffix
// instructions applied.
func (pb *pointBuilder) formatFieldName(attribute, path string) string {
	fieldName := attribute
	fieldPrefix := pb.metric.FieldPrefix
	fieldSeparator := pb.metric.FieldSeparator

	if fieldPrefix != "" {
		fieldName = fieldPrefix + fieldName
	}

	if path != "" {
		fieldName = fieldName + fieldSeparator + strings.Replace(path, "/", fieldSeparator, -1)
	}

	return fieldName
}

// fillFields recurses into the supplied value object, generating a named field
// for every value it discovers.
func (pb *pointBuilder) fillFields(name string, value interface{}, fieldMap map[string]interface{}) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		// keep going until we get to something that is not a map
		for key, innerValue := range valueMap {
			if _, ok := innerValue.([]interface{}); ok {
				continue
			}

			var innerName string
			if name == "" {
				innerName = pb.metric.FieldPrefix + key
			} else {
				innerName = name + pb.metric.FieldSeparator + key
			}

			pb.fillFields(innerName, innerValue, fieldMap)
		}

		return
	}

	if _, ok := value.([]interface{}); ok {
		return
	}

	if pb.metric.FieldName != "" {
		name = pb.metric.FieldName
		if prefix := pb.metric.FieldPrefix; prefix != "" {
			name = prefix + name
		}
	}

	if name == "" {
		name = defaultFieldName
	}

	fieldMap[name] = value
}

// applySubstitutions updates all the keys in the supplied map
// of fields to account for $1-style substitution instructions.
func (pb *pointBuilder) applySubstitutions(mbean string, fieldMap map[string]interface{}) {
	properties := makePropertyMap(mbean, pb.metric.Mbean)

	for i, subKey := range pb.substitutions[1:] {
		symbol := fmt.Sprintf("$%d", i+1)
		substitution := properties[subKey]

		for fieldName, fieldValue := range fieldMap {
			newFieldName := strings.Replace(fieldName, symbol, substitution, -1)
			if fieldName != newFieldName {
				fieldMap[newFieldName] = fieldValue
				delete(fieldMap, fieldName)
			}
		}
	}
}

// makePropertyMap returns a the mbean property-key list as
// a dictionary. foo:x=y becomes map[string]string { "x": "y" }
func makePropertyMap(mbean, metricMbean string) map[string]string {
	props := make(map[string]string)
	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	// 如若 `Catalina:name="http*nio-*",type=ThreadPool"` 含多个 * 只取第一个 sub match；
	// 测试发现从 jolokia api 获取数据时，mbean 为 Catalina:name="http*nio-* 会报错，
	// 似乎 * 在 "" 内时通配符 * 不会匹配第二个 "
	metricDomain, metricRegexMap := makeTagValueRegexMap(metricMbean)
	if metricDomain != domain { // domain should equal
		metricRegexMap = make(map[string]*regexp.Regexp)
	}

	if domain != "" && len(object) == 2 {
		list := object[1]

		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			if key := pair[0]; key != "" {
				props[key] = pair[1]

				// 取 match[0][1] 作为 tag 值
				if v, ok := metricRegexMap[key]; ok && v != nil {
					match := v.FindAllStringSubmatch(pair[1], -1)
					if len(match) >= 1 && len(match[0]) > 1 {
						props[key] = match[0][1]
					}
				}
			}
		}
	}

	return props
}

// makeSubstitutionList returns an array of values to
// use as substitutions when renaming fields
// with the $1..$N syntax. The first item in the list
// is always the mbean domain.
func makeSubstitutionList(mbean string) []string {
	subs := make([]string, 0)

	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]

	if domain != "" && len(object) == 2 {
		subs = append(subs, domain)
		list := object[1]

		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)

			if len(pair) != 2 {
				continue
			}

			key := pair[0]
			if key == "" {
				continue
			}

			property := pair[1]
			if !strings.Contains(property, "*") {
				continue
			}

			subs = append(subs, key)
		}
	}

	return subs
}

// mbean 里的 * 替换为 (.*) 并编译正则表达式, 贪婪匹配
func makeTagValueRegexMap(mbean string) (string, map[string]*regexp.Regexp) {
	subs := make(map[string]*regexp.Regexp)
	object := strings.SplitN(mbean, ":", 2)
	domain := object[0]
	if domain != "" && len(object) == 2 {
		list := object[1]
		for _, keyProperty := range strings.Split(list, ",") {
			pair := strings.SplitN(keyProperty, "=", 2)
			if len(pair) == 2 && pair[0] != "" {
				// default nil
				subs[pair[0]] = nil
				property := pair[1]
				if strings.Contains(property, "*") {
					property = strings.Replace(property, "*", "(.*)", -1)
					if r, err := regexp.Compile(property); err == nil {
						// if successful
						subs[pair[0]] = r
					}
				}
			}
		}
	}
	return domain, subs
}

// ------------------------------------------------------------------------------
// ------------------------------------ client ----------------------------------
type Client struct {
	URL    string
	client *http.Client
	config *ClientConfig
}

type ClientConfig struct {
	ResponseTimeout time.Duration
	Username        string
	Password        string
	ProxyConfig     *ProxyConfig
	tls.ClientConfig
}

type ProxyConfig struct {
	DefaultTargetUsername string
	DefaultTargetPassword string
	Targets               []ProxyTargetConfig
}

type ProxyTargetConfig struct {
	Username string
	Password string
	URL      string
}

type ReadRequest struct {
	Mbean      string
	Attributes []string
	Path       string
}

type ReadResponse struct {
	Status            int
	Value             interface{}
	RequestMbean      string
	RequestAttributes []string
	RequestPath       string
	RequestTarget     string
}

// Jolokia JSON request object. Example: {
//   "type": "read",
//   "mbean: "java.lang:type="Runtime",
//   "attribute": "Uptime",
//   "target": {
//     "url: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//   }
// }
type jolokiaRequest struct {
	Type      string         `json:"type"`
	Mbean     string         `json:"mbean"`
	Attribute interface{}    `json:"attribute,omitempty"`
	Path      string         `json:"path,omitempty"`
	Target    *jolokiaTarget `json:"target,omitempty"`
}

type jolokiaTarget struct {
	URL      string `json:"url"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// Jolokia JSON response object. Example: {
//   "request": {
//     "type": "read"
//     "mbean": "java.lang:type=Runtime",
//     "attribute": "Uptime",
//     "target": {
//       "url": "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//     }
//   },
//   "value": 1214083,
//   "timestamp": 1488059309,
//   "status": 200
// }
type jolokiaResponse struct {
	Request jolokiaRequest `json:"request"`
	Value   interface{}    `json:"value"`
	Status  int            `json:"status"`
}

func NewClient(url string, config *ClientConfig) (*Client, error) {
	tlsConfig, err := config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.ResponseTimeout,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.ResponseTimeout,
	}

	return &Client{
		URL:    url,
		config: config,
		client: client,
	}, nil
}

func (c *Client) read(requests []ReadRequest) ([]ReadResponse, error) {
	jRequests := makeJolokiaRequests(requests, c.config.ProxyConfig)
	requestBody, err := json.Marshal(jRequests)
	if err != nil {
		return nil, err
	}

	requestURL, err := formatReadURL(c.URL, c.config.Username, c.config.Password)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		//err is not contained in returned error - it may contain sensitive data (password) which should not be logged
		return nil, fmt.Errorf("unable to create new request for: '%s'", c.URL)
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
			c.URL, resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jResponses []jolokiaResponse
	decoder := json.NewDecoder(bytes.NewReader(responseBody))
	decoder.UseNumber()
	err = decoder.Decode(&jResponses)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON response: %s: %s", err, responseBody)
	}
	return makeReadResponses(jResponses), nil
}

func makeJolokiaRequests(rrequests []ReadRequest, proxyConfig *ProxyConfig) []jolokiaRequest {
	jrequests := make([]jolokiaRequest, 0)
	if proxyConfig == nil {
		for _, rr := range rrequests {
			jrequests = append(jrequests, makeJolokiaRequest(rr, nil))
		}
	} else {
		for _, t := range proxyConfig.Targets {
			if t.Username == "" {
				t.Username = proxyConfig.DefaultTargetUsername
			}
			if t.Password == "" {
				t.Password = proxyConfig.DefaultTargetPassword
			}

			for _, rr := range rrequests {
				jtarget := &jolokiaTarget{
					URL:      t.URL,
					User:     t.Username,
					Password: t.Password,
				}

				jrequests = append(jrequests, makeJolokiaRequest(rr, jtarget))
			}
		}
	}

	return jrequests
}

func makeJolokiaRequest(rrequest ReadRequest, jtarget *jolokiaTarget) jolokiaRequest {
	jrequest := jolokiaRequest{
		Type:   "read",
		Mbean:  rrequest.Mbean,
		Path:   rrequest.Path,
		Target: jtarget,
	}

	if len(rrequest.Attributes) == 1 {
		jrequest.Attribute = rrequest.Attributes[0]
	}
	if len(rrequest.Attributes) > 1 {
		jrequest.Attribute = rrequest.Attributes
	}

	return jrequest
}

func makeReadResponses(jresponses []jolokiaResponse) []ReadResponse {
	rresponses := make([]ReadResponse, 0)

	for _, jr := range jresponses {
		rrequest := ReadRequest{
			Mbean:      jr.Request.Mbean,
			Path:       jr.Request.Path,
			Attributes: []string{},
		}

		attrValue := jr.Request.Attribute
		if attrValue != nil {
			attribute, ok := attrValue.(string)
			if ok {
				rrequest.Attributes = []string{attribute}
			} else {
				attributes, _ := attrValue.([]interface{})
				rrequest.Attributes = make([]string, len(attributes))
				for i, attr := range attributes {
					rrequest.Attributes[i] = attr.(string)
				}
			}
		}
		rresponse := ReadResponse{
			Value:             jr.Value,
			Status:            jr.Status,
			RequestMbean:      rrequest.Mbean,
			RequestAttributes: rrequest.Attributes,
			RequestPath:       rrequest.Path,
		}
		if jtarget := jr.Request.Target; jtarget != nil {
			rresponse.RequestTarget = jtarget.URL
		}

		rresponses = append(rresponses, rresponse)
	}

	return rresponses
}

func formatReadURL(configURL, username, password string) (string, error) {
	parsedURL, err := url.Parse(configURL)
	if err != nil {
		return "", err
	}

	readURL := url.URL{
		Host:   parsedURL.Host,
		Scheme: parsedURL.Scheme,
	}

	if username != "" || password != "" {
		readURL.User = url.UserPassword(username, password)
	}

	readURL.Path = path.Join(parsedURL.Path, "read")
	readURL.Query().Add("ignoreErrors", "true")
	return readURL.String(), nil
}
