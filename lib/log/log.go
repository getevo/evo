package log

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Level represents the severity level of the log message.
type Level int32

// Logging levels as constants. Each constant represents a severity level.
const (
	CriticalLevel Level = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

// Format represents the log output format.
type Format int32

const (
	TextFormat Format = iota
	JSONFormat
)

// Field represents a structured key-value pair in a log entry.
type Field struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// Entry represents a single log message, including metadata such as timestamp, file, line number, and severity level.
type Entry struct {
	Level   string    `json:"level"`   // The severity level as a string (e.g., "Error", "Info")
	Date    time.Time `json:"date"`    // The timestamp when this log entry was created
	File    string    `json:"file"`    // The source file that generated the log entry
	Line    int       `json:"line"`    // The line number in the source file
	Message string    `json:"message"` // The formatted log message
	Fields  []Field   `json:"fields,omitempty"`
}

type contextKey struct{}

var (
	wd              string
	level           atomic.Int32
	format          atomic.Int32
	stackTraceLevel atomic.Int32
	writersMu       sync.RWMutex
	writers         []func(*Entry)
	levels          = []string{"", "Critical", "Error", "Warning", "Notice", "Info", "Debug"}
)

// StdWriter is the default log writer that outputs to stdout.
var StdWriter = func(entry *Entry) {
	if Format(format.Load()) == JSONFormat {
		data, _ := json.Marshal(entry)
		fmt.Fprintln(os.Stdout, string(data))
		return
	}
	var b strings.Builder
	b.WriteString(entry.Date.Format("15:04:05"))
	b.WriteString(" [")
	b.WriteString(entry.Level)
	b.WriteString("] ")
	b.WriteString(entry.File)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(entry.Line))
	b.WriteString(" ")
	b.WriteString(entry.Message)
	for _, f := range entry.Fields {
		b.WriteString(" ")
		b.WriteString(f.Key)
		b.WriteString("=")
		b.WriteString(formatValue(f.Value))
	}
	fmt.Fprintln(os.Stdout, b.String())
}

func init() {
	wd, _ = os.Getwd()
	level.Store(int32(WarningLevel))
	format.Store(int32(TextFormat))
	writers = []func(*Entry){StdWriter}
}

// --- Configuration ---

// ParseLevel converts a string expression (like "error", "warn") into a corresponding Level constant.
// If it cannot find a match, it defaults to NoticeLevel.
func ParseLevel(expr string) Level {
	switch strings.TrimSpace(strings.ToLower(expr)) {
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

// SetLevel updates the global logging level.
// Messages below this level will not be output.
func SetLevel(lvl Level) {
	level.Store(int32(lvl))
}

// GetLevel returns the current global logging level.
func GetLevel() Level {
	return Level(level.Load())
}

// SetFormat sets the global log output format (TextFormat or JSONFormat).
func SetFormat(f Format) {
	format.Store(int32(f))
}

// GetFormat returns the current global log output format.
func GetFormat() Format {
	return Format(format.Load())
}

// SetStackTrace configures how many additional stack frames are skipped
// when determining the source file and line number for log messages.
func SetStackTrace(lvl int) {
	stackTraceLevel.Store(int32(lvl))
}

// AddWriter adds one or more writer functions to the global list of writers.
func AddWriter(input ...func(*Entry)) {
	writersMu.Lock()
	writers = append(writers, input...)
	writersMu.Unlock()
}

// SetWriters replaces the current list of writers with the provided ones.
func SetWriters(input ...func(*Entry)) {
	writersMu.Lock()
	writers = input
	writersMu.Unlock()
}

// --- Context ---

// WithContextFields returns a new context with the given key-value fields attached for logging.
func WithContextFields(ctx context.Context, keysAndValues ...any) context.Context {
	existing := ContextFields(ctx)
	fields := append(existing, parseFields(keysAndValues)...)
	return context.WithValue(ctx, contextKey{}, fields)
}

// ContextFields returns the log fields stored in the context.
func ContextFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	if fields, ok := ctx.Value(contextKey{}).([]Field); ok {
		cp := make([]Field, len(fields))
		copy(cp, fields)
		return cp
	}
	return nil
}

// --- Core ---

func doLog(lvl Level, skip int, ctx context.Context, message string, fields []Field) {
	if lvl > Level(level.Load()) {
		return
	}
	_, file, line, _ := runtime.Caller(skip + int(stackTraceLevel.Load()))
	file = shortenPath(file)

	entry := &Entry{
		Level:   levels[lvl],
		Date:    time.Now(),
		File:    file,
		Line:    line,
		Message: message,
		Fields:  fields,
	}

	if ctx != nil {
		if ctxFields := ContextFields(ctx); len(ctxFields) > 0 {
			entry.Fields = append(ctxFields, entry.Fields...)
		}
	}

	writersMu.RLock()
	ws := make([]func(*Entry), len(writers))
	copy(ws, writers)
	writersMu.RUnlock()

	for _, w := range ws {
		w(entry)
	}
}

// msg handles structured logging with key-value params.
func msg(message any, lvl Level, params ...any) {
	if message == nil {
		return
	}
	doLog(lvl, 3, nil, fmt.Sprint(message), parseFields(params))
}

// msgf handles printf-style formatting.
func msgf(message any, lvl Level, params ...any) {
	if message == nil {
		return
	}
	doLog(lvl, 3, nil, fmt.Sprintf(fmt.Sprint(message), params...), nil)
}

// msgCtx handles context-aware structured logging.
func msgCtx(ctx context.Context, message any, lvl Level, params ...any) {
	if message == nil {
		return
	}
	doLog(lvl, 3, ctx, fmt.Sprint(message), parseFields(params))
}

// --- Helpers ---

func parseFields(params []any) []Field {
	if len(params) == 0 {
		return nil
	}
	fields := make([]Field, 0, (len(params)+1)/2)
	i := 0
	for i < len(params) {
		if key, ok := params[i].(string); ok {
			if i+1 < len(params) {
				fields = append(fields, Field{Key: key, Value: params[i+1]})
				i += 2
			} else {
				fields = append(fields, Field{Key: key, Value: nil})
				i++
			}
		} else {
			fields = append(fields, Field{Key: "!BADKEY", Value: params[i]})
			i++
		}
	}
	return fields
}

func formatValue(v any) string {
	if v == nil {
		return "nil"
	}
	ref := reflect.ValueOf(v)
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			return "nil"
		}
		ref = ref.Elem()
	}
	switch ref.Kind() {
	case reflect.String:
		return strconv.Quote(ref.String())
	case reflect.Bool:
		return strconv.FormatBool(ref.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(ref.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(ref.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(ref.Float(), 'f', -1, 64)
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Array:
		data, err := json.Marshal(ref.Interface())
		if err != nil {
			return fmt.Sprint(ref.Interface())
		}
		return string(data)
	default:
		return fmt.Sprint(ref.Interface())
	}
}

func shortenPath(file string) string {
	if wd != "" {
		if rel, err := filepath.Rel(wd, file); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return file
}

// --- Logging functions ---

// Critical logs a message at the Critical level with optional key-value fields.
func Critical(message any, params ...any) { msg(message, CriticalLevel, params...) }

// Criticalf logs a printf-formatted message at the Critical level.
func Criticalf(message any, params ...any) { msgf(message, CriticalLevel, params...) }

// CriticalF is an alias for Criticalf.
func CriticalF(message any, params ...any) { msgf(message, CriticalLevel, params...) }

// CriticalContext logs at Critical level with context fields.
func CriticalContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, CriticalLevel, params...)
}

// Error logs a message at the Error level with optional key-value fields.
func Error(message any, params ...any) { msg(message, ErrorLevel, params...) }

// Errorf logs a printf-formatted message at the Error level.
func Errorf(message any, params ...any) { msgf(message, ErrorLevel, params...) }

// ErrorF is an alias for Errorf.
func ErrorF(message any, params ...any) { msgf(message, ErrorLevel, params...) }

// ErrorContext logs at Error level with context fields.
func ErrorContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, ErrorLevel, params...)
}

// Warning logs a message at the Warning level with optional key-value fields.
func Warning(message any, params ...any) { msg(message, WarningLevel, params...) }

// Warningf logs a printf-formatted message at the Warning level.
func Warningf(message any, params ...any) { msgf(message, WarningLevel, params...) }

// WarningF is an alias for Warningf.
func WarningF(message any, params ...any) { msgf(message, WarningLevel, params...) }

// WarningContext logs at Warning level with context fields.
func WarningContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, WarningLevel, params...)
}

// Notice logs a message at the Notice level with optional key-value fields.
func Notice(message any, params ...any) { msg(message, NoticeLevel, params...) }

// Noticef logs a printf-formatted message at the Notice level.
func Noticef(message any, params ...any) { msgf(message, NoticeLevel, params...) }

// NoticeF is an alias for Noticef.
func NoticeF(message any, params ...any) { msgf(message, NoticeLevel, params...) }

// NoticeContext logs at Notice level with context fields.
func NoticeContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, NoticeLevel, params...)
}

// Info logs a message at the Info level with optional key-value fields.
func Info(message any, params ...any) { msg(message, InfoLevel, params...) }

// Infof logs a printf-formatted message at the Info level.
func Infof(message any, params ...any) { msgf(message, InfoLevel, params...) }

// InfoF is an alias for Infof.
func InfoF(message any, params ...any) { msgf(message, InfoLevel, params...) }

// InfoContext logs at Info level with context fields.
func InfoContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, InfoLevel, params...)
}

// Debug logs a message at the Debug level with optional key-value fields.
func Debug(message any, params ...any) { msg(message, DebugLevel, params...) }

// Debugf logs a printf-formatted message at the Debug level.
func Debugf(message any, params ...any) { msgf(message, DebugLevel, params...) }

// DebugF is an alias for Debugf.
func DebugF(message any, params ...any) { msgf(message, DebugLevel, params...) }

// DebugContext logs at Debug level with context fields.
func DebugContext(ctx context.Context, message any, params ...any) {
	msgCtx(ctx, message, DebugLevel, params...)
}

// Fatal logs a message at the Critical level and exits with code 1.
func Fatal(message any, params ...any) {
	msg(message, CriticalLevel, params...)
	os.Exit(1)
}

// Fatalf logs a printf-formatted message at the Critical level and exits with code 1.
func Fatalf(message any, params ...any) {
	msgf(message, CriticalLevel, params...)
	os.Exit(1)
}

// FatalF is an alias for Fatalf.
func FatalF(message any, params ...any) {
	msgf(message, CriticalLevel, params...)
	os.Exit(1)
}

// Panic logs a message at the Critical level and panics.
func Panic(message any, params ...any) {
	msg(message, CriticalLevel, params...)
	panic(fmt.Sprint(message))
}

// Panicf logs a printf-formatted message at the Critical level and panics.
func Panicf(message any, params ...any) {
	msgf(message, CriticalLevel, params...)
	panic(fmt.Sprintf(fmt.Sprint(message), params...))
}

// PanicF is an alias for Panicf.
func PanicF(message any, params ...any) {
	msgf(message, CriticalLevel, params...)
	panic(fmt.Sprintf(fmt.Sprint(message), params...))
}

// --- Scoped logger ---

// Logger is a scoped logger with pre-attached fields.
// Use WithField or WithFields to create one, then chain further calls or call a log-level method.
type Logger struct {
	fields []Field
}

// WithField returns a Logger with a single key-value field pre-attached.
// The returned Logger can be chained with additional WithField / WithFields calls.
func WithField(key string, value any) *Logger {
	return &Logger{fields: []Field{{Key: key, Value: value}}}
}

// WithFields returns a Logger with all entries from the given map pre-attached as fields.
func WithFields(fields map[string]any) *Logger {
	f := make([]Field, 0, len(fields))
	for k, v := range fields {
		f = append(f, Field{Key: k, Value: v})
	}
	return &Logger{fields: f}
}

// WithField returns a new Logger with an additional key-value field merged in.
func (l *Logger) WithField(key string, value any) *Logger {
	fields := make([]Field, len(l.fields)+1)
	copy(fields, l.fields)
	fields[len(l.fields)] = Field{Key: key, Value: value}
	return &Logger{fields: fields}
}

// WithFields returns a new Logger with additional map fields merged in.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	merged := make([]Field, len(l.fields), len(l.fields)+len(fields))
	copy(merged, l.fields)
	for k, v := range fields {
		merged = append(merged, Field{Key: k, Value: v})
	}
	return &Logger{fields: merged}
}

// lMsg is the shared structured helper.
// Call chain: public method → lMsg → doLog  (skip=3 puts Caller at the public method's caller).
func (l *Logger) lMsg(lvl Level, ctx context.Context, message any, params []any) {
	if message == nil {
		return
	}
	extra := parseFields(params)
	all := make([]Field, 0, len(l.fields)+len(extra))
	all = append(all, l.fields...)
	all = append(all, extra...)
	doLog(lvl, 3, ctx, fmt.Sprint(message), all)
}

// lMsgf is the shared printf helper.
func (l *Logger) lMsgf(lvl Level, message any, params []any) {
	if message == nil {
		return
	}
	doLog(lvl, 3, nil, fmt.Sprintf(fmt.Sprint(message), params...), l.fields)
}

func (l *Logger) Critical(message any, params ...any) { l.lMsg(CriticalLevel, nil, message, params) }
func (l *Logger) Criticalf(message any, params ...any) { l.lMsgf(CriticalLevel, message, params) }
func (l *Logger) CriticalF(message any, params ...any)  { l.lMsgf(CriticalLevel, message, params) }
func (l *Logger) CriticalContext(ctx context.Context, message any, params ...any) {
	l.lMsg(CriticalLevel, ctx, message, params)
}

func (l *Logger) Error(message any, params ...any) { l.lMsg(ErrorLevel, nil, message, params) }
func (l *Logger) Errorf(message any, params ...any) { l.lMsgf(ErrorLevel, message, params) }
func (l *Logger) ErrorF(message any, params ...any)  { l.lMsgf(ErrorLevel, message, params) }
func (l *Logger) ErrorContext(ctx context.Context, message any, params ...any) {
	l.lMsg(ErrorLevel, ctx, message, params)
}

func (l *Logger) Warning(message any, params ...any) { l.lMsg(WarningLevel, nil, message, params) }
func (l *Logger) Warningf(message any, params ...any) { l.lMsgf(WarningLevel, message, params) }
func (l *Logger) WarningF(message any, params ...any)  { l.lMsgf(WarningLevel, message, params) }
func (l *Logger) WarningContext(ctx context.Context, message any, params ...any) {
	l.lMsg(WarningLevel, ctx, message, params)
}

func (l *Logger) Notice(message any, params ...any) { l.lMsg(NoticeLevel, nil, message, params) }
func (l *Logger) Noticef(message any, params ...any) { l.lMsgf(NoticeLevel, message, params) }
func (l *Logger) NoticeF(message any, params ...any)  { l.lMsgf(NoticeLevel, message, params) }
func (l *Logger) NoticeContext(ctx context.Context, message any, params ...any) {
	l.lMsg(NoticeLevel, ctx, message, params)
}

func (l *Logger) Info(message any, params ...any) { l.lMsg(InfoLevel, nil, message, params) }
func (l *Logger) Infof(message any, params ...any) { l.lMsgf(InfoLevel, message, params) }
func (l *Logger) InfoF(message any, params ...any)  { l.lMsgf(InfoLevel, message, params) }
func (l *Logger) InfoContext(ctx context.Context, message any, params ...any) {
	l.lMsg(InfoLevel, ctx, message, params)
}

func (l *Logger) Debug(message any, params ...any) { l.lMsg(DebugLevel, nil, message, params) }
func (l *Logger) Debugf(message any, params ...any) { l.lMsgf(DebugLevel, message, params) }
func (l *Logger) DebugF(message any, params ...any)  { l.lMsgf(DebugLevel, message, params) }
func (l *Logger) DebugContext(ctx context.Context, message any, params ...any) {
	l.lMsg(DebugLevel, ctx, message, params)
}

func (l *Logger) Fatal(message any, params ...any) {
	l.lMsg(CriticalLevel, nil, message, params)
	os.Exit(1)
}
func (l *Logger) Fatalf(message any, params ...any) {
	l.lMsgf(CriticalLevel, message, params)
	os.Exit(1)
}
func (l *Logger) Panic(message any, params ...any) {
	l.lMsg(CriticalLevel, nil, message, params)
	panic(fmt.Sprint(message))
}
func (l *Logger) Panicf(message any, params ...any) {
	l.lMsgf(CriticalLevel, message, params)
	panic(fmt.Sprintf(fmt.Sprint(message), params...))
}
