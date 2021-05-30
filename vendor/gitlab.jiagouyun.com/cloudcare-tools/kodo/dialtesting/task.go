package dialtesting

import (
	"time"
)

const (
	StatusStop = "stop"

	ClassHTTP     = "HTTP"
	ClassTCP      = "TCP"
	ClassDNS      = "DNS"
	ClassHeadless = "HEADLESS"

	ClassOther = "OTHER"
)

type Task interface {
	ID() string
	Status() string
	Run() error
	Init() error
	CheckResult() []string
	Class() string
	GetResults() (map[string]string, map[string]interface{})
	PostURLStr() string
	MetricName() string
	Stop() error
	RegionName() string
	AccessKey() string
	Check() error
	UpdateTimeUs() int64
	GetFrequency() string
	GetOwnerExternalID() string
	SetOwnerExternalID(string)
	GetLineData() string

	SetRegionId(string)
	SetAk(string)
	SetStatus(string)
	SetUpdateTime(int64)

	Ticker() *time.Ticker
}
