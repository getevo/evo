package main

import (
	"fmt"
	"github.com/getevo/evo/lib/gpath"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func format() {
	cfg := Build{}
	cfg.WorkingDir = gpath.WorkingDir()
	filepath.Walk(cfg.WorkingDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			b, err := gpath.ReadFile(path)
			if err == nil {
				code := string(b)
				code = formatStruct(code)
				f, err := gpath.Open(path)
				if err != nil {
					panic(err)
				}
				f.WriteString(code)
			}
		}
		return nil
	})
}

var structRegex = regexp.MustCompile(`(?smU)type\s+(\w+)\s+struct\s*{\s*?(.+)?\s*\}`)
var structLineRegex = regexp.MustCompile(`(?Us)(\S+)\s+(\S+?)(.*?)`)

func formatStruct(s string) string {
	maxFieldLen := 0
	maxFieldTypeLen := 0
	for _, row := range structRegex.FindAllStringSubmatch(s, -1) {

		if len(row) != 3 {
			continue
		}
		lines := strings.Split(row[2], "\n")
		for _, line := range lines {
			res := structLineRegex.FindAllStringSubmatch(line, 1)
			if len(res) == 1 && len(res[0]) >= 3 {
				field := res[0][1]
				fieldType := res[0][2]
				if len(fieldType) > maxFieldTypeLen {
					maxFieldTypeLen = len(fieldType)
				}
				if len(field) > maxFieldLen {
					maxFieldLen = len(field)
				}
				/*fieldTag := ""
				if len(res[0]) == 4{
					fieldTag = res[0][3]
				}*/
			}
		}

	}

	s = structRegex.ReplaceAllStringFunc(s, func(s string) string {
		parts := structRegex.FindStringSubmatch(s)
		if len(parts) != 3 {
			fmt.Println(parts)
			return s
		}
		if strings.TrimSpace(parts[2]) == "" {
			return "type " + parts[1] + " struct{}"
		}

		head := "type " + parts[1] + " struct{"
		var inner = ""
		lines := strings.Split(parts[2], "\n")
		for _, line := range lines {
			res := structLineRegex.FindAllStringSubmatch(line, 1)
			if len(res) == 1 && len(res[0]) >= 3 {
				field := addSpace(res[0][1], maxFieldLen)

				fieldType := addSpace(res[0][2], maxFieldTypeLen)
				fieldTag := ""
				if len(res[0]) == 4 {
					fieldTag = strings.TrimSpace(res[0][3])
				}
				inner += "\n\t" + field + "\t" + fieldType + "\t" + fieldTag
			} else if strings.TrimSpace(line) != "" {
				inner += "\n" + line
			}
		}

		tail := "\n}"

		return head + inner + tail
	})

	return s
}

func addSpace(s string, i int) string {
	for len(s) < i {
		s = s + " "
	}
	return s
}
