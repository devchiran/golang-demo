package runtime

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/twitsprout/tools/buffer"
)

var ignoreFrames = [...]string{
	"runtime.main",
	"runtime.goexit",
}

// Stacktrace returns a formatted stack trace skipping the provided number of
// functions.
func Stacktrace(skip int) string {
	var scratch [64]byte
	buf := buffer.Get()
	defer buffer.Put(buf)

	var n int
	callers := make([]uintptr, 12)
	for {
		n = runtime.Callers(skip+2, callers)
		if n < len(callers) {
			break
		}
		callers = make([]uintptr, len(callers)*2)
	}

	var i int
	frames := runtime.CallersFrames(callers[0:n])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if ignoreFrame(frame.Function) {
			continue
		}
		if i > 0 {
			buf.WriteByte('\n')
		}
		i++
		buf.WriteString(frame.Function)
		buf.WriteByte('\n')
		buf.WriteByte('\t')
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		line := strconv.AppendInt(scratch[:0], int64(frame.Line), 10)
		buf.Write(line)
	}

	return buf.String()
}

func ignoreFrame(function string) bool {
	for _, f := range ignoreFrames {
		if strings.HasPrefix(function, f) {
			return true
		}
	}
	return false
}
