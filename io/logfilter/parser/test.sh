#!/bin/bash
go test -test.v -timeout 1h -run TestLexer &&
go test -test.v -timeout 1h -run TestParseBinaryExpr &&
go test -test.v -timeout 1h -run TestParseQuery &&
go test -test.v -timeout 1h -run TestDQL2Influx &&
go test -bench BenchmarkParser &&
go test -cover
