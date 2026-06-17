// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ibm_i

const sampleConfig = `
[[inputs.external]]
  daemon = true
  name   = "ibm_i"
  cmd    = "/usr/local/datakit/externals/ibm_i"

  ## Set true to enable election.
  election = true

  args = [
    "--interval", "60s",
    "--host", "<ibm-i-host>",
    "--username", "DKUSER",

    ## Optional: IBM i Access ODBC driver name registered in odbcinst.ini.
    # "--driver", "IBM i Access ODBC Driver 64-bit",

    ## Optional: use a raw ODBC connection string instead of host/username/password/driver.
    # "--dsn", "Driver={IBM i Access ODBC Driver 64-bit};System=<ibm-i-host>;UID=DKUSER;PWD=<password>;",

    ## Optional query and timeout settings.
    # "--query-timeout", "30s",
    # "--job-query-timeout", "240s",
    # "--system-mq-query-timeout", "80s",
    # "--severity-threshold", "50",

    ## Repeat these arguments as needed.
    ## When --query is omitted, all default queries are enabled.
    ## On large systems, start with base queries and enable job detail
    ## queries only when needed.
    # "--query", "disk_usage",
    # "--query", "cpu_usage",
    # "--query", "memory_info",
    # "--query", "subsystem",
    # "--query", "job_queue",

    ## Limit message_queue_info to selected queues to reduce target load.
    # "--message-queue", "QSYSOPR",
    # "--message-queue", "QSYSMSG",
  ]
  envs = [
    "ENV_INPUT_IBM_I_PASSWORD=<password>",
    "LD_LIBRARY_PATH=/opt/ibm/iaccess/lib64:$LD_LIBRARY_PATH",
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
