package viewfn

import (
	"github.com/CloudyKit/jet"
	"github.com/disintegration/imaging"
	"github.com/iesreza/io/lib/gpath"
	"github.com/iesreza/io/lib/log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
)

var Functions = map[string]jet.Func{
	"thumb": Thumb,
}

var WorkingDir = gpath.WorkingDir()

// Bind bind global useful functions to view
func Bind(views *jet.Set, fn ...string) {

	for _, item := range fn {
		if v, ok := Functions[item]; ok {
			views.AddGlobalFunc(item, v)
		}

	}

}

func Thumb(arguments jet.Arguments) reflect.Value {
	var path string
	var width = -1
	var height = -1

	arguments.RequireNumOfArguments("thumb", 1, 3)
	for i := 0; i < arguments.NumOfArguments(); i++ {
		switch arguments.Get(i).Kind() {
		case reflect.String:
			path = arguments.Get(i).Interface().(string)

			break
		case reflect.Struct:
			if v, ok := arguments.Get(i).Interface().(interface{ Image() string }); ok {
				path = v.Image()
			}
			break
		case reflect.Ptr:
			if v, ok := arguments.Get(i).Elem().Interface().(interface{ Image() string }); ok {
				path = v.Image()
			} else if v, ok := arguments.Get(i).Interface().(interface{ Image() string }); ok {
				path = v.Image()
			}
			break
		case reflect.Int:
			if width == -1 {
				width = arguments.Get(i).Interface().(int)
			} else {
				height = arguments.Get(i).Interface().(int)
			}
			break
		}

	}

	if width < 1 {
		width = 128
	}
	if height < 1 {
		height = 128
	}
	srcPath := WorkingDir + "/httpdocs/" + path
	srcFileName := filepath.Base(path)
	fi, err := os.Stat(srcPath)
	if err != nil {
		log.Error(err)
		return reflect.ValueOf("/files/images/404.png")
	}
	destFileName := strconv.Itoa(width) + "x" + strconv.Itoa(height) + "s" + strconv.FormatInt(fi.Size(), 16) + "_" + srcFileName
	destPath := WorkingDir + "/httpdocs/files/thumb/" + destFileName
	if !gpath.IsFileExist(destPath) {
		src, err := imaging.Open(srcPath)
		if err != nil {
			log.Error(err)
			return reflect.ValueOf("/files/images/404.png")
		}
		src = imaging.Thumbnail(src, width, height, imaging.Lanczos)
		err = imaging.Save(src, destPath)

		if err != nil {
			log.Error(err)
			return reflect.ValueOf("/files/images/404.png")
		}
	}
	return reflect.ValueOf("files/thumb/" + destFileName)
}
