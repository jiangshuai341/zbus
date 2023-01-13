package logger

import "testing"

var longStr = string(make([]byte, 1024))

func BenchmarkLogger(b *testing.B) {
	lg := GetLogger("TestLog")
	for i := 0; i < b.N; i++ {
		lg.Debug("debugxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		lg.Info(longStr)
		lg.Warn("warn")
		lg.Error("ERROR")
	}
	//FlushLogger()
}
