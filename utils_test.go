package datakit

import (
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	//"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

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
	fmt.Println(d.UnitString(time.Second))
	fmt.Println(d.UnitString(time.Nanosecond))
	fmt.Println(d.UnitString(time.Microsecond))
	fmt.Println(d.UnitString(time.Millisecond))
	fmt.Println(d.UnitString(time.Minute))
	fmt.Println(d.UnitString(time.Hour))
}
