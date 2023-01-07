package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	currUnixTime int64
	currDateTime string
	currDateHour string
	currDateDay  string
)

func init() {
	now := time.Now()
	currUnixTime = now.Unix()
	currDateTime = now.Format("2006-01-02 15:04:05")
	currDateHour = now.Format("2006010215")
	currDateDay = now.Format("20060102")
	go func() {
		tm := time.NewTimer(time.Second)
		for {
			now := time.Now()
			d := time.Second - time.Duration(now.Nanosecond())
			tm.Reset(d)
			<-tm.C
			now = time.Now()
			currUnixTime = now.Unix()
			currDateTime = now.Format("2006-01-02 15:04:05")
			currDateHour = now.Format("2006010215")
			currDateDay = now.Format("20060102")
		}
	}()
}

// day for rotate log by day
const (
	day dateType = iota
	hour
)

// logWriter is interface for different writer.
type logWriter interface {
	Write(v []byte)
	NeedPrefix() bool
	getOutFile() *os.File
}

// consoleWriter writes the logs to the console.
type consoleWriter struct {
}

func (c *consoleWriter) getOutFile() *os.File {
	return os.Stdout
}

// rollFileWriter struct for rotate logs by file size.
type rollFileWriter struct {
	logpath  string
	name     string
	num      int
	size     int64
	currSize int64
	currFile *os.File
	openTime int64
}

func (w *rollFileWriter) getOutFile() *os.File {
	return w.currFile
}

// dateWriter rotate logs by date.
type dateWriter struct {
	logpath   string
	name      string
	dateType  dateType
	num       int
	currDate  string
	currFile  *os.File
	openTime  int64
	hasPrefix bool
}

func (w *dateWriter) getOutFile() *os.File {
	return w.currFile
}

// dateType is uint8
type dateType uint8

func reOpenFile(path string, currFile **os.File, openTime *int64) {
	*openTime = currUnixTime
	if *currFile != nil {
		_ = (*currFile).Close()
	}
	of, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		*currFile = of
	} else {
		fmt.Println("open log file error", err)
	}
}
func (w *consoleWriter) Write(v []byte) {
	_, _ = os.Stdout.Write(v)
}

// NeedPrefix shows whether needs the prefix for the console writer.
func (w *consoleWriter) NeedPrefix() bool {
	return true
}

// Write for writing []byte to the writter.
func (w *rollFileWriter) Write(v []byte) {
	if w.currFile == nil || w.openTime+10 < currUnixTime {
		fullPath := filepath.Join(w.logpath, w.name+".log")
		reOpenFile(fullPath, &w.currFile, &w.openTime)
	}
	if w.currFile == nil {
		return
	}
	n, _ := w.currFile.Write(v)
	w.currSize += int64(n)
	if w.currSize >= w.size {
		w.currSize = 0
		for i := w.num - 1; i >= 1; i-- {
			var n1, n2 string
			if i > 1 {
				n1 = strconv.Itoa(i - 1)
			}
			n2 = strconv.Itoa(i)
			p1 := filepath.Join(w.logpath, w.name+n1+".log")
			p2 := filepath.Join(w.logpath, w.name+n2+".log")
			if _, err := os.Stat(p1); !os.IsNotExist(err) {
				_ = os.Rename(p1, p2)
			}
		}
		fullPath := filepath.Join(w.logpath, w.name+".log")
		reOpenFile(fullPath, &w.currFile, &w.openTime)
	}
}

// newRollFileWriter returns a rollFileWriter, rotate logs in sizeMB , and num files are keeped.
func newRollFileWriter(logpath, name string, num, sizeMB int) *rollFileWriter {
	w := &rollFileWriter{
		logpath: logpath,
		name:    name,
		num:     num,
		size:    int64(sizeMB) * 1024 * 1024,
	}
	fullPath := filepath.Join(logpath, name+".log")
	st, _ := os.Stat(fullPath)
	if st != nil {
		w.currSize = st.Size()
	}
	return w
}

// NeedPrefix shows need prefix or not.
func (w *rollFileWriter) NeedPrefix() bool {
	return true
}

func (w *dateWriter) isExpired() bool {
	currDate := w.getCurrDate()
	return w.currDate != currDate
}

// Write method implement for the dateWriter
func (w *dateWriter) Write(v []byte) {
	if w.isExpired() {
		w.currDate = w.getCurrDate()
		//w.cleanOldLogs()
		fullPath := filepath.Join(w.logpath, w.name+"_"+w.currDate+".log")
		reOpenFile(fullPath, &w.currFile, &w.openTime)
	}
	if w.currFile == nil || w.openTime+10 < currUnixTime {
		fullPath := filepath.Join(w.logpath, w.name+"_"+w.currDate+".log")
		reOpenFile(fullPath, &w.currFile, &w.openTime)
	}
	if w.currFile == nil {
		return
	}
	_, _ = w.currFile.Write(v)
}

// NeedPrefix shows whether needs prefix info for dateWriter or not.
func (w *dateWriter) NeedPrefix() bool {
	return w.hasPrefix
}

func (w *dateWriter) SetPrefix(enable bool) {
	w.hasPrefix = enable
}

// newDateWriter returns a writer which keeps logs in hours or day format.
func newDateWriter(logpath, name string, dateType dateType, num int) *dateWriter {
	w := &dateWriter{
		logpath:   logpath,
		name:      name,
		num:       num,
		dateType:  dateType,
		hasPrefix: true,
	}
	w.currDate = w.getCurrDate()
	return w
}

func (w *dateWriter) cleanOldLogs() {
	format := "20060102"
	duration := -time.Hour * 24
	if w.dateType == hour {
		format = "2006010215"
		duration = -time.Hour
	}

	t := time.Now()
	t = t.Add(duration * time.Duration(w.num))
	for i := 0; i < 30; i++ {
		t = t.Add(duration)
		k := t.Format(format)
		fullPath := filepath.Join(w.logpath, w.name+"_"+k+".log")
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			_ = os.Remove(fullPath)
		}
	}
	return
}

func (w *dateWriter) getCurrDate() string {
	if w.dateType == hour {
		return currDateHour
	}
	return currDateDay // day
}
