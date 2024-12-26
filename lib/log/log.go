package log

import (
	"fmt"
	"github.com/getevo/json"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Level represents the severity level of the log message.
// It is defined as an integer type and uses constants for named levels.
type Level int

// wd stores the current working directory path.
// Used to shorten file paths in log entries.
var wd, _ = os.Getwd()

// level stores the current global logging level.
// Only messages at or above this level will be logged.
var level = WarningLevel

// stackTraceLevel indicates how many additional stack frames to skip when determining the caller location.
var stackTraceLevel = 0

// writers holds a list of functions that process log entries.
// Each writer is a function that takes an *Entry and outputs it (e.g., to console, file, etc.).
var writers []func(log *Entry) = []func(log *Entry){
	StdWriter,
}

// StdWriter is a default writer function that prints the log message to stdout.
var StdWriter = func(log *Entry) {
	fmt.Println(log.Date.Format("15:04:05"), "["+log.Level+"]", log.File+":"+strconv.Itoa(log.Line), log.Message)
}

// levels maps the Level constants to their string representations.
// The index corresponds to the Level value (with a placeholder at index 0).
var levels = []string{"", "Critical", "Error", "Warning", "Notice", "Info", "Debug"}

// Logging levels as constants. Each constant represents a severity level.
const (
	CriticalLevel Level = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

// ParseLevel converts a string expression (like "error", "warn") into a corresponding Level constant.
// If it cannot find a match, it defaults to NoticeLevel.
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

// AddWriter adds one or more writer functions to the global list of writers.
// These writers are called for each log message.
func AddWriter(input ...func(message *Entry)) {
	writers = append(writers, input...)
}

// SetWriters replaces the current list of writers with the provided ones.
// It sets the global writers slice to only the specified functions.
func SetWriters(input ...func(message *Entry)) {
	writers = input
}

// SetLevel updates the global logging level.
// Messages below this level will not be output.
func SetLevel(lvl Level) {
	level = lvl
}

// SetStackTrace configures how many additional stack frames are skipped
// when determining the source file and line number for log messages.
func SetStackTrace(lvl int) {
	stackTraceLevel = lvl
}

// Entry represents a single log message, including metadata such as timestamp, file, line number, and severity level.
type Entry struct {
	Level   string    `json:"level"`   // The severity level as a string (e.g., "Error", "Info")
	Date    time.Time `json:"date"`    // The timestamp when this log entry was created
	File    string    `json:"file"`    // The source file that generated the log entry
	Line    int       `json:"line"`    // The line number in the source file
	Message string    `json:"message"` // The formatted log message
}

// msg is an internal function that creates a log Entry and passes it to all configured writers.
// It uses runtime.Caller to determine file and line number and applies message formatting.
func msg(message any, level Level, params ...any) {
	if message == nil {
		return
	}
	_, file, line, _ := runtime.Caller(2 + stackTraceLevel)

	entry := Entry{
		Level:   levels[level],
		Date:    time.Now(),
		File:    file,
		Line:    line,
		Message: fmt.Sprintf(fmt.Sprint(message), params...),
	}

	for _, writer := range writers {
		writer(&entry)
	}
}

// toValue converts various parameter types into a string representation.
// For strings, it wraps the value in quotes, and for complex types, it attempts JSON marshaling.
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

// quote wraps a string in double quotes and escapes any existing double quotes within the string.
func quote(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
}

// Fatal logs a message at the Critical level and then exits the program.
// It behaves like Critical but calls os.Exit(1) after logging.
func Fatal(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// FatalF logs a formatted message at the Critical level and then exits the program.
// It uses formatting similar to fmt.Printf before logging.
func FatalF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// Fatalf logs a formatted message at the Critical level and then exits the program.
// It is similar to FatalF, provided for compatibility.
func Fatalf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// Panic logs a message at the Critical level and then calls panic().
// It behaves like Critical but ends with a panic.
func Panic(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// PanicF logs a formatted message at the Critical level and then calls panic().
// It uses formatting similar to fmt.Printf before logging.
func PanicF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// Panicf logs a formatted message at the Critical level and then calls panic().
// It behaves like PanicF, provided for compatibility.
func Panicf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
	os.Exit(128)
}

// Critical logs a message at the Critical level.
// These messages typically indicate severe errors or conditions.
func Critical(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
}

// CriticalF logs a formatted message at the Critical level using fmt.Printf-style formatting.
func CriticalF(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
}

// Criticalf logs a formatted message at the Critical level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like CriticalF.
func Criticalf(message any, params ...any) {
	if level >= CriticalLevel {
		msg(message, CriticalLevel, params...)
	}
}

// Error logs a message at the Error level.
// These messages represent errors that may require attention, but are not as severe as Critical.
func Error(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}
}

// ErrorF logs a formatted message at the Error level using fmt.Printf-style formatting.
func ErrorF(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}
}

// Errorf logs a formatted message at the Error level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like ErrorF.
func Errorf(message any, params ...any) {
	if level >= ErrorLevel {
		msg(message, ErrorLevel, params...)
	}
}

// Warning logs a message at the Warning level.
// Warnings indicate potentially problematic situations that deserve attention.
func Warning(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// WarningF logs a formatted message at the Warning level using fmt.Printf-style formatting.
func WarningF(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// Warningf logs a formatted message at the Warning level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like WarningF.
func Warningf(message any, params ...any) {
	if level >= WarningLevel {
		msg(message, WarningLevel, params...)
	}
}

// Notice logs a message at the Notice level.
// Notices provide informational messages that are more important than Info but not as severe as Warnings.
func Notice(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// NoticeF logs a formatted message at the Notice level using fmt.Printf-style formatting.
func NoticeF(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// Noticef logs a formatted message at the Notice level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like NoticeF.
func Noticef(message any, params ...any) {
	if level >= NoticeLevel {
		msg(message, NoticeLevel, params...)
	}
}

// Info logs a message at the Info level.
// Info messages provide general operational information.
func Info(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// InfoF logs a formatted message at the Info level using fmt.Printf-style formatting.
func InfoF(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// Infof logs a formatted message at the Info level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like InfoF.
func Infof(message any, params ...any) {
	if level >= InfoLevel {
		msg(message, InfoLevel, params...)
	}
}

// Debug logs a message at the Debug level.
// Debug messages are used for internal testing and troubleshooting.
func Debug(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}

// DebugF logs a formatted message at the Debug level using fmt.Printf-style formatting.
func DebugF(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}

// Debugf logs a formatted message at the Debug level using fmt.Printf-style formatting.
// Provided for compatibility; behaves like DebugF.
func Debugf(message any, params ...any) {
	if level >= DebugLevel {
		msg(message, DebugLevel, params...)
	}
}
