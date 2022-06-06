package socket

import (
	"time"
)

const (
	StatusStop = "stop"
	MaxMsgSize = 15 * 1024 * 1024
	ClassTCP   = "TCP"
)

type Task interface {
	ID() string
	Status() string
	Run() error
	Init() error
	Class() string
	GetResults() (map[string]string, map[string]interface{})

	Ticker() *time.Ticker
}
