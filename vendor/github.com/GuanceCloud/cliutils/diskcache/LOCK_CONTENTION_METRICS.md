# Lock Contention Metrics Implementation

This document explains how lock contention metrics are measured and used in the diskcache module.

## Overview

The enhanced diskcache now tracks lock contention across three critical lock types:
- **write lock** (`wlock`): Excludes concurrent Put operations
- **read lock** (`rlock`): Excludes concurrent Get operations  
- **rw lock** (`rwlock`): Excludes structural operations (rotate, switch, drop, close)

## Instrumentation Strategy

### 1. Instrumented Locks

Instead of using `sync.Mutex` and `sync.RWMutex`, we now use:

```go
type InstrumentedMutex struct {
    mu          sync.Mutex
    lockType     LockType
    path         string
    lockWaitTime *prometheus.HistogramVec
    contention    *prometheus.CounterVec
}

type InstrumentedRWMutex struct {
    mu           sync.RWMutex
    path         string
    lockWaitTime *prometheus.HistogramVec
    contention    *prometheus.CounterVec
}
```

### 2. Contention Detection

Contention is detected using `TryLock()`:

```go
func (im *InstrumentedMutex) Lock() {
    start := time.Now()
    
    // Check if mutex is already locked (contention)
    if im.mu.TryLock() {
        // No contention - immediate acquisition
        im.observeLockTime(start, false)
        return
    }
    
    // Contention occurred - wait for lock
    im.observeContention()
    im.mu.Lock()
    im.observeLockTime(start, true)
}
```

## Metrics Exposed

### 1. Lock Wait Time Histogram
```
diskcache_lock_wait_seconds{lock_type="write",path="/cache/path"} 0.001
diskcache_lock_wait_seconds{lock_type="read",path="/cache/path"} 0.0001
diskcache_lock_wait_seconds{lock_type="rw",path="/cache/path"} 0.05
```

**Buckets**: `[0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5]` seconds

### 2. Lock Contention Counter
```
diskcache_lock_contention_total{lock_type="write",path="/cache/path"} 15
diskcache_lock_contention_total{lock_type="read",path="/cache/path"} 3
diskcache_lock_contention_total{lock_type="rw",path="/cache/path"} 1
```

## Usage Examples

### 1. Monitoring Lock Contention Rate

```promql
# Contention rate per second by lock type
rate(diskcache_lock_contention_total[5m]) by (lock_type, path)

# Percentile of lock wait times
histogram_quantile(0.95, rate(diskcache_lock_wait_seconds_bucket[5m])) by (lock_type, path)
```

### 2. Alerting on High Contention

```promql
# Alert when lock wait times exceed threshold
histogram_quantile(0.95, diskcache_lock_wait_seconds) > 0.1

# Alert when contention rate is high
rate(diskcache_lock_contention_total[5m]) > 10
```

### 3. Performance Analysis

```promql
# Lock wait time distribution
histogram_quantile(0.50, diskcache_lock_wait_seconds) by (lock_type)
histogram_quantile(0.95, diskcache_lock_wait_seconds) by (lock_type)
histogram_quantile(0.99, diskcache_lock_wait_seconds) by (lock_type)

# Correlate contention with operations
rate(diskcache_lock_contention_total[5m]) / 
rate(diskcache_put_total[5m] + diskcache_get_total[5m])
```

## Implementation Benefits

### 1. **Early Detection of Performance Issues**
- Identifies lock contention before it becomes a bottleneck
- Shows which lock types are under stress
- Helps optimize lock granularity and usage patterns

### 2. **Correlation with System Load**
- Can correlate contention spikes with:
  - High put/get rates
  - Disk I/O delays
  - Memory pressure
  - File system issues

### 3. **Capacity Planning**
- Understanding contention patterns helps with:
  - Sizing thread pools
  - Configuring batch sizes
  - Optimizing cache operations
  - Planning resource allocation

### 4. **Troubleshooting**
- Quickly identify if performance issues are due to:
  - Write contention (many puts)
  - Read contention (many gets)
  - Structural operations (rotations, switches)

## Integration with Existing Metrics

The lock contention metrics integrate seamlessly with existing diskcache metrics:

```
# Overall view of cache performance
diskcache_put_latency_seconds{path="/cache"} 0.001
diskcache_get_latency_seconds{path="/cache"} 0.0005
diskcache_lock_wait_seconds{lock_type="write",path="/cache"} 0.0002
diskcache_lock_wait_seconds{lock_type="read",path="/cache"} 0.0001

# Identify if locks are the bottleneck
diskcache_lock_wait_seconds / diskcache_put_latency_seconds
```

## Performance Overhead

The instrumentation adds minimal overhead:
- **Contention detection**: One `TryLock()` call before actual lock
- **Timing**: One `time.Since()` call per lock acquisition
- **Metrics update**: One histogram and counter observation per lock

The overhead is negligible compared to typical I/O operations and provides valuable observability.

## Production Deployment

### 1. **Monitoring Dashboard**
Create Grafana panels showing:
- Lock wait time percentiles by type
- Contention rates over time
- Lock wait time vs operation latency
- Contention heat maps by cache path

### 2. **Alerting Rules**
Set up alerts for:
- High p95/p99 lock wait times (>100ms)
- Elevated contention rates (>1/sec sustained)
- Sudden spikes in wait times

### 3. **Capacity Analysis**
Use metrics to:
- Identify optimal thread counts
- Tune batch sizes
- Plan for scale increases
- Optimize lock granularity