package go_websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	Log *Logger = DefaultLogger()
)

type LogLevel int8

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelPanic
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	case LogLevelPanic:
		return "panic"
	}
	return ""
}

type LogFields map[string]interface{}

type Logger struct {
	log     *log.Logger
	ctx     context.Context
	fields  LogFields
	callers []string
}

func DefaultLogger() *Logger {
	return NewLogger(os.Stdout, "", log.LstdFlags)
}

func NewLogger(out io.Writer, prefix string, flag int) *Logger {
	logger := log.New(out, prefix, flag)
	return &Logger{log: logger}
}

// 克隆实例
func (l *Logger) clone() *Logger {
	cl := *l
	return &cl
}

// 添加字段
func (l *Logger) WithFields(f LogFields) *Logger {
	nl := l.clone()
	if nl.fields == nil {
		nl.fields = make(LogFields)
	}
	for k, v := range f {
		nl.fields[k] = v
	}
	return nl
}

// 添加上下文
func (l *Logger) WithContext(ctx context.Context) *Logger {
	nl := l.clone()
	nl.ctx = ctx
	return nl
}

// 添加某一层的调用栈信息
func (l *Logger) WithCaller(skip int) *Logger {
	nl := l.clone()
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		forPC := runtime.FuncForPC(pc)
		nl.callers = []string{
			fmt.Sprintf("%s: %d %s", file, line, forPC.Name()),
		}
	}
	return nl
}

// 添加整个调用栈信息
func (l *Logger) WithCallersFrames() *Logger {
	maxCallerDepth := 25
	minCallerDepth := 1
	callers := make([]string, 0)
	pcs := make([]uintptr, maxCallerDepth)
	depth := runtime.Callers(minCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		s := fmt.Sprintf("%s: %d %s", frame.File, frame.Line, frame.Function)
		callers = append(callers, s)
		if !more {
			break
		}
	}
	nl := l.clone()
	nl.callers = callers
	return nl
}

// JSON格式化
func (l *Logger) JSONFormat(level LogLevel, msg string) map[string]interface{} {
	data := make(LogFields, len(l.fields)+4)
	data["level"] = level.String()
	data["time"] = time.Now().Local().Format("2006-01-02 15:04:05")
	data["msg"] = msg
	data["callers"] = l.callers
	if len(l.fields) > 0 {
		for k, v := range l.fields {
			if _, ok := data[k]; !ok {
				data[k] = v
			}
		}
	}
	return data
}

// 输出
func (l *Logger) Output(level LogLevel, msg string) {
	body, _ := json.Marshal(l.JSONFormat(level, msg))
	content := string(body)
	switch level {
	case LogLevelDebug:
		l.log.Print(content)
	case LogLevelInfo:
		l.log.Print(content)
	case LogLevelWarn:
		l.log.Print(content)
	case LogLevelError:
		l.log.Print(content)
	case LogLevelFatal:
		l.log.Fatal(content)
	case LogLevelPanic:
		l.log.Panic(content)
	}
}

func (l *Logger) Debug(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelDebug, fmt.Sprint(v...))
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelInfo, fmt.Sprint(v...))
}

func (l *Logger) Infof(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelWarn, fmt.Sprint(v...))
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelError, fmt.Sprint(v...))
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelError, fmt.Sprintf(format, v...))
}

func (l *Logger) Fatal(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelFatal, fmt.Sprint(v...))
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelFatal, fmt.Sprintf(format, v...))
}

func (l *Logger) Panic(ctx context.Context, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelPanic, fmt.Sprint(v...))
}

func (l *Logger) Panicf(ctx context.Context, format string, v ...interface{}) {
	l.WithContext(ctx).Output(LogLevelPanic, fmt.Sprintf(format, v...))
}
