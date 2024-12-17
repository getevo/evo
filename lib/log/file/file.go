package file

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/getevo/evo/v2/lib/log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Path       string                        // Directory for log files
	FileName   string                        // Filename template with wildcards (e.g., log_%y-%m-%d.log)
	Expiration time.Duration                 // Expiration in days, if <=0 no cleanup
	LogFormat  func(entry *log.Entry) string // Function to format log entry
}

// fileLogger is the internal structure for the file logger
type fileLogger struct {
	config      Config
	file        *os.File
	filePath    string
	mutex       sync.Mutex
	expiryMutex sync.Mutex
	currentDate string
}

// NewFileLogger creates a file logger compatible with the log package
func NewFileLogger(config ...Config) func(log *log.Entry) {
	c := Config{}
	if len(config) > 0 {
		c = config[0]
	}

	if c.Path == "" {
		c.Path, _ = os.Getwd()
	}
	if c.FileName == "" {
		execName := filepath.Base(os.Args[0])
		c.FileName = fmt.Sprintf("%s.log", execName)
	}
	if c.LogFormat == nil {
		c.LogFormat = defaultLogFormat
	}

	logger := &fileLogger{
		config:      c,
		currentDate: time.Now().Format("2006-01-02"),
	}

	logger.openLogFile()
	go logger.startLogRotation()

	return func(log *log.Entry) {
		logger.writeLog(log)
	}
}

// openLogFile opens or creates the log file with append mode
func (f *fileLogger) openLogFile() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if !gpath.IsDirExist(f.config.Path) {
		err := gpath.MakePath(f.config.Path)
		log.Fatalf("failed to open log dir: %v", err)
	}
	f.filePath = f.getFilePath()
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	/*if f.file != nil {
		_ = f.file.Close()
	}*/
	f.file = file
}

// getFilePath generates the log file path with the current date
func (f *fileLogger) getFilePath() string {
	template := f.config.FileName
	now := time.Now()
	fileName := strings.ReplaceAll(template, "%y", now.Format("2006"))
	fileName = strings.ReplaceAll(fileName, "%mm", now.Format("04"))
	fileName = strings.ReplaceAll(fileName, "%m", now.Format("01"))
	fileName = strings.ReplaceAll(fileName, "%d", now.Format("02"))
	fileName = strings.ReplaceAll(fileName, "%hh", now.Format("15"))

	return filepath.Join(f.config.Path, fileName)
}

// writeLog safely writes the log entry to the file
func (f *fileLogger) writeLog(entry *log.Entry) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	logString := f.config.LogFormat(entry)
	_, err := f.file.WriteString(logString + "\r\n")
	if err != nil {
		log.Error("failed to write to log file: %v", err)
	}
}

// startLogRotation rotates the log at midnight
func (f *fileLogger) startLogRotation() {
	for {
		nextMidnight := time.Now().Truncate(60 * time.Second).Add(60 * time.Second)
		time.Sleep(time.Until(nextMidnight))
		f.rotateLog()
	}
}

// rotateLog closes the current log file and opens a new one
func (f *fileLogger) rotateLog() {
	f.currentDate = time.Now().Format("2006-01-02")
	if f.getFilePath() != f.filePath {
		// release the old file
		_ = f.file.Close()

		// open new file
		f.openLogFile()
	}
	f.cleanupOldLogs()
}

// cleanupOldLogs removes expired log files if expiration is set
func (f *fileLogger) cleanupOldLogs() {
	if f.config.Expiration <= 1 {
		return
	}

	expirationDate := time.Now().Add(-f.config.Expiration)
	files, _ := filepath.Glob(filepath.Join(f.config.Path, "*.log"))
	for _, file := range files {
		if file == f.filePath {
			continue
		}
		if stat, err := os.Stat(file); err == nil && stat.ModTime().Before(expirationDate) {
			_ = os.Remove(file)
		}
	}
}

// defaultLogFormat is the default formatter for log entries
func defaultLogFormat(e *log.Entry) string {
	return fmt.Sprintf("%s [%s] %s:%d %s", e.Date.Format("2006-01-02 15:04:05"), e.Level, e.File, e.Line, e.Message)
}
