# bpf-netlog host veth peer capture plan

Date: 2026-04-16 20:13:59 CST

## Step 1. Current behavior review

- `bpf-netlog` currently builds capture state as `netns -> NIC -> raw socket -> TCPConns`.
- For container net namespaces, each discovered non-loopback NIC can allocate its own `TPacket` ring and `TCPConns` state.
- Host namespace capture intentionally skips most virtual NICs, so host-side veth peers are not used for container capture.
- Direction and metric aggregation rely on the capture interface looking like the local endpoint. The code currently approximates that with `MAC == interface MAC`.

Relevant files:

- `internal/plugins/externals/ebpf/internal/l4log/nic_monitor.go`
- `internal/plugins/externals/ebpf/internal/l4log/net_ns.go`
- `internal/plugins/externals/ebpf/internal/l4log/capture.go`

## Step 2. Design decision

Implement a minimal, traceable optimization:

1. Detect when a container NIC can be mapped to a host-side peer.
2. Prefer opening the raw socket on the host peer for that NIC.
3. Keep the logical ownership on the container:
   - keep container `nsUID`
   - keep container `containerID`
   - keep container K8s tags
   - keep container NIC IP list for metric ownership
4. Fall back to the original container-netns capture path when peer mapping is unavailable or unsafe.

This change does not attempt a full CNI classifier. Instead it uses a conservative rule:

- container NIC is eligible only when:
  - `iflink != ifindex`
  - a host NIC with `ifindex == iflink` exists
  - the resolved host NIC is virtual
  - the resolved host NIC is not loopback

This intentionally targets common `veth`/`cali*` style peers and avoids force-routing `ipvlan/macvlan/SR-IOV/unknown` through the host-peer path.

## Step 3. Implementation plan

1. Extend NIC metadata with traceable capture attributes:
   - `IfIndex`
   - `IfLink`
   - `IsLoopback`
   - `IsVirtual`
   - capture target fields
2. Build a host NIC index that includes virtual interfaces for peer resolution only.
3. Compute a capture plan per container NIC:
   - `host-peer` when the conservative rule matches
   - `in-netns` otherwise
4. Open sockets according to the selected capture namespace/interface.
5. Mark host-peer captures as trusted local interfaces so direction/aggregation does not depend on runtime MAC equality.

## Step 4. Validation target

- `go test ./internal/plugins/externals/ebpf/internal/l4log`
- inspect logs to confirm `host-peer` plans are selected only for eligible NICs
- verify fallback still works when peer resolution fails

## Step 5. Implementation log

- 2026-04-16 20:13:59 CST: Recorded current behavior and selected the conservative host-peer design.
- 2026-04-16 20:13:59 CST: Implemented NIC metadata enrichment, conservative host-peer planning, and trusted-local fallback handling.
- 2026-04-16 20:13:59 CST: Corrected socket-open scoping so host-peer plans open sockets in the host namespace instead of the container namespace.

## Step 6. Validation log

- `env GOCACHE=/tmp/datakit-go-cache go test ./internal/plugins/externals/ebpf/internal/l4log`
  - ran without sudo first
  - package build succeeded, but existing `TestPortListen` failed in sandbox with `route ip+net: netlinkrib: operation not permitted`
- `sudo ... /home/vircoys/.g/go/bin/go test ./internal/plugins/externals/ebpf/internal/l4log -run "TestResolveCapturePlan|TestDetectProto|TestMatch|TestRetrans"`
  - passed
- `sudo ... /home/vircoys/.g/go/bin/go test ./internal/plugins/externals/ebpf/internal/l4log`
  - passed

Validation conclusion:

- the new host-peer planning logic compiles and passes package tests
- the non-sudo failure was caused by an existing privileged network inspection test, not by the host-peer change

## Step 7. Next incremental change

Add explicit capture-plan decision logging so runtime behavior is traceable without attaching a debugger.

Target:

- log when a NIC is resolved to `host-peer`
- log when a NIC falls back to `in-netns`
- log enough identifiers to correlate decisions with container netns and interface names

## Step 8. Next incremental change

Add host-peer conflict handling.

Problem:

- host-peer planning is conservative, but if multiple container NICs ever resolve to the same host-side interface, the current per-namespace loop can still attempt duplicate capture ownership.

Decision:

- iterate namespaces in a stable order
- keep the first host-peer claimant for a given host interface
- log a warning for later claimants
- force later claimants back to `in-netns`

This keeps runtime behavior deterministic and traceable.

## Step 9. Next incremental change

Add in-memory capture-plan counters and periodic summary logging.

Goal:

- keep decision-level logs for each NIC
- also provide coarse runtime totals so operators can see whether host-peer optimization is actually taking effect

Initial counters:

- `host_peer_selected`
- `fallback_in_netns`
- `fallback_conflict`
- `host_namespace_direct`

Reporting:

- log a summary periodically from the monitor loop
- keep the implementation local to `l4log` without introducing external telemetry dependencies

## Step 10. Next incremental change

Normalize capture-plan reasons into stable reason codes.

Goal:

- avoid free-form fallback reasons drifting over time
- make per-NIC decision logs and periodic summaries align on the same reason vocabulary

Initial reason codes:

- `host_namespace_or_missing_nic`
- `iflink_unavailable`
- `host_peer_not_found`
- `host_peer_not_eligible`
- `host_peer_selected`
- `host_peer_conflict`

## Step 12. Next incremental change

Replace unbounded decision counters with:

- current snapshot counters
- per-period delta counters

Reason:

- the monitor loop rescans interfaces every 20 seconds
- simple cumulative counters drift upward forever and are less useful for runtime diagnosis

New reporting shape:

- `snapshot_*`: current effective capture-plan distribution
- `delta_*`: plan changes observed since the last summary tick

## Step 13. Next incremental change

Reduce steady-state allocations for namespace NIC IP lists.

Problem:

- every monitor iteration rebuilds `nicIP []string` from interface addresses
- in stable environments this repeats identical work and allocations

Decision:

- cache namespace NIC IP lists on the retained `netnsInformation`
- rebuild only when the effective NIC/address fingerprint changes

## Step 14. Next incremental change

Stabilize host-peer ownership across monitor iterations.

Problem:

- without a persistent owner cache, host-peer ownership is effectively re-elected every cycle
- stable namespace ordering already helps, but ownership is still derived fresh each time

Decision:

- persist host-peer owners in the monitor state
- release owners when the owning namespace disappears
- prefer the cached owner when the same namespace still claims the interface

## Step 11. Next incremental change

Prioritize stability and lower steady-state overhead.

Changes:

- only emit per-NIC capture-plan logs when the effective plan changes
- collapse host NIC discovery into a single scan per monitor iteration, then reuse it for:
  - host-peer resolution
  - host namespace capture NIC selection

Expected effect:

- much less repetitive logging during stable runtime
- one less host netns NIC enumeration per monitor cycle

## Step 15. Next incremental change

Reduce formatting overhead in steady-state plan tracking.

Problem:

- `capturePlanKey()` and `capturePlanFingerprint()` run in the monitor hot path
- both were using `fmt.Sprintf`, which adds avoidable formatting overhead and small allocations

Decision:

- replace `fmt.Sprintf` with lightweight string building and integer/bool append helpers

## Step 16. Next incremental change

Stop holding container netns file descriptors across monitor iterations.

Problem:

- an open netns file descriptor does not block container process exit
- but it can keep the network namespace object alive longer than necessary
- current monitor state stores container netns handles persistently

Decision:

- keep a persistent handle only for the host namespace
- for container namespaces:
  - record `nsUID`
  - record container pids
  - close the discovered netns fd immediately
  - reopen a temporary netns handle from a live pid only when needed for:
    - NIC enumeration
    - raw socket creation in container netns

## Step 17. Next incremental change

Add a short-lived NIC enumeration cache for container namespaces.

Problem:

- after removing persistent container netns handles, stable containers still pay the full
  `GetFromPid + setns + net.Interfaces()` cost every monitor cycle

Decision:

- cache enumerated container NICs on `netnsInformation`
- attach a short TTL so the cache is refreshed periodically
- continue reopening netns only when the cache is stale

## Step 18. Next incremental change

Prune stale pids from namespace ownership state.

Problem:

- container namespaces are tracked by a pid set
- exited pids can remain in that set until the namespace disappears from discovery
- temporary netns reopen may repeatedly try dead pids first

Decision:

- prune dead pids when selecting a pid for temporary namespace access
- keep at least one currently alive pid when available

## Step 19. Next incremental change

Add short failure backoff for temporary container netns access.

Problem:

- if all tracked pids in a namespace are stale, reopen attempts can fail repeatedly
- repeated failures cause repeated `/proc` probing and noisy logs every monitor cycle

Decision:

- cache the last temporary netns access error timestamp per namespace
- skip repeated reopen attempts during a short backoff window
- keep the behavior local to container namespaces only

## Step 20. Next incremental change

Make container namespace handles explicitly non-persistent.

Problem:

- container namespace tracking still carries a `netnsHandle` struct
- after earlier cleanup changes, that struct no longer needs a live fd for containers
- keeping a closed fd value inside the struct is confusing and risks future misuse

Decision:

- keep a real persistent handle only for the host namespace
- explicitly mark container namespace handles as invalid after discovery
- ensure close paths only close open handles

## Step 21. Next incremental change

Add a shared host-side packet ring for `host-peer` capture plans.

Problem:

- after moving `veth` containers onto host-peer capture plans, every selected host peer still
  opened its own `TPacket` ring
- that preserves interface isolation, but it keeps the number of host-side sockets and rings
  proportional to the number of selected container peers
- the main optimization target has shifted from container netns overhead to host-side ring count

Decision:

- keep host physical interfaces on their existing per-interface raw sockets
- collapse only `host-peer` capture plans onto one shared host namespace `TPacket`
- route packets in user space by `CaptureInfo.InterfaceIndex`
- keep container ownership on the existing `TCPConns` objects, so shared host capture only changes
  the packet ingress path, not the logical aggregation owner
- record the selected host peer `ifindex` on each capture plan for traceability

Implementation notes:

- extract per-packet decode/update logic from `TCPConns.CapturePacket()` into a reusable helper
- introduce `hostPeerSharedCapture` to own the single shared `TPacket`
- build an `ifindex -> TCPConns` route table from active `host-peer` interface plans
- do not start dedicated `CapturePacket()` goroutines for `host-peer` connections
- continue running `Gather()` and port-listen watchers per logical container owner

Validation:

- unit tests still pass for the `l4log` package subset used in earlier iterations
- added route-table coverage with `TestBuildHostPeerSharedRoutes`

## Step 22. Next incremental change

Add kernel-side interface filtering for the shared host-peer socket.

Problem:

- the shared host-peer socket reduced the number of host-side rings
- but because it was opened without binding to a specific interface, it could still receive
  packets from unrelated host interfaces
- user space routing dropped those packets later, but they still consumed ring space and reader CPU

Decision:

- add a host-peer interface whitelist at the socket filter layer
- prefer an eBPF `BPF_PROG_TYPE_SOCKET_FILTER` program on kernels that support it
- fall back to classic BPF (`SO_ATTACH_FILTER`) on older kernels or if eBPF attachment fails
- keep the whitelist semantics the same on both paths: only packets whose `ifindex` belongs to
  the active shared host-peer route table are accepted

Implementation notes:

- shared socket route changes now trigger filter resync
- eBPF path reads `__sk_buff.ifindex`
- classic BPF path uses `bpf.ExtInterfaceIndex`
- the current filter scope is interface whitelist only; protocol filtering for the shared socket
  is still handled after packet delivery

Validation:

- added unit tests for ifindex normalization
- added unit tests for classic BPF whitelist generation
- added unit tests for eBPF socket filter spec generation

## Step 23. Next incremental change

Combine interface whitelist and protocol filter on the shared host-peer socket.

Problem:

- Step 22 blocked unrelated physical interfaces from entering the shared socket
- but accepted packets from the whitelisted host-peer interfaces still reached user space even when
  they were outside the existing `tcp or udp port 8472 or udp port 4789` capture scope
- that left avoidable packet delivery overhead inside the shared ring path

Decision:

- keep the interface whitelist from Step 22
- combine it with the existing protocol filter semantics on both kernel-side filter paths:
  - eBPF socket filter
  - classic BPF fallback
- preserve the old protocol behavior instead of inventing a new narrower rule set

Implementation notes:

- classic BPF now does:
  - load interface index
  - check whitelist
  - run the pre-existing protocol filter program
- eBPF now does:
  - read `__sk_buff.ifindex`
  - check whitelist
  - parse Ethernet / IPv4 / IPv6 headers using socket-filter-safe load instructions
  - accept TCP plus VXLAN UDP ports `8472` and `4789`
- shared socket resync still re-attaches the filter whenever the active host-peer route set changes

Validation:

- updated unit tests for combined classic BPF generation
- updated unit tests for combined eBPF program generation

## Step 24. Next incremental change

Surface shared socket filter mode and active route fingerprint in logs.

Problem:

- after adding dual-path kernel filters, the runtime can select either:
  - eBPF socket filter
  - classic BPF fallback
- without explicit logging, it is hard to verify which path is active on a given host
  and which host-peer ifindexes are currently installed into the shared filter

Decision:

- add a stable shared-socket summary string
- include:
  - namespace id
  - active filter mode
  - route count
  - installed ifindex fingerprint
- emit that summary in:
  - shared socket periodic stats logs
  - monitor periodic summary logs

Validation:

- added unit test coverage for the shared summary formatter

## Step 25. Next incremental change

Track shared filter attach outcomes in runtime summaries.

Problem:

- Step 24 exposed the current shared filter mode and active ifindex fingerprint
- but it still did not show whether the process:
  - consistently succeeds with eBPF
  - repeatedly falls back to classic BPF
  - occasionally fails both paths and retries later

Decision:

- add attach outcome counters to the shared socket runtime state
- include:
  - total filter sync count
  - eBPF success count
  - classic BPF success count
  - eBPF failure count
- expose those counters through the existing shared socket summary string

Validation:

- updated unit tests for shared summary output to cover attach counters

## Step 26. Runtime validation on current host

Validate the shared host-peer filter path with root privileges on the current machine.

Observed environment:

- kernel: `Linux 6.8.0-49-generic`
- `CONFIG_BPF=y`
- `CONFIG_BPF_SYSCALL=y`
- `CONFIG_BPF_JIT=y`
- `CONFIG_BPF_JIT_ALWAYS_ON=y`

Observed result:

- privileged runtime attach succeeds overall because the implementation falls back to classic BPF
- the current eBPF socket-filter program is still rejected by the verifier on this host
- the verifier reports:
  `invalid bpf_context access off=76 size=4`
- the runtime therefore selects `mode=cbpf`

Implication:

- feature probing alone is not enough to conclude that the current eBPF socket-filter implementation
  is accepted by the verifier on all supported kernels
- the fallback path is not theoretical; it is the active path on this host today

## Step 27. Replace hand-written socket-filter asm with C eBPF plus map updates

Replace the shared host-peer socket filter implementation with a compiled C eBPF program and
an ifindex whitelist map.

Problem:

- the hand-written Go asm socket filter from Steps 22-26 still failed verifier checks on the current
  `Linux 6.8.0-49-generic` host
- shared host-peer route changes forced a filter re-attach on every ifindex-set change
- the verifier-sensitive packet parsing logic was harder to maintain in hand-written Go asm

Decision:

- replace the hand-written eBPF program with a C socket filter at:
  `internal/c/l4log/host_peer_filter.c`
- keep the classic BPF path as the old-kernel / load-failure fallback
- use a BPF hash map named `bpfmap_host_peer_ifindex` as the active host-peer ifindex whitelist
- attach the eBPF socket filter once and update only the whitelist map when the route set changes
- preserve the existing packet scope:
  - TCP
  - VXLAN UDP ports `8472` and `4789`

Implementation notes:

- added `host_peer_filter.o` build support in `internal/plugins/externals/ebpf/Makefile`
- added bindata accessors for `HostPeerFilterBin()` on linux amd64 and arm64
- replaced `filter_socket_linux.go` with a runtime-backed implementation that:
  - loads the compiled C object through the existing `bpfutil.Runtime`
  - attaches `socket__host_peer_filter` to the shared host-peer socket
  - looks up `bpfmap_host_peer_ifindex`
  - updates only map entries during later route changes
  - falls back to `cbpf` if C eBPF load or attach fails
- updated `hostPeerSharedCapture` to own and close the new shared filter runtime cleanly

Validation:

- built `internal/c/elf/linux_amd64/host_peer_filter.o` locally
- updated unit tests:
  - removed the old hand-written asm program spec assertion
  - added a stable fingerprint test for shared host-peer filter state
- non-root unit tests passed:
  `go test ./internal/plugins/externals/ebpf/internal/l4log -run 'TestResolveCapturePlan|TestFallbackCapturePlan|TestCapturePlanStatsSummary|TestBuildHostPeerSharedRoutes|TestNormalizeIfIndexes|TestNewSharedHostPeerCBPFFilter|TestSharedHostPeerFilterFingerprint|TestHostPeerSharedSummary|TestBuildNICIPsAndFingerprint|TestAnyPID|TestAnyPIDPrunesInvalidEntries|TestCloneNICInfos|TestContainerNICCacheTTLConstant|TestContainerNetNSErrorBackoffConstant|TestTempNetNSErrorBackoffHelpers|TestNetnsHandleCloseWithClosedFD|TestDetectProto|TestMatch|TestRetrans'`
- privileged runtime validation now succeeds with eBPF on this host:
  `sudo /home/vircoys/.g/go/bin/go test ./internal/plugins/externals/ebpf/internal/l4log -run 'TestAttachSharedHostPeerSocketFilterRuntime' -v`
  observed result:
  `runtime attach result: mode=ebpf ebpf_failed=false`

Open point:

- local arm64 object generation for `host_peer_filter.o` was not completed on this amd64 machine because
  the installed arm64 kernel header set is incomplete for cross-builds
- arm64 runtime therefore still needs either:
  - a proper arm64 build environment to produce `elf/linux_arm64/host_peer_filter.o`
  - or continued `cbpf` fallback on that platform until the object is generated

## Step 28. Tighten namespace runtime metadata and listener/watcher stability

Synchronize namespace runtime metadata across monitor rounds and remove stale pid snapshots from the
port-listen watcher path.

Problem:

- `CmpAndAddNIC()` reused the previous namespace object once a namespace already existed
- but the fresh per-round namespace object carried the new:
  - `pid` set
  - `container_id`
  - k8s tags
- those values were not copied back into the reused namespace state
- as a result:
  - the port-listen watcher could continue scanning an old pid snapshot
  - long-lived namespaces could drift away from current runtime metadata
  - if the watcher exited on `ctx.Done()`, the `portListenRunner` flag stayed set and prevented restart

Decision:

- add explicit namespace runtime metadata sync on every reused namespace
- store and read pid sets through helpers instead of direct map access
- make the port-listen watcher ask for the latest pid snapshot on every polling cycle
- ensure the watcher always clears its runner flag on exit
- avoid launching duplicate watcher goroutines when one is already active

Implementation notes:

- added `snapshotPIDs()`, `replacePIDs()`, and `syncRuntimeMetadata()` on `netnsInformation`
- switched container netns reopen helpers and watcher startup to use fresh pid snapshots
- added `portListenWatching()` to `netnsHandle`
- changed `tcpPortListenWatcher()` to accept a pid-provider callback instead of a frozen pid map
- added `defer atomic.StoreInt64(..., 0)` so the watcher can restart cleanly after exit

Validation:

- added unit coverage for namespace runtime metadata sync
- updated targeted `l4log` tests passed successfully

## Step 29. Add short TTL cache for host namespace NIC inventory

Cache host namespace NIC enumeration just like container NIC enumeration.

Problem:

- `buildHostNICInventory()` still enumerated host namespace NICs on every monitor pass
- after earlier container-side caching, host namespace enumeration became one of the remaining
  steady-state repeated scans
- that path rebuilds:
  - peer lookup map
  - capture NIC slice
  on every 20-second monitor loop

Decision:

- add a short TTL cache for host namespace NIC inventory
- cache:
  - host peer map by `ifindex`
  - host capture NIC slice
- keep the cache local to the host namespace object and reuse it until TTL expiry

Implementation notes:

- introduced `hostNICCacheTTL = 15s`
- stored cached host peer map and capture NIC slice on `netnsInformation`
- added `cloneNICInfoMap()` for safe cached reuse

Validation:

- added unit coverage for `hostNICCacheTTL`
- updated targeted `l4log` tests passed successfully

## Step 30. Replace linear port-listen scans with indexed lookup

Turn TCP listen-direction checks into indexed lookups instead of repeated linear scans.

Problem:

- `portListen.Query()` previously scanned the full per-namespace listen slice on every direction probe
- that cost grows linearly with the number of listening addresses and ports in the namespace
- in steady state, this path is exercised on hot packet-processing flows whenever connection direction
  is still unknown

Decision:

- keep the same external behavior
- replace the internal slice scan with per-namespace indexes keyed by:
  - wildcard listen port
  - IPv6-only wildcard listen port
  - exact `port -> ip` mapping

Implementation notes:

- replaced `map[string][]*tcpPortInf` with `map[string]*nsPortListenIndex`
- `Update()` now builds the namespace index once per watcher refresh
- `Query()` now checks:
  - wildcard port map
  - IPv6 wildcard port map
  - exact `port -> ip` map
- preserved the existing `macEQ` gate and IPv6 wildcard behavior

Validation:

- added `TestPortListenIndexQuery`
- updated targeted `l4log` tests passed successfully
