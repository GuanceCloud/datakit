// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// TLSClientConfig represents the standard client TLS config.
type TLSClientConfig struct {
	CaCerts            []string `json:"ca_certs" toml:"ca_certs"`
	Cert               string   `json:"cert" toml:"cert"`
	CertKey            string   `json:"cert_key" toml:"cert_key"`
	CaCertsBase64      []string `json:"ca_certs_base64" toml:"ca_certs_base64"`
	CertBase64         string   `json:"cert_base64" toml:"cert_base64"`
	CertKeyBase64      string   `json:"cert_key_base64" toml:"cert_key_base64"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify" toml:"insecure_skip_verify"`
	ServerName         string   `json:"server_name" toml:"server_name"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not configured.
func (c *TLSClientConfig) TLSConfig() (*tls.Config, error) {
	// This check returns a nil (aka, "use the default")
	// tls.Config if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security, or
	//     * an SNI server name.

	if len(c.CaCerts) == 0 &&
		c.CertKey == "" &&
		c.Cert == "" &&
		!c.InsecureSkipVerify && //nolint:gosec
		c.ServerName == "" {
		return nil, nil
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec
		Renegotiation:      tls.RenegotiateNever,
	}

	if len(c.CaCerts) != 0 {
		pool, err := makeCertPool(c.CaCerts)
		if err != nil {
			return nil, err
		}

		tlsConfig.RootCAs = pool
	}

	if c.Cert != "" && c.CertKey != "" {
		if err := loadCertificate(tlsConfig, c.Cert, c.CertKey); err != nil {
			return nil, err
		}
	}

	tlsConfig.ServerName = c.ServerName

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := os.ReadFile(filepath.Clean(certFile))
		if err != nil {
			return nil, fmt.Errorf("could not read certificate %q: %w", certFile, err)
		}

		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("could not parse any PEM certificates %q: %w", certFile, err)
		}
	}

	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("could not load keypair %s:%s: %w", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()

	return nil
}

// MergeTLSConfig merge TLS config info. insecureSkipVerify default to false.
func MergeTLSConfig(t *TLSClientConfig,
	cacertFiles []string,
	certFile,
	keyFile string,
	tlsOpen,
	insecureSkipVerify bool,
) *TLSClientConfig {
	if t != nil {
		// Because ipt.InsecureSkipVerify is priority ipt.TLSClientConfig.InsecureSkipVerify
		if insecureSkipVerify {
			t.InsecureSkipVerify = true
		}

		return t
	}

	if t == nil && !tlsOpen {
		return nil
	}

	newCacertFiles := make([]string, 0)
	for _, s := range cacertFiles {
		if s != "" {
			newCacertFiles = append(newCacertFiles, s)
		}
	}
	insecure := insecureSkipVerify || len(newCacertFiles) == 0

	return &TLSClientConfig{
		CaCerts:            newCacertFiles,
		Cert:               certFile,
		CertKey:            keyFile,
		InsecureSkipVerify: insecure,
	}
}

// Base64ToTLSFiles returns TLS files from base64.
//
// Example: used by redis-cli.
//
//	if ipt.TLSClientConfig != nil && (ipt.TLSClientConfig.CertBase64 != "" || ipt.TLSClientConfig.CertKeyBase64 != "") {
//		   caCerts, cert, certKey, err := ipt.TLSClientConfig.Base64ToTLSFiles()
//		   if err != nil {
//		   	   l.Errorf("Collect: %s", err)
//		   	   return
//		   }
//		   ipt.TLSClientConfig.CaCerts = caCerts
//		   ipt.TLSClientConfig.Cert = cert
//		   ipt.TLSClientConfig.CertKey = certKey
//		   for _, caCert := range caCerts {
//		   	   defer os.Remove(caCert) // nolint:errcheck
//		   }
//		   defer os.Remove(cert)    // nolint:errcheck
//		   defer os.Remove(certKey) // nolint:errcheck
//	}
func (c *TLSClientConfig) Base64ToTLSFiles() (caCerts []string, cert, certKey string, err error) {
	if c == nil {
		return []string{}, "", "", fmt.Errorf("TLSClientConfig is nil")
	}

	if c.Cert != "" || c.CertKey != "" {
		return []string{}, "", "", fmt.Errorf("TLSClientConfig Cert or CertKey is not nil")
	}

	if c.CertBase64 == "" || c.CertKeyBase64 == "" {
		return []string{}, "", "", fmt.Errorf("TLSClientConfig CertBase64 or CertKeyBase64 is nil")
	}

	for _, p := range c.CaCertsBase64 {
		if b, err := base64.StdEncoding.DecodeString(p); err != nil {
			return []string{}, "", "", err
		} else {
			caFile, err := createFile(b, "ca.crt")
			if err != nil {
				return []string{}, "", "", err
			}
			caCerts = append(caCerts, caFile)
		}
	}

	if b, err := base64.StdEncoding.DecodeString(c.CertBase64); err != nil {
		return []string{}, "", "", err
	} else {
		cert, err = createFile(b, "cert.crt")
		if err != nil {
			return []string{}, "", "", err
		}
	}

	if b, err := base64.StdEncoding.DecodeString(c.CertKeyBase64); err != nil {
		return []string{}, "", "", err
	} else {
		certKey, err = createFile(b, "cert.key")
		if err != nil {
			return []string{}, "", "", err
		}
	}

	return caCerts, cert, certKey, nil
}

func createFile(b []byte, fileName string) (string, error) {
	f, err := os.CreateTemp("", "temp_*"+fileName)
	if err != nil {
		return "", err
	}

	_, err = f.Write(b)
	if err != nil {
		return "", err
	}

	err = f.Chmod(0o666)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

// TLSConfigWithBase64 returns a tls.Config, may be nil without error if TLS is not configured.
func (c *TLSClientConfig) TLSConfigWithBase64() (*tls.Config, error) {
	// This check returns a nil (aka, "use the default")
	// tls.Config if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security, or
	//     * an SNI server name.

	if c == nil {
		return nil, nil
	}

	var err error

	caCertsBlock := []([]byte){}
	certBlock := []byte{}
	certKeyBlock := []byte{}

	if len(c.CaCertsBase64) == 0 && c.CertBase64 == "" && c.CertKeyBase64 == "" {
		// load from file
		for _, p := range c.CaCerts {
			if b, err := os.ReadFile(filepath.Clean(p)); err != nil {
				return nil, err
			} else {
				caCertsBlock = append(caCertsBlock, b)
			}
		}
		if c.Cert != "" {
			if certBlock, err = os.ReadFile(c.Cert); err != nil {
				return nil, err
			}
		}

		if c.CertKey != "" {
			if certKeyBlock, err = os.ReadFile(c.CertKey); err != nil {
				return nil, err
			}
		}
	} else {
		// load from base64
		for _, p := range c.CaCertsBase64 {
			if b, err := base64.StdEncoding.DecodeString(p); err != nil {
				return nil, err
			} else {
				caCertsBlock = append(caCertsBlock, b)
			}
		}

		if c.CertBase64 != "" {
			if certBlock, err = base64.StdEncoding.DecodeString(c.CertBase64); err != nil {
				return nil, err
			}
		}

		if c.CertKeyBase64 != "" {
			if certKeyBlock, err = base64.StdEncoding.DecodeString(c.CertKeyBase64); err != nil {
				return nil, err
			}
		}
	}

	if len(caCertsBlock) == 0 {
		c.InsecureSkipVerify = true
	}

	if len(caCertsBlock) == 0 &&
		len(certKeyBlock) == 0 &&
		len(certBlock) == 0 &&
		!c.InsecureSkipVerify && //nolint:gosec
		c.ServerName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec
		Renegotiation:      tls.RenegotiateNever,
	}

	if len(caCertsBlock) != 0 {
		pool, err := makeCertPoolWithBase64(caCertsBlock)
		if err != nil {
			return nil, err
		}

		tlsConfig.RootCAs = pool
	}

	if len(certBlock) != 0 && len(certKeyBlock) != 0 {
		if err := loadCertificateWithBase64(tlsConfig, certBlock, certKeyBlock); err != nil {
			return nil, err
		}
	}

	tlsConfig.ServerName = c.ServerName

	return tlsConfig, nil
}

func makeCertPoolWithBase64(certInfos [][]byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, block := range certInfos {
		if ok := pool.AppendCertsFromPEM(block); !ok {
			return nil, fmt.Errorf("could not parse any PEM certificates %q", string(block))
		}
	}

	return pool, nil
}

func loadCertificateWithBase64(config *tls.Config, certBlock, keyBlock []byte) error {
	cert, err := tls.X509KeyPair(certBlock, keyBlock)
	if err != nil {
		return fmt.Errorf("could not load key pair %s:%s: %w", string(certBlock), string(keyBlock), err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()

	return nil
}

// TODO ...

func LoadTLSConfigByBase64(caCerts []string, cert, certKey string, insecureSkipVerify bool, serverName string) (*tls.Config, error) {
	if len(caCerts) == 0 &&
		cert == "" &&
		certKey == "" &&
		!insecureSkipVerify && //nolint:gosec
		serverName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		Renegotiation:      tls.RenegotiateNever,
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec
		ServerName:         serverName,
	}

	if len(caCerts) != 0 {
		pool := x509.NewCertPool()
		for _, caCert := range caCerts {
			caCertPEMBlock, err := base64.StdEncoding.DecodeString(caCert)
			if err != nil {
				return nil, fmt.Errorf("could not read caCert data %w", err)
			}
			if ok := pool.AppendCertsFromPEM(caCertPEMBlock); !ok {
				return nil, fmt.Errorf("could not parse any PEM certificates request %w", err)
			}
		}
		tlsConfig.RootCAs = pool
	}

	if cert != "" && certKey != "" {
		certPEMBlock, err := base64.StdEncoding.DecodeString(cert)
		if err != nil {
			return nil, fmt.Errorf("could not read cert data %w", err)
		}
		keyPEMBlock, err := base64.StdEncoding.DecodeString(certKey)
		if err != nil {
			return nil, fmt.Errorf("could not read certKey data %w", err)
		}

		cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			return nil, fmt.Errorf("could not load keypair %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	return tlsConfig, nil
}

func DefaultTLSConfigWithInsecureSkipVerify(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		Renegotiation:      tls.RenegotiateNever,
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec
	}
}
