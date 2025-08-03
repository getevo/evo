# Version Library

The Version library provides comprehensive version string handling capabilities for Go applications. It offers functions for comparing, normalizing, and validating version strings in various formats, making it easy to implement version-based logic in your applications.

## Installation

```go
import "github.com/getevo/evo/v2/lib/version"
```

## Features

- **Version Comparison**: Compare version strings with support for various operators
- **Version Normalization**: Normalize version strings to a standard format
- **Version Constraints**: Define and check version constraints (e.g., ">= 1.0.0, < 2.0.0")
- **Semantic Versioning Support**: Handle semantic versioning with pre-release and build metadata
- **Special Version Formats**: Support for development, alpha, beta, RC versions
- **Wildcard Support**: Handle wildcard patterns in version constraints
- **Tilde Operator**: Support for next significant release operator (~)

## Usage Examples

### Basic Version Comparison

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Compare two version strings
    result := version.Compare("2.3.4", "v3.1.2", "<")
    fmt.Printf("Is 2.3.4 < v3.1.2? %t\n", result) // true
    
    // Compare with different operators
    fmt.Printf("Is 1.0rc1 >= 1.0? %t\n", version.Compare("1.0rc1", "1.0", ">=")) // false
    fmt.Printf("Is 1.0.1 > 1.0.0? %t\n", version.Compare("1.0.1", "1.0.0", ">")) // true
    fmt.Printf("Is 2.0.0 = 2.0? %t\n", version.Compare("2.0.0", "2.0", "=")) // true
    fmt.Printf("Is 1.5.2 != 1.5.3? %t\n", version.Compare("1.5.2", "1.5.3", "!=")) // true
}
```

### Simple Version Comparison

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // CompareSimple returns:
    // -1 if version1 < version2
    //  0 if version1 = version2
    //  1 if version1 > version2
    
    result := version.CompareSimple("1.2", "1.0.1")
    fmt.Printf("Comparing 1.2 and 1.0.1: %d\n", result) // 1 (1.2 > 1.0.1)
    
    result = version.CompareSimple("1.0rc1", "1.0")
    fmt.Printf("Comparing 1.0rc1 and 1.0: %d\n", result) // -1 (1.0rc1 < 1.0)
    
    result = version.CompareSimple("2.0.0", "2.0")
    fmt.Printf("Comparing 2.0.0 and 2.0: %d\n", result) // 0 (equal)
}
```

### Version Normalization

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Normalize version strings to a standard format
    fmt.Printf("10.4.13-b normalized: %s\n", version.Normalize("10.4.13-b")) // 10.4.13.0-beta
    fmt.Printf("1.0.0alpha1 normalized: %s\n", version.Normalize("1.0.0alpha1")) // 1.0.0.0-alpha1
    fmt.Printf("2.0.0-rc.1 normalized: %s\n", version.Normalize("2.0.0-rc.1")) // 2.0.0.0-rc1
    fmt.Printf("v1.2 normalized: %s\n", version.Normalize("v1.2")) // 1.2.0.0
    fmt.Printf("1.0-dev normalized: %s\n", version.Normalize("1.0-dev")) // 1.0.0.0-dev
    
    // Master-like branches are normalized to a very high version
    fmt.Printf("master normalized: %s\n", version.Normalize("master")) // 9999999-dev
    fmt.Printf("trunk normalized: %s\n", version.Normalize("trunk")) // 9999999-dev
}
```

### Version Constraints

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Create a single constraint
    constraint := version.NewConstrain(">=", "1.0.0")
    fmt.Printf("Does 1.1.0 match >= 1.0.0? %t\n", constraint.Match("1.1.0")) // true
    fmt.Printf("Does 0.9.0 match >= 1.0.0? %t\n", constraint.Match("0.9.0")) // false
    
    // Create a constraint group from a string
    group := version.NewConstrainGroupFromString(">2.0,<=3.0")
    fmt.Printf("Does 2.5.0 match >2.0,<=3.0? %t\n", group.Match("2.5.0")) // true
    fmt.Printf("Does 3.5.0 match >2.0,<=3.0? %t\n", group.Match("3.5.0")) // false
    fmt.Printf("Does 1.9.0 match >2.0,<=3.0? %t\n", group.Match("1.9.0")) // false
    
    // Using the tilde operator for next significant release
    group = version.NewConstrainGroupFromString("~1.2.3")
    fmt.Printf("Does 1.2.3 match ~1.2.3? %t\n", group.Match("1.2.3")) // true
    fmt.Printf("Does 1.2.9 match ~1.2.3? %t\n", group.Match("1.2.9")) // true
    fmt.Printf("Does 1.3.0 match ~1.2.3? %t\n", group.Match("1.3.0")) // false
    
    // Using wildcards
    group = version.NewConstrainGroupFromString("1.0.*")
    fmt.Printf("Does 1.0.5 match 1.0.*? %t\n", group.Match("1.0.5")) // true
    fmt.Printf("Does 1.1.0 match 1.0.*? %t\n", group.Match("1.1.0")) // false
}
```

### Handling Pre-release Versions

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Pre-release versions are ordered: dev < alpha/a < beta/b < RC/rc < stable < patch/pl/p
    
    fmt.Printf("Is 1.0-dev < 1.0-alpha? %t\n", version.Compare("1.0-dev", "1.0-alpha", "<")) // true
    fmt.Printf("Is 1.0-alpha < 1.0-beta? %t\n", version.Compare("1.0-alpha", "1.0-beta", "<")) // true
    fmt.Printf("Is 1.0-beta < 1.0-rc? %t\n", version.Compare("1.0-beta", "1.0-rc", "<")) // true
    fmt.Printf("Is 1.0-rc < 1.0? %t\n", version.Compare("1.0-rc", "1.0", "<")) // true
    fmt.Printf("Is 1.0 < 1.0-p1? %t\n", version.Compare("1.0", "1.0-p1", "<")) // true
    
    // Numeric suffixes are compared numerically
    fmt.Printf("Is 1.0-alpha1 < 1.0-alpha2? %t\n", version.Compare("1.0-alpha1", "1.0-alpha2", "<")) // true
    fmt.Printf("Is 1.0-beta3 < 1.0-beta10? %t\n", version.Compare("1.0-beta3", "1.0-beta10", "<")) // true
}
```

### Validating Version Format

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Check if a string is in a valid version format
    fmt.Printf("Is '1.0.0' a valid version? %t\n", version.ValidSimpleVersionFormat("1.0.0")) // true
    fmt.Printf("Is 'v2.1.3-rc1' a valid version? %t\n", version.ValidSimpleVersionFormat("v2.1.3-rc1")) // true
    fmt.Printf("Is '1.x.3' a valid version? %t\n", version.ValidSimpleVersionFormat("1.x.3")) // false
    fmt.Printf("Is 'not-a-version' a valid version? %t\n", version.ValidSimpleVersionFormat("not-a-version")) // false
}
```

## Advanced Usage

### Creating Custom Version Constraints

You can create custom version constraints by combining multiple constraints:

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Create a custom constraint group
    group := version.NewConstrainGroup()
    
    // Add constraints to the group
    group.AddConstraint(version.NewConstrain(">=", "1.0.0"))
    group.AddConstraint(version.NewConstrain("<", "2.0.0"))
    group.AddConstraint(version.NewConstrain("!=", "1.5.0"))
    
    // Check versions against the constraint group
    fmt.Printf("Does 1.2.0 match the constraints? %t\n", group.Match("1.2.0")) // true
    fmt.Printf("Does 1.5.0 match the constraints? %t\n", group.Match("1.5.0")) // false (excluded)
    fmt.Printf("Does 2.0.0 match the constraints? %t\n", group.Match("2.0.0")) // false (too high)
    fmt.Printf("Does 0.9.0 match the constraints? %t\n", group.Match("0.9.0")) // false (too low)
}
```

### Working with Version Ranges

The library supports various ways to specify version ranges:

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/version"
)

func main() {
    // Exact version
    group := version.NewConstrainGroupFromString("1.0.2")
    fmt.Printf("Does 1.0.2 match '1.0.2'? %t\n", group.Match("1.0.2")) // true
    fmt.Printf("Does 1.0.3 match '1.0.2'? %t\n", group.Match("1.0.3")) // false
    
    // Range with comparison operators
    group = version.NewConstrainGroupFromString(">=1.0,<2.0")
    fmt.Printf("Does 1.5.0 match '>=1.0,<2.0'? %t\n", group.Match("1.5.0")) // true
    
    // Wildcard
    group = version.NewConstrainGroupFromString("1.0.*")
    fmt.Printf("Does 1.0.5 match '1.0.*'? %t\n", group.Match("1.0.5")) // true
    
    // Tilde operator (next significant release)
    group = version.NewConstrainGroupFromString("~1.2")
    fmt.Printf("Does 1.2.9 match '~1.2'? %t\n", group.Match("1.2.9")) // true
    fmt.Printf("Does 1.3.0 match '~1.2'? %t\n", group.Match("1.3.0")) // false
}
```

## How It Works

The Version library provides a set of functions for working with version strings:

1. **Normalization**: The `Normalize` function converts various version formats into a standardized format for comparison. It handles different version string formats, including those with pre-release identifiers and build metadata.

2. **Comparison**: The `Compare` and `CompareSimple` functions compare two version strings. They first normalize the versions and then compare them component by component, taking into account the special ordering of pre-release identifiers.

3. **Constraints**: The `Constraint` struct represents a single version constraint with an operator and a version. The `ConstraintGroup` struct represents a group of constraints that must all be satisfied.

4. **Parsing**: The library includes functions for parsing version strings and constraint expressions, handling various formats and syntaxes.

The library follows these rules for version comparison:

- Version components are compared numerically (e.g., 1.10 > 1.2)
- Pre-release versions are ordered: dev < alpha/a < beta/b < RC/rc < stable < patch/pl/p
- Numeric suffixes are compared numerically (e.g., beta2 < beta10)
- Master-like branches (master, trunk, default) are treated as very high versions
