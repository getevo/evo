package stract

import (
	"regexp"
	"strings"
)

func (ctx *Context) GetChildren() []*Context {
	if ctx.Children == nil {
		return []*Context{}
	}
	return ctx.Children
}

func (ctx *Context) GetVaryDicts() []VaryDict {
	if ctx.VaryDict == nil {
		return []VaryDict{}
	}
	return ctx.VaryDict
}

func (c *Context) Get(s string) (bool, []string) {
	for _, vary := range c.VaryDict {
		if vary.Key == s {
			return true, vary.Values
		}
	}
	return false, nil
}

func (c *Context) GetSingleValue(s string) string {
	for _, vary := range c.VaryDict {
		if vary.Key == s {
			if len(vary.Values) > 0 {
				return vary.Values[0]
			}
			return ""
		}
	}
	return ""
}

func (ctx *Context) GetChild(name string) (bool, *Context) {
	for idx, item := range ctx.Children {
		if item.Name == name {
			return true, ctx.Children[idx]
		}
	}
	return false, nil
}

func (ctx *Context) GetVaryDict(name string) (bool, []string) {
	return ctx.Get(name)
}

func (ctx *Context) VaryDictHas(name string, value string) bool {
	exists, vary := ctx.Get(name)
	if !exists {
		return false
	}
	for _, item := range vary {
		if item == value {
			return true
		}
	}
	return false
}

func (ctx *Context) VaryDictMatch(name string, regex *regexp.Regexp) (bool, [][]string) {
	exists, vary := ctx.Get(name)
	if !exists {
		return false, nil
	}
	for _, item := range vary {
		var matches = regex.FindAllStringSubmatch(item, -1)
		if len(matches) > 0 {
			return true, matches
		}
	}
	return false, nil
}

func (ctx *Context) VaryDictContains(name string, value string) (bool, string) {
	exists, vary := ctx.Get(name)
	if !exists {
		return false, ""
	}
	for _, item := range vary {
		if strings.Contains(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (ctx *Context) VaryDictStartsWith(name string, value string) (bool, string) {
	exists, vary := ctx.Get(name)
	if !exists {
		return false, ""
	}
	for _, item := range vary {
		if strings.HasPrefix(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (ctx *Context) VaryDictEndsWith(name string, value string) (bool, string) {
	exists, vary := ctx.Get(name)
	if !exists {
		return false, ""
	}
	for _, item := range vary {
		if strings.HasSuffix(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (vary *VaryDict) VaryDictHas(value string) bool {
	for _, item := range vary.Values {
		if item == value {
			return true
		}
	}
	return false
}

func (vary *VaryDict) VaryDictMatch(regex *regexp.Regexp) (bool, [][]string) {
	for _, item := range vary.Values {
		var matches = regex.FindAllStringSubmatch(item, -1)
		if len(matches) > 0 {
			return true, matches
		}
	}
	return false, nil
}

func (vary *VaryDict) VaryDictContains(value string) (bool, string) {
	for _, item := range vary.Values {
		if strings.Contains(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (vary *VaryDict) VaryDictStartsWith(value string) (bool, string) {
	for _, item := range vary.Values {
		if strings.HasPrefix(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (vary VaryDict) VaryDictEndsWith(value string) (bool, string) {
	for _, item := range vary.Values {
		if strings.HasSuffix(item, value) {
			return true, item
		}
	}
	return false, ""
}

func (vary *VaryDict) VaryDictGetVar(value string) string {
	for _, item := range vary.Values {
		if strings.HasPrefix(item, value) {
			if len(item) > len(value) {
				var c = item[len(value)]
				if c == ':' {
					return item[len(value)+1:]
				} else if c == '(' {
					return item[len(value)+1 : len(item)-1]
				}
			}
			return ""
		}
	}
	return ""
}

func ParseVar(input string, key string) string {
	if strings.HasPrefix(input, key) {
		if len(input) > len(key) {
			var c = input[len(key)]
			if c == ':' {
				var v = input[len(key)+1:]
				return v
			} else if c == '(' {
				return input[len(key)+1 : len(input)-1]
			}
		}

	}

	return ""
}
