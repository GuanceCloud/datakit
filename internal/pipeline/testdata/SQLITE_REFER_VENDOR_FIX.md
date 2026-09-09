# SQLite refer dependency fix

The local fix is in `vendor/github.com/GuanceCloud/pipeline-go/ptinput/refertable/table_sqlite_other.go`
(upstream module v1.4.3). Do not regenerate vendor without preserving/reapplying
the change or adopting an upstream release containing it.

It serializes refresh against reads, returns commit failures, publishes table
metadata only after successful commit, and discards connections on commit or
rollback failure so an unfinished SQLite transaction is not returned to the pool.

Use `-mod=vendor` to exercise this fix. `-mod=mod` intentionally continues to
exercise unmodified upstream and may fail the retained reproducer.

Regression entry points in `internal/pipeline/jit` (requires `pipeline_jit` build tag):

- `TestReferTableSQLiteReadTransactionRecovery`: no SO required; external read transaction blocks refresh.
- `TestReferTableSQLiteDiskRefreshLifecycle`: real SO via `PLATYPUS_JIT_RUNTIME`.
- `TestReferTableSQLiteRefreshLifecycle`: in-memory SQLite.

The vendor patch also adds `ReferTable.Close` and SQLite backend `Close`, and
stops the pull worker's ticker on return. Owners must cancel and join the pull
worker and drain script users before closing the shared service. Tests use this
API when available and retain subprocess isolation for unmodified-upstream runs.
PullWorker also propagates its context to HTTP refresh requests and checks
cancellation before updating tables; a blocked HTTP request can now be canceled.

The `:memory:` SQLite pool is limited to one open/idle connection: each physical
connection otherwise owns a different database, so concurrent readers can report
`no such table` after successful initialization. This intentionally serializes
SQLite access; it does not change the non-SQLite in-memory backend or disk pools.
Throughput under contention still needs measurement. Update failures now return
to PullWorker and do not close the initialization channel; a later successful
refresh completes initialization. InitFinished also stops its waiting ticker.

Additional regression entry points (no SO needed):

- `TestSQLiteReferInitializationRequiresSuccessfulUpdate`: invalid schema must not
  signal ready; a subsequent valid response recovers.
- `TestSQLiteReferMemoryConcurrentQueries`: 32 readers, each doing 50 query/stats
  pairs, all must see the initialized database. Both tests failed before this fix
  and passed three times with Go race after it.

These tests do not establish a complete production lifecycle: DataKit's config
replacement does not yet coordinate old worker shutdown and final service close.
Deployment/version, concurrency and sustained resource gates remain outstanding.
