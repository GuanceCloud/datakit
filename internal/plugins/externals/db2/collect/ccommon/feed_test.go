// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ccommon

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

func TestFeedLastErrorLoop(t *testing.T) {
	const defaultErrInterval = time.Second * 30

	const inputName = "db2"
	l := logger.SLogger(inputName)
	DatakitLastErrURL = GetLastErrorURL("127.0.0.1", 9529)

	defer func() {
		// Restore variables' value.
		DatakitLastErrURL = ""
	}()

	type args struct {
		errString string
		ch        chan os.Signal
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "normal",
			args: args{
				errString: "something",
				ch:        make(chan os.Signal),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				feedErrInterval = time.Second * 3
				defer func() {
					// Restore default values.
					feedErrInterval = defaultErrInterval
				}()

				timeout := time.NewTicker(time.Second * 10)
				defer timeout.Stop()

				for range timeout.C {
					tt.args.ch <- os.Interrupt
					break
				}
			}()

			FeedLastErrorLoop(inputName, l, tt.args.errString, tt.args.ch)
		})
	}
}

func TestPrint(t *testing.T) {
	err := fmt.Errorf("some error")

	t.Run("Println", func(t *testing.T) {
		fmt.Println("failed:", err.Error())
	})
}

func TestFeedEvent(t *testing.T) {
	const inputName = "db2"
	t.Run("normal", func(t *testing.T) {
		l := logger.SLogger(inputName)

		event := &DFEvent{
			Name:        inputName,
			EventPrefix: EventPrefixDb2,
			Tags: map[string]string{
				"host":   "1.2.3.4",
				"source": inputName,
			},
			Date:    time.Now().Unix(),
			Status:  "warning",
			Title:   "Table space state change",
			Message: fmt.Sprintf("State of `%s` changed from `%s` to `%s`.", "test", "NORMAL", "DOWN"),
		}

		datakitPostEventURL := GetPostURL(false, CategoryEvent, inputName, "127.0.0.1", 9529)
		FeedEvent(l, event, datakitPostEventURL)
	})
}
