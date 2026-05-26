# Netlog Rewrite Performance Report

## Scope

This report captures the `netlog` hot-path performance after the staged rewrite work on April 9, 2026.

The measured areas are:

- NIC packet decode object reuse
- legacy TCP packet sorter versus the newer runtime path
- HTTP/1 request detection and lightweight parsing
- HTTP response status parsing
- the new bounded stream reassembler baseline

## Runtime Status

At the time of this report:

- packet decode reuse is active in `internal/l4log/capture.go`
- HTTP/1 header parsing is fed by the new stream reassembler in `internal/l4log/l7_http.go`
- HTTP/2 frame input is fed by the shared flow tracker in `internal/l4log/l7_http2.go`
- runtime retransmit classification in `internal/l4log/l4_tcp.go` uses the shared `tcpFlowTracker`
- large continuity gaps now force a new TCP chunk using tracker semantics
- the old sorter remains in `internal/l4log/l4_tcp_legacy_sort.go` for compatibility tests and benchmark comparison

## Verification

### Unit tests

```bash
go test ./internal/plugins/externals/ebpf/internal/l4log
```

### Focused benchmark command

```bash
go test -run '^$' -bench 'Benchmark(HTTPReqOrResp(Legacy|Fast)|ParseHTTPRequestMeta(Legacy|Fast)|ParseHTTPResponseStatus(Legacy|Fast)|DecodePacket(Legacy|ReuseDecoder)|TCPRetransInsert(LegacyOrdered|Ordered|LegacyReorder|Reorder)|StreamReassemblerPush(InOrder|Reorder))$' -benchmem ./internal/plugins/externals/ebpf/internal/l4log
```

## Benchmark Environment

- Date: `2026-04-09`
- Host CPU: `AMD Ryzen 7 9700X 8-Core Processor`
- OS: `linux/amd64`

## Results

### Packet Decode

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkDecodePacketLegacy` | `581.3` | `1928` | `34` |
| `BenchmarkDecodePacketReuseDecoder` | `28.61` | `0` | `0` |

Summary:

- decoder reuse is about `20.3x` faster on this microbench
- decode-path allocations were removed in the measured benchmark

### HTTP Request Detection

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkHTTPReqOrRespLegacy` | `34.53` | `80` | `1` |
| `BenchmarkHTTPReqOrRespFast` | `8.757` | `0` | `0` |

Summary:

- request detection is about `3.9x` faster
- per-call allocation is removed in the measured benchmark

### HTTP Request Parsing

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkParseHTTPRequestMetaLegacy` | `2021` | `6265` | `30` |
| `BenchmarkParseHTTPRequestMetaFast` | `204.0` | `208` | `6` |

Summary:

- request header parsing is about `9.9x` faster
- request parsing allocation dropped by about `96.7%`

### HTTP Response Parsing

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkParseHTTPResponseStatusLegacy` | `1432` | `5241` | `20` |
| `BenchmarkParseHTTPResponseStatusFast` | `10.23` | `0` | `0` |

Summary:

- response status parsing is about `140x` faster
- response parsing allocation is removed in the measured benchmark

### Legacy Sorter Comparison

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkTCPRetransInsertLegacyOrdered` | `5567` | `7480` | `138` |
| `BenchmarkTCPRetransInsertOrdered` | `2648` | `12368` | `131` |
| `BenchmarkTCPRetransInsertLegacyReorder` | `13989` | `7480` | `138` |
| `BenchmarkTCPRetransInsertReorder` | `12760` | `12368` | `131` |

Summary:

- the newer ordered path benchmark is about `2.1x` faster than the legacy ordered sorter
- the newer reorder path benchmark is about `1.1x` faster than the legacy reorder sorter
- these sorter comparison numbers remain useful as a compatibility reference, but the production runtime is no longer driven by the legacy sorter

### Stream Reassembler Baseline

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkStreamReassemblerPushInOrder` | `1024` | `1706` | `64` |
| `BenchmarkStreamReassemblerPushReorder` | `842.0` | `2565` | `41` |

Summary:

- this is the current baseline for the shared bounded reassembler
- it is now the main building block for HTTP/1, HTTP/2, and runtime retransmit classification

## Overall Readout

The biggest wins in this round came from:

1. reusing decode objects on the packet capture path
2. replacing heavy HTTP header parsing with lightweight parsing on contiguous stream input
3. moving runtime retransmit and flow ordering logic toward the shared tracker/reassembler model

The main remaining work is not basic parser speed anymore. The remaining opportunity is to keep migrating TCP chunk construction and event shaping away from the old packet-oriented chunk model so the runtime path depends even less on legacy structures.
