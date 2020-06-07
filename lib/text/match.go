package text

import "path/filepath"

func Match(input, pattern string) bool {
	res, _ := filepath.Match(pattern, input)
	return res
}
