package file

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/getevo/evo/v2/lib/log"
)

// Config defines the file logger configuration.
type Config struct {
	Path       string                        // Directory for log files
	FileName   string                        // Filename template with wildcards (%y, %m, %d, %hh, %mm)
	Expiration time.Duration                 // Delete logs older than this duration (<=0 means no cleanup)
	MaxSize    int64                         // Max file size in bytes before rotation (0 = no size limit)
	LogFormat  func(entry *log.Entry) string // Custom format function
}

// FileLogger wraps a file-based logger with lifecycle management.
type FileLogger struct {
	*fileLogger
}

// fileLogger is the internal structure for the file logger.
type fileLogger struct {
	config      Config
	file        *os.File
	writer      *bufio.Writer
	filePath    string
	fileSize    int64
	mutex       sync.Mutex
	currentDate string
	done        chan struct{}
}

// NewFileLogger creates a file logger writer function.
func NewFileLogger(config ...Config) func(*log.Entry) {
	return newFileLogger(config...).writeLog
}

// NewFileLoggerWithHandle creates a file logger with lifecycle management (Flush, Close).
func NewFileLoggerWithHandle(config ...Config) *FileLogger {
	return &FileLogger{newFileLogger(config...)}
}

// Writer returns the writer function for use with log.AddWriter.
func (fl *FileLogger) Writer() func(*log.Entry) {
	return fl.writeLog
}

func newFileLogger(config ...Config) *fileLogger {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Path == "" {
		cfg.Path, _ = os.Getwd()
	}
	if cfg.FileName == "" {
		cfg.FileName = filepath.Base(os.Args[0]) + ".log"
	}
	if cfg.LogFormat == nil {
		cfg.LogFormat = defaultLogFormat
	}

	f := &fileLogger{
		config:      cfg,
		currentDate: time.Now().Format("2006-01-02"),
		done:        make(chan struct{}),
	}

	f.openLogFile()
	go f.startLogRotation()
	go f.startAutoFlush()

	return f
}

// openLogFile opens or creates the log file with append mode.
func (f *fileLogger) openLogFile() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.openLogFileLocked()
}

func (f *fileLogger) openLogFileLocked() {
	path := f.getFilePath()
	if path == f.filePath && f.file != nil {
		return
	}

	// Flush and close old file
	if f.writer != nil {
		f.writer.Flush()
	}
	if f.file != nil {
		f.file.Close()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "file logger: failed to create directory %s: %v\n", dir, err)
		return
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file logger: failed to open %s: %v\n", path, err)
		return
	}

	stat, _ := file.Stat()
	f.fileSize = stat.Size()
	f.file = file
	f.writer = bufio.NewWriterSize(file, 4096)
	f.filePath = path
}

// getFilePath generates the log file path with the current date/time.
func (f *fileLogger) getFilePath() string {
	now := time.Now()
	name := f.config.FileName
	// Replace longer patterns first to avoid partial matches
	name = strings.ReplaceAll(name, "%mm", now.Format("04"))
	name = strings.ReplaceAll(name, "%hh", now.Format("15"))
	name = strings.ReplaceAll(name, "%y", now.Format("2006"))
	name = strings.ReplaceAll(name, "%m", now.Format("01"))
	name = strings.ReplaceAll(name, "%d", now.Format("02"))
	return filepath.Join(f.config.Path, name)
}

// writeLog safely writes the log entry to the file.
func (f *fileLogger) writeLog(entry *log.Entry) {
	line := f.config.LogFormat(entry) + "\n"

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.writer == nil {
		return
	}

	n, err := f.writer.WriteString(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file logger: write error: %v\n", err)
		return
	}
	f.fileSize += int64(n)

	// Flush immediately for Error level and above
	if entry.Level == "Critical" || entry.Level == "Error" {
		f.writer.Flush()
	}

	// Check size-based rotation
	if f.config.MaxSize > 0 && f.fileSize >= f.config.MaxSize {
		f.rotateBySizeLocked()
	}
}

// rotateBySizeLocked rotates the file when it exceeds MaxSize. Must be called with mutex held.
func (f *fileLogger) rotateBySizeLocked() {
	if f.writer != nil {
		f.writer.Flush()
	}
	if f.file != nil {
		f.file.Close()
	}

	// Rename current file with timestamp suffix
	ts := time.Now().Format("150405")
	os.Rename(f.filePath, f.filePath+"."+ts)

	// Open new file at same path
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file logger: rotate error: %v\n", err)
		return
	}
	f.file = file
	f.writer = bufio.NewWriterSize(file, 4096)
	f.fileSize = 0
}

// Flush flushes buffered data to disk.
func (f *fileLogger) Flush() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.writer != nil {
		f.writer.Flush()
	}
}

// Close flushes and closes the file logger.
func (f *fileLogger) Close() {
	close(f.done)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.writer != nil {
		f.writer.Flush()
	}
	if f.file != nil {
		f.file.Close()
	}
}

// startAutoFlush periodically flushes the write buffer.
func (f *fileLogger) startAutoFlush() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			f.mutex.Lock()
			if f.writer != nil {
				f.writer.Flush()
			}
			f.mutex.Unlock()
		case <-f.done:
			return
		}
	}
}

// startLogRotation rotates the log file at midnight.
func (f *fileLogger) startLogRotation() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		select {
		case <-time.After(next.Sub(now)):
			f.rotateLog()
		case <-f.done:
			return
		}
	}
}

// rotateLog closes the current log file and opens a new one if the filename changed.
func (f *fileLogger) rotateLog() {
	f.currentDate = time.Now().Format("2006-01-02")

	// Flush before rotating
	f.mutex.Lock()
	if f.writer != nil {
		f.writer.Flush()
	}
	f.mutex.Unlock()

	f.openLogFile()
	f.cleanupOldLogs()
}

// cleanupOldLogs removes expired log files matching the filename pattern.
func (f *fileLogger) cleanupOldLogs() {
	if f.config.Expiration <= 0 {
		return
	}

	expirationDate := time.Now().Add(-f.config.Expiration)

	// Build a glob pattern scoped to the configured filename pattern
	pattern := f.config.FileName
	for _, wc := range []string{"%mm", "%hh", "%y", "%m", "%d"} {
		pattern = strings.ReplaceAll(pattern, wc, "*")
	}

	// Match both regular and size-rotated files (*.HHMMSS suffix)
	matches, _ := filepath.Glob(filepath.Join(f.config.Path, pattern))
	rotated, _ := filepath.Glob(filepath.Join(f.config.Path, pattern+".*"))
	matches = append(matches, rotated...)

	currentPath := f.filePath
	for _, match := range matches {
		if match == currentPath {
			continue
		}
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if info.ModTime().Before(expirationDate) {
			os.Remove(match)
		}
	}
}

// defaultLogFormat is the default formatter for log entries.
func defaultLogFormat(entry *log.Entry) string {
	var b strings.Builder
	b.WriteString(entry.Date.Format("2006-01-02 15:04:05"))
	b.WriteString(" [")
	b.WriteString(entry.Level)
	b.WriteString("] ")
	b.WriteString(entry.File)
	b.WriteString(":")
	b.WriteString(fmt.Sprint(entry.Line))
	b.WriteString(" ")
	b.WriteString(entry.Message)
	for _, f := range entry.Fields {
		b.WriteString(" ")
		b.WriteString(f.Key)
		b.WriteString("=")
		b.WriteString(fmt.Sprint(f.Value))
	}
	return b.String()
}
