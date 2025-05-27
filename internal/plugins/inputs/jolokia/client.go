// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type jclient struct {
	url     string
	client  *http.Client
	config  *jclientConfig
	upState int
}

type jclientConfig struct {
	respTimeout time.Duration
	username    string
	password    string
	proxyCfg    *proxyConfig
	tls.ClientConfig
}

type proxyConfig struct {
	defTargetUsername string
	defTargetPassword string
	targets           []proxyTargetConfig
}

type proxyTargetConfig struct {
	username string
	password string
	url      string
}

type readRequest struct {
	mbean      string
	attributes []string
	path       string
}

type readResponse struct {
	status            int
	value             interface{}
	requestMbean      string
	requestAttributes []string
	requestPath       string
	requestTarget     string
}

//	Jolokia JSON request object. Example: {
//	  "type": "read",
//	  "mbean: "java.lang:type="Runtime",
//	  "attribute": "Uptime",
//	  "target": {
//	    "url: "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//	  }
//	}.
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

//	jolokiaResponseJolokia JSON response object. Example: {
//	  "request": {
//	    "type": "read"
//	    "mbean": "java.lang:type=Runtime",
//	    "attribute": "Uptime",
//	    "target": {
//	      "url": "service:jmx:rmi:///jndi/rmi://target:9010/jmxrmi"
//	    }
//	  },
//	  "value": 1214083,
//	  "timestamp": 1488059309,
//	  "status": 200
//	}.
type jolokiaResponse struct {
	Request jolokiaRequest `json:"request"`
	Value   interface{}    `json:"value"`
	Status  int            `json:"status"`
}

func newClient(url string, config *jclientConfig) (*jclient, error) {
	tlsConfig, err := config.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		ResponseHeaderTimeout: config.respTimeout,
		TLSClientConfig:       tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.respTimeout,
	}

	return &jclient{
		url:    url,
		config: config,
		client: client,
	}, nil
}

func (c *jclient) read(requests []readRequest) ([]readResponse, error) {
	jRequests := makeJolokiaRequests(requests, c.config.proxyCfg)
	requestBody, err := json.Marshal(jRequests)
	if err != nil {
		return nil, err
	}
	// kafka requestURL eg http://localhost:8080/jolokia/read
	requestURL, err := formatReadURL(c.url, c.config.username, c.config.password)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(requestBody))
	if err != nil {
		// err is not contained in returned error - it may contain sensitive data (password) which should not be logged
		return nil, fmt.Errorf("unable to create new request for: '%s'", c.url)
	}

	req.Header.Add("Content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
			c.url, resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
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
		return nil, fmt.Errorf("decoding JSON response: %w: %s", err, responseBody)
	}
	return makeReadResponses(jResponses), nil
}

func makeJolokiaRequests(rrequests []readRequest, proxyConfig *proxyConfig) []jolokiaRequest {
	jrequests := make([]jolokiaRequest, 0)
	if proxyConfig == nil {
		for _, rr := range rrequests {
			jrequests = append(jrequests, makeJolokiaRequest(rr, nil))
		}
	} else {
		for _, t := range proxyConfig.targets {
			if t.username == "" {
				t.username = proxyConfig.defTargetUsername
			}
			if t.password == "" {
				t.password = proxyConfig.defTargetPassword
			}

			for _, rr := range rrequests {
				jtarget := &jolokiaTarget{
					URL:      t.url,
					User:     t.username,
					Password: t.password,
				}

				jrequests = append(jrequests, makeJolokiaRequest(rr, jtarget))
			}
		}
	}

	return jrequests
}

func makeJolokiaRequest(rrequest readRequest, jtarget *jolokiaTarget) jolokiaRequest {
	jrequest := jolokiaRequest{
		Type:   "read",
		Mbean:  rrequest.mbean,
		Path:   rrequest.path,
		Target: jtarget,
	}

	if len(rrequest.attributes) == 1 {
		jrequest.Attribute = rrequest.attributes[0]
	}
	if len(rrequest.attributes) > 1 {
		jrequest.Attribute = rrequest.attributes
	}

	return jrequest
}

func makeReadResponses(jresponses []jolokiaResponse) []readResponse {
	rresponses := make([]readResponse, 0)

	for _, jr := range jresponses {
		rrequest := readRequest{
			mbean:      jr.Request.Mbean,
			path:       jr.Request.Path,
			attributes: []string{},
		}

		attrValue := jr.Request.Attribute
		if attrValue != nil {
			attribute, ok := attrValue.(string)
			if ok {
				rrequest.attributes = []string{attribute}
			} else {
				attributes, _ := attrValue.([]interface{})
				rrequest.attributes = make([]string, len(attributes))
				for i, attr := range attributes {
					if s, ok := attr.(string); ok {
						rrequest.attributes[i] = s
					} else {
						log.Warnf("attr expect to be string, go %s", reflect.TypeOf(attr).String())
					}
				}
			}
		}
		rresponse := readResponse{
			value:             jr.Value,
			status:            jr.Status,
			requestMbean:      rrequest.mbean,
			requestAttributes: rrequest.attributes,
			requestPath:       rrequest.path,
		}
		if jtarget := jr.Request.Target; jtarget != nil {
			rresponse.requestTarget = jtarget.URL
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
