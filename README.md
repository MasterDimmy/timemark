# timemark
Golang functions time marker to detect too huge time function consumers when it's not comfortable with pprof.

![coverage](https://img.shields.io/badge/coverage-100%25-yellowgreen)

Usage:
```
func timeMarkAlert(a *timemark.AlertData) {
	fmt.Printf("timemark 3 seconds: %+v", a)
	fmt.Printf("timemark callers: %s", a.CallersTree(2))
}

var timemark_2second = timemark.New(timeMarkAlert).AlertIfMore(3 * time.Second)

//this should end in 3 seconds
func someLongFunction {
	defer timemark_2second.Get().Check()
	
        time.Sleep(time.Second*5)
}
```


Bix example of usage:

```

type AlertData struct {
	AlertType    int
	AlertTypeStr string

	Function string    //invoker's function name
	Callers  []uintptr //stack frame  for runtime.CallersFrame for future investigation 

	When  time.Time //when happened
	Spent time.Duration
}

//called on error behaviour
func alert_function(a *AlertData) {
	fmt.Printf("%+v", a)
	fmt.Printf("callers: %s", a.CallersTree(2)
}

//time mark controller 1
tm1 := timemark.New(alert_function).AlertIfMore(2*time.Second)


//time mark controller 2
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
