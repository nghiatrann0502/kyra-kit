package errors

import "runtime"

type Frame struct {
	File     string
	Line     int
	Function string
}

func captureStack(skip int) []Frame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := make([]Frame, 0, n)

	callersFrames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := callersFrames.Next()
		frames = append(frames, Frame{
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		})
		if !more {
			break
		}
	}
	return frames
}
