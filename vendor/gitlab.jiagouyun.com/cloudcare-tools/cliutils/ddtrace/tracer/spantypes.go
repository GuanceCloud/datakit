// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package tracer

// Span types have similar behavior to "app types" and help categorize
// traces in the Datadog application. They can also help fine grain agent
// level bahviours such as obfuscation and quantization, when these are
// enabled in the agent's configuration.
const (
	// SpanTypeWeb marks a span as an HTTP server request.
	SpanTypeWeb SpanType = "web"

	// SpanTypeHTTP marks a span as an HTTP client request.
	SpanTypeHTTP SpanType = "http"

	// SpanTypeSQL marks a span as an SQL operation. These spans may
	// have an "sql.command" tag.
	SpanTypeSQL SpanType = "sql"

	// SpanTypeCassandra marks a span as a Cassandra operation. These
	// spans may have an "sql.command" tag.
	SpanTypeCassandra SpanType = "cassandra"

	// SpanTypeRedis marks a span as a Redis operation. These spans may
	// also have a "redis.raw_command" tag.
	SpanTypeRedis SpanType = "redis"

	// SpanTypeMemcached marks a span as a memcached operation.
	SpanTypeMemcached SpanType = "memcached"

	// SpanTypeMongoDB marks a span as a MongoDB operation.
	SpanTypeMongoDB SpanType = "mongodb"

	// SpanTypeElasticSearch marks a span as an ElasticSearch operation.
	// These spans may also have an "elasticsearch.body" tag.
	SpanTypeElasticSearch SpanType = "elasticsearch"

	// SpanTypeLevelDB marks a span as a leveldb operation
	SpanTypeLevelDB SpanType = "leveldb"

	// SpanTypeDNS marks a span as a DNS operation.
	SpanTypeDNS SpanType = "dns"

	// SpanTypeMessageConsumer marks a span as a queue operation
	SpanTypeMessageConsumer SpanType = "queue"

	// SpanTypeMessageProducer marks a span as a queue operation.
	SpanTypeMessageProducer SpanType = "queue"

	// SpanTypeConsul marks a span as a Consul operation.
	SpanTypeConsul SpanType = "consul"
)

type SpanType string
