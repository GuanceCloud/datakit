package iploc

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ip2location/ip2location-go"
	"github.com/stretchr/testify/assert"
)

var iplocationMockResult ip2location.IP2Locationrecord

type mockDB struct{}

func (db *mockDB) Get_all(ip string) (ip2location.IP2Locationrecord, error) {
	return iplocationMockResult, nil
}

var mockOpenDB = func(f string) (DB, error) {
	return &mockDB{}, nil
}

func TestIPloc(t *testing.T) {
	openDB = mockOpenDB
	iplocInstance := &IPloc{}
	datadir, err := ioutil.TempDir(os.TempDir(), "iploc")
	if err != nil {
		t.Fatalf("tmp dir create error")
	}
	defer os.RemoveAll(datadir) //nolint:errcheck
	iplocFile := "ip.bin"
	ispFile := "isp.txt"

	iplocDir := filepath.Join(datadir, "ipdb/iploc")
	err = os.MkdirAll(iplocDir, os.ModePerm)
	if err != nil {
		l.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(iplocDir, iplocFile), []byte(""), fs.ModePerm)
	if err != nil {
		t.Fatalf("iploc file create errror")
	}

	err = ioutil.WriteFile(filepath.Join(iplocDir, ispFile), []byte("221.0.0.0/13 unicom"), fs.ModePerm)
	if err != nil {
		t.Fatalf("isp file create errror")
	}

	iplocInstance.Init(datadir, map[string]string{"isp_file": ispFile, "iploc_file": iplocFile})

	t.Run("SearchIsp", func(t *testing.T) {
		assert.Equal(t, "unicom", iplocInstance.SearchIsp("221.0.0.0"))
		assert.Equal(t, "unknown", iplocInstance.SearchIsp("0.0.0.0"))
	})

	t.Run("Geo", func(t *testing.T) {
		iplocationMockResult = ip2location.IP2Locationrecord{Country_short: "CN", City: ErrInvalidIP, Region: ErrInvalidIP}
		ipInfo, err := iplocInstance.Geo("0.0.0.0")
		assert.NoError(t, err)
		assert.Equal(t, "CN", ipInfo.Country)
		assert.Equal(t, "unknown", ipInfo.City)

		backDB := iplocInstance.db
		iplocInstance.db = nil
		_, err = iplocInstance.Geo("0.0.0.0")

		assert.NoError(t, err)

		iplocInstance.db = backDB
	})
}
