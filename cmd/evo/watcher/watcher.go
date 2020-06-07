package watcher

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)
import "github.com/getevo/evo/lib/text"
import "github.com/getevo/evo/lib/gpath"

var mu sync.Mutex
var files = map[string]int64{}
var dirs = map[string]bool{}
var goPath = build.Default.GOPATH
var ds = "/"
var src = ""
var ignore = regexp.MustCompile(`(\\|\/)\.\w+`)

func NewWatcher(dir string, callback func()) {
	if runtime.GOOS == OSWindows {
		ds = `\`
	}
	src = build.Default.GOPATH + ds + "src"

	fmt.Println("Start first scan")
	firstScan(dir, 0)
	fmt.Println("Finish first scan")
	go func() {
		for {
			time.Sleep(1 * time.Second)

			if changeScan() {
				callback()
				//go firstScan(dir, 0)
			}
		}
	}()
}

func changeScan() bool {
	res := false
	mu.Lock()
	for path, fingerprint := range files {
		info, err := os.Stat(path)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			info, err = os.Stat(path)
			if err != nil {
				delete(files, path)
				continue
			}
		}
		lastMod := info.Size() + info.ModTime().Unix()
		if fingerprint != lastMod {
			files[path] = lastMod
			fmt.Println("CHANGED: " + path)
			res = true
		}
	}
	mu.Unlock()
	return res
}

func firstScan(dir string, level int) {
	dir = strings.TrimRight(dir, ds+".")
	if ignore.MatchString(dir) {
		return
	}
	if _, ok := dirs[dir]; ok {
		return
	}

	res, err := ioutil.ReadDir(dir)
	if err == nil {
		for _, info := range res {
			path := dir + ds + info.Name()
			//fmt.Println(path)
			if info.IsDir() && path != src {
				dirs[path] = true
				if gpath.IsFileExist(path + ds + ".ignore") {
					continue
				}
				//fmt.Println(path)
				//fmt.Println(path,info.Name())
				if path != dir {
					firstScan(path, level)
				}
				continue
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				if level < 4 {
					mu.Lock()
					if _, ok := files[path]; !ok {
						//fmt.Println("ADD "+path)
						files[path] = info.ModTime().Unix() + info.Size()
					}
					mu.Unlock()
					parseImports(path, level)
				}
			}
		}
	}

	/*	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if ignore.MatchString(path){
			return nil
		}
		fmt.Println(path)
		if info.IsDir() && path != src{
			if gpath.IsFileExist(path+"/.ignore"){
				return nil
			}
			fmt.Println(path)
			fmt.Println(path,info.Name())
			if path != dir {
				firstScan(path,level)
			}
			return nil
		}
		if strings.HasSuffix(info.Name(),".go"){
			if level < 4 {
				mu.Lock()
				if _,ok := files[path]; !ok{
					fmt.Println("ADD "+path)
					files[path] = info.ModTime().Unix()+info.Size()
				}
				mu.Unlock()
				parseImports(path, level)
			}
		}

		return nil
	})*/
}

func parseImports(s string, level int) {
	if strings.HasSuffix(s, "_test.go") {
		return
	}

	content := strings.Split(string(gpath.SafeFileContent(s)), "\n")
	capture := false
	for _, line := range content {
		line = strings.TrimSpace(line)
		if !capture && text.Match(line, "import*(") {
			capture = true
			continue
		}
		if !capture && strings.HasPrefix(line, "import ") {
			line = strings.TrimSpace(line[6:])
			pack := strings.Trim(line, `"`)
			if !strings.HasPrefix(pack, "golang.org") && gpath.IsDirExist(goPath+ds+"src"+ds+pack) {
				if len(pack) > 3 {
					firstScan(goPath+ds+"src"+ds+pack, level+1)
				}

			}
			continue
		}
		if capture && line == ")" {
			capture = false
			continue
		}

		if capture {
			pack := strings.Trim(line, `"`)

			if !strings.HasPrefix(pack, "golang.org") && gpath.IsDirExist(goPath+ds+"src"+ds+pack) {
				if len(pack) > 3 {
					firstScan(goPath+ds+"src"+ds+pack, level+1)
				}
			}
		}
	}
}
