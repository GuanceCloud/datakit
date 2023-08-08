// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

const packageName = "traps"

var (
	l = logger.DefaultSLogger(packageName)
	g = datakit.G("snmp_traps")
)

type TrapsServerOpt struct {
	Enabled               bool
	BindHost              string
	Port                  uint16
	Namespace             string
	CommunityStrings      []string
	Users                 []UserV3
	StopTimeout           int
	Election              bool
	InputTags             map[string]string
	Feeder                dkio.Feeder
	Tagger                dkpt.GlobalTagger
	authoritativeEngineID string
}

// UserV3 contains the definition of one SNMPv3 user with its username and its auth
// parameters.
type UserV3 struct {
	Username     string
	AuthKey      string
	AuthProtocol string
	PrivKey      string
	PrivProtocol string
}

func (tso *TrapsServerOpt) Addr() string {
	return fmt.Sprintf("%s:%d", tso.BindHost, tso.Port)
}

//------------------------------------------------------------------------------
