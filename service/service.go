package service

import (
	"context"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/uploader"
)

type (
	Service interface {
		Start(context.Context, uploader.IUploader) error
	}

	Creator func(log.Logger) Service
)

var Services = map[string]Creator{}

func Add(name string, creator Creator) {
	Services[name] = creator
}
