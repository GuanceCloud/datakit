// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logging collects host logging data.
package logging

var sampleCfg = `
# Log collection configuration example
[[inputs.logging]]
  # ========== File Configuration ==========
  # List of log file paths, supports glob patterns for batch specification
  # Recommended to use absolute paths with file extensions to avoid collecting unexpected files
  logfiles = [
    # Linux/Unix examples:
    # "/var/log/*.log",     # All .log files in directory
    # "/var/log/syslog",    # System log file
    # "/var/log/nginx/*",   # Nginx log directory
    # "/opt/app/logs/*.txt", # Application log files

    # Windows examples:
    # "C:/logs/*.log",
    # "D:/app/logs/*.txt",
    # '''C:\\Program Files\\App\\logs\\*.log''', # Use triple quotes for paths with spaces
  ]

  # Socket log reception, supports tcp/udp protocols
  # Recommended to use internal network ports for security
  sockets = [
    # "tcp://0.0.0.0:9540",  # TCP listening port
    # "udp://0.0.0.0:9541",  # UDP listening port
  ]

  # File path filtering, files matching these patterns will be ignored
  ignore = [
    # "*.tmp",
    # "*.swp",
    # "/var/log/old/*",
  ]

  # ========== Log Processing Configuration ==========
  # Log source identifier, defaults to 'default'
  source = ""

  # Service name, defaults to source value
  service = ""

  # Pipeline script path, uses <source>.p if empty
  pipeline = ""

  # Storage index name
  storage_index = ""

  # Ignored log levels
  # Supported levels: emerg/alert/critical/error/warning/info/debug/OK
  ignore_status = []

  # Character encoding, supports: utf-8/utf-16le/gbk/gb18030
  character_encoding = ""

  # Whether to remove ANSI escape codes (like terminal colors)
  remove_ansi_escape_codes = false

  # ========== Multiline Log Configuration ==========
  # Multiline log splitting regex pattern
  # Tip: use triple quotes '''regexp''' to avoid escaping
  # Regex reference: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\d{4}-\d{2}-\d{2}'''

  # Enable automatic multiline detection
  auto_multiline_detection = true

  # Additional multiline splitting patterns
  auto_multiline_extra_patterns = []

  # ========== Performance Configuration ==========
  # Maximum number of open files limit, default is 500
  # Global configuration, maximum value is used when multiple collectors configure this
  # max_open_files = 500

  # Time to ignore inactive files, default is "1h"
  ignore_dead_log = "1h"

  # Whether to read from the beginning of log files
  from_beginning = false

  # ========== Custom Tags ==========
  [inputs.logging.tags]
  # environment = "production"
  # region = "us-east-1"
  # team = "backend"
`
