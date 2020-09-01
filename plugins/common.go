package plugins

// If inputs and externals share some code, put them here. because inputs imports all telegraf and datakit inputs,
// DO NOT import inputs from externals, this makes external's binary too bigger.
//
// For the same reason, also do not import inputs from independent binary within datakit(i.e., installer.go).

import (
	"database/sql"
	"time"
)

type OracleMonitor struct {
	LibPath  string `json:"libPath" toml:"libPath"`
	Metric   string `json:"metricName" toml:"metricName"`
	Interval string `json:"interval" toml:"interval"`

	InstanceId string `json:"instanceId" toml:"instanceId"`
	User       string `json:"username" toml:"username"`
	Password   string `json:"password" toml:"password"`
	Desc       string `json:"instanceDesc" toml:"instanceDesc"`
	Host       string `json:"host" toml:"host"`
	Port       string `json:"port" toml:"port"`
	Server     string `json:"server" toml:"server"`
	Type       string `json:"type" toml:"type"`

	Tags map[string]string `json:"tags" toml:"tags"`

	DB               *sql.DB       `json:"-" json:"-"`
	IntervalDuration time.Duration `json:"-" json:"-"`
}
