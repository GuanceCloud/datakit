// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package operator implement API request to operator.
package operator

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var log = logger.DefaultSLogger("operator")

type Operator struct {
	address     string
	routes      map[string]string
	httpTimeout time.Duration
	httpCli     *http.Client
}

func NewOperatorClientFromEnv() *Operator {
	op := &Operator{
		routes:      map[string]string{},
		httpTimeout: time.Second * 3,
		httpCli: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					//nolint
					InsecureSkipVerify: true,
				},
			},
		},
	}

	if operatorAddress := os.Getenv("ENV_DATAKIT_OPERATOR"); operatorAddress != "" {
		u, err := url.Parse(operatorAddress)
		if err != nil {
			log.Error(err)
		} else {
			op.address = u.String()
			op.routes[datakit.Ping] = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, "/ping")
			op.routes[datakit.Election] = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, "/v1/datakit/election")
		}
	}

	return op
}

func (op *Operator) Ping() error {
	if op.address == "" {
		return fmt.Errorf("invalid operator address")
	}

	requrl := op.routes[datakit.Ping]
	log.Debugf("ping sending %s", requrl)

	req, err := http.NewRequest("HEAD", requrl, nil)
	if err != nil {
		return err
	}

	resp, err := op.httpCli.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	switch resp.StatusCode / 100 {
	case 2:
		return nil
	default:
		return fmt.Errorf("ping failed")
	}
}

func (op *Operator) Election(namespace, id string, reqBody io.Reader) ([]byte, error) {
	requrl := op.routes[datakit.Election]
	log.Debugf("election sending %s", requrl)

	req, err := http.NewRequest("POST", requrl, reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := op.httpCli.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2:
		log.Debugf("election %s ok", requrl)
		return body, nil
	default:
		log.Debugf("election failed: %d", resp.StatusCode)
		return nil, fmt.Errorf("election failed: %s", string(body))
	}
}

func (op *Operator) ElectionHeartbeat(namespace, id string, req io.Reader) ([]byte, error) {
	return nil, fmt.Errorf("undefined interface")
}
