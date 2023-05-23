# Template Renderer

A Go package `tpl` that provides a template rendering function `Render` to replace placeholders in a source string with corresponding values from a set of parameters.

## Usage

To use this package, import it in your Go code:

```go
import "github.com/getevo/evo/v2/lib/tpl"
```

### Function: Render

The **`Render`** function replaces placeholders in the source string with values from the provided parameters. Placeholders are identified using the following regular expression: (?mi)\$([a-z\.\_\[\]0-9]+)*.

Parameters
- **`src`** (string): The source string containing placeholders to replace.
- **`params`** (variadic interface{}): Parameters used to replace the placeholders. Can be any type of value.

Return Value
- **`string`**: The modified source string with placeholders replaced by corresponding parameter values.

### Example
Here's an example usage of the Render function:

```go

package main

import (
	"fmt"
     "github.com/getevo/evo/v2/lib/tpl"
)

type User struct {
	Name   string
	Family string
}

func main() {
	var text = `Hello $title $user.Name $user.Family you have $sender[0] email From $sender[2][from]($sender[2][user].Name $sender[2][user].Family) at $date[0]:$date[1]:$date[2]`
	fmt.Println(tpl.Render(text, map[string]interface{}{
		"title":  "Mrs",
		"user":   User{Name: "Maria", Family: "Rossy"},
		"sender": []interface{}{1, "empty!", map[string]interface{}{"from": "example@example.com", "user": User{Name: "Marco", Family: "Pollo"}}},
		"date":   []int{10, 15, 20},
	}))
}

```

In the above example, the placeholders in the text string are replaced with the corresponding values from the provided parameters.