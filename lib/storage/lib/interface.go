package lib

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

type Driver interface {
	Init(params string) error
	SetWorkingDir(path string) error
	WorkingDir() string
	Touch(path string) error
	WriteJson(path string, content any) error
	Write(path string, content any) error
	Append(path string, content any) error
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
	storage Driver
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

func (f FileInfo) Append(content any) error {
	if f.IsDir() {
		return fmt.Errorf("cant write on directory")
	}
	return f.storage.Append(f.path, content)
}

func (f FileInfo) Write(content any) error {
	if f.IsDir() {
		return fmt.Errorf("cant write on directory")
	}
	return f.storage.Write(f.path, content)
}

func (f FileInfo) Remove() error {
	if f.IsDir() {
		return fmt.Errorf("cant delete a directory")
	}
	return f.storage.Remove(f.path)
}

func (f FileInfo) RemoveDir() error {
	if !f.IsDir() {
		return fmt.Errorf("cant remove dir a file")
	}
	return f.storage.RemoveAll(f.path)
}

func (f FileInfo) Touch() error {
	if f.IsDir() {
		return fmt.Errorf("cant touch a dir")
	}
	return f.storage.Touch(f.path)
}

func NewFileInfo(path string, size int64, mode fs.FileMode, modTime time.Time, isDir bool, sys any, storage Driver) FileInfo {
	return FileInfo{
		path:    path,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
		sys:     sys,
		storage: storage,
	}
}
