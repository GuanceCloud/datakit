// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package chrony

const sampleCfg = `
[[inputs.chrony]]
  ## (Optional) Collect interval, default is 10 seconds
  # interval = "10s"

  ## (Optional) Exec chronyc timeout, default is 8 seconds
  # timeout = "8s"

  ## (Optional) The binPath of chrony
  bin_path = "chronyc"

  ## (Optional) Remote chrony servers
  ## If use remote chrony servers, election must be true
  ## If use remote chrony servers, bin_paths should be shielded
  # remote_addrs = ["<ip>:22"]
  # remote_users = ["<remote_login_name>"]
  # remote_passwords = ["<remote_login_password>"]
  ## If use remote_rsa_path, remote_passwords should be shielded
  # remote_rsa_paths = ["/home/<your_name>/.ssh/id_rsa"]
  # remote_command = "chronyc -n tracking"

  ## Set true to enable election
  election = true

[inputs.chrony.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
