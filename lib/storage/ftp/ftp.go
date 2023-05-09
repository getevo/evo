package ftp

import (
	"bytes"
	"encoding/json"
	"fmt"
	glob "github.com/ganbarodigital/go_glob"
	"github.com/getevo/evo/v2/lib/storage/lib"
	ftp "github.com/jlaffaye/ftp"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type Driver struct {
	Dir      string
	name     string
	host     string
	username string
	password string
}

func (driver *Driver) SetName(name string) {
	driver.name = name
}

func (driver *Driver) Name() string {
	return driver.name
}

func (driver *Driver) Type() string {
	return "ftp"
}

func (driver *Driver) Remove(path string) error {
	path = driver.getRealPath(path)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return err
	}
	return connection.Delete(path)
}

func (driver *Driver) RemoveAll(path string) error {
	path = driver.getRealPath(path)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return err
	}

	return connection.RemoveDirRecur(path)
}

func (driver *Driver) List(path string, recursive ...bool) ([]lib.FileInfo, error) {
	var relative = path
	path = driver.getRealPath(path)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	var result []lib.FileInfo
	if err != nil {
		return result, err
	}

	list, err := connection.List(path)

	if err != nil {
		return result, err
	}

	for _, item := range list {

		result = append(result, lib.NewFileInfo(path+"/"+item.Name, int64(item.Size), 0644, item.Time, item.Type == ftp.EntryTypeFolder, nil, driver))

		if item.Type == ftp.EntryTypeFolder && len(recursive) > 0 && recursive[0] == true {
			var items, _ = driver.List(relative+"/"+item.Name, true)
			result = append(result, items...)
		}
	}
	return result, err
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

var settings = lib.StorageSettings(`^(?P<proto>ftp):\/\/(?P<username>[\S\s]+)\:(?P<password>[\S\s]+)@(?P<host>[a-zA-Z0-9\-\_\.\:]+)(\/(?P<dir>[a-zA-Z0-9\-\_\.\/]*)){0,1}`)

func (driver *Driver) Init(input string) error {
	var params, err = settings.Parse(input)
	driver.username = params["username"]
	driver.password = params["password"]
	if !strings.Contains(params["host"], ":") {
		params["host"] += ":21"
	}
	driver.host = params["host"]
	if err != nil {
		return err
	}

	return driver.SetWorkingDir(params["dir"])
}
func (driver *Driver) SetWorkingDir(path string) error {
	path = driver.getRealPath(path)
	var conn, err = driver.getConnection()
	defer driver.Close(conn)
	if err != nil {
		return err
	}
	err = conn.ChangeDir(path)
	if err != nil {
		return err
	}
	driver.Dir, err = conn.CurrentDir()
	return err
}

func (driver *Driver) WorkingDir() string {
	return driver.Dir
}

func (driver *Driver) Touch(path string) error {
	return driver.Write(path, "")
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
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return err
	}
	var data *bytes.Reader
	switch v := content.(type) {
	case string:
		data = bytes.NewReader([]byte(v))
	case []byte:
		data = bytes.NewReader(v)
	default:
		return fmt.Errorf("invalid content type")
	}
	return connection.Append(path, data)
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
	driver.MkdirAll(dest)

	src = driver.getRealPath(src)
	dest = driver.getRealPath(dest)

	c1, err := driver.getConnection()
	defer driver.Close(c1)
	if err != nil {
		return err
	}
	c2, err := driver.getConnection()
	defer driver.Close(c2)
	if err != nil {
		return err
	}

	c3, err := driver.getConnection()
	defer driver.Close(c3)
	if err != nil {
		return err
	}

	var walker = c1.Walk(src)
	for walker.Next() {
		t := walker.Stat().Type
		if t == ftp.EntryTypeFolder {
			fmt.Println("make dir:", dest+walker.Path()[len(src):])
			c2.MakeDir(dest + walker.Path()[len(src):])
		} else if t == ftp.EntryTypeFile {

			stream, err := c3.Retr(walker.Path())
			if err != nil {
				return err
			}
			fmt.Println("copy:", walker.Path(), dest+walker.Path()[len(src):])
			err = c2.Stor(dest+walker.Path()[len(src):], stream)
			stream.Close()
			if err != nil {
				return err
			}

		}
		//walker.Path()
	}

	return nil
}

func (driver *Driver) CopyFile(src, dest string) error {
	src = driver.getRealPath(src)
	dest = driver.getRealPath(dest)

	c1, err := driver.getConnection()
	defer driver.Close(c1)
	if err != nil {
		return err
	}
	c2, err := driver.getConnection()
	defer driver.Close(c2)
	if err != nil {
		return err
	}
	stream, err := c1.Retr(src)
	if err != nil {
		return err
	}

	err = c2.Stor(dest, stream)
	stream.Close()
	if err != nil {
		return err
	}
	return nil

}

func (driver *Driver) Stat(path string) (lib.FileInfo, error) {
	var dir = filepath.Dir(path)
	var base = filepath.Base(path)
	var ls, _ = driver.List(dir, false)
	for _, item := range ls {
		if item.Name() == base {
			return item, nil
		}
	}
	return lib.FileInfo{}, fmt.Errorf("file not found")
}

func (driver *Driver) IsFileExists(path string) bool {
	var dir = filepath.Dir(path)
	var base = filepath.Base(path)
	var ls, _ = driver.List(dir, false)
	for _, item := range ls {

		if item.Name() == base && !item.IsDir() {
			return true
		}
	}
	return false
}

func (driver *Driver) IsDirExists(path string) bool {
	var dir = filepath.Dir(path)
	var base = filepath.Base(path)
	var ls, _ = driver.List(dir, false)
	for _, item := range ls {
		if item.Name() == base && item.IsDir() {
			return true
		}
	}
	return false
}

func (driver *Driver) IsDir(path string) bool {
	return driver.IsDirExists(path)
}

func (driver *Driver) ReadAll(path string) ([]byte, error) {
	path = driver.getRealPath(path)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return nil, err
	}

	stream, err := connection.Retr(path)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	stream.Close()
	return buf.Bytes(), nil
}

func (driver *Driver) ReadAllString(path string) (string, error) {
	var content, err = driver.ReadAll(path)
	return string(content), err
}

func (driver *Driver) Mkdir(path string, perm ...fs.FileMode) error {
	path = driver.getRealPath(path)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return err
	}
	return connection.MakeDir(path)

}

func (driver *Driver) MkdirAll(path string, perm ...fs.FileMode) error {
	var relative = filepath.Clean(path)
	parts := strings.Split(strings.Replace(relative, `/`, `\`, -1), `\`)
	var connection, err = driver.getConnection()
	defer driver.Close(connection)
	if err != nil {
		return err
	}
	var build = ""
	for _, item := range parts {
		build += "/" + item
		var p = driver.getRealPath(build)
		err := connection.MakeDir(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (driver *Driver) Write(path string, content interface{}) error {
	var conn, err = driver.getConnection()
	defer driver.Close(conn)
	if err != nil {
		return err
	}
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
	return conn.Stor(path, data)
}

func (driver *Driver) getRealPath(path string) string {
	path = filepath.Clean(filepath.Join(driver.Dir, path))
	path = strings.Replace(path, `\`, `/`, -1)
	return path
}

func (driver *Driver) getConnection() (*ftp.ServerConn, error) {
	c, err := ftp.Dial(driver.host, ftp.DialWithDisabledEPSV(true), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		fmt.Println("connection  err:", err)
		return nil, err
	}
	err = c.Login(driver.username, driver.password)
	if err != nil {
		fmt.Println("login  err:", err)
		return nil, err
	}

	return c, nil
}

func (driver *Driver) Close(connection *ftp.ServerConn) {
	if connection != nil {
		connection.Quit()
		connection = nil
	}
}
