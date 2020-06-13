package log

import (
	"fmt"
	"github.com/getevo/evo/lib/date"
	"github.com/getevo/evo/lib/log/logger"
	"github.com/wzshiming/ctc"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var log = logger.New()
var writer *os.File
var writerLock sync.Mutex
var logFileName string
var settings = &Logger{
	WriteToFile: true,
	Concurrent:  true,
	Level:       DebugLevel,
	MaxSize:     50,
	MaxAge:      60,
	Path:        "./logs",
}

// LogLevel type
type Level int

const (
	CriticalLevel Level = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

type Logger struct {
	WriteToFile bool
	Concurrent  bool
	Level       Level
	MaxSize     int
	MaxAge      int
	Path        string
}

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

func SetSettings(logSettings *Logger) {
	settings = logSettings
	log.SetLogLevel(logger.LogLevel(settings.Level))
}

func Register(logSettings *Logger) {
	if logSettings != nil {
		SetSettings(logSettings)
	}
	os.Mkdir(settings.Path, 755)
	Rotate()
}

func Rotate() {
	if settings.WriteToFile && settings.Path+"/"+time.Now().Format("2006-01-02")+".log" != logFileName {
		writerLock.Lock()
		if writer != nil {
			writer.Close()
		}
		logFileName = settings.Path + "/" + time.Now().Format("2006-01-02") + ".log"
		var err error
		writer, err = os.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("Cannot write log file:")
			panic(err)
		}
		writerLock.Unlock()
		cleanOldFiles()
		log.Info("Log Rotated")
		go func() {
			duration, _ := date.Now().DiffExpr("tomorrow midnight")
			time.Sleep(duration)
			time.Sleep(1 * time.Second)
			Rotate()
		}()
	}
}

func cleanOldFiles() {
	var dirSize int64
	filepath.Walk(settings.Path, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSize += file.Size()
		}
		return nil
	})
	sizeMB := int(float64(dirSize) / 1024.0 / 1024.0)

	if sizeMB > settings.MaxSize || true {
		filepath.Walk(settings.Path, func(path string, file os.FileInfo, err error) error {
			if !file.IsDir() {
				if int((time.Now().Unix()-file.ModTime().Unix())/86400) > settings.MaxAge {
					err := os.Remove(path)
					if err != nil {
						Error("unable to remove log file:", err)
					} else {
						log.Info("Log File Deleted: " + path)
					}
				}
			}
			return nil
		})
	}

}

// Fatal is just like func l.Critical logger except that it is followed by exit to program
func Fatal(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	log.LogWrapped(logger.CriticalLevel, msg)
	os.Exit(1)
}

// FatalF is just like func l.CriticalF logger except that it is followed by exit to program
func FatalF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	os.Exit(1)
}

// FatalF is just like func l.CriticalF logger except that it is followed by exit to program
func Fatalf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	os.Exit(1)
}

// Panic is just like func l.Critical except that it is followed by a call to panic
func Panic(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	panic(msg)
}

// PanicF is just like func l.CriticalF except that it is followed by a call to panic
func PanicF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	panic(msg)
}

// PanicF is just like func l.CriticalF except that it is followed by a call to panic
func Panicf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	panic(msg)
}

// Critical logs a message at a Critical Level
func Critical(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
}

// CriticalF logs a message at Critical level using the same syntax and options as fmt.Printf
func CriticalF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))
	log.LogWrapped(logger.CriticalLevel, msg)
}

// CriticalF logs a message at Critical level using the same syntax and options as fmt.Printf
func Criticalf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(CriticalLevel, log.LogWrapped(logger.CriticalLevel, msg))

}

// Error logs a message at Error level
func Error(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(ErrorLevel, log.LogWrapped(logger.ErrorLevel, msg))

}

// ErrorF logs a message at Error level using the same syntax and options as fmt.Printf
func ErrorF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(ErrorLevel, log.LogWrapped(logger.ErrorLevel, msg))

}

// ErrorF logs a message at Error level using the same syntax and options as fmt.Printf
func Errorf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(ErrorLevel, log.LogWrapped(logger.ErrorLevel, msg))

}

// Warning logs a message at Warning level
func Warning(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(WarningLevel, log.LogWrapped(logger.WarningLevel, msg))

}

// WarningF logs a message at Warning level using the same syntax and options as fmt.Printf
func WarningF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(WarningLevel, log.LogWrapped(logger.WarningLevel, msg))

}

// WarningF logs a message at Warning level using the same syntax and options as fmt.Printf
func Warningf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(WarningLevel, log.LogWrapped(logger.WarningLevel, msg))

}

// Notice logs a message at Notice level
func Notice(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(NoticeLevel, log.LogWrapped(logger.NoticeLevel, msg))

}

// NoticeF logs a message at Notice level using the same syntax and options as fmt.Printf
func NoticeF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(NoticeLevel, log.LogWrapped(logger.NoticeLevel, msg))

}

// NoticeF logs a message at Notice level using the same syntax and options as fmt.Printf
func Noticef(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(NoticeLevel, log.LogWrapped(logger.NoticeLevel, msg))

}

// Info logs a message at Info level
func Info(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(InfoLevel, log.LogWrapped(logger.InfoLevel, msg))

}

// InfoF logs a message at Info level using the same syntax and options as fmt.Printf
func InfoF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(InfoLevel, log.LogWrapped(logger.InfoLevel, msg))

}

// InfoF logs a message at Info level using the same syntax and options as fmt.Printf
func Infof(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(InfoLevel, log.LogWrapped(logger.InfoLevel, msg))

}

// Debug logs a message at Debug level
func Debug(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(DebugLevel, log.LogWrapped(logger.DebugLevel, msg))

}

// DebugF logs a message at Debug level using the same syntax and options as fmt.Printf
func DebugF(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(DebugLevel, log.LogWrapped(logger.DebugLevel, msg))

}

// DebugF logs a message at Debug level using the same syntax and options as fmt.Printf
func Debugf(format interface{}, a ...interface{}) {
	msg := fmt.Sprintf(fmt.Sprint(format), a...)
	writeToFile(DebugLevel, log.LogWrapped(logger.DebugLevel, msg))

}

func writeToFile(level Level, message string) {
	if level <= settings.Level {
		writerLock.Lock()
		writer.WriteString("\r\n" + message)
		writerLock.Unlock()
	}
}

func Read(lines int, level Level) {
	if !settings.WriteToFile {
		log.Error("unable to show history of logs. write log to file is disabled")
		return
	}

	printable := ""
	var cursor int64 = 0
	stat, _ := writer.Stat()
	filesize := stat.Size()
	lastPiece := io.SeekEnd
	breakLines := false
	for i := 0; i < lines; i++ {
		line := ""
		for {
			cursor -= 1
			if cursor <= -filesize { // stop if we are at the beginning
				breakLines = true
				break
			}
			writer.Seek(cursor, lastPiece)

			char := make([]byte, 1)
			writer.Read(char)

			if cursor != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
				break
			}

			line = fmt.Sprintf("%s%s", string(char), line) // there is more efficient way

			if cursor <= -filesize { // stop if we are at the beginning
				breakLines = true
				break
			}
		}

		if len(line) > 5 {
			msgLevel := ParseLevel(line[1:5])
			if msgLevel <= level {

				printable = "\r\n" + fmt.Sprintln(logger.Colors[logger.LogLevel(msgLevel)], line, ctc.Reset, printable)
				if breakLines {
					break
				}
				continue
			}

		}
		if breakLines {
			break
		}
		if i > 0 {
			i--
		}
	}

	fmt.Println(printable)

}

func Clear() {
	writer.Truncate(0)
	writer.Seek(0, 0)
	writer.Sync()
}

func ClearAll() {
	Clear()
	filepath.Walk(settings.Path, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			os.Remove(path)
		}
		return nil
	})
}
