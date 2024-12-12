package log

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/text"
	"github.com/getevo/json"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// LogLevel type
type Level int

var wd, _ = os.Getwd()
var level = WarningLevel
var stackTraceLevel = 0
var writers []func(log string) = []func(log string){
	stdWriter,
}

var Serializer = func(v *Entry) string { return text.ToJSON(v) }

func stdWriter(log string) {
	fmt.Println(log)
}

var levels = []string{"", "Critical", "Error", "Warning", "Notice", "Info", "Debug"}

const (
	CriticalLevel Level = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

func ParseLevel(expr string) Level {
	expr = strings.TrimSpace(strings.ToLower(expr))
	switch expr {
	case "critical", "crit":
		return CriticalLevel
	case "error", "erro":
		return ErrorLevel
	case "warning", "warn":
		return WarningLevel
	case "notice", "noti":
		return NoticeLevel
	case "info":
		return InfoLevel
	case "debug", "debu":
		return DebugLevel
	default:
		return NoticeLevel
	}
}

func AddWriter(input ...func(message string)) {
	writers = append(writers, input...)
}

func SetWriters(input ...func(message string)) {
	writers = input
}

func SetLevel(lvl Level) {
	level = lvl
}

func SetStackTrace(lvl int) {
	stackTraceLevel = lvl
}

type Entry struct {
	Level   string `json:"level"`
	Date    string `json:"date"`
	File    string `json:"file"`
	Message string `json:"message"`
}

func msg(message any, level Level, params ...any) {
	if message == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2)
	entry := Entry{
		Level:   levels[level],
		Date:    time.Now().Format("2006-01-02 15:04:05"),
		File:    file[len(wd)+1:] + ":" + strconv.Itoa(line),
		Message: fmt.Sprintf(fmt.Sprint(message), params...),
	}

	for _, writer := range writers {
		writer(Serializer(&entry))
	}
}

func toValue(param any) string {
	var ref = reflect.ValueOf(param)
	for ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	switch ref.Kind() {
	case reflect.String:
		return strconv.Quote(ref.Interface().(string))
	case reflect.Bool:
		if ref.Interface().(bool) {
			return "true"
		} else {
			return "false"
		}
	case reflect.Float64, reflect.Float32, reflect.Int, reflect.Int64, reflect.Int32, reflect.Int8, reflect.Uint16, reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint8:
		return fmt.Sprint(ref.Interface())
	case reflect.Array, reflect.Slice, reflect.Struct, reflect.Map:
		var b, _ = json.Marshal(ref.Interface())

		return string(b)
	default:
		return quote(fmt.Sprint(ref.Interface()))
	}
}

func quote(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
}

// Fatal is just like func l.Critical logger except that it is followed by exit to program
func Fatal(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(1)
}

// FatalF is just like func l.CriticalF logger except that it is followed by exit to program
func FatalF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(1)
}

// FatalF is just like func l.CriticalF logger except that it is followed by exit to program
func Fatalf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(1)
}

// Panic is just like func l.Critical except that it is followed by a call to panic
func Panic(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	panic(msg)
}

// PanicF is just like func l.CriticalF except that it is followed by a call to panic
func PanicF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	panic(msg)
}

// PanicF is just like func l.CriticalF except that it is followed by a call to panic
func Panicf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	panic(msg)
}

// Critical logs a message at a Critical Level
func Critical(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
}

// CriticalF logs a message at Critical level using the same syntax and options as fmt.Printf
func CriticalF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
}

// CriticalF logs a message at Critical level using the same syntax and options as fmt.Printf
func Criticalf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}

}

// Error logs a message at Error level
func Error(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}

}

// ErrorF logs a message at Error level using the same syntax and options as fmt.Printf
func ErrorF(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}
}

// ErrorF logs a message at Error level using the same syntax and options as fmt.Printf
func Errorf(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}
}

// Warning logs a message at Warning level
func Warning(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// WarningF logs a message at Warning level using the same syntax and options as fmt.Printf
func WarningF(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// WarningF logs a message at Warning level using the same syntax and options as fmt.Printf
func Warningf(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// Notice logs a message at Notice level
func Notice(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// NoticeF logs a message at Notice level using the same syntax and options as fmt.Printf
func NoticeF(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// NoticeF logs a message at Notice level using the same syntax and options as fmt.Printf
func Noticef(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// Info logs a message at Info level
func Info(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// InfoF logs a message at Info level using the same syntax and options as fmt.Printf
func InfoF(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// InfoF logs a message at Info level using the same syntax and options as fmt.Printf
func Infof(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// Debug logs a message at Debug level
func Debug(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}

// DebugF logs a message at Debug level using the same syntax and options as fmt.Printf
func DebugF(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}

// DebugF logs a message at Debug level using the same syntax and options as fmt.Printf
func Debugf(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}
