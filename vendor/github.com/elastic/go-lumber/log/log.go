// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package log provides logging functionality used in go-lumber.
//
// The log package provides replaceable logging for use from within go-lumber.
// Overwrite Logging variable with custom Logging implementation for integrating
// go-lumber logging with applications logging strategy.
package log

import "log"

// Logging interface custom loggers must implement.
type Logging interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Print(...interface{})
}

type defaultLogger struct{}

// Logger provides the global logger used by go-lumber.
var Logger Logging = defaultLogger{}

// Printf calls Logger.Printf to print to the standard logger. Arguments are
// handled in the manner of fmt.Printf.
func Printf(format string, args ...interface{}) {
	Logger.Printf(format, args...)
}

// Println calls Logger.Println to print to the standard logger. Arguments are
// handled in the manner of fmt.Println.
func Println(args ...interface{}) {
	Logger.Println(args...)
}

// Print calls Logger.Print to print to the standard logger. Arguments are
// handled in the manner of fmt.Print.
func Print(args ...interface{}) {
	Logger.Print(args...)
}

func (defaultLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (defaultLogger) Println(args ...interface{}) {
	log.Println(args...)
}

func (defaultLogger) Print(args ...interface{}) {
	log.Print(args...)
}
