// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import "os"

// ENV_DEFAULT_ENCODING: use protobuf point instead of lineprotocol(default use lineprotocol)

const (
	EnvDefaultEncoding = "ENV_DEFAULT_ENCODING"
)

func loadEnvs() {
	if v, ok := os.LookupEnv(EnvDefaultEncoding); ok {
		DefaultEncoding = EncodingStr(v)
	}
}
