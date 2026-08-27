# WinNetFlow (Windows ETW netflow) collector

`winnetflow` collects per-process L4 network flow metrics on Windows by
consuming TCP/UDP events from the `Microsoft-Windows-TCPIP` ETW provider. The
core measurement schema (`netflow`) is compatible with the Linux eBPF netflow
collector, including normalized client/server endpoint roles and byte fields.

It also collects aggregated HTTP request metrics (`httpflow`, enabled by
default) from the `Microsoft-Windows-HttpService` ETW provider, covering IIS,
HttpListener and other HTTP.sys based servers.

## Architecture

```
ETW kernel provider (Microsoft-Windows-TCPIP)
   |  (kernel-side event ID + keyword filtering at enable time)
   v
ProcessTrace callback thread          [etw_windows.go: handleRecord]
   |  copies header + user data (pooled), pushes to channel
   v
1 ordered decode worker               [etw_windows.go: decodeLoop]
   |  fast-path: layout-based parsing (no syscalls) for known event versions
   |  fallback:  TDH property decoding for unknown versions / rare events
   v
flowAggregator                        [flow.go]
   |  per-(pid,tuple,direction) counters; lifecycle de-dup; burst replay
   v
Input.Run ticker -> point.Network points
```

Every point carries the merged tag set (host tags + `[inputs.winnetflow.tags]`),
injected via `point.WithExtraTags` at flush time — the same contract as other
DataKit inputs.

Session lifecycle:

1. `StartTraceW` (real-time mode, QPC clock). A stale session with the same
   name is stopped first (`ERROR_ALREADY_EXISTS` handling).
2. `EnableTraceEx2` with `EVENT_FILTER_TYPE_EVENT_ID` (only the events we
   consume) plus keyword filtering to keep volume proportional to traffic.
3. `OpenTraceW` + `ProcessTrace` on a dedicated goroutine; `CloseTrace` +
   `ControlTrace(STOP)` on shutdown.
4. Every ~10 minutes the input logs a self-telemetry summary (decoded /
   dropped / parse errors / session `EventsLost` / `RealTimeBuffersLost`), and
   `FeedLastError` if `ProcessTrace` exits unexpectedly.

The L7 httpflow pipeline mirrors this but keeps its own session
(`datakit-winnetflow-http`), its own single decode worker and its own
aggregator:

```
Microsoft-Windows-HttpService provider
   |  (event ID + keyword filter: 0x136, level 4)
   v
ProcessTrace callback thread        [http_etw_windows.go: handleRecord]
   |  copies header (ActivityID/RelatedActivityID + user data), pooled
   v
1 decode worker                     [http_etw_windows.go: decodeLoop]
   |  fast path for version 0 layouts (verified below)
   |  TDH fallback for unknown versions
   v
httpAggregator                      [httpflow.go]
   |  groups by (method,path,status,tuple,pid); truncates long paths
   v
Input.Run ticker -> point.Network points ("httpflow")
```

The single decode worker is a correctness requirement, not a performance
choice: HTTP correlation is strictly order-sensitive (connection before
request, request before its parse/response/terminal events). Parallel
consumers of the raw channel let a later event overtake an earlier one and
produce missed correlations under load (observed as `missed_conn`/`missed_req`
in the failure stats).

## Supported events and layouts

Provider GUID: `{2f07e2ee-15db-40f1-90ef-9d7ba282188a}`.

| Event | Meaning | Fast-path versions | TDH fallback |
| --- | --- | --- | --- |
| 1017 | TCP accept complete (server side, PID) | - | always |
| 1033 | TCP connect complete (client side, PID) | - | always |
| 1034/1045 | connect failure / timeout | - | always |
| 1040/1043 | abort / disconnect (tuple + PID) | - | always |
| 1184 | connection terminated by RST | - | always |
| 1300 | connection rundown (existing conns at enable) | - | always |
| 1051 | TCP state change | v0 | others |
| 1073 | TCP data send (legacy) | v0 | others |
| 1332 | TCP data send (current) | v0-v5 | others |
| 1074 | TCP data receive (NumPkt on v1+) | v0-v1 | others |
| 1341 | TCP RTT sample | v0 | others |
| 1351 | TCP retransmit round (RexmitCount) | v0-v1 | others |
| 1169/1170 | UDP send/receive (messages+bytes+PID) | v0-v1 | others |

Fast-path decoders read fixed offsets from the template layouts (verified
against the OS manifest, see below) and are pinned by synthetic-data unit tests
(`TestFastDecode*`). Unknown versions automatically fall back to TDH so a
future OS layout change degrades to correct-but-slower instead of misparsing.

## How the event layouts were extracted

The authoritative field names/order come from the installed OS manifest, not
from memory:

```powershell
# Event IDs, keywords and message placeholders:
wevtutil gp Microsoft-Windows-TCPIP /ge:true /gm:true > tcpip_manifest.xml

# Field-level templates (property names, inTypes, lengths):
$pm = New-Object System.Diagnostics.Eventing.Reader.ProviderMetadata("Microsoft-Windows-TCPIP")
$pm.Events | Where-Object { $_.Id -in 1017,1033,1043,1074,1169,1170,1300,1332 } | ForEach-Object { $_.Template }
```

When adding support for a new Windows version:

1. Dump the manifest on that version and diff the templates of the events
   above against the current `u32At`/`u64At` offsets in `etw_windows.go`.
2. If offsets changed, either extend the fast path for the new version (update
   the version gate + add a synthetic-data test) or leave it to the TDH
   fallback (correct by construction).
3. Re-run the matrix script on that OS (below).

## HTTP.sys (L7) events and correlation

Provider GUID: `{dd5ef90a-6398-47a4-ad34-4dcecdef795f}`. The collector enables
keywords `0x136` (request 0x002 | response 0x004 | connection 0x010 | cache
0x020 | request queue 0x100) at level 4, with an event ID filter that keeps
only the events below.

| Event | Meaning | Fast-path versions | TDH fallback |
| --- | --- | --- | --- |
| 21 | new connection (local/remote endpoints) | v0 | others |
| 24 | connection cleanup (always last) | v0 | others |
| 1 | request received; ActivityID = connection, RelatedActivityID = request | v0 | others |
| 2 | verb (HTTP_VERB enum) + UTF-16 URL | v0 | others |
| 3 | deliver: site id, request queue (app pool), URL | v0 | others |
| 4 | response received (status + ASCII verb) | v0 | others |
| 8 | fast response (status + ASCII verb) | v0 | others |
| 10 | send complete (terminal) | v0 | others |
| 11 | cached and send (terminal) | v0 | others |
| 12 | fast send (terminal) | v0 | others |
| 16 | served from cache (terminal, BytesSent) | v0 | others |

Correlation keys:

- Event 21 registers a connection under its ActivityID.
- Event 1 carries the connection as ActivityID and the request as
  `EVENT_HEADER_EXT_TYPE_RELATED_ACTIVITYID` (0x0001) extended data. The
  request is registered under the RelatedActivityID.
- Events 2, 3, 4, 8, 10, 11, 12 and 16 are request-scoped: their ActivityID is
  the request GUID.
- A request completes on the first terminal event (10/11/12/16) and is emitted
  with the connection tuple + process name; event 24 drops any still-pending
  requests on that connection.

Verified layouts (Windows 10.0.26200; offsets relative to UserData):

| Event | Layout |
| --- | --- |
| 21 | 0 ConnectionObj u64; 8 LocalAddrLength u32; 12 sockaddr; +len RemoteAddrLength u32; sockaddr |
| 1 | 0 RequestId u64; 8 ConnectionId u64; 16 RemoteAddrLength u32; 20 sockaddr_in (IPv4) |
| 2 | 0 RequestObj u64; 8 HttpVerb u32; 12 URL UTF-16 null-terminated |
| 3 | 0 RequestObj u64; 8 RequestId u64; 16 SiteId u32; 20 QueueName UTF-16; URL UTF-16; Status u32 |
| 4/8 | 0 RequestId u64; 8 ConnectionId u64; 16 StatusCode u16; 18 ASCII verb; trailing fields |
| 10/11/12 | 0 RequestId u64; 8 HttpStatus u16 |
| 16 | 0 RequestObj u64; 8 SiteId u32; 12 BytesSent u32 |
| 24 | 0 ConnectionObj u64 |

Event 1 layouts differ across OS generations (Server <= 2019 keeps a 32-bit
address length, newer versions moved to 64-bit with an extra padding field);
the fast path covers the version seen on 10.0.26200 and unknown versions
degrade to TDH. The `EVENT_HEADER_EXTENDED_DATA_ITEM` layout
(`Reserved1/ExtType/Linkage/DataSize/DataPtr`, 16 bytes) was verified against
`evntcons.h` and a live capture.

The path is normalized via `net/url.Parse`; scheme, host, and query string are
excluded to avoid leaking secrets and creating unbounded metric cardinality.
Paths longer than `httpflow_path_limit` are truncated on a valid UTF-8 boundary
and flagged `truncated`.
The `httpflow` measurement uses `point.CommonLoggingOptions` because
`DefaultMetricOptions` drops string fields.

The TDH fallback decoders are exercised against real events in the smoke suite
(`TestHTTPTDHPathWithRealEvents` / `TestL4TDHPathWithRealEvents`): raw events
captured on the current OS are replayed through every TDH decoder, proving the
property names resolve against a real manifest (HttpVerb/Url/SiteId/
RequestQueueName/StatusCode/Verb/HttpStatus/BytesSent/LocalAddr/RemoteAddr for
HttpService; Tcb/BytesSent/NumBytes/NumPkt/SRtt/RttVar/RexmitCount/OldState/
NewState/NumMessages/Pid/LocalSockAddr/RemoteSockAddr for TCPIP). This is the
only local coverage for the cross-OS safety net, because on the reference OS
every event version is handled by a fast path and TDH never runs in
production.

## Verification (production gate)

```powershell
# On each target OS, elevated shell, inside the repo checkout:
powershell -ExecutionPolicy Bypass -File internal/plugins/inputs/winnetflow/verify/verify-winnetflow.ps1
```

The script runs: unit tests -> real-ETW session lifecycle -> loopback smoke ->
200-connection stress (byte accuracy + event drops) -> L7 HTTP.sys session
lifecycle + HttpListener traffic smoke -> Input.Run end-to-end (periodic feed,
production-default loopback filtering, and shutdown final flush) -> 60s soak
with non-empty L4/L7 output -> session cleanup, and optionally `go test -race`
when a C toolchain is available. The soak can be extended with
`DK_WINNETFLOW_SOAK_SECONDS=600` for a longer run. Record the printed
`MATRIX: <OS> <arch>` line plus the stress ratio for the version matrix.

Known matrix entries:

| OS | Arch | Result |
| --- | --- | --- |
| Windows 10.0.26200 (Win11/Server 2025) | amd64 | 100.09% byte accuracy, 0 drops (200 conns, 100 MB x 2 directions) |
| Windows Server 2022 (10.0.20348) | amd64 | 99.97% byte accuracy, 0 drops (200 conns, 100 MB x 2); full suite incl. real-event TDH replay (L4 + HTTP.sys) green |

Pending: Server 2016 / 2019 and Win10 21H2+ entries.

Verification note for slow VMs (QEMU TCG): the HTTP TDH replay test generates
traffic via a .NET HttpListener; under software emulation the listener can
exceed the 15 s startup grace period in `gen-httpsvc-traffic.ps1` (raise it)
and .NET's thread pool can throw an AccessViolationException that is
environmental, not a winnetflow defect; retry the test in that case.

TCG note for Server 2016/2019: under QEMU software emulation (no WHPX/Hyper-V
on this host), the first boot of the applied eval images grinds for hours and
the unattended provisioning (FirstLogonCommands) did not complete in repeated
attempts, leaving no SSH/WinRM-Basic channel. The same answer file completes
in ~1 h on Server 2022 under TCG. If OpenSSH.Server cannot be installed via
Add-WindowsCapability on 2016/2019 eval media, ship `OpenSSH-Win64.zip` on the
answer-file media and run `install-sshd.ps1` from the first-logon command
(see the v6 autounattend variant used for the VM matrix runs).

## Tuning

Config (`[[inputs.winnetflow]]`):

| Option | Default | Meaning |
| --- | --- | --- |
| `interval` | `60s` | flow aggregation / report interval (minimum `60s`, matching the Linux eBPF collector) |
| `etw_buffer_size_kb` | `64` | ETW session buffer size in KB (clamped to 4-1024) |
| `etw_min_buffers` | `8` | minimum ETW buffers (clamped to 2-4096) |
| `etw_max_buffers` | `256` | maximum ETW buffers; clamped to 2-4096 and a 256 MiB/session budget |
| `max_flows` | `65536` | max flows tracked per interval; excess new flows are dropped and counted (`flows_skipped`) |
| `enable_httpflow` | `true` | collect L7 HTTP request metrics from the HTTP.sys provider |
| `max_http_requests` | `65536` | max in-flight HTTP requests tracked; excess requests are dropped and counted in the periodic summary |
| `httpflow_path_limit` | `256` | max request path length; longer paths are truncated and flagged `truncated` |

Env equivalents: `ENV_INPUT_WINNETFLOW_INTERVAL`,
`ENV_INPUT_WINNETFLOW_ETW_BUFFER_SIZE_KB`,
`ENV_INPUT_WINNETFLOW_ETW_MIN_BUFFERS`,
`ENV_INPUT_WINNETFLOW_ETW_MAX_BUFFERS`, `ENV_INPUT_WINNETFLOW_MAX_FLOWS`,
`ENV_INPUT_WINNETFLOW_ENABLE_HTTPFLOW`, `ENV_INPUT_WINNETFLOW_MAX_HTTP_REQUESTS`,
`ENV_INPUT_WINNETFLOW_HTTPFLOW_PATH_LIMIT`, `ENV_INPUT_WINNETFLOW_TAGS`.

## Known limitations

- TCP send events carry no packet count, so `packets_written` is only populated
  for UDP (message count as an approximation).
- UDP direction is heuristic: bound sockets on non-ephemeral ports (< 49152)
  are treated as `incoming`, everything else `outgoing`.
- Existing TCP connections and rare inline tuple recovery use a TCP listener
  snapshot for direction; the snapshot refreshes every 30 seconds.
- Windows ETW has no netns/Kubernetes/NAT metadata. Core endpoint-role tags are
  compatible; NAT tags are `N/A` and platform-specific Linux metadata is not
  synthesized.
- `tcp_close_wait` / `tcp_last_ack` / `tcp_time_wait` are TCP state *transition
  counts* per interval (schema-compatible with ebpf-net/netflow), not durations.
- Under extreme load ETW can drop events; counters are exposed in the periodic
  summary and `FeedLastError` fires if the session dies.
- Pure loopback flows and invalid/unspecified endpoints are not reported. The
  periodic summary exposes them as `loopback_filtered` and `invalid_filtered`.
- Requires administrator rights for real-time ETW sessions.
- Per-TCP-endpoint flows: a connection produces one point per socket endpoint.
  Client ports in the eBPF-compatible rollup range (32768-65535) are normalized to
  `*` before interval output aggregation; server ports remain exact.
- The fast path covers events listed above; anything else (new event IDs or
  unknown versions) is dropped or TDH-decoded and counted as parse errors.
- L4 uses 4 parallel decode workers to keep up with data-event volume; the
  pending-replay design absorbs reordered data events (they are buffered by
  TCB until the tuple is known), but in a rare scheduling race a data event
  can be processed after the TCB close deleted the tuple and its bytes are
  then dropped. The impact is bounded to the last few events of a connection
  and is visible in the parse-error/dropped telemetry; modern TCPIP versions
  (1332 v5+) additionally self-heal via inline addresses.
- `httpflow` only sees HTTP traffic that goes through HTTP.sys; self-contained
  socket servers are invisible to it.
- HTTP.sys events carry no HTTP version or request body size: `http_version`
  is always empty and `bytes_read` always 0. `bytes_written` is only available
  on cache-served responses (event 16).
- Request latency is measured from event 1 (receive) to the terminal event, so
  it excludes the client-to-server network path and queueing before HTTP.sys
  accepts the request.
