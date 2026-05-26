// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDorisFEMetrics(t *testing.T) {
	text := `
# HELP doris_fe_qps qps
# TYPE doris_fe_qps gauge
doris_fe_qps 3
# HELP doris_fe_query_latency_ms latency
# TYPE doris_fe_query_latency_ms summary
doris_fe_query_latency_ms{quantile="0.5"} 10
doris_fe_query_latency_ms{quantile="0.75"} 20
doris_fe_query_latency_ms_sum 120
doris_fe_query_latency_ms_count 4
`

	metrics, err := parseDorisFEMetrics(strings.NewReader(text))
	require.NoError(t, err)

	assert.True(t, metrics.HasQPS)
	assert.Equal(t, float64(3), metrics.QPS)
	assert.True(t, metrics.HasAvgQueryTimeMS)
	assert.Equal(t, float64(30), metrics.AvgQueryTimeMS)
}

func TestSelectCurrentFrontend(t *testing.T) {
	frontends := []map[string]string{
		{
			"Name":             "fe-01",
			"Alive":            "true",
			"IsMaster":         "true",
			"CurrentConnected": "false",
			"Version":          "2.0.0-master",
			"HostName":         "doris-fe-01",
		},
		{
			"Name":             "fe-02",
			"Alive":            "true",
			"IsMaster":         "false",
			"CurrentConnected": "true",
			"Version":          "2.0.0-connected",
			"HostName":         "doris-fe-02",
		},
	}

	currentFE := selectCurrentFrontend(frontends)

	require.Equal(t, "fe-02", currentFE["Name"])
	assert.Equal(t, "2.0.0-connected", selectDorisVersion(currentFE))
	assert.Equal(t, "doris-fe-02", selectDorisHostname(currentFE))
}

func TestSelectCurrentFrontendNoFallback(t *testing.T) {
	frontends := []map[string]string{
		{
			"Name":     "fe-01",
			"Alive":    "true",
			"IsMaster": "false",
			"Version":  "2.0.0-follower",
			"IP":       "172.20.80.2",
		},
		{
			"Name":     "fe-master",
			"Alive":    "true",
			"IsMaster": "true",
			"Version":  "2.0.0-master",
			"IP":       "172.20.80.3",
		},
	}

	currentFE := selectCurrentFrontend(frontends)

	assert.Nil(t, currentFE)
	assert.Equal(t, "", selectDorisVersion(currentFE))
	assert.Equal(t, "", selectDorisHostname(currentFE))
}

func TestInitDatabaseInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.setup()
	ipt.db = db

	rows := sqlmock.NewRows([]string{"Name", "CurrentConnected", "Version", "HostName"}).
		AddRow("fe-01", "false", "2.0.0-master", "doris-fe-01").
		AddRow("fe-02", "true", "2.0.0-current", "doris-fe-02")
	mock.ExpectQuery(regexp.QuoteMeta(dorisFrontendsQuery)).WillReturnRows(rows)

	require.NoError(t, ipt.initDatabaseInfo(context.Background()))

	assert.Equal(t, "2.0.0-current", ipt.version)
	assert.Equal(t, "doris-fe-02", ipt.databaseInstance)
	assert.Equal(t, "127.0.0.1:9030", ipt.mergedTags["server"])
	assert.Equal(t, "doris-fe-02", ipt.mergedTags["database_instance"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInitDatabaseInfoKeepsConfiguredDatabaseInstance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.Tags["database_instance"] = "doris-prod"
	ipt.Tags["server"] = "doris-main"
	ipt.setup()
	ipt.db = db

	rows := sqlmock.NewRows([]string{"Name", "CurrentConnected", "Version", "HostName"}).
		AddRow("fe-01", "true", "2.0.0-current", "doris-fe-01")
	mock.ExpectQuery(regexp.QuoteMeta(dorisFrontendsQuery)).WillReturnRows(rows)

	require.NoError(t, ipt.initDatabaseInfo(context.Background()))

	assert.Equal(t, "2.0.0-current", ipt.version)
	assert.Equal(t, "doris-prod", ipt.databaseInstance)
	assert.Equal(t, "doris-main", ipt.mergedTags["server"])
	assert.Equal(t, "doris-prod", ipt.mergedTags["database_instance"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCollectDatabaseObjectReturnsFrontendError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.setup()
	ipt.db = db

	mock.ExpectQuery(regexp.QuoteMeta(dorisFrontendsQuery)).WillReturnError(errors.New("query failed"))

	err = ipt.collectDatabaseObject()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "show frontends")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInitDatabaseInfoReturnsFrontendError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.setup()
	ipt.db = db

	mock.ExpectQuery(regexp.QuoteMeta(dorisFrontendsQuery)).WillReturnError(errors.New("query failed"))

	err = ipt.initDatabaseInfo(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "show frontends")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestValidateCustomQueryRequiresFields(t *testing.T) {
	require.Error(t, validateCustomQuery(&customQuery{
		SQL:    "SELECT 1",
		Metric: "doris_custom",
	}))

	require.NoError(t, validateCustomQuery(&customQuery{
		SQL:    "SELECT 1 AS value",
		Metric: "doris_custom",
		Fields: []string{"value"},
	}))
}

func TestQueryCustomRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck

	ipt := defaultInput()
	ipt.db = db

	rows := sqlmock.NewRows([]string{"table_schema", "table_count", "state"}).
		AddRow("default", []byte("3"), []byte("NORMAL"))
	mock.ExpectQuery("SELECT custom").WillReturnRows(rows)

	result, err := ipt.queryCustomRows(context.Background(), "SELECT custom")
	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "default", result[0]["table_schema"])
	assert.Equal(t, "3", result[0]["table_count"])
	assert.Equal(t, "NORMAL", result[0]["state"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCustomQueryPoints(t *testing.T) {
	ipt := defaultInput()
	ipt.setup()
	ipt.mergedTags["database_instance"] = "doris-fe-01"

	query := &customQuery{
		Metric: "doris_custom",
		Tags:   []string{"table_schema"},
		Fields: []string{"table_count"},
	}
	rows := []map[string]interface{}{
		{
			"table_schema": "default",
			"table_count":  "3",
		},
	}

	pts := ipt.getCustomQueryPoints(query, rows, time.Now())
	require.Len(t, pts, 1)

	assert.Equal(t, "127.0.0.1:9030", pts[0].GetTag("server"))
	assert.Equal(t, "doris-fe-01", pts[0].GetTag("database_instance"))
	assert.Equal(t, "default", pts[0].GetTag("table_schema"))
	assert.Equal(t, float64(3), pts[0].Get("table_count"))
}

func TestDorisDSNStringWithTLS(t *testing.T) {
	ipt := defaultInput()
	ipt.setup()
	ipt.TLS = &TLS{InsecureSkipVerify: true}

	dsn, err := ipt.getDSNString()
	require.NoError(t, err)

	assert.Contains(t, dsn, "tls=doris-127.0.0.1%3A9030")
}

func TestCreateTLSConf(t *testing.T) {
	cfg, err := createTLSConf("", "", "")
	require.NoError(t, err)

	assert.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
}

func TestCreateTLSConfWithClientCertOnly(t *testing.T) {
	certFile, keyFile := createTestCertPair(t)

	cfg, err := createTLSConf("", certFile, keyFile)
	require.NoError(t, err)

	assert.Len(t, cfg.Certificates, 1)
	assert.Nil(t, cfg.RootCAs)
}

func TestCreateTLSConfRequiresCertAndKey(t *testing.T) {
	certFile, _ := createTestCertPair(t)

	_, err := createTLSConf("", certFile, "")
	require.Error(t, err)
}

func createTestCertPair(t *testing.T) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "doris-test",
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), 0o600))

	return certFile, keyFile
}

func TestMetricHTTPClientWithTLS(t *testing.T) {
	ipt := defaultInput()
	ipt.MetricTLS = &TLS{
		InsecureSkipVerify: true,
		AllowTLS10:         true,
	}

	client, err := ipt.metricHTTPClient()
	require.NoError(t, err)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	assert.Equal(t, uint16(tls.VersionTLS10), transport.TLSClientConfig.MinVersion)
}
