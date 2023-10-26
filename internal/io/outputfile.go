// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"io"
)

func (x *dkIO) matchOutputFileInput(feedName string) bool {
	if len(x.outputFileInputs) == 0 { // no inputs configure, all inputs matched
		return true
	}

	for _, v := range x.outputFileInputs {
		if v == feedName {
			return true
		}
	}
	return false
}

func (x *dkIO) fileOutput(d *iodata) error {
	// concurrent write
	x.lock.Lock()
	defer x.lock.Unlock()

	if n, err := x.fd.WriteString("# " + d.from + " > " + d.category.String() + "\n"); err != nil {
		return err
	} else {
		x.outputFileSize += int64(n)
	}

	for _, pt := range d.points {
		if n, err := x.fd.WriteString(pt.LineProto() + "\n"); err != nil {
			return err
		} else {
			x.outputFileSize += int64(n)
			if x.outputFileSize > 32*1024*1024 { // truncate file on 32MB
				if err := x.fd.Truncate(0); err != nil {
					return fmt.Errorf("truncate error: %w", err)
				}

				if _, err := x.fd.Seek(0, io.SeekStart); err != nil {
					return fmt.Errorf("seek error: %w", err)
				}

				x.outputFileSize = 0
			}
		}
	}

	return nil
}
