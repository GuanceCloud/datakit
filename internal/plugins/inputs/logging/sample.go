// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logging collects host logging data.
package logging

var sampleCfg = `
[[inputs.logging]]
  # List of log files, supports batch specification using glob patterns.
  # It is recommended to use absolute paths and specify file extensions.
  # Narrow the scope as much as possible to avoid collecting unexpected files.
  logfiles = [
    # UNIX-like log path example:
    # "/var/log/*.log",  # All log files in the directory
    # "/var/log/*.txt",  # All txt files in the directory
    # "/var/log/sys*",   # All files prefixed with sys in the directory
    # "/var/log/syslog", # Unix-style file path

    # Windows log path example:
    # "C:/path/to/*.txt",
    # or like this(with space in path):
    # '''C:\\path\\to some\\*.txt''',
  ]

  ## Socket currently supports two protocols: tcp/udp. It is recommended to use internal
  ## network ports for security.
  socket = [
   #"tcp://0.0.0.0:9540",
   #"udp://0.0.0.0:9541"
  ]

  # File path filtering using glob patterns, any file matching these patterns will not be collected
  ignore = [""]

  # Logging source, defaults to 'default' if empty
  source = ""

  # Logging service, defaults to source name if empty
  service = ""

  # Pipeline script path, uses <source>.p if empty.
  pipeline = ""

  # Set index name.
  storage_index = ""

  # Ignore logging levels(status):
  # allowed levels: emerg/alert/critical/error/warning/info/debug/OK
  ignore_status = []

  # Select logging encoding
  #    utf-8/utf-16le/utf-16le/gbk/gb18030
  character_encoding = ""

  # Regexp to split multiline log.
  # Tips: use three single quotes '''this-regexp''' to avoid escaping.
  # Regex reference: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  # Enable automatic multiline mode
  auto_multiline_detection = true

  # Add more multiline split patterns
  auto_multiline_extra_patterns = []

  ## Whether to remove ANSI escape codes, such as text colors in standard output
  remove_ansi_escape_codes = false

  ## Limit the maximum number of open files, default is 500
  ## This is a global configuration; if multiple collectors configure this, the maximum value will be used
  # max_open_files = 500

  ## Ignore inactive files
  ignore_dead_log = "1h"

  ## Whether to read from the beginning of log files
  from_beginning = false

  # Custom tags
  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`
