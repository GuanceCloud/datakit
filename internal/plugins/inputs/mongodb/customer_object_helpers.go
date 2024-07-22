// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (ipt *Input) setIptCOStatus() {
	ipt.CollectCoStatus = "OK"
}

func (ipt *Input) setIptErrCOStatus() {
	ipt.CollectCoStatus = "NotOK"
}

func (ipt *Input) setIptErrCOMsg(s string) {
	ipt.CollectCoErrMsg = s
}

func (ipt *Input) setIptLastCOInfoByErr() {
	ipt.LastCustomerObject.tags["reason"] = ipt.CollectCoErrMsg
	ipt.LastCustomerObject.tags["col_co_status"] = ipt.CollectCoStatus
}

func (ipt *Input) getCoPointByColErr() []*gcPoint.Point {
	ms := []inputs.MeasurementV2{}

	// Extract host and port from HostPort
	host, portStr, err := net.SplitHostPort(ipt.HostPort)
	if err != nil {
		log.Errorf("Error splitting HostPort: %v", err)
		return nil
	}
	port, _ := strconv.Atoi(portStr) // Convert port string to int

	if ipt.LastCustomerObject == nil {
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", host, port),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf(inputName+"-%s:%d", host, port),
			"host":          host,
			"ip":            fmt.Sprintf("%s:%d", host, port),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "database",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		log.Debug("m.Point().LineProto:", m.Point().LineProto())
		ipt.LastCustomerObject = m
		ms = append(ms, m)
	} else {
		ipt.setIptLastCOInfoByErr()
		ms = append(ms, ipt.LastCustomerObject)
	}
	pts := getPointsFromMeasurement(ms)

	return pts
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) GetMongoCoInfo(cli *mongo.Client) error {
	// 定义用于保存 MongoDB 服务器状态的结构体
	serverStatus := struct {
		Version string `bson:"version"`
		Uptime  int    `bson:"uptime"`
	}{}

	// 执行 serverStatus 命令以获取 MongoDB 服务器的状态信息
	err := cli.Database("admin").RunCommand(context.TODO(), bson.D{
		bson.E{Key: "serverStatus", Value: 1},
	}).Decode(&serverStatus)
	if err != nil {
		return fmt.Errorf("failed to get MongoDB server status: %w", err)
	}

	// 将 MongoDB 的版本和运行时间赋值给 ipt 结构体中的字段
	ipt.Version = serverStatus.Version
	ipt.Uptime = serverStatus.Uptime

	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}
	host, portStr, err := net.SplitHostPort(ipt.HostPort)
	if err != nil {
		log.Errorf("Error splitting HostPort: %v", err)
		return []*gcPoint.Point{}, nil
	}
	port, _ := strconv.Atoi(portStr) // Convert port string to int
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", host, port),
		"uptime":       ipt.Uptime,
		"version":      ipt.Version,
	}
	tags := map[string]string{
		"name":          fmt.Sprintf(inputName+"-%s:%d", host, port),
		"host":          host,
		"ip":            fmt.Sprintf("%s:%d", host, port),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "database",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ipt.LastCustomerObject = m
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) FeedCoErr(err error) {
	ipt.setIptErrCOMsg(err.Error())
	ipt.setIptErrCOStatus()
	pts := ipt.getCoPointByColErr()
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			dkio.WithLastErrorInput(inputName),
			dkio.WithLastErrorCategory(gcPoint.CustomObject),
		)
		log.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByPts() {
	for _, mgocli := range ipt.mgoSvrs {
		err := ipt.GetMongoCoInfo(mgocli.cli)
		if err != nil {
			log.Errorf("GetMongoCoInfo: %s", err)
			ipt.FeedCoErr(err)
		} else {
			pts, _ := ipt.collectCustomerObjectMeasurement()
			if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
				dkio.WithCollectCost(time.Since(time.Now())),
				dkio.WithElection(ipt.Election),
				dkio.WithInputName(inputName),
			); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(gcPoint.CustomObject),
				)
				log.Errorf("feed : %s", err)
			}
		}
	}
}
