package mongodb

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// TlsClientConfig represents the standard client TLS config.
type TlsClientConfig struct {
	CaCerts            []string `json:"ca_certs" toml:"ca_certs"`
	Cert               string   `json:"cert" toml:"cert"`
	CertKey            string   `json:"cert_key" toml:"cert_key"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify" toml:"insecure_skip_verify"`
	ServerName         string   `json:"server_name" toml:"server_name"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (this *TlsClientConfig) TlsConfig() (*tls.Config, error) {
	// This check returns a nil (aka, "use the default")
	// tls.Config if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security, or
	//     * an SNI server name.
	if len(this.CaCerts) == 0 && this.CertKey == "" && this.Cert == "" && !this.InsecureSkipVerify && this.ServerName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: this.InsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if len(this.CaCerts) != 0 {
		if pool, err := makeCertPool(this.CaCerts); err != nil {
			return nil, err
		} else {
			tlsConfig.RootCAs = pool
		}
	}

	if this.Cert != "" && this.CertKey != "" {
		if err := loadCertificate(tlsConfig, this.Cert, this.CertKey); err != nil {
			return nil, err
		}
	}

	if this.ServerName != "" {
		tlsConfig.ServerName = this.ServerName
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

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
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
