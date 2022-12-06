package logger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DEBUG loglevel
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	OFF
)

var (
	logLevel = DEBUG
	colored  = false

	logQueue    = make(chan *logValue, 10000)
	loggerMutex sync.Mutex
	loggerMap   = make(map[string]*Logger)
	writeDone   = make(chan bool)
	callerSkip  = 2
	callerFlag  = true

	waitFlushTimeout       = time.Second
	syncDone, syncCancel   = context.WithCancel(context.Background())
	asyncDone, asyncCancel = context.WithCancel(context.Background())
)

// Logger is the struct with name and wirter.
type Logger struct {
	name   string
	writer LogWriter
	depth  int
}

// LogLevel is uint8 type
type LogLevel uint8

type logValue struct {
	level  LogLevel
	value  []byte
	fileNo string
	writer LogWriter
}

func init() {
	go flushLog()
}

// String turns the LogLevel to string.
func (lv *LogLevel) String() string {
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

func GetLogger(name string) *Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if lg, ok := loggerMap[name]; ok {
		return lg
	}
	lg := &Logger{
		name:   name,
		writer: &ConsoleWriter{},
	}
	loggerMap[name] = lg
	err := lg.SetFileRoller("./logs", 10, 100)
	if err != nil {
		return nil
	}
	return lg
}

func SetLevel(level LogLevel) {
	logLevel = level
}

func GetLogLevel() LogLevel {
	return logLevel
}

func GetLevel() string {
	return logLevel.String()
}

func Colored() {
	colored = true
}

func StringToLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return DEBUG
	}
}

func SetCallerSkip(skip int) {
	callerSkip = skip
}

func SetCallerFlag(flag bool) {
	callerFlag = flag
}

func (l *Logger) SetLogName(name string) {
	l.name = name
}

func (l *Logger) SetFileRoller(logpath string, num int, sizeMB int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		panic(err)
	}
	w := NewRollFileWriter(logpath, l.name, num, sizeMB)
	l.writer = w
	return nil
}

func (l *Logger) IsConsoleWriter() bool {
	if reflect.TypeOf(l.writer) == reflect.TypeOf(&ConsoleWriter{}) {
		return true
	}
	return false
}

func (l *Logger) SetWriter(w LogWriter) {
	l.writer = w
}

func (l *Logger) SetDayRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := NewDateWriter(logpath, l.name, DAY, num)
	l.writer = w
	return nil
}

func (l *Logger) SetHourRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := NewDateWriter(logpath, l.name, HOUR, num)
	l.writer = w
	return nil
}

// SetConsole sets the logger with console writer.
func (l *Logger) SetConsole() {
	l.writer = &ConsoleWriter{}
}

func (l *Logger) SetDepth(in int) {
	l.depth = in
}

// Debug logs interface in debug loglevel.
func (l *Logger) Debug(v ...any) {
	l.Writef(l.depth, DEBUG, "", v)
}

// Info logs interface in Info loglevel.
func (l *Logger) Info(v ...any) {
	l.Writef(l.depth, INFO, "", v)
}

// Warn logs interface in warning loglevel
func (l *Logger) Warn(v ...any) {
	l.Writef(l.depth, WARN, "", v)
}

// Error logs interface in Error loglevel
func (l *Logger) Error(v ...any) {
	l.Writef(l.depth, ERROR, "", v)
}

// Debugf logs interface in debug loglevel with formating string
func (l *Logger) Debugf(format string, v ...any) {
	l.Writef(l.depth, DEBUG, format, v)
}

// Infof logs interface in Infof loglevel with formating string
func (l *Logger) Infof(format string, v ...any) {
	l.Writef(l.depth, INFO, format, v)
}

// Warnf logs interface in warning loglevel with formating string
func (l *Logger) Warnf(format string, v ...any) {
	l.Writef(l.depth, WARN, format, v)
}

// Errorf logs interface in Error loglevel with formating string
func (l *Logger) Errorf(format string, v ...any) {
	l.Writef(l.depth, ERROR, format, v)
}

func (l *Logger) Writef(depth int, level LogLevel, format string, v []any) {
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
		if colored && l.IsConsoleWriter() {
			buf.WriteString(level.coloredString())
		} else {
			buf.WriteString(level.String())
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
