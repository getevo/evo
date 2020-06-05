package gpath

import (
	"encoding/json"
	"os"
	"time"
)

var DefaultTimeout = 2 * time.Second

type file struct {
	timeout    time.Duration
	lastaccess int64
	fp         *os.File
	mode       int
	closed     bool
	path       string
	observer   bool
	tests      string
}

// Open opens a file
func Open(path string) (*file, error) {
	f := file{timeout: DefaultTimeout, path: path}
	var err error
	if !IsFileExist(path) {
		f.fp, err = os.Create(path)
	} else {
		f.mode = os.O_RDWR
		f.fp, err = os.OpenFile(path, f.mode, 0644)
	}
	if err != nil {
		return &f, err
	}

	f.SetLastAccess()
	f.initTimeout()
	return &f, nil
}

// SetLastAccess sets last access time of a file
func (f *file) SetLastAccess() {
	f.lastaccess = time.Now().UnixNano()
}

func (f *file) access(mode int) error {
	var err error
	f.initTimeout()
	if mode != f.mode {
		f.mode = mode
		if !f.closed {
			f.Close()
		}
	}
	if f.closed {
		f.fp, err = os.OpenFile(f.path, f.mode, 0644)
		if err != nil {
			return err
		}
	}
	f.SetLastAccess()
	return nil
}

// WriteString writes string to file
func (f *file) WriteString(v string) error {
	return f.Write([]byte(v))
}

// WriteJson take struct and write to file as json
func (f *file) WriteJson(v interface{}, pretty bool) error {
	if pretty {
		b, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			return err
		}
		return f.Write(b)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return f.Write(b)
}

// Write writes bytes to file
func (f *file) Write(v []byte) error {
	if err := f.access(os.O_RDWR); err != nil {
		return err
	}
	f.Truncate()
	_, err := f.fp.Write(v)
	return err
}

// Append appends bytes to file
func (f *file) Append(v []byte) error {
	if err := f.access(os.O_APPEND | os.O_WRONLY); err != nil {
		return err
	}
	_, err := f.fp.Write(v)
	return err
}

// AppendString appends string to file
func (f *file) AppendString(v string) error {
	return f.Append([]byte(v))
}

// ReadAll read file to bytes
func (f *file) ReadAll() ([]byte, error) {
	if err := f.access(os.O_RDWR); err != nil {
		return []byte{}, err
	}

	fileinfo, err := f.fp.Stat()
	if err != nil {
		return []byte{}, err
	}
	filesize := fileinfo.Size()

	buffer := make([]byte, filesize)
	f.fp.Seek(0, 0)
	_, err = f.fp.Read(buffer)
	return buffer, err
}

// ReadAllString read file to string
func (f *file) ReadAllString() (string, error) {
	b, err := f.ReadAll()
	return string(b), err
}

// UnmarshalJson read file and parse as json
func (f *file) UnmarshalJson(v interface{}) error {
	b, err := f.ReadAll()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// Truncate empty the file
func (f *file) Truncate() error {
	if err := f.access(os.O_RDWR); err != nil {
		return err
	}
	err := f.fp.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.fp.Seek(0, 0)
	return err
}

// SetTimeout closes the file if not being modified after duration
func (f *file) SetTimeout(d time.Duration) {
	f.timeout = d
}

// Close releases file resources
func (f *file) Close() {
	f.fp.Close()
	f.closed = true
}

func (f *file) setObserver(v bool) {
	f.observer = v
}

func (f *file) initTimeout() {

	go func() {
		if f.observer == true {
			return
		}
		f.observer = true
		for {
			time.Sleep(f.timeout)
			if f.timeout == 0 || f.closed {
				break
			}
			if time.Now().UnixNano()-f.lastaccess > int64(f.timeout) {
				f.Close()
				break
			}
		}
		f.observer = false

	}()
}
