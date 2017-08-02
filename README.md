# Lo

[![license](https://img.shields.io/github/license/dyson/lo.svg)](https://github.com/dyson/lo/blob/master/LICENSE)

A minimal logger for Go

Lo is a fork and modification to the official Go flag package (https://golang.org/pkg/flag/). It's retained only the Printf function for writing messages and has added three levels of logging (none, info, and debug). Lo's been influenced by a number of articles, discussion and personal experience.

Things to note:
- With the log level set to lo.LevelNone, lo isn't nop. Using a single Printf function this isn't possible but it is minimal. 
- With the log level set to info, debug messages aren't nop for the same reason above.
- INFO and DEBUG are prefixed to the format string for easier log passing.
- Lo only implements Printf as this is the minimal interface a logger needs. This makes switching out this logger for something else in the event you need a different features easy. [(discussion)](https://groups.google.com/forum/#!msg/golang-dev/F3l9Iz1JX4g/szAb07lgFAAJ)
- Lo only has INFO and DEBUG levels. From experience that's all I use. [(Dave Cheney)](https://dave.cheney.net/2015/11/05/lets-talk-about-logging)
- The package level logger has been removed so you need to create a lo logger and pass that around your app. [(Peter Bourgon)](https://peter.bourgon.org/blog/2017/06/09/theory-of-modern-go.html)

## Installation
Using dep for dependency management (https://github.com/golang/dep):
```
dep ensure github.com/dyson/lo
```

Using go get:
```
$ go get github.com/dyson/lo
```
## Usage

```
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dyson/lo"
)

func main() {
	logger := lo.New(os.Stdout, "", log.LstdFlags)

	fmt.Println("show only info messages by default:")
	logger.Printf("info message")
	logger.Printf("debug: debug message with a space")
	logger.Printf("debug:debug message without a space")

	fmt.Println("\nshow info and debug messages:")
	logger.SetLevel(lo.LevelDebug)
	logger.Printf("info message")
	logger.Printf("debug: debug message with a space")
	logger.Printf("debug:debug message without a space")

	fmt.Println("\ndisabled logging:")
	logger.SetLevel(lo.LevelNone)
	logger.Printf("info message")
	logger.Printf("debug: debug message with a space")
	logger.Printf("debug:debug message without a space")
}
```

Running example:
```
$ go run main.go
go run main.go
show only info messages by default:
2017/08/02 11:15:07 INFO info message

show info and debug messages:
2017/08/02 11:15:07 INFO info message
2017/08/02 11:15:07 DEBUG debug message with a space
2017/08/02 11:15:07 DEBUG debug message without a space

disabled logging:
```

When passing around the logger you should accept an interface: [(Dave Cheney)](https://dave.cheney.net/2017/01/23/the-package-level-logger-anti-pattern)

```
type logger interface {
	Printf(string, ...interface{})
}
```

Using the logger interface above you can easily implement a nop logger if you want to disable logging:
```
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dyson/lo"
)

type logger interface {
	Printf(string, ...interface{})
}

type nopLogger struct{}

func (l *nopLogger) Printf(format string, v ...interface{}) {}

func main() {
	var logger logger
	disableLogging := true // set from ENV var or command line flag, etc
	if disableLogging {
		logger = &nopLogger{}
	} else {
		logger = lo.New(os.Stdout, "", log.LstdFlags)
	}

	fmt.Println("nop logger:")
	logger.Printf("info message")
	logger.Printf("debug: debug message")

	disableLogging = false
	if disableLogging {
		logger = &nopLogger{}
	} else {
		logger = lo.New(os.Stdout, "", log.LstdFlags)
	}

	fmt.Println("\nlo logger:")
	logger.Printf("info message")
	logger.Printf("debug: debug message")

}
```

Running example:
```
$ go run main.go
nop logger:

lo logger:
2017/08/02 13:39:14 INFO info message
```
## License
See LICENSE file.