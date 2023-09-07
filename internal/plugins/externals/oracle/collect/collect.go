// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package collect contains Oracle collect implement.
package collect

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect/ccommon"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect/coracle"
)

func Run(opt *ccommon.Option) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	ccommon.DatakitLastErrURL = ccommon.GetLastErrorURL(opt.DatakitHTTPHost, opt.DatakitHTTPPort)
	PrintInfof("Datakit lastError URL: %s", ccommon.DatakitLastErrURL)

	// If singleton, use function pointer's list: "[]func(string) ccommon.IInput"
	// If multiple instance, use interface's list: "[]ccommon.IInput"
	//
	// For now, we use function pointer's list because of performance improving,
	// e.g. if not specified, it should not be initialized.
	var definedInputList []func([]string, *ccommon.Option) ccommon.IInput
	definedInputList = append(definedInputList, coracle.NewInput)

	ch := make(chan struct{}, len(definedInputList))

	for _, inputFunc := range definedInputList {
		input := inputFunc(infoMsgs, opt)
		if input != nil {
			go func() {
				input.Run()
				ch <- struct{}{}
			}()
		}
	}

	count := 0
	select {
	case <-ch:
		count++
		if count == len(definedInputList) {
			close(ch)
		}
	case <-sigs:
		return
	}
}

var infoMsgs []string

func PrintInfof(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)

	fmt.Println(str)

	infoMsgs = append(infoMsgs, str)
}
