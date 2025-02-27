// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package xfsquota

const (
	inputName = "xfsquota"

	sampleConfig = `
[[inputs.xfsquota]]
    ## Path to the xfs_quota binary.
    binary_path = "/usr/sbin/xfs_quota"

    ## (Optional) Collect interval: (defaults to "1m").
    interval = "1m"

    ## Require.
    ## Filesystem path to which the quota will be applied.
    filesystem_path = "/hana"

    ## Specifies the version of the parsing format.
    parser_version = "v1"

    [inputs.xfsquota.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)
