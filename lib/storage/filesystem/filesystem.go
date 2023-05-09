package filesystem

import (
	"encoding/json"
	"fmt"
	"github.com/getevo/evo/v2/lib/gpath"
	"github.com/getevo/evo/v2/lib/storage/lib"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Driver struct {
	Dir  string
	name string
}

func (driver *Driver) SetName(name string) {
	driver.name = name
}

func (driver *Driver) Name() string {
	return driver.name
}

func (driver *Driver) Type() string {
	return "fs"
}

func (driver *Driver) Remove(path string) error {
	path = driver.getRealPath(path)
	return os.Remove(path)
}

func (driver *Driver) RemoveAll(path string) error {
	path = driver.getRealPath(path)
	return os.RemoveAll(path)
}

func (driver *Driver) List(path string, recursive ...bool) ([]lib.FileInfo, error) {
	path = driver.getRealPath(path)
	if len(recursive) == 0 || recursive[0] == false {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		list, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return nil, err
		}

		var finfo []lib.FileInfo
		for _, item := range list {
			finfo = append(finfo, lib.NewFileInfo(path+"/"+item.Name(), item.Size(), item.Mode(), item.ModTime(), item.IsDir(), item.Sys(), driver))
		}
		return finfo, nil
	}

	var result []lib.FileInfo
	err := filepath.Walk(path, func(path string, item os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		result = append(result, lib.NewFileInfo(path, item.Size(), item.Mode(), item.ModTime(), item.IsDir(), item.Sys(), driver))
		return nil
	})
	return result, err
}

func (driver *Driver) Search(match string) ([]lib.FileInfo, error) {
	var result []lib.FileInfo
	match = strings.TrimRight(driver.Dir, "/") + "/" + strings.TrimLeft(match, "/")
	var files, err = filepath.Glob(match)
	if err != nil {
		return result, err
	}
	for _, item := range files {
		f, err := os.Stat(item)
		if err != nil {
			return result, err
		}
		result = append(result, lib.NewFileInfo(filepath.Clean(item), f.Size(), f.Mode(), f.ModTime(), f.IsDir(), f.Sys(), driver))
	}
	return result, err
}

var settings = lib.StorageSettings(`(^(?P<proto>[a-zA-Z]+)://(?P<dir>.*))`)

func (driver *Driver) Init(input string) error {
	var params, err = settings.Parse(input)
	if err != nil {
		return err
	}
	return driver.SetWorkingDir(params["dir"])
}
func (driver *Driver) SetWorkingDir(path string) error {
	path = driver.getRealPath(path)
	driver.Dir = path
	if !gpath.IsDir(driver.Dir) {
		return fmt.Errorf("invalid dir %s", driver.Dir)
	}
	return nil
}

func (driver *Driver) WorkingDir() string {
	return driver.Dir
}

func (driver *Driver) Touch(path string) error {
	path = driver.getRealPath(path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func (driver *Driver) WriteJson(path string, content interface{}) error {
	var b, err = json.Marshal(content)
	if err != nil {
		return err
	}
	return driver.Write(path, b)
}

func (driver *Driver) Append(path string, content interface{}) error {
	path = driver.getRealPath(path)
	return gpath.Append(path, content)
}

func (driver *Driver) SetMetadata(path string, meta lib.Metadata) error {
	path = driver.getRealPath(path)
	return nil
}

func (driver *Driver) GetMetadata(path string) (*lib.Metadata, error) {
	var meta = lib.Metadata{}
	var stat, err = driver.Stat(path)
	if err != nil {
		return nil, err
	}
	meta["name"] = stat.Name()
	meta["size"] = fmt.Sprint(stat.Size())
	meta["is_dir"] = fmt.Sprint(stat.IsDir())
	meta["mode"] = stat.Mode().String()
	meta["mod_time"] = fmt.Sprint(stat.ModTime().Unix())

	return &meta, nil
}

func (driver *Driver) CopyDir(src, dest string) error {
	src = filepath.Join(driver.Dir, src)
	dest = filepath.Join(driver.Dir, dest)
	return gpath.CopyDir(src, dest)
}

func (driver *Driver) CopyFile(src, dest string) error {
	src = filepath.Join(driver.Dir, src)
	dest = filepath.Join(driver.Dir, dest)
	return gpath.CopyFile(src, dest)
}

func (driver *Driver) Stat(path string) (lib.FileInfo, error) {
	path = driver.getRealPath(path)
	var stat, err = os.Stat(path)
	if err != nil {
		return lib.FileInfo{}, err
	}
	return lib.NewFileInfo(path, stat.Size(), stat.Mode(), stat.ModTime(), stat.IsDir(), stat.Sys(), driver), nil
}

func (driver *Driver) IsFileExists(path string) bool {
	return gpath.IsFileExist(path)
}

func (driver *Driver) IsDirExists(path string) bool {
	path = driver.getRealPath(path)
	return gpath.IsDirExist(path)
}

func (driver *Driver) IsDir(path string) bool {
	path = driver.getRealPath(path)
	return gpath.IsDir(path)
}

func (driver *Driver) ReadAll(path string) ([]byte, error) {
	path = driver.getRealPath(path)
	return gpath.ReadFile(path)
}

func (driver *Driver) ReadAllString(path string) (string, error) {
	var content, err = driver.ReadAll(path)
	return string(content), err
}

func (driver *Driver) Mkdir(path string, perm ...fs.FileMode) error {
	path = driver.getRealPath(path)
	if len(perm) == 0 {
		perm = []fs.FileMode{0755}
	}
	return os.Mkdir(path, perm[0])
}

func (driver *Driver) MkdirAll(path string, perm ...fs.FileMode) error {
	path = driver.getRealPath(path)
	if len(perm) == 0 {
		perm = []fs.FileMode{0755}
	}
	return os.MkdirAll(path, perm[0])
}

func (driver *Driver) Write(path string, content interface{}) error {
	return gpath.Write(path, content)
}

func (driver *Driver) getRealPath(path string) string {
	path, _ = filepath.Abs(filepath.Clean(filepath.Join(driver.Dir, path)))
	return path
}
