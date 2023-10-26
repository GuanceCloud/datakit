// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nvidiasmi

const sampleCfg = `
[[inputs.gpu_smi]]

  ##(Optional) Collect interval, default is 10 seconds
  interval = "10s"

  ##The binPath of gpu-smi

  ##If nvidia GPU
  #(Example & default) bin_paths = ["/usr/bin/nvidia-smi"]
  #(Example windows) bin_paths = ["nvidia-smi"]

  ##If lluvatar GPU
  #(Example) bin_paths = ["/usr/local/corex/bin/ixsmi"]
  #(Example) envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]
  ##(Optional) Exec gpu-smi envs, default is []
  #envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]

  ##If remote GPU servers collected
  ##If use remote GPU servers, election must be true
  ##If use remote GPU servers, bin_paths should be shielded
  #(Example) remote_addrs = ["192.168.1.1:22"]
  #(Example) remote_users = ["remote_login_name"]
  ##If use remote_rsa_path, remote_passwords should be shielded
  #(Example) remote_passwords = ["remote_login_password"]
  #(Example) remote_rsa_paths = ["/home/your_name/.ssh/id_rsa"]
  #(Example) remote_command = "nvidia-smi -x -q"

  ##(Optional) Exec gpu-smi timeout, default is 5 seconds
  timeout = "5s"
  ##(Optional) Feed how much log data for ProcessInfos, default is 10. (0: 0 ,-1: all)
  process_info_max_len = 10
  ##(Optional) GPU drop card warning delay, default is 300 seconds
  gpu_drop_warning_delay = "300s"

  ## Set true to enable election
  election = false

[inputs.gpu_smi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
