// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package collect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/net/context/ctxhttp"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

// feedPoint sends points to DataKit using point package.
func (ipt *Input) feedPoint(points []*point.Point) error {
	if len(points) == 0 {
		return nil
	}

	if ipt.dkURLPath == "" {
		for _, pt := range points {
			cp.Infof("%s", pt.Pretty())
		}
		return nil
	}

	// Convert points to line protocol
	// TODO: we should use encode(in protobuf or line-proto) in reserved buffer
	var data []byte
	for _, pt := range points {
		data = append(data, []byte(pt.LineProto()+"\n")...)
	}

	if err := ipt.writeData(data); err != nil {
		return fmt.Errorf("failed to write points: %w", err)
	}

	return nil
}

func (ipt *Input) writeData(data []byte) error {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	httpReq, err := http.NewRequest("POST", ipt.dkURLPath, bytes.NewBuffer(data))
	if err != nil {
		l.Errorf(err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf(err.Error())
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
