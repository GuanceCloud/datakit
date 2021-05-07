package mongodb

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
)

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TlsCAs             []string `json:"tls_cas" toml:"tls_cas"`
	TlsCert            string   `json:"tls_cert" toml:"tls_cert"`
	TlsKey             string   `json:"tls_key" toml:"tls_key"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify" toml:"insecure_skip_verify"`
	ServerName         string   `json:"tls_server_name" toml:"tls_server_name"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (this *ClientConfig) TlsConfig() (*tls.Config, error) {
	// This check returns a nil (aka, "use the default")
	// tls.Config if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security, or
	//     * an SNI server name.
	if len(this.TlsCAs) == 0 && this.TlsKey == "" && this.TlsCert == "" && !this.InsecureSkipVerify && this.ServerName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: this.InsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if len(this.TlsCAs) != 0 {
		if pool, err := MakeCertPool(this.TlsCAs); err != nil {
			return nil, err
		} else {
			tlsConfig.RootCAs = pool
		}
	}

	if this.TlsCert != "" && this.TlsKey != "" {
		if err := loadCertificate(tlsConfig, this.TlsCert, this.TlsKey); err != nil {
			return nil, err
		}
	}

	if this.ServerName != "" {
		tlsConfig.ServerName = this.ServerName
	}

	return tlsConfig, nil
}

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TlsCert           string   `json:"tls_cert" toml:"tls_cert"`
	TlsKey            string   `json:"tls_key" toml:"tls_key"`
	TlsAllowedCACerts []string `json:"tls_allowed_ca_certs" toml:"tls_allowed_ca_certs"`
	TlsCipherSuites   []string `json:"tls_cipher_suites" toml:"tls_cipher_suites"`
	TlsMinVersion     string   `json:"tls_min_version" toml:"tls_min_version"`
	TlsMaxVersion     string   `json:"tls_max_version" toml:"tls_max_version"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (this *ServerConfig) TlsConfig() (*tls.Config, error) {
	if this.TlsCert == "" && this.TlsKey == "" && len(this.TlsAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(this.TlsAllowedCACerts) != 0 {
		if pool, err := MakeCertPool(this.TlsAllowedCACerts); err != nil {
			return nil, err
		} else {
			tlsConfig.ClientCAs = pool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	if this.TlsCert != "" && this.TlsKey != "" {
		if err := loadCertificate(tlsConfig, this.TlsCert, this.TlsKey); err != nil {
			return nil, err
		}
	}

	if len(this.TlsCipherSuites) != 0 {
		if cipherSuites, err := ParseCiphers(this.TlsCipherSuites); err != nil {
			return nil, fmt.Errorf("could not parse server cipher suites %s: %v", strings.Join(this.TlsCipherSuites, ","), err)
		} else {
			tlsConfig.CipherSuites = cipherSuites
		}
	}

	if this.TlsMaxVersion != "" {
		if version, err := ParseTLSVersion(this.TlsMaxVersion); err != nil {
			return nil, fmt.Errorf("could not parse tls max version %q: %v", this.TlsMaxVersion, err)
		} else {
			tlsConfig.MaxVersion = version
		}
	}
	if this.TlsMinVersion != "" {
		if version, err := ParseTLSVersion(this.TlsMinVersion); err != nil {
			return nil, fmt.Errorf("could not parse tls min version %q: %v", this.TlsMinVersion, err)
		} else {
			tlsConfig.MinVersion = version
		}
	}

	if tlsConfig.MinVersion != 0 && tlsConfig.MaxVersion != 0 && tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return nil, fmt.Errorf("tls min version %q can't be greater than tls max version %q", tlsConfig.MinVersion, tlsConfig.MaxVersion)
	}

	return tlsConfig, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	if cert, err := tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		return fmt.Errorf("could not load keypair %s:%s: %v\n", certFile, keyFile, err)
	} else {
		config.Certificates = []tls.Certificate{cert}
		config.BuildNameToCertificate()

		return nil
	}
}

func MakeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		if pem, err := ioutil.ReadFile(certFile); err != nil {
			return nil, fmt.Errorf("could not read certificate %q: %v", certFile, err)
		} else {
			if ok := pool.AppendCertsFromPEM(pem); !ok {
				return nil, fmt.Errorf("could not parse any PEM certificates %q: %v", certFile, err)
			}
		}
	}

	return pool, nil
}

func ParseCiphers(ciphers []string) ([]uint16, error) {
	var (
		supported = tls.CipherSuites()
		suites    []uint16
	)
	for _, cipher := range ciphers {
		for _, suite := range supported {
			if cipher == suite.Name {
				suites = append(suites, suite.ID)
			} else {
				return nil, fmt.Errorf("unsupported cipher %q", cipher)
			}
		}
	}

	return suites, nil
}

// "TLS10": tls.VersionTLS10
// "TLS11": tls.VersionTLS11
// "TLS12": tls.VersionTLS12
// "TLS13": tls.VersionTLS13
// "TLS30": tls.VersionTLS30
func ParseTLSVersion(version string) (uint16, error) {
	switch version {
	case "TLS10":
		return tls.VersionTLS10, nil
	case "TLS11":
		return tls.VersionTLS11, nil
	case "TLS12":
		return tls.VersionTLS12, nil
	case "TLS13":
		return tls.VersionTLS13, nil
	// case "TLS30":
	// 	return tls.VersionSSL30, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %q", version)
	}
}
