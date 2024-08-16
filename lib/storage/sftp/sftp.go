package sftp

import (
	"bytes"
	"fmt"
	glob "github.com/ganbarodigital/go_glob"
	"github.com/getevo/evo/v2/lib/storage/lib"
	"github.com/getevo/evo/v2/lib/storage/sftp/client"
	"github.com/getevo/json"
	"sync"
	"time"

	"io/fs"
	"path/filepath"
	"strings"
)

var mu sync.Mutex

type Driver struct {
	Dir      string
	name     string
	host     string
	username string
	password string
	client   *client.Client
}

func (driver *Driver) SetName(name string) {
	driver.name = name
}

func (driver *Driver) Name() string {
	return driver.name
}

func (driver *Driver) Type() string {
	return "sftp"
}

func (driver *Driver) Remove(path string) error {
	path = driver.getRealPath(path)
	return driver.client.Remove(path)
}

func (driver *Driver) RemoveAll(path string) error {
	files, err := driver.List(path, true)
	if err != nil {
		return err
	}
	for i := len(files) - 1; i > -1; i-- {
		var item = files[i]
		if item.IsDir() {
			driver.client.RemoveDir(item.Path())
		} else {
			driver.client.Remove(item.Path())
		}
	}
	return driver.client.RemoveDir(driver.getRealPath(path))

}

func (driver *Driver) List(path string, recursive ...bool) ([]lib.FileInfo, error) {
	var relative = path
	path = driver.getRealPath(path)
	files, err := driver.client.List(path)
	if err != nil {
		return nil, err
	}
	var result []lib.FileInfo
	for _, item := range files {
		result = append(result, lib.NewFileInfo(path+"/"+item.Name(), item.Size(), item.Mode(), item.ModTime(), item.IsDir(), item.Sys(), driver))
		if item.IsDir() && len(recursive) > 0 && recursive[0] == true {
			var ls, _ = driver.List(relative+"/"+item.Name(), true)
			result = append(result, ls...)
		}
	}
	return result, nil
}

func (driver *Driver) Search(match string) ([]lib.FileInfo, error) {
	var result []lib.FileInfo
	var globPath = driver.getRealPath(match)
	g := glob.NewGlob(globPath)
	var parts = strings.Split(match, "*")
	var files, err = driver.List(parts[0], true)
	if err != nil {
		return result, err
	}
	for idx, _ := range files {
		if ok, _ := g.Match(files[idx].Path()); ok {
			result = append(result, files[idx])
		}
	}
	return result, err
}

var settings = lib.StorageSettings(`^(?P<proto>sftp):\/\/(?P<username>[\S\s]+)\:(?P<password>[\S\s]+)@(?P<host>[a-zA-Z0-9\-\_\.\:]+)(\/(?P<dir>[a-zA-Z0-9\-\_\.\/]*)){0,1}`)

func (driver *Driver) Init(input string) error {

	var params, err = settings.Parse(input)
	driver.username = params["username"]
	driver.password = params["password"]
	if !strings.Contains(params["host"], ":") {
		params["host"] += ":22"
	}
	driver.host = params["host"]
	if err != nil {
		return err
	}

	var cfg = client.Config{
		Username: params["username"],
		Password: params["password"],
		Server:   params["host"],
		Timeout:  3 * time.Second,
	}
	driver.client, err = client.New(cfg)
	if err != nil {
		return err
	}
	return driver.SetWorkingDir(params["dir"])
}
func (driver *Driver) SetWorkingDir(path string) error {
	driver.Dir = path
	return nil
}

func (driver *Driver) WorkingDir() string {
	return driver.Dir
}

func (driver *Driver) Touch(path string) error {
	return driver.client.Touch(path)
}

func (driver *Driver) WriteJson(path string, content any) error {
	var b, err = json.Marshal(content)
	if err != nil {
		return err
	}
	return driver.Write(path, b)
}

func (driver *Driver) Append(path string, content any) error {
	path = driver.getRealPath(path)
	var data *bytes.Reader
	switch v := content.(type) {
	case string:
		data = bytes.NewReader([]byte(v))
	case []byte:
		data = bytes.NewReader(v)
	default:
		return fmt.Errorf("invalid content type")
	}
	return driver.client.Append(path, data)
}

func (driver *Driver) SetMetadata(path string, meta lib.Metadata) error {
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
	var list, err = driver.List(src, true)
	if err != nil {
		return err
	}
	var path = driver.getRealPath(src)
	driver.MkdirAll(dest)

	for _, item := range list {
		var destination = filepath.Join(dest, item.Path()[len(path):])
		if item.IsDir() {
			driver.Mkdir(destination)
		} else {
			stream, err := driver.client.Download(item.Path())
			if err != nil {
				return err
			}
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(stream)
			if err != nil {
				return err
			}
			stream.Close()
			err = driver.Write(destination, buf.Bytes())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (driver *Driver) CopyFile(src, dest string) error {
	var content, err = driver.ReadAll(src)
	if err != nil {
		return err
	}
	return driver.Write(dest, content)

}

func (driver *Driver) Stat(path string) (lib.FileInfo, error) {
	path = driver.getRealPath(path)
	info, err := driver.client.Stat(path)
	if err != nil {
		return lib.FileInfo{}, fmt.Errorf("file not found")
	}
	return lib.NewFileInfo(path, info.Size(), info.Mode(), info.ModTime(), info.IsDir(), info.Sys(), driver), nil
}

func (driver *Driver) IsFileExists(path string) bool {
	path = driver.getRealPath(path)
	info, err := driver.client.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && true
}

func (driver *Driver) IsDirExists(path string) bool {
	path = driver.getRealPath(path)
	info, err := driver.client.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir() && true
}

func (driver *Driver) IsDir(path string) bool {
	return driver.IsDirExists(path)
}

func (driver *Driver) ReadAll(path string) ([]byte, error) {
	path = driver.getRealPath(path)

	stream, err := driver.client.Download(path)

	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(stream)
	if err != nil {
		return nil, err
	}
	stream.Close()

	return buf.Bytes(), nil
}

func (driver *Driver) ReadAllString(path string) (string, error) {
	var content, err = driver.ReadAll(path)
	return string(content), err
}

func (driver *Driver) Mkdir(path string, perm ...fs.FileMode) error {
	path = driver.getRealPath(path)
	return driver.client.Mkdir(path)

}

func (driver *Driver) MkdirAll(path string, perm ...fs.FileMode) error {
	path = driver.getRealPath(path)
	return driver.client.MkdirAll(path)
}

func (driver *Driver) Write(path string, content any) error {
	path = driver.getRealPath(path)
	var data *bytes.Reader
	switch v := content.(type) {
	case string:
		data = bytes.NewReader([]byte(v))
	case []byte:
		data = bytes.NewReader(v)
	default:
		return fmt.Errorf("invalid content type")
	}
	return driver.client.Write(path, data)

}

func (driver *Driver) getRealPath(path string) string {
	path = filepath.Clean(filepath.Join(driver.Dir, path))
	path = strings.Replace(path, `\`, `/`, -1)
	return path
}
