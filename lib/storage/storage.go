package storage

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/getevo/evo/v2/lib/storage/filesystem"
	"github.com/getevo/evo/v2/lib/storage/ftp"
	"github.com/getevo/evo/v2/lib/storage/lib"
	"github.com/getevo/evo/v2/lib/storage/s3"
	"github.com/getevo/evo/v2/lib/storage/sftp"
	"io/fs"
	"reflect"
	"regexp"
)

// A FileInfo describes a file and is returned by Stat and Lstat.
type FileInfo = lib.FileInfo

// A FileMode represents a file's mode and permission bits.
// The bits have the same definition on all systems, so that
// information about files can be moved from one system
// to another portably. Not all bits apply to all systems.
// The only required bit is ModeDir for directories.
type FileMode = fs.FileMode

// The defined file mode bits are the most significant bits of the FileMode.
// The nine least-significant bits are the standard Unix rwxrwxrwx permissions.
// The values of these bits should be considered part of the public API and
// may be used in wire protocols or disk representations: they must not be
// changed, although new bits might be added.
const (
	// ModeDir The single letters are the abbreviations
	// used by the String method's formatting.
	ModeDir        = fs.ModeDir        // d: is a directory
	ModeAppend     = fs.ModeAppend     // a: append-only
	ModeExclusive  = fs.ModeExclusive  // l: exclusive use
	ModeTemporary  = fs.ModeTemporary  // T: temporary file; Plan 9 only
	ModeSymlink    = fs.ModeSymlink    // L: symbolic link
	ModeDevice     = fs.ModeDevice     // D: device file
	ModeNamedPipe  = fs.ModeNamedPipe  // p: named pipe (FIFO)
	ModeSocket     = fs.ModeSocket     // S: Unix domain socket
	ModeSetuid     = fs.ModeSetuid     // u: setuid
	ModeSetgid     = fs.ModeSetgid     // g: setgid
	ModeCharDevice = fs.ModeCharDevice // c: Unix character device, when ModeDevice is set
	ModeSticky     = fs.ModeSticky     // t: sticky
	ModeIrregular  = fs.ModeIrregular  // ?: non-regular file; nothing else is known about this file

	// ModeType Mask for the type bits. For regular files, none will be set.
	ModeType = fs.ModeType

	ModePerm = fs.ModePerm // Unix permission bits, 0o777
)

var availableDrivers = []lib.Driver{&filesystem.Driver{}, &s3.Driver{}, &ftp.Driver{}, &sftp.Driver{}}
var Pool = map[string]lib.Driver{}

func Drivers() []lib.Driver {
	return availableDrivers
}

var configRegex = regexp.MustCompile(`(?m)^([a-zA-Z0-9\_\-]+)://`)

func NewStorage(tag string, storage string) (*lib.Driver, error) {
	var config = configRegex.FindAllStringSubmatch(storage, -1)
	if len(config) == 0 {
		return nil, fmt.Errorf("invalid storage config string %s, required  proto://...", storage)
	}
	var driver = config[0][1]

	for _, item := range availableDrivers {
		if item.Type() == driver {
			var t = generic.Parse(item).IndirectType()
			var obj = reflect.New(t).Interface().(lib.Driver)
			obj.SetName(tag)
			err := obj.Init(storage)
			if err != nil {
				return nil, err
			}
			Pool[tag] = obj
			return &obj, nil
		}
	}
	return nil, fmt.Errorf("invalid driver %s", driver)
}

func GetStorage(tag string) lib.Driver {
	var storage, ok = Pool[tag]
	if !ok {
		return nil
	}
	return storage
}
