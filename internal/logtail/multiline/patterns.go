// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package multiline

var GlobalDigitPatterns = []string{
	// time.RFC3339, "2006-01-02T15:04:05Z07:00"
	`^\d+-\d+-\d+T\d+:\d+:\d+(\.\d+)?(Z\d*:?\d*)?`,
	// time.RFC822, "02 Jan 06 15:04 MST"
	`^\d+ [A-Za-z_]+ \d+ \d+:\d+ [A-Za-z_]+`,
	// time.RFC822Z, "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	`^\d+ [A-Za-z_]+ \d+ \d+:\d+ -\d+`,
	// time.RFC3339Nano, "2006-01-02T15:04:05.999999999Z07:00"
	`^\d+-\d+-\d+[A-Za-z_]+\d+:\d+:\d+\.\d+[A-Za-z_]+\d+:\d+`,
	// 2021-07-08 05:08:19,214
	`^\d+-\d+-\d+ \d+:\d+:\d+(,\d+)?`,
	// 2021-07-08 05:08:19.214, used by log4j, postgresql, rabbitmq, jenkins.
	`^\d{4}-\d{2}-\d{2}[ T]\d{1,2}:\d{2}:\d{2}\.\d+([ \t]+[A-Z]{2,5})?`,
	// 2021-07-08,05:08:19.214
	`^\d{4}-\d{2}-\d{2},\d{1,2}:\d{2}:\d{2}([,.]\d+)?`,
	// 2021.07.08 05:08:19
	`^\d{4}\.\d{1,2}\.\d{1,2}[ T]\d{1,2}:\d{2}:\d{2}([,.]\d+)?`,
	// 210708 05:08:19
	`^\d{6} +\d{2}:\d{2}:\d{2}`,
	// 10/Oct/2000:13:55:36 -0700
	`^\d{1,2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}`,
	// Slash-separated dates used by nginx error logs and many application logs.
	`^\d{4}/\d{1,2}/\d{1,2}[ T-]\d{1,2}:\d{2}:\d{2}([,.]\d+)?`,
	// US/EU slash dates, "6/25/2024 2:30:15 PM - INFO ..."
	`^\d{1,2}[/-]\d{1,2}[/-]\d{2,4}[ T]\d{1,2}:\d{2}:\d{2}([,.]\d+)?( (AM|PM))?`,
	// 02-Jan-2006 15:04:05.123, used by Tomcat and Oracle-style logs.
	`^\d{1,2}-[A-Za-z]{3}-\d{4} \d{2}:\d{2}:\d{2}([,.]\d+)?`,
	// Compact timestamps, "20240708 050819" and "20240708T05:08:19".
	`^\d{8}[ T]\d{2}:?\d{2}:?\d{2}([,.]\d+)?`,
	// Time-only logs are accepted only when immediately followed by a level.
	`^\d{2}:\d{2}:\d{2}([,.]\d+)?[ \t]+(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\b`,
	// Unix epoch seconds/milliseconds followed by a level.
	`^\d{10,13}([,.]\d+)?[ \t]+(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\b`,
	// Apache/nginx access log with HTTPDATE after client/ident/auth fields.
	`^[0-9A-Fa-f:.]+ [^ ]+ [^ ]+ \[\d{1,2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\]`,
	// Android logcat, "01-02 15:04:05.123  123  456 I/Tag: message"
	`^\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} +\d+ +\d+ [VDIWEF]/`,
	// 2021-01-31 - with stricter matching around the months/days
	`^\d{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])`,
	// MongoDB text logs, "2021-07-08T05:08:19.123+0000 I CONTROL ..."
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?[+-]\d{4}[ \t]+[A-Z][ \t]+`,
	// Redis server log, "31350:M 23 Jan 2020 11:45:04.030 * Ready"
	`^\d+:[A-Z] \d{1,2} [A-Za-z]{3} \d{4} \d{2}:\d{2}:\d{2}\.\d+`,
}

var GlobalLetterPatterns = []string{
	// time.ANSIC, "Mon Jan _2 15:04:05 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+ \d+`,
	// time.RubyDate, "Mon Jan 02 15:04:05 -0700 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [\-\+]\d+ \d+`,
	// time.UnixDate, "Mon Jan _2 15:04:05 MST 2006"
	`^[A-Za-z_]+ [A-Za-z_]+ +\d+ \d+:\d+:\d+( [A-Za-z_]+ \d+)?`,
	// time.RFC850, "Monday, 02-Jan-06 15:04:05 MST"
	`^[A-Za-z_]+, \d+-[A-Za-z_]+-\d+ \d+:\d+:\d+ [A-Za-z_]+`,
	// time.RFC1123, "Mon, 02 Jan 2006 15:04:05 MST"
	`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ [A-Za-z_]+`,
	// time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	`^[A-Za-z_]+, \d+ [A-Za-z_]+ \d+ \d+:\d+:\d+ -\d+`,
	// Syslog RFC3164, "Jan 02 15:04:05 host app[123]:"
	`^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) +\d{1,2} \d{2}:\d{2}:\d{2}[ \t]`,
	// Syslog-like timestamp with fractional seconds.
	`^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) +\d{1,2} \d{2}:\d{2}:\d{2}\.\d+[ \t]`,
	// Apache/nginx access log with HTTPDATE after client/ident/auth fields.
	`^[0-9A-Fa-f:.]+ [^ ]+ [^ ]+ \[\d{1,2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\]`,
	// Kubernetes/glog/klog, "E0123 12:34:56.789012 1 file.go:123] ..."
	`^[IWEF]\d{4} \d{2}:\d{2}:\d{2}\.\d{6} +\d+ `,
	// Docker daemon and logfmt-style timestamps, `time="2021-07-08T05:08:19Z"`.
	`^(time|ts|timestamp)=["']?\d{4}-\d{2}-\d{2}[T ]\d{2}:?\d{2}:?\d{2}`,
	// logfmt with level first, `level=info ts=2021-07-08T05:08:19Z`.
	`^level=(trace|debug|info|notice|warn|warning|error|fatal|panic)[ \t]+(time|ts|timestamp)=`,
	// logfmt with level first and no timestamp field, `level=info msg=...`.
	`^(level|severity)=(trace|debug|info|notice|warn|warning|error|fatal|panic)\b`,
	// Level-first logs, "INFO 2021-07-08 05:08:19 ..." or "[ERROR] 2021/07/08 ..."
	`^\[?(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\]?[: ]+\d{4}[-/]\d{1,2}[-/]\d{1,2}`,
	// Level-only starts, "INFO [main] ..." or "ERROR app.module: ...".
	`^(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)[ \t]+(\[|\(|[A-Za-z0-9_.-]+[: -])`,
	// Pipe-separated level-first logs, common in Loguru-style output.
	`^(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)[ \t]*\|`,
	// Python logging defaults, "ERROR:root:message".
	`^(DEBUG|INFO|WARNING|ERROR|CRITICAL):[A-Za-z0-9_.-]+:`,
	// Default java logging SimpleFormatter date format
	`^[A-Za-z_]+ \d+, \d+ \d+:\d+:\d+ (AM|PM)`,
	// Ruby logger, "I, [2021-07-08T05:08:19.123456 #123] INFO -- : msg"
	`^[A-Z], \[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?`,
	// Go stack starts. Python/Java exception markers are often continuation
	// lines after a timestamped error header, so they are intentionally omitted.
	`^(panic:|fatal error:)`,
	// Java exception root line. Keep "Caused by:" out: it is normally a continuation.
	`^Exception in thread "[^"]+" `,
	// Fully-qualified Java/.NET exception root line, such as "java.lang.IllegalStateException:".
	`^([A-Za-z_$][A-Za-z0-9_$]*\.)+[A-Za-z_$][A-Za-z0-9_$]*(Exception|Error)(:|$)`,
	// AWS Lambda runtime markers.
	`^(START|END|REPORT) RequestId:`,
}

var GlobalSymbolPatterns = []string{
	// Promtail-style and Elasticsearch-style bracketed timestamps:
	// [2020-12-03 11:36:20], [2021-06-01T11:56:06,712]
	`^\[\d{4}-\d{2}-\d{2}[T ]\d{1,2}:\d{2}:\d{2}([,.]\d+)?(Z|[+-]\d{2}:?\d{2})?\]`,
	// Bracketed date starts used by Filebeat examples and Elasticsearch logs.
	`^\[\d{4}-\d{2}-\d{2}\]([ \t]|$)`,
	// (2020-12-03 11:36:20) and (2020-12-03T11:36:20Z)
	`^\(\d{4}-\d{2}-\d{2}[T ]\d{1,2}:\d{2}:\d{2}([,.]\d+)?(Z|[+-]\d{2}:?\d{2})?\)`,
	// Bracketed application/date prefixes, e.g. [beat-logstash-2015.11.28].
	`^\[[^]\s]+-\d{4}\.\d{2}\.\d{2}\]`,
	// [10/Oct/2000:13:55:36 -0700]
	`^\[\d{1,2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\]`,
	// [02-Jan-2006 15:04:05], common in PHP and application logs.
	`^\[\d{1,2}-[A-Za-z]{3}-\d{4} \d{2}:\d{2}:\d{2}([,.]\d+)?\]`,
	// [15:04:05.123] INFO ...
	`^\[\d{2}:\d{2}:\d{2}([,.]\d+)?\][ \t]+`,
	// Syslog RFC5424/RFC3164 with priority, "<34>1 2021-07-08T05:08:19Z ..."
	`^<\d{1,5}>(1 )?`,
	// Apache error log, "[Tue Jan 02 15:04:05.123456 2006] ..."
	`^\[[A-Za-z]{3} [A-Za-z]{3} +\d{1,2} \d{2}:\d{2}:\d{2}(\.\d+)? \d{4}\]`,
	// Apache/nginx access log with HTTPDATE after client/ident/auth fields.
	`^[0-9A-Fa-f:.]+ [^ ]+ [^ ]+ \[\d{1,2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2} [+-]\d{4}\]`,
	// Kernel ring buffer, "[12345.678901] message"
	`^\[\s*\d+\.\d+\]`,
	// Level-first logs, "INFO 2021-07-08 05:08:19 ..." or "[ERROR] 2021/07/08 ..."
	`^\[?(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\]?[: ]+\d{4}[-/]\d{1,2}[-/]\d{1,2}`,
	// Bracketed level starts, "[INFO] ..." and "(ERROR) ...".
	`^\[(TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\][ \t]+`,
	`^\((TRACE|DEBUG|INFO|NOTICE|WARN|WARNING|ERROR|ERR|CRITICAL|CRIT|FATAL|PANIC)\)[ \t]+`,
	// Short bracketed levels, "[I] ...", "[E] ...".
	`^\[[TDIWEF]\][ \t]+`,
	// gin log, [GIN] 2006/01/02 - 08:53:39
	`^\[GIN\] \d+/\d+/\d+ - \d+:\d+:\d+`,
	// RabbitMQ/Erlang reports, "=INFO REPORT==== 2-Jan-2024::03:04:05 ==="
	`^=(INFO|ERROR|WARNING|DEBUG|CRASH|SUPERVISOR) REPORT====`,
	// slow log, # Time: 2026-03-25T13:30:55.776770239Z
	`^# Time: \d{4}`,
	// JSON object/array format. Avoid treating generic bracket-prefixed stack lines
	// such as "[signal SIGSEGV...]" as a new multiline entry.
	`^\s*(\{|\[\s*[\{"])`,
}
