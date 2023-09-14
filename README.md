<h1 align="center">timemark</h1>

<p align="center">
    <a href="#">
        <img src="https://img.shields.io/badge/coverage-100%25-yellowgreen" alt="Coverage">
    </a>
</p>

<p align="center">
    A Golang utility for marking function execution times. It assists in identifying functions that consume excessive time, especially when using pprof is not convenient.
</p>

---

### Installation

```go
import (
	"github.com/MasterDimmy/timemark"
)
```

### Quick Start

To create a timemark alert that warns you if a function runs longer than expected:

```go
func timeMarkAlert(a *timemark.AlertData) {
	fmt.Printf("timemark 3 seconds: %+v", a)
	fmt.Printf("timemark caller: %s:%s", a.File, a.Function)
}

var timemark_2second = timemark.New(timeMarkAlert).AlertIfMore(3 * time.Second)

// This function should complete in 3 seconds
func someLongFunction() {
	defer timemark_2second.Get().Check()
	
    time.Sleep(time.Second*5)
}
```

### Advanced Usage

For a more complex example with multiple configurations:

```go
type AlertData struct {
	AlertType    int
	AlertTypeStr string
	Function     string      // Invoker's function name
	Callers      []uintptr   // Stack frame for runtime.CallersFrame for future investigation 
	When         time.Time   // Timestamp of the alert
	Spent        time.Duration
}

// Called when a timemark alert is triggered
func alert_function(a *AlertData) {
	fmt.Printf("%+v", a)
	fmt.Printf("callers: %s", a.CallersTree(2)
}

// Time mark controller 1
tm1 := timemark.New(alert_function).AlertIfMore(2*time.Second)

// Time mark controller 2
tm2 := timemark.New(alert_function).AlertIfLess(2*time.Second).AlertAtStart().AlertAtEnd()

func b() {
	defer tm1.Get().Check()
	
	c()
	c()
	time.Sleep(1*time.Second)
}

func c() {
	defer tm2.Get().Check()

	time.Sleep(1*time.Second)
}

func a() {
	defer tm1.Get().Check()

	b()
	c()
}
```

---

### Contribution

If you'd like to contribute, please fork the repository and make changes as you'd like. Pull requests are warmly welcome.

---

### License

MIT &copy; MasterDimmy
```
