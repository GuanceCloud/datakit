package prom

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

func getCaCertPool(caFile string) (*x509.CertPool, error) {
	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM")
	}
	return caCertPool, nil
}

func TLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	if len(caFile) > 0 && len(certFile) > 0 && len(keyFile) > 0 {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}

		caCertPool, err := getCaCertPool(caFile)

		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			MinVersion:   tls.VersionTLS10,
		}
		tlsConfig.BuildNameToCertificate()

		return tlsConfig, nil
	} else if len(caFile) > 0 {
		caCertPool, err := getCaCertPool(caFile)
		if err != nil {
			return nil, err
		}

		tlsConfig := &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS10,
		}
		tlsConfig.BuildNameToCertificate()
		return tlsConfig, nil

	} else {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS10,
		}
		tlsConfig.BuildNameToCertificate()
		return tlsConfig, nil
	}

}
