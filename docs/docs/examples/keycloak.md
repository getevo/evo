---
layout: default
title: Keycloak Example
parent: Examples
---

# Keycloak


Example usage of keycloak:

```go
package main

import (
	"github.com/getevo/evo"
	"github.com/getevo/evo/apps/keycloak"
)

func main() {

	evo.Setup()
	keycloak.Register("https://authfootters.ies-italia.it", "footters", "dev")
	evo.Run()
}
```


[Keycloak Example Project](https://github.com/getevo/examples/tree/master/keycloak_example)
