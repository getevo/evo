package lib

import (
	"io"
	"io/fs"
	"path/filepath"
	"time"
)

type File interface {
	WriteBytes(bytes []byte) (error, *File)
	WriteString(content string) (error, *File)
	WriteJson(content interface{}) (error, *File)
	Write(reader io.Reader) (error, *File)
	AppendBytes(bytes []byte) (error, *File)
	AppendString(content string) (error, *File)
	SetMetadata(meta Metadata)
	Truncate() error
	Close()
}

type Driver interface {
	Init(params string) error
	SetWorkingDir(path string) error
	WorkingDir() string
	File(path string) (error, *File)
	Touch(path string) error
	WriteJson(path string, content interface{}) error
	Write(path string, content interface{}) error
	Append(path string, content interface{}) error
	SetMetadata(path string, meta Metadata) error
	GetMetadata(path string) (*Metadata, error)
	CopyDir(src, dest string) error
	CopyFile(src, dest string) error
	Stat(path string) (FileInfo, error)
	IsFileExists(path string) bool
	IsDirExists(path string) bool
	IsDir(path string) bool
	ReadAll(path string) ([]byte, error)
	ReadAllString(path string) (string, error)
	Mkdir(path string, perm ...fs.FileMode) error
	MkdirAll(path string, perm ...fs.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
	List(path string, recursive ...bool) ([]FileInfo, error)
	Search(match string) ([]FileInfo, error)
	Name() string
	SetName(name string)
	Type() string
}

type FileInfo struct {
	path    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

func (f FileInfo) Name() string {
	return filepath.Base(f.path)
}

func (f FileInfo) Dir() string {
	if f.isDir {
		return f.path
	}
	return filepath.Dir(f.path)
}

func (f FileInfo) Path() string {
	return f.path
}

func (f FileInfo) Extension() string {
	return filepath.Ext(f.path)
}

func (f FileInfo) Size() int64 {
	return f.size
}

func (f FileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f FileInfo) ModTime() time.Time {
	return f.modTime
}

func (f FileInfo) IsDir() bool {
	return f.isDir
}

func (f FileInfo) Sys() any {
	return f.sys
}

func NewFileInfo(path string, size int64, mode fs.FileMode, modTime time.Time, isDir bool, sys any) FileInfo {
	return FileInfo{
		path:    path,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
		sys:     sys,
	}
}
