package logger

import "testing"

func BenchmarkRogger(b *testing.B) {
	SetLevel(DEBUG)
	longStr := string(make([]byte, 1024))
	lg := GetLogger("TestLog")
	lg.SetFileRoller("./logs", 10, 100)
	for i := 0; i < b.N; i++ {
		lg.Debug("debugxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		lg.Info(longStr)
		lg.Warn("warn")
		lg.Error("ERROR")
	}
	FlushLogger()
}
