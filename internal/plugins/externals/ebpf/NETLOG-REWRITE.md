# Netlog Rewrite Plan

## Why

The current `netlog` hot path is still fundamentally packet-centric:

1. capture packets from NICs
2. sort packets by TCP sequence
3. infer retransmit / keepalive / direction
4. parse L7 protocol directly from packet payloads

That model has two problems:

- correctness gets harder as jitter, mild reordering, GRO/TSO, and retransmit overlap increase
- performance work increasingly fights the packet sorter instead of simplifying the pipeline

The rewrite direction is to move from a packet sorter to a passive stream reassembler.

## Non-goals

- do not reimplement the Linux TCP stack
- do not copy kernel congestion control, RTO, RACK, TLP, or socket receive queue logic
- do not require perfect loss recovery from passive observation

## What To Borrow From Linux

Borrow the model, not the implementation:

- track `next_expected_seq` per flow direction
- keep a bounded out-of-order queue
- use ACK progress as evidence for RTT / retransmit inference
- classify gaps as `suspected` rather than claiming exact packet loss
- prefer small, bounded state over unbounded buffering

## Target Architecture

### 1. Capture

Keep NIC capture lightweight.

- decode only L2/L3/L4 metadata
- normalize into a small packet event
- do not do expensive protocol parsing here

### 2. Flow Tracker

Maintain per-flow state keyed by:

- netns
- VNI when present
- src/dst IP
- src/dst port
- protocol / direction

Responsibilities:

- handshake state
- direction-specific TCP sequence progress
- ACK-driven RTT hints
- retransmit and overlap classification
- bounded buffering decisions

### 3. Stream Reassembler

Per direction, assemble only contiguous byte ranges for upper layers.

- in-order segments advance immediately
- small out-of-order segments are buffered
- duplicate / fully-covered segments count as retransmits
- large gaps or buffer overflow trigger resync with an explicit `gap` marker

Upper-layer parsers should consume contiguous stream chunks, not raw packets.

### 4. Protocol Parsers

- HTTP/1.x parser consumes stream bytes
- HTTP/2/gRPC parser consumes stream bytes or framed chunks from the reassembler output
- protocol logic no longer needs to know TCP packet ordering details

## Rollout Plan

### Phase 1

Add reusable core types without changing runtime behavior.

- bounded stream reassembler
- overlap/retransmit/gap classification
- focused tests and benchmarks

Status: implemented in `internal/l4log/stream_reassembly.go`

### Phase 2

Introduce a parallel flow tracker behind the existing `TCPLog` path.

- build packet events from capture loop
- feed both legacy logic and new tracker in tests/benchmarks
- compare output behavior on replayed traffic

Status: the reusable tracker core now exists in `internal/l4log/tcp_flow_tracker.go` and is used by `l7_http.go`, `l7_http2.go`, and the runtime retransmit classification path in `l4_tcp.go`

### Phase 3

Switch HTTP/1.x netlog generation to reassembled stream chunks.

- keep legacy fallback
- measure allocation and CPU impact on mixed traffic

Status: packet delivery for HTTP/1.x now goes through the stream reassembler before header parsing; HTTP/2 also uses the same tracker for contiguous frame input after connection preface detection. TCP connection state is now separate from the old packet sorter, runtime retransmit classification no longer depends on the legacy sorter path, and large continuity gaps now force a new TCP chunk using tracker semantics

### Phase 4

Move HTTP/2 and gRPC handling to the new flow/reassembly path.

## Guardrails

- bounded memory per flow and per direction
- no unbounded reorder queue
- explicit resync on large gaps
- copy payload only when buffering is required
- keep unit tests for overlap, retransmit, out-of-order drain, and gap recovery
