// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package datakit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/GuanceCloud/cliutils"
	tu "github.com/GuanceCloud/cliutils/testutil"
)

func TestZstdzip(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
	}{
		{
			name: "common-1k",
			in:   []byte(cliutils.CreateRandomString(1024)),
		},

		{
			name: "common-4k",
			in:   []byte(cliutils.CreateRandomString(1024 * 4)),
		},

		{
			name: "common-1024k",
			in:   []byte(cliutils.CreateRandomString(1024 * 1024)),
		},

		{
			name: "zstd-readme",
			in:   []byte(zstdReadme),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			out, err := Zstdzip(tc.in)
			if err != nil {
				assert.NoError(t, err)
			}

			t.Logf("zstd compressed %d/%d(%f), cost %v", len(tc.in), len(out), float64(len(out))/float64(len(tc.in)), time.Since(start))

			start = time.Now()
			out, err = Zstdzip2(tc.in)
			if err != nil {
				assert.NoError(t, err)
			}

			t.Logf("zstd2 compressed %d/%d(%f), cost %v", len(tc.in), len(out), float64(len(out))/float64(len(tc.in)), time.Since(start))

			start = time.Now()
			out, err = GZip(tc.in)
			if err != nil {
				assert.NoError(t, err)
			}

			t.Logf("gzip compressed %d/%d(%f), cost %v", len(tc.in), len(out), float64(len(out))/float64(len(tc.in)), time.Since(start))
		})
	}
}

func BenchmarkGzipString(b *testing.B) {
	data := cliutils.CreateRandomString(1024)
	for i := 0; i < b.N; i++ {
		_, err := GZipStr(data)
		if err != nil {
			b.Fatal(err)
		}
	}

	// -------------- mem bench ----------------
	// goos: darwin
	// goarch: amd64
	// pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit
	// BenchmarkGzipString-4               4778            237426 ns/op
	// PASS
	// ok      gitlab.jiagouyun.com/cloudcare-tools/datakit    1.722s

	// Type: alloc_space
	// Time: Mar 24, 2021 at 2:25pm (CST)
	// Entering interactive mode (type "help" for commands, "o" for options)
	// (pprof) top
	// Showing nodes accounting for 5182.03MB, 99.58% of 5204.07MB total
	// Dropped 12 nodes (cum <= 26.02MB)
	// Showing top 10 nodes out of 18
	//       flat  flat%   sum%        cum   cum%
	//  4239.69MB 81.47% 81.47%  5141.93MB 98.81%  compress/flate.NewWriter
	//   873.72MB 16.79% 98.26%   873.72MB 16.79%  compress/flate.(*compressor).initDeflate (inline)
	//    46.60MB   0.9% 99.15%    46.60MB   0.9%  compress/flate.(*huffmanEncoder).generate
	//    16.51MB  0.32% 99.47%    28.52MB  0.55%  compress/flate.newHuffmanBitWriter (inline)
	//     5.51MB  0.11% 99.58%  5147.44MB 98.91%  io.WriteString
	//          0     0% 99.58%    54.11MB  1.04%  compress/flate.(*Writer).Flush
	//          0     0% 99.58%    54.11MB  1.04%  compress/flate.(*compressor).deflate
	//          0     0% 99.58%   902.25MB 17.34%  compress/flate.(*compressor).init
	//          0     0% 99.58%    54.11MB  1.04%  compress/flate.(*compressor).syncFlush
	//          0     0% 99.58%    54.11MB  1.04%  compress/flate.(*compressor).writeBlock

	// -------------- cpu bench ----------------
	// goos: darwin
	// goarch: amd64
	// pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit
	// BenchmarkGzipString-4               4855            229180 ns/op
	// PASS
	// ok      gitlab.jiagouyun.com/cloudcare-tools/datakit    1.849s
	//
	// Type: cpu
	// Time: Mar 24, 2021 at 2:30pm (CST)
	// Duration: 1.81s, Total samples = 3.22s (178.10%)
	// Entering interactive mode (type "help" for commands, "o" for options)
	// (pprof) topo
	// unrecognized command: "topo"
	// (pprof) top
	// Showing nodes accounting for 2730ms, 84.78% of 3220ms total
	// Dropped 35 nodes (cum <= 16.10ms)
	// Showing top 10 nodes out of 100
	//       flat  flat%   sum%        cum   cum%
	//      910ms 28.26% 28.26%      910ms 28.26%  runtime.pthread_cond_wait
	//      810ms 25.16% 53.42%      810ms 25.16%  runtime.kevent
	//      290ms  9.01% 62.42%      290ms  9.01%  runtime.pthread_kill
	//      160ms  4.97% 67.39%      160ms  4.97%  runtime.memclrNoHeapPointers
	//      160ms  4.97% 72.36%      240ms  7.45%  runtime.scanobject
	//      120ms  3.73% 76.09%      120ms  3.73%  runtime.memmove
	//       80ms  2.48% 78.57%       90ms  2.80%  runtime.findObject
	//       80ms  2.48% 81.06%       80ms  2.48%  runtime.pthread_cond_signal
	//       60ms  1.86% 82.92%       60ms  1.86%  runtime.madvise
	//       60ms  1.86% 84.78%       70ms  2.17%  runtime.nanotime1
}

func BenchmarkGzipBytes(b *testing.B) {
	data := cliutils.CreateRandomString(1024)
	for i := 0; i < b.N; i++ {
		_, err := GZip([]byte(data))
		if err != nil {
			b.Fatal(err)
		}
	}

	// goos: darwin
	// goarch: amd64
	// pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit
	// BenchmarkGzipBytes-4        5211            245549 ns/op
	// PASS
	// ok      gitlab.jiagouyun.com/cloudcare-tools/datakit    2.834s
	// Mac :) (doc-integration) gtp mem.out
	// Type: alloc_space
	// Time: Mar 24, 2021 at 2:26pm (CST)
	// Entering interactive mode (type "help" for commands, "o" for options)
	// (pprof) top
	// Showing nodes accounting for 8.46GB, 99.54% of 8.49GB total
	// Dropped 28 nodes (cum <= 0.04GB)
	// Showing top 10 nodes out of 17
	//			flat  flat%   sum%        cum   cum%
	//		6.90GB 81.20% 81.20%     8.39GB 98.72%  compress/flate.NewWriter
	//		1.44GB 17.00% 98.20%     1.44GB 17.00%  compress/flate.(*compressor).initDeflate (inline)
	//		0.07GB  0.87% 99.07%     0.07GB  0.87%  compress/flate.(*huffmanEncoder).generate
	//		0.02GB  0.28% 99.34%     0.04GB  0.52%  compress/flate.newHuffmanBitWriter (inline)
	//		0.02GB   0.2% 99.54%     8.49GB   100%  gitlab.jiagouyun.com/cloudcare-tools/datakit.BenchmarkGzipBytes
	//				 0     0% 99.54%     0.09GB  1.04%  compress/flate.(*Writer).Flush (inline)
	//				 0     0% 99.54%     0.09GB  1.04%  compress/flate.(*compressor).deflate
	//				 0     0% 99.54%     1.49GB 17.52%  compress/flate.(*compressor).init
	//				 0     0% 99.54%     0.09GB  1.04%  compress/flate.(*compressor).syncFlush
	//				 0     0% 99.54%     0.09GB  1.04%  compress/flate.(*compressor).writeBlock

	// ---------------- cpu bench ----------------
	// goos: darwin
	// goarch: amd64
	// pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit
	// BenchmarkGzipBytes-4        4754            234517 ns/op
	// PASS
	// ok      gitlab.jiagouyun.com/cloudcare-tools/datakit    1.992s
	//
	// Type: cpu
	// Time: Mar 24, 2021 at 2:33pm (CST)
	// Duration: 1.90s, Total samples = 3.29s (172.73%)
	// Entering interactive mode (type "help" for commands, "o" for options)
	// (pprof) top
	// Showing nodes accounting for 2880ms, 87.54% of 3290ms total
	// Dropped 36 nodes (cum <= 16.45ms)
	// Showing top 10 nodes out of 111
	//       flat  flat%   sum%        cum   cum%
	//     1090ms 33.13% 33.13%     1090ms 33.13%  runtime.pthread_cond_wait
	//      700ms 21.28% 54.41%      700ms 21.28%  runtime.kevent
	//      240ms  7.29% 61.70%      240ms  7.29%  runtime.pthread_kill
	//      210ms  6.38% 68.09%      210ms  6.38%  runtime.pthread_cond_signal
	//      200ms  6.08% 74.16%      280ms  8.51%  runtime.scanobject
	//      120ms  3.65% 77.81%      120ms  3.65%  runtime.memclrNoHeapPointers
	//      110ms  3.34% 81.16%      110ms  3.34%  runtime.pthread_cond_timedwait_relative_np
	//       80ms  2.43% 83.59%       80ms  2.43%  runtime.usleep
	//       70ms  2.13% 85.71%       70ms  2.13%  runtime.memmove
	//       60ms  1.82% 87.54%       60ms  1.82%  runtime.madvise
}

func TestDuration(t *testing.T) {
	d := Duration{Duration: time.Second}
	tu.Equals(t, "1s", d.UnitString(time.Second))
	tu.Equals(t, "1000000000ns", d.UnitString(time.Nanosecond))
	tu.Equals(t, "1000000mics", d.UnitString(time.Microsecond))
	tu.Equals(t, "1000ms", d.UnitString(time.Millisecond))
	tu.Equals(t, "0m", d.UnitString(time.Minute))
	tu.Equals(t, "0h", d.UnitString(time.Hour))
}

var zstdReadme = `
# zstd

[Zstandard](https://facebook.github.io/zstd/) is a real-time compression algorithm, providing high compression ratios.
It offers a very wide range of compression / speed trade-off, while being backed by a very fast decoder.
A high performance compression algorithm is implemented. For now focused on speed.

This package provides [compression](#Compressor) to and [decompression](#Decompressor) of Zstandard content.

This package is pure Go and without use of "unsafe".

The zstd package is provided as open source software using a Go standard license.

Currently the package is heavily optimized for 64 bit processors and will be significantly slower on 32 bit processors.

## Installation

Install using go get -u github.com/klauspost/compress. The package is located in github.com/klauspost/compress/zstd.

[![Go Reference](https://pkg.go.dev/badge/github.com/klauspost/compress/zstd.svg)](https://pkg.go.dev/github.com/klauspost/compress/zstd)

## Compressor

### Status:

STABLE - there may always be subtle bugs, a wide variety of content has been tested and the library is actively
used by several projects. This library is being [fuzz-tested](https://github.com/klauspost/compress-fuzz) for all updates.

There may still be specific combinations of data types/size/settings that could lead to edge cases,
so as always, testing is recommended.

For now, a high speed (fastest) and medium-fast (default) compressor has been implemented.

* The "Fastest" compression ratio is roughly equivalent to zstd level 1.
* The "Default" compression ratio is roughly equivalent to zstd level 3 (default).
* The "Better" compression ratio is roughly equivalent to zstd level 7.
* The "Best" compression ratio is roughly equivalent to zstd level 11.

In terms of speed, it is typically 2x as fast as the stdlib deflate/gzip in its fastest mode.
The compression ratio compared to stdlib is around level 3, but usually 3x as fast.


### Usage

An Encoder can be used for either compressing a stream via the
io.WriteCloser interface supported by the Encoder or as multiple independent
tasks via the EncodeAll function.
Smaller encodes are encouraged to use the EncodeAll function.
Use NewWriter to create a new instance that can be used for both.

To create a writer with default options, do like this:

Go
// Compress input to output.
func Compress(in io.Reader, out io.Writer) error {
    enc, err := zstd.NewWriter(out)
    if err != nil {
        return err
    }
    _, err = io.Copy(enc, in)
    if err != nil {
        enc.Close()
        return err
    }
    return enc.Close()
}


Now you can encode by writing data to enc. The output will be finished writing when Close() is called.
Even if your encode fails, you should still call Close() to release any resources that may be held up.

The above is fine for big encodes. However, whenever possible try to *reuse* the writer.

To reuse the encoder, you can use the Reset(io.Writer) function to change to another output.
This will allow the encoder to reuse all resources and avoid wasteful allocations.

Currently stream encoding has 'light' concurrency, meaning up to 2 goroutines can be working on part
of a stream. This is independent of the WithEncoderConcurrency(n), but that is likely to change
in the future. So if you want to limit concurrency for future updates, specify the concurrency
you would like.

If you would like stream encoding to be done without spawning async goroutines, use WithEncoderConcurrency(1)
which will compress input as each block is completed, blocking on writes until each has completed.

You can specify your desired compression level using WithEncoderLevel() option. Currently only pre-defined
compression settings can be specified.

#### Future Compatibility Guarantees

This will be an evolving project. When using this package it is important to note that both the compression efficiency and speed may change.

The goal will be to keep the default efficiency at the default zstd (level 3).
However the encoding should never be assumed to remain the same,
and you should not use hashes of compressed output for similarity checks.

The Encoder can be assumed to produce the same output from the exact same code version.
However, the may be modes in the future that break this,
although they will not be enabled without an explicit option.

This encoder is not designed to (and will probably never) output the exact same bitstream as the reference encoder.

Also note, that the cgo decompressor currently does not [report all errors on invalid input](https://github.com/DataDog/zstd/issues/59),
[omits error checks](https://github.com/DataDog/zstd/issues/61), [ignores checksums](https://github.com/DataDog/zstd/issues/43)
and seems to ignore concatenated streams, even though [it is part of the spec](https://github.com/facebook/zstd/blob/dev/doc/zstd_compression_format.md#frames).

#### Blocks

For compressing small blocks, the returned encoder has a function called EncodeAll(src, dst []byte) []byte.

EncodeAll will encode all input in src and append it to dst.
This function can be called concurrently.
Each call will only run on a same goroutine as the caller.

Encoded blocks can be concatenated and the result will be the combined input stream.
Data compressed with EncodeAll can be decoded with the Decoder, using either a stream or DecodeAll.

Especially when encoding blocks you should take special care to reuse the encoder.
This will effectively make it run without allocations after a warmup period.
To make it run completely without allocations, supply a destination buffer with space for all content.

Go
import "github.com/klauspost/compress/zstd"

// Create a writer that caches compressors.
// For this operation type we supply a nil Reader.
var encoder, _ = zstd.NewWriter(nil)

// Compress a buffer.
// If you have a destination buffer, the allocation in the call can also be eliminated.
func Compress(src []byte) []byte {
    return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}


You can control the maximum number of concurrent encodes using the WithEncoderConcurrency(n)
option when creating the writer.

Using the Encoder for both a stream and individual blocks concurrently is safe.

### Performance

I have collected some speed examples to compare speed and compression against other compressors.

* file is the input file.
* out is the compressor used. zskp is this package. zstd is the Datadog cgo library. gzstd/gzkp is gzip standard and this library.
* level is the compression level used. For zskp level 1 is "fastest", level 2 is "default"; 3 is "better", 4 is "best".
* insize/outsize is the input/output size.
* millis is the number of milliseconds used for compression.
* mb/s is megabytes (2^20 bytes) per second.


Silesia Corpus:
http://sun.aei.polsl.pl/~sdeor/corpus/silesia.zip

This package:
file    out     level   insize      outsize     millis  mb/s
silesia.tar zskp    1   211947520   73821326    634     318.47
silesia.tar zskp    2   211947520   67655404    1508    133.96
silesia.tar zskp    3   211947520   64746933    3000    67.37
silesia.tar zskp    4   211947520   60073508    16926   11.94

cgo zstd:
silesia.tar zstd    1   211947520   73605392    543     371.56
silesia.tar zstd    3   211947520   66793289    864     233.68
silesia.tar zstd    6   211947520   62916450    1913    105.66
silesia.tar zstd    9   211947520   60212393    5063    39.92

gzip, stdlib/this package:
silesia.tar gzstd   1   211947520   80007735    1498    134.87
silesia.tar gzkp    1   211947520   80088272    1009    200.31

GOB stream of binary data. Highly compressible.
https://files.klauspost.com/compress/gob-stream.7z

file        out     level   insize  outsize     millis  mb/s
gob-stream  zskp    1   1911399616  233948096   3230    564.34
gob-stream  zskp    2   1911399616  203997694   4997    364.73
gob-stream  zskp    3   1911399616  173526523   13435   135.68
gob-stream  zskp    4   1911399616  162195235   47559   38.33

gob-stream  zstd    1   1911399616  249810424   2637    691.26
gob-stream  zstd    3   1911399616  208192146   3490    522.31
gob-stream  zstd    6   1911399616  193632038   6687    272.56
gob-stream  zstd    9   1911399616  177620386   16175   112.70

gob-stream  gzstd   1   1911399616  357382013   9046    201.49
gob-stream  gzkp    1   1911399616  359136669   4885    373.08

The test data for the Large Text Compression Benchmark is the first
10^9 bytes of the English Wikipedia dump on Mar. 3, 2006.
http://mattmahoney.net/dc/textdata.html

file    out level   insize      outsize     millis  mb/s
enwik9  zskp    1   1000000000  343833605   3687    258.64
enwik9  zskp    2   1000000000  317001237   7672    124.29
enwik9  zskp    3   1000000000  291915823   15923   59.89
enwik9  zskp    4   1000000000  261710291   77697   12.27

enwik9  zstd    1   1000000000  358072021   3110    306.65
enwik9  zstd    3   1000000000  313734672   4784    199.35
enwik9  zstd    6   1000000000  295138875   10290   92.68
enwik9  zstd    9   1000000000  278348700   28549   33.40

enwik9  gzstd   1   1000000000  382578136   8608    110.78
enwik9  gzkp    1   1000000000  382781160   5628    169.45

Highly compressible JSON file.
https://files.klauspost.com/compress/github-june-2days-2019.json.zst

file                        out level   insize      outsize     millis  mb/s
github-june-2days-2019.json zskp    1   6273951764  697439532   9789    611.17
github-june-2days-2019.json zskp    2   6273951764  610876538   18553   322.49
github-june-2days-2019.json zskp    3   6273951764  517662858   44186   135.41
github-june-2days-2019.json zskp    4   6273951764  464617114   165373  36.18

github-june-2days-2019.json zstd    1   6273951764  766284037   8450    708.00
github-june-2days-2019.json zstd    3   6273951764  661889476   10927   547.57
github-june-2days-2019.json zstd    6   6273951764  642756859   22996   260.18
github-june-2days-2019.json zstd    9   6273951764  601974523   52413   114.16

github-june-2days-2019.json gzstd   1   6273951764  1164397768  26793   223.32
github-june-2days-2019.json gzkp    1   6273951764  1120631856  17693   338.16

VM Image, Linux mint with a few installed applications:
https://files.klauspost.com/compress/rawstudio-mint14.7z

file                    out level   insize      outsize     millis  mb/s
rawstudio-mint14.tar    zskp    1   8558382592  3718400221  18206   448.29
rawstudio-mint14.tar    zskp    2   8558382592  3326118337  37074   220.15
rawstudio-mint14.tar    zskp    3   8558382592  3163842361  87306   93.49
rawstudio-mint14.tar    zskp    4   8558382592  2970480650  783862  10.41

rawstudio-mint14.tar    zstd    1   8558382592  3609250104  17136   476.27
rawstudio-mint14.tar    zstd    3   8558382592  3341679997  29262   278.92
rawstudio-mint14.tar    zstd    6   8558382592  3235846406  77904   104.77
rawstudio-mint14.tar    zstd    9   8558382592  3160778861  140946  57.91

rawstudio-mint14.tar    gzstd   1   8558382592  3926234992  51345   158.96
rawstudio-mint14.tar    gzkp    1   8558382592  3960117298  36722   222.26

CSV data:
https://files.klauspost.com/compress/nyc-taxi-data-10M.csv.zst

file                    out level   insize      outsize     millis  mb/s
nyc-taxi-data-10M.csv   zskp    1   3325605752  641319332   9462    335.17
nyc-taxi-data-10M.csv   zskp    2   3325605752  588976126   17570   180.50
nyc-taxi-data-10M.csv   zskp    3   3325605752  529329260   32432   97.79
nyc-taxi-data-10M.csv   zskp    4   3325605752  474949772   138025  22.98

nyc-taxi-data-10M.csv   zstd    1   3325605752  687399637   8233    385.18
nyc-taxi-data-10M.csv   zstd    3   3325605752  598514411   10065   315.07
nyc-taxi-data-10M.csv   zstd    6   3325605752  570522953   20038   158.27
nyc-taxi-data-10M.csv   zstd    9   3325605752  517554797   64565   49.12

nyc-taxi-data-10M.csv   gzstd   1   3325605752  928654908   21270   149.11
nyc-taxi-data-10M.csv   gzkp    1   3325605752  922273214   13929   227.68

## Decompressor

Staus: STABLE - there may still be subtle bugs, but a wide variety of content has been tested.

This library is being continuously [fuzz-tested](https://github.com/klauspost/compress-fuzz),
kindly supplied by [fuzzit.dev](https://fuzzit.dev/).
The main purpose of the fuzz testing is to ensure that it is not possible to crash the decoder,
or run it past its limits with ANY input provided.

### Usage

The package has been designed for two main usages, big streams of data and smaller in-memory buffers.
There are two main usages of the package for these. Both of them are accessed by creating a Decoder.

For streaming use a simple setup could look like this:

Go
import "github.com/klauspost/compress/zstd"

func Decompress(in io.Reader, out io.Writer) error {
    d, err := zstd.NewReader(in)
    if err != nil {
        return err
    }
    defer d.Close()

    // Copy content...
    _, err = io.Copy(out, d)
    return err
}


It is important to use the "Close" function when you no longer need the Reader to stop running goroutines,
when running with default settings.
Goroutines will exit once an error has been returned, including io.EOF at the end of a stream.

Streams are decoded concurrently in 4 asynchronous stages to give the best possible throughput.
However, if you prefer synchronous decompression, use WithDecoderConcurrency(1) which will decompress data
as it is being requested only.

For decoding buffers, it could look something like this:

Go
import "github.com/klauspost/compress/zstd"

// Create a reader that caches decompressors.
// For this operation type we supply a nil Reader.
var decoder, _ = zstd.NewReader(nil, WithDecoderConcurrency(0))

// Decompress a buffer. We don't supply a destination buffer,
// so it will be allocated by the decoder.
func Decompress(src []byte) ([]byte, error) {
    return decoder.DecodeAll(src, nil)
}


Both of these cases should provide the functionality needed.
The decoder can be used for *concurrent* decompression of multiple buffers.
By default 4 decompressors will be created.

It will only allow a certain number of concurrent operations to run.
To tweak that yourself use the WithDecoderConcurrency(n) option when creating the decoder.
It is possible to use WithDecoderConcurrency(0) to create GOMAXPROCS decoders.

### Dictionaries

Data compressed with [dictionaries](https://github.com/facebook/zstd#the-case-for-small-data-compression) can be decompressed.

Dictionaries are added individually to Decoders.
Dictionaries are generated by the zstd --train command and contains an initial state for the decoder.
To add a dictionary use the WithDecoderDicts(dicts ...[]byte) option with the dictionary data.
Several dictionaries can be added at once.

The dictionary will be used automatically for the data that specifies them.
A re-used Decoder will still contain the dictionaries registered.

When registering multiple dictionaries with the same ID, the last one will be used.

It is possible to use dictionaries when compressing data.

To enable a dictionary use WithEncoderDict(dict []byte). Here only one dictionary will be used
and it will likely be used even if it doesn't improve compression.

The used dictionary must be used to decompress the content.

For any real gains, the dictionary should be built with similar data.
If an unsuitable dictionary is used the output may be slightly larger than using no dictionary.
Use the [zstd commandline tool](https://github.com/facebook/zstd/releases) to build a dictionary from sample data.
For information see [zstd dictionary information](https://github.com/facebook/zstd#the-case-for-small-data-compression).

For now there is a fixed startup performance penalty for compressing content with dictionaries.
This will likely be improved over time. Just be aware to test performance when implementing.

### Allocation-less operation

The decoder has been designed to operate without allocations after a warmup.

This means that you should *store* the decoder for best performance.
To re-use a stream decoder, use the Reset(r io.Reader) error to switch to another stream.
A decoder can safely be re-used even if the previous stream failed.

To release the resources, you must call the Close() function on a decoder.
After this it can *no longer be reused*, but all running goroutines will be stopped.
So you *must* use this if you will no longer need the Reader.

For decompressing smaller buffers a single decoder can be used.
When decoding buffers, you can supply a destination slice with length 0 and your expected capacity.
In this case no unneeded allocations should be made.

### Concurrency

The buffer decoder does everything on the same goroutine and does nothing concurrently.
It can however decode several buffers concurrently. Use WithDecoderConcurrency(n) to limit that.

The stream decoder will create goroutines that:

1) Reads input and splits the input into blocks.
2) Decompression of literals.
3) Decompression of sequences.
4) Reconstruction of output stream.

So effectively this also means the decoder will "read ahead" and prepare data to always be available for output.

The concurrency level will, for streams, determine how many blocks ahead the compression will start.

Since "blocks" are quite dependent on the output of the previous block stream decoding will only have limited concurrency.

In practice this means that concurrency is often limited to utilizing about 3 cores effectively.

### Benchmarks

The first two are streaming decodes and the last are smaller inputs.

Running on AMD Ryzen 9 3950X 16-Core Processor. AMD64 assembly used.


BenchmarkDecoderSilesia-32    	                   5	 206878840 ns/op	1024.50 MB/s	   49808 B/op	      43 allocs/op
BenchmarkDecoderEnwik9-32                          1	1271809000 ns/op	 786.28 MB/s	   72048 B/op	      52 allocs/op

Concurrent blocks, performance:

BenchmarkDecoder_DecodeAllParallel/kppkn.gtb.zst-32         	   67356	     17857 ns/op	10321.96 MB/s	        22.48 pct	     102 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/geo.protodata.zst-32     	  266656	      4421 ns/op	26823.21 MB/s	        11.89 pct	      19 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/plrabn12.txt.zst-32      	   20992	     56842 ns/op	8477.17 MB/s	        39.90 pct	     754 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/lcet10.txt.zst-32        	   27456	     43932 ns/op	9714.01 MB/s	        33.27 pct	     524 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/asyoulik.txt.zst-32      	   78432	     15047 ns/op	8319.15 MB/s	        40.34 pct	      66 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/alice29.txt.zst-32       	   65800	     18436 ns/op	8249.63 MB/s	        37.75 pct	      88 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/html_x_4.zst-32          	  102993	     11523 ns/op	35546.09 MB/s	         3.637 pct	     143 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/paper-100k.pdf.zst-32    	 1000000	      1070 ns/op	95720.98 MB/s	        80.53 pct	       3 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/fireworks.jpeg.zst-32    	  749802	      1752 ns/op	70272.35 MB/s	       100.0 pct	       5 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/urls.10K.zst-32          	   22640	     52934 ns/op	13263.37 MB/s	        26.25 pct	    1014 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/html.zst-32              	  226412	      5232 ns/op	19572.27 MB/s	        14.49 pct	      20 B/op	       0 allocs/op
BenchmarkDecoder_DecodeAllParallel/comp-data.bin.zst-32     	  923041	      1276 ns/op	3194.71 MB/s	        31.26 pct	       0 B/op	       0 allocs/op


This reflects the performance around May 2022, but this may be out of date.

## Zstd inside ZIP files

It is possible to use zstandard to compress individual files inside zip archives.
While this isn't widely supported it can be useful for internal files.

To support the compression and decompression of these files you must register a compressor and decompressor.

It is highly recommended registering the (de)compressors on individual zip Reader/Writer and NOT
use the global registration functions. The main reason for this is that 2 registrations from
different packages will result in a panic.

It is a good idea to only have a single compressor and decompressor, since they can be used for multiple zip
files concurrently, and using a single instance will allow reusing some resources.

See [this example](https://pkg.go.dev/github.com/klauspost/compress/zstd#example-ZipCompressor) for
how to compress and decompress files inside zip archives.

# Contributions

Contributions are always welcome.
For new features/fixes, remember to add tests and for performance enhancements include benchmarks.

For general feedback and experience reports, feel free to open an issue or write me on [Twitter](https://twitter.com/sh0dan).

This package includes the excellent [github.com/cespare/xxhash](https://github.com/cespare/xxhash) package Copyright (c) 2016 Caleb Spare.
	`
