# args Package

The Args package is used to check or retrieve the values of application arguments.

## Usage

To use this package, import it in your Go code:
```go
import "github.com/getevo/evo/v2/lib/args"
```

### Functions

It provides two simple functions:

```go
//$ ./myapp -arg1 hello
// This function returns the value of the argument.
var value = args.Get("arg1")
// The output should be "hello".
fmt.Println(value)

// This function returns true if the argument exists.
if args.Exists("arg2") {
    fmt.Println("arg2 exsits!")
}
```
In the above code snippet, we have an example usage of the args package. The first function, args.Get, retrieves the value of the "arg1" argument, which is expected to be passed when running the application. The retrieved value is then printed using fmt.Println.

The second function, args.Exists, checks if the "arg2" argument exists. If it does exist, you can add your desired code or logic within the if statement block.

Please note that the code provided is a simplified example, and you may need to import the necessary packages and initialize args before using it.

#### [< Table of Contents](https://github.com/getevo/evo#table-of-contents)