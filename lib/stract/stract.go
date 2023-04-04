package stract

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

const none rune = '\x10'

var varRegex = regexp.MustCompile(`(?m)\$\{{0,1}([0-9a-zA-Z\_\-]+)\}{0,1}`)

type Context struct {
	Parent   *Context   `json:"-"`
	Children []*Context `json:"children"`
	Name     string     `json:"name"`
	VaryDict []VaryDict `json:"vary_dict"`
}

func (c *Context) getVaryDictRecursive(key string) (bool, []string) {
	for _, vary := range (*c).VaryDict {
		if vary.Key == key {
			return true, vary.Values
		}
	}
	if c.Parent != nil {
		return c.Parent.getVaryDictRecursive(key)
	}
	return false, nil
}

func (c *Context) getContextRecursive(key string) (bool, *Context) {
	for _, child := range (*c).Children {
		if child.Name == key {
			return true, child
		}
	}
	if c.Parent != nil {
		return c.Parent.getContextRecursive(key)
	}
	return false, nil
}

func (c *Context) tryResolve(key string) string {
	key = varRegex.ReplaceAllStringFunc(key, func(s string) string {
		var match = varRegex.FindAllStringSubmatch(s, 1)
		var ptr = c
		for {
			ok, vary := ptr.Get(match[0][1])
			if ok {
				return vary[0]
			}
			if ptr.Parent == nil {
				break
			}
			ptr = ptr.Parent
		}
		return s
	})

	return key
}

type VaryDict struct {
	Key    string
	Values []string
}
type Token struct {
	Value string
	Pos   int
}

func OpenAndParse(path string) (*Context, error) {
	var dir = filepath.Dir(path)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(file)
	var lines = strings.Split(content, "\n")
	var root = Context{
		Name:     "root",
		VaryDict: []VaryDict{},
	}
	var ptr = &root
	var stack = runeStack{}
	var previous = ""
	for pos, line := range lines {
		var tokens = getTokens(line)
		var varyDict = VaryDict{}
		for _, token := range tokens {
			var last = string(stack.Last())

			switch token.Value {
			case "{":
				var node = &Context{
					Name:   previous,
					Parent: ptr,
				}
				ptr.Children = append(ptr.Children, node)
				ptr = node
				stack.push('}')

			case "}":
				if last == "}" {
					stack.pop()
					ptr = ptr.Parent
				}
			case "@import":
				var context, err = OpenAndParse(filepath.Join(dir, tokens[pos+1].Value))
				if err != nil {
					return nil, err
				}
				merge(ptr, context)
			default:
				if varyDict.Key == "" {

					if len(tokens) > 1 {
						varyDicts, err := processVaryDict(&token, ptr)
						if err != nil {
							return ptr, err
						}
						varyDict.Key = varyDicts[0].Value
						if len(varyDicts) > 1 {
							for _, item := range varyDicts[1:] {
								varyDict.Values = append(varyDict.Values, item.Value)
							}
						}
					} else {
						ctx, err := processKey(&token, ptr)
						if err != nil {
							return ptr, err
						}
						if ctx != nil {
							merge(ptr, ctx)
							continue
						} else {
							varyDict.Key = token.Value
						}
					}

				} else {
					varyDicts, err := processVaryDict(&token, ptr)
					if err != nil {
						return ptr, err
					}
					for _, item := range varyDicts {
						varyDict.Values = append(varyDict.Values, item.Value)
					}
				}

			}
			previous = token.Value
		}
		if len(tokens) == 2 && tokens[1].Value == "{" {
			continue
		}
		if varyDict.Key != "" {
			ptr.VaryDict = append(ptr.VaryDict, varyDict)
		}
		finisher(ptr)
	}
	//fmt.Println(PrettyStruct(root))
	return &root, nil
}

func finisher(context *Context) {
	for idx, _ := range context.VaryDict {
		context.VaryDict[idx].Key = context.tryResolve(context.VaryDict[idx].Key)
		for k, _ := range context.VaryDict[idx].Values {
			context.VaryDict[idx].Values[k] = context.tryResolve(context.VaryDict[idx].Values[k])
		}
	}
	for idx, _ := range context.Children {
		context.Children[idx].Name = context.tryResolve(context.Children[idx].Name)
	}
}

func processKey(token *Token, ptr *Context) (*Context, error) {
	var ocs = varRegex.FindAllStringSubmatch(token.Value, -1)
	if len(ocs) > 0 {
		if len(ocs) == 1 {
			var found, ctx = ptr.getContextRecursive(ocs[0][1])
			if !found {
				return nil, nil
			}
			return ctx, nil
		}
	}
	return nil, nil
}

func processVaryDict(token *Token, ptr *Context) ([]Token, error) {
	var result []Token
	var ocs = varRegex.FindAllStringSubmatch(token.Value, -1)
	if len(ocs) > 0 {
		for _, oc := range ocs {
			var found, tokens = ptr.getVaryDictRecursive(oc[1])

			if found {
				if len(ocs) > 1 {
					token.Value = strings.Replace(token.Value, oc[0], tokens[len(tokens)-1], 1)
				} else {
					for _, t := range tokens {
						result = append(result, Token{
							Value: t,
						})
					}
				}

			} else {
				//return result, fmt.Errorf("invalid parameter %s", oc[0])
				return []Token{*token}, nil
			}
		}
		if len(ocs) > 1 {
			return []Token{*token}, nil
		}
	}
	return []Token{
		*token,
	}, nil
}

func merge(src *Context, context *Context) {
	for _, item := range context.VaryDict {
		src.VaryDict = append(src.VaryDict, item)
	}
	for _, item := range context.Children {
		src.Children = append(src.Children, item)
	}
}

func getTokens(line string) []Token {
	var tokens []Token
	var buff = ""
	var stack = runeStack{}
	var previous rune
	for pos, char := range line {
		if char == '#' {
			break
		}

		var last = stack.Last()
		if char == ' ' || char == '\t' || char == '\r' {
			if last == none {
				if buff != "" {
					tokens = append(tokens, Token{
						Value: cleanText(buff),
						Pos:   pos - len(buff),
					})
				}
				buff = ""
				continue
			}
		}
		buff += string(char)
		if previous == '\\' && (char == '"' || char == '\'') {
			continue
		}

		switch char {
		case '(':
			stack.push(')')
		case '`':
			if last != '`' {
				stack.push('`')
			}
		case '"':
			if last != '"' {
				stack.push('"')
			}
		case '\'':
			if last != '\'' {
				stack.push('\'')
			}
		case '[':
			stack.push(']')
		case '{':
			if !stack.isEmpty() {
				stack.push('}')
			}

		}

		if last != none && last == char {
			stack.pop()
		}
		previous = char
	}
	if buff != "" {
		tokens = append(tokens, Token{
			Value: cleanText(buff),
			Pos:   len(line) - len(buff),
		})
	}
	return tokens
}

func cleanText(t string) string {
	t = strings.TrimSpace(t)
	var length = len(t)
	if length > 2 {
		if (t[0] == '"' && t[length-1] == '"') ||
			(t[0] == '\'' && t[length-1] == '\'') ||
			(t[0] == '(' && t[length-1] == ')') ||
			(t[0] == '`' && t[length-1] == '`') ||
			(t[0] == '<' && t[length-1] == '>') {
			t = t[1 : length-1]
		}
	}
	return t
}

func PrettyStruct(data interface{}) string {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(val)
}
