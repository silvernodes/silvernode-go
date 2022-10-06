package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	DEBUG int = 0
	INFO      = 1
	WARN      = 2
	ERROR     = 3
	FATAL     = 4
)

func LevelToString(lv int) string {
	if lv == DEBUG {
		return "DEBUG"
	} else if lv == INFO {
		return "INFO"
	} else if lv == WARN {
		return "WARN"
	} else if lv == ERROR {
		return "ERROR"
	} else {
		return "FATAL"
	}
}

func StringToLevel(lv string) int {
	infos := strings.Split(lv, "-")
	if len(infos) <= 0 {
		return DEBUG
	}
	var lvl int = DEBUG
	switch infos[0] {
	case "DEBUG":
		lvl = DEBUG
	case "INFO":
		lvl = INFO
	case "WARNING":
		lvl = WARN
	case "ERROR":
		lvl = ERROR
	case "FATAL":
		lvl = FATAL
	default:
		lvl = DEBUG
	}
	return lvl
}

type Logger struct {
	category  string
	level     int
	logWriter LogWriter
}

type LogWriter interface {
	Write(info *LogInfo)
	Close()
}

func NewLogger(category string, loglevel int, logWriter LogWriter) *Logger {
	l := new(Logger)
	l.category = category
	l.level = loglevel
	if logWriter != nil {
		l.logWriter = logWriter
	} else {
		l.logWriter = NewConsoleLogWriter()
	}
	return l
}

func (l *Logger) BindWriter(writer LogWriter) {
	if l.logWriter != nil {
		l.logWriter.Close()
	}
	l.logWriter = writer
}

func (l *Logger) Source(callstack int) string {
	if callstack < 0 {
		return ""
	}
	src := ""
	if callstack+1 >= 0 {
		pc, _, lineno, ok := runtime.Caller(callstack + 1)
		if ok {
			src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
		}
	}
	return src
}

func (l *Logger) doLog(lvl int, callstack int, any interface{}, args ...interface{}) {
	if lvl < l.level {
		return
	}
	src := l.Source(callstack + 1)
	var msg string = ""
	switch any.(type) {
	case string:
		msg = any.(string)
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
	case error:
		msg = any.(error).Error()
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
	default:
		msg = fmt.Sprint(any)
	}
	info := new(LogInfo)
	info.Category = l.category
	info.Level = lvl
	info.Message = msg
	info.Source = src
	info.SetCreated(time.Now())
	if lvl <= DEBUG {
		info.Println() // DEBUG always only to console
	} else {
		l.logWriter.Write(info)
	}
}

func (l *Logger) Log(lvl int, arg0 interface{}, args ...interface{}) {
	l.doLog(lvl, -99, arg0, args...)
}

func (l *Logger) Debug(arg0 interface{}, args ...interface{}) {
	l.doLog(DEBUG, 1, arg0, args...)
}

func (l *Logger) Info(arg0 interface{}, args ...interface{}) {
	l.doLog(INFO, 1, arg0, args...)
}

func (l *Logger) Warn(arg0 interface{}, args ...interface{}) {
	l.doLog(WARN, 1, arg0, args...)
}

func (l *Logger) Error(arg0 interface{}, args ...interface{}) {
	l.doLog(ERROR, 1, arg0, args...)
}

func (l *Logger) Fatal(arg0 interface{}, args ...interface{}) {
	l.doLog(FATAL, 1, arg0, args...)
}
