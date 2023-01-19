package logger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func init() {
	go flushLog()
}

type LogLevel uint8

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	OFF
)

type LogMode uint8

const (
	ScrollByFileSize LogMode = iota
	ScrollByDay
	ScrollByHour
	Console
)

// global Config
var (
	logLevel            = DEBUG
	logMode     LogMode = Console
	colored             = false
	maxFileNum          = 10
	maxFileSize         = 100 //MB
	logDirPath          = "./log"

	callerSkip = 2
	callerFlag = true

	logQueue    = make(chan *logValue, 10000)
	loggerMutex sync.Mutex
	loggerMap   = make(map[string]*Logger)
	writeDone   = make(chan bool)

	waitFlushTimeout       = time.Second
	syncDone, syncCancel   = context.WithCancel(context.Background())
	asyncDone, asyncCancel = context.WithCancel(context.Background())
)

type Logger struct {
	name   string
	writer logWriter
	depth  int
}

type logValue struct {
	level  LogLevel
	value  []byte
	fileNo string
	writer logWriter
}

var confOnce sync.Once

func SetGlobalConfig(level LogLevel, mode LogMode, maxLogFileNum int, maxLogFileSize int, isColor bool, dirPath string) {
	var isDo bool = false
	confOnce.Do(func() {
		logLevel = level
		logMode = mode
		maxFileNum = maxLogFileNum
		maxFileSize = maxLogFileSize
		colored = isColor
		logDirPath = dirPath

		isDo = true
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000") + "| SetGlobalConfig Succ")
	})

	if !isDo {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05.000") + "| SetGlobalConfig Failed , please set config only once")
	}
}
func RedirectStdout(name string) {
	lg := GetLogger("name")
	os.Stdout = lg.writer.getOutFile()
}
func GetLogger(name string) *Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if lg, ok := loggerMap[name]; ok {
		return lg
	}
	lg := &Logger{
		name: name,
	}
	loggerMap[name] = lg

	var err error = nil

	switch logMode {
	case ScrollByFileSize:
		err = lg.setFileRoller("./logs", maxFileNum, maxFileSize)
	case ScrollByDay:
		err = lg.setDayRoller("./logs", maxFileNum)
	case ScrollByHour:
		err = lg.setHourRoller("./logs", maxFileNum)
	case Console:
		lg.setConsole()
	}
	if err != nil {
		fmt.Println("[GetLogger] Failed Err:", err)
		return nil
	}
	return lg
}

func (l *Logger) SetDepth(in int) {
	l.depth = in
}

// Debug logs interface in debug loglevel.
func (l *Logger) Debug(v ...any) {
	l.writef(l.depth, DEBUG, "", v)
}

// Info logs interface in Info loglevel.
func (l *Logger) Info(v ...any) {
	l.writef(l.depth, INFO, "", v)
}

// Warn logs interface in warning loglevel
func (l *Logger) Warn(v ...any) {
	l.writef(l.depth, WARN, "", v)
}

// Error logs interface in Error loglevel
func (l *Logger) Error(v ...any) {
	l.writef(l.depth, ERROR, "", v)
}

// Debugf logs interface in debug loglevel with formating string
func (l *Logger) Debugf(format string, v ...any) {
	l.writef(l.depth, DEBUG, format, v)
}

// Infof logs interface in Infof loglevel with formating string
func (l *Logger) Infof(format string, v ...any) {
	l.writef(l.depth, INFO, format, v)
}

// Warnf logs interface in warning loglevel with formating string
func (l *Logger) Warnf(format string, v ...any) {
	l.writef(l.depth, WARN, format, v)
}

// Errorf logs interface in Error loglevel with formating string
func (l *Logger) Errorf(format string, v ...any) {
	l.writef(l.depth, ERROR, format, v)
}

// WriteLog write log into log files ignore the log level and log prefix
func (l *Logger) WriteLog(msg []byte) {
	logQueue <- &logValue{value: msg, writer: l.writer}
}

// FlushLogger flushs all log to disk.
func FlushLogger() {
	syncCancel()
	select {
	case <-time.After(waitFlushTimeout):
	case <-asyncDone.Done():
	}
}

// String turns the LogLevel to string.
func (lv *LogLevel) string() string {
	switch *lv {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Colored enable colored level string when use console writer
func (lv *LogLevel) coloredString() string {
	switch *lv {
	case DEBUG:
		return "\x1b[34mDEBUG\x1b[0m" //blue
	case INFO:
		return "\x1b[32mINFO\x1b[0m" //green
	case WARN:
		return "\x1b[33mWARN\x1b[0m" // yellow
	case ERROR:
		return "\x1b[31mERROR\x1b[0m" //cred
	default:
		return "\x1b[37mUNKNOWN\x1b[0m" // white
	}
}

func (l *Logger) setFileRoller(logpath string, num int, sizeMB int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		panic(err)
	}
	w := newRollFileWriter(logpath, l.name, num, sizeMB)
	l.writer = w
	return nil
}

func (l *Logger) setDayRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := newDateWriter(logpath, l.name, day, num)
	l.writer = w
	return nil
}

func (l *Logger) setHourRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := newDateWriter(logpath, l.name, hour, num)
	l.writer = w
	return nil
}

// setConsole sets the logger with console writer.
func (l *Logger) setConsole() {
	l.writer = &consoleWriter{}
}

func (l *Logger) writef(depth int, level LogLevel, format string, v []any) {
	if level < logLevel {
		return
	}

	buf := bytes.NewBuffer(nil)
	if l.writer.NeedPrefix() {
		fmt.Fprintf(buf, "%s|", time.Now().Format("2006-01-02 15:04:05.000"))

		if callerFlag {
			pc, file, line, ok := runtime.Caller(depth + callerSkip)
			if !ok {
				file = "???"
				line = 0
			} else {
				file = filepath.Base(file)
			}
			fmt.Fprintf(buf, "%s:%s:%d|", file, getFuncName(runtime.FuncForPC(pc).Name()), line)
		}
		if colored && logMode == Console {
			buf.WriteString(level.coloredString())
		} else {
			buf.WriteString(level.string())
		}
		buf.WriteByte('|')
	}

	if format == "" {
		fmt.Fprint(buf, v...)
	} else {
		fmt.Fprintf(buf, format, v...)
	}
	if l.writer.NeedPrefix() {
		buf.WriteByte('\n')
	}
	logQueue <- &logValue{value: buf.Bytes(), writer: l.writer}
}

func getFuncName(name string) string {
	idx := strings.LastIndexByte(name, '/')
	if idx != -1 {
		name = name[idx:]
		idx = strings.IndexByte(name, '.')
		if idx != -1 {
			name = strings.TrimPrefix(name[idx:], ".")
		}
	}
	return name
}

func flushLog() {
	for {
		select {
		case v := <-logQueue:
			v.writer.Write(v.value)
		default:
			select {
			case v := <-logQueue:
				v.writer.Write(v.value)
			case <-syncDone.Done():
				asyncCancel()
				return
			}
		}
	}
}
