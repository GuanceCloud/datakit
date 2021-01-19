package rtpanic

import (
	"runtime"
)

const (
	StackTraceSize = 4096
)

// 所有 agent 的 panic 信息都需要上报给 csos
type RecoverCallback func(info []byte, err error)

// @recoverCallback: 复活函数, 即如果某个 goroutine panic 后, 可以指定某个函数, 继续复活该 goroutine
// @cleanupCallback: 清理/善后回调函数, 比如上报 panic 信息, 现场清理等等
// TRICK: 尽量将这些回调定义成本地函数, 这样便于处理现场, 比如 recoverCallback:
func Recover(recoverCallback, cleanupCallback RecoverCallback) {
	r := recover()

	// 通过判断 recover() 的返回情况, 确定 goroutine 是正常退出还是被 panic 了
	if err, ok := r.(error); ok {
		buf := make([]byte, StackTraceSize)
		runtime.Stack(buf, false)

		if cleanupCallback != nil {
			cleanupCallback(buf, err)
		}

		if recoverCallback != nil {
			recoverCallback(buf, err) // 将 panic 信息回送给复活函数处理
		}
	}
}
