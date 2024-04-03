//go:build linux
// +build linux

package run

import (
	"fmt"
	"time"

	manager "github.com/DataDog/ebpf-manager"

	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/offset"
)

func guessOffsetConntrack(guessed *dkoffset.OffsetConntrackC) (
	[]manager.ConstantEditor, *dkoffset.OffsetConntrackC, error,
) {
	var err error
	var constEditor []manager.ConstantEditor
	var ctOffset *dkoffset.OffsetConntrackC
	loopCount := 8

	for i := 0; i < loopCount; i++ {
		constEditor, ctOffset, err = dkoffset.GuessOffsetConntrack(guessed)
		if err == nil {
			return constEditor, ctOffset, nil
		}
		time.Sleep(time.Second * 5)
	}

	return constEditor, ctOffset, err
}

func guessOffsetHTTP(status *dkoffset.OffsetGuessC) ([]manager.ConstantEditor, error) {
	var err error
	var constEditor []manager.ConstantEditor
	loopCount := 5

	for i := 0; i < loopCount; i++ {
		constEditor, err = dkoffset.GuessOffsetHTTPFlow(status)
		if err == nil {
			return constEditor, err
		}
		time.Sleep(time.Second * 5)
	}
	return constEditor, err
}

func getOffset(saved *dkoffset.OffsetGuessC) (*dkoffset.OffsetGuessC, error) {
	bpfManager, err := dkoffset.NewGuessManger()
	if err != nil {
		return nil, fmt.Errorf("new offset manger: %w", err)
	}
	// Start the manager
	if err := bpfManager.Start(); err != nil {
		return nil, err
	}
	loopCount := 5
	defer bpfManager.Stop(manager.CleanAll) //nolint:errcheck
	for i := 0; i < loopCount; i++ {
		status, err := dkoffset.GuessOffset(bpfManager, saved, ipv6Disabled)
		if err != nil {
			saved = nil
			if i == loopCount-1 {
				return nil, err
			}
			log.Error(err)
			continue
		}

		constEditor := dkoffset.NewConstEditor(status)

		if enableTrace && enableHTTPFlow {
			_, offsetSeq, err := dkoffset.GuessOffsetTCPSeq(constEditor)
			if err != nil {
				saved = nil
				if i == loopCount-1 {
					return nil, err
				}
				log.Error(err)
				continue
			}

			dkoffset.SetTCPSeqOffset(status, offsetSeq)
		}

		return status, nil
	}
	return nil, err
}
