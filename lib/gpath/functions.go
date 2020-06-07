package gpath

import (
	"github.com/getevo/evo/lib/text"
	copy2 "github.com/otiai10/copy"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// MakePath create path recursive
func MakePath(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// Parent return root of path
func Parent(s string) string {
	list := text.SplitAny(s, "\\/")
	return strings.Join(list[0:len(list)-1], "/")
}

// WorkingDir return current working directory
func WorkingDir() string {
	path, _ := os.Getwd()
	return path
}

// RSlash trim right slash
func RSlash(path string) string {

	return strings.TrimRight(strings.TrimSpace(path), "/")
}

// IsDirExist checks if is directory exist
func IsDirExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) || info == nil {
		return false
	}
	return info.IsDir()
}

// IsDir checks if  path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// IsFileExist checks if file exist
func IsFileExist(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirEmpty checks if directory is empty
func IsDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true
	}
	return false // Either not empty or error, suits both cases
}

// Stat return information about path
func Stat(path string) *os.FileInfo {
	fileStat, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return &fileStat
}

type internalPathInfo struct {
	FileName  string
	Path      string
	Extension string
}

// PathInfo return detailed information of a path
func PathInfo(path string) internalPathInfo {
	info := internalPathInfo{
		FileName:  filepath.Base(path),
		Path:      filepath.Dir(path),
		Extension: filepath.Ext(path),
	}
	return info
}

// CopyDir copy directory with its contents to destination
func CopyDir(src, dest string) error {
	return copy2.Copy(src, dest)
}

// CopyFile copies file to destination
func CopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

// SymLink create SymLinks
func SymLink(src, dest string) error {
	return os.Link(src, dest)
}

// Remove removes a path or file
func Remove(path string) error {
	if IsDir(path) {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// SafeFileContent return content of a file assumed is accessible
func SafeFileContent(path string) []byte {
	data, _ := ioutil.ReadFile(path)
	return data
}

// ReadFile reads file to bytes
func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
