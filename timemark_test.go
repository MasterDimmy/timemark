package timemark

import (
	"fmt"
	"testing"
	"time"
)

var alerts_called = 0

//called on error behaviour
var alert_function = func(s string) func(a *AlertData) {
	return func(a *AlertData) {
		alerts_called++

		fmt.Printf("[%s]\n%+v\n",
			s,
			a)

		fmt.Println(a.CallersTree(3))
		fmt.Println(a.CallersTree(2))
	}
}

//time mark controller 1
var tm1 = New(alert_function("more")).AlertIfMore(200 * time.Millisecond)

//time mark controller 2
var tm2 = New(alert_function("start and less")).AlertAtStart().AlertIfLess(2 * time.Second)

//time marker, no alerts
var tm3 = New(alert_function("start")).AlertAtStart()

func b() {
	defer tm1.Get().Check()

	c2()
	c2()
	time.Sleep(100 * time.Millisecond)
}

type obj struct {
}

var ob obj

func (o *obj) c() {
	defer tm1.Get().Check()
	defer tm2.Get().Check()

	time.Sleep(100 * time.Millisecond)
}

func c1() {
	ob.c()
}

func c2() {
	c1()
}

func a() {
	defer tm1.Get().Check()
	defer tm3.Get().Check()

	b()
	c2()
}

func Test_TimeMarker(t *testing.T) {
	go c2()

	a()

	//fmt.Printf("alerts: %d\n", alerts_called)

	if alerts_called != 11 {
		t.Fatal()
	}
}
