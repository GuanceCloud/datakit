// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"path/filepath"
)

func TLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	// Load CA cert
	caCert, err := ioutil.ReadFile(filepath.Clean(caFile))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM")
	}

	tlsConfig := &tls.Config{ //nolint:gosec
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS10,
	}
	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
