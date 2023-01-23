package timemark

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

func Test_NeverAlert(t *testing.T) {
	var neverAlert = func(t *testing.T) func(a *AlertData) {
		return func(a *AlertData) {
			t.Fatal()
		}
	}

	tm1 := New(neverAlert(t))
	defer tm1.Get().Check()
}

var alerts_called = 0
var alert_function = func(s string, t *testing.T) func(a *AlertData) {
	var alertOnTimeMarkExceeding = func(s string, a *AlertData) {
		fmt.Printf("[NOT DEFAULT,  type: %s, where: %s] %s:%d function [%s] worked %s\n", a.AlertTypeStr, s, a.File, a.Line, a.Function, a.Spent.Truncate(time.Millisecond).String())
	}
	return func(a *AlertData) {
		alerts_called++
		alertOnTimeMarkExceeding(s, a)
		if !strings.Contains(a.File, "timemark_test.go") {
			callers := make([]uintptr, 30)
			wr := runtime.Callers(0, callers)

			for i := 0; i < wr; i++ {
				fn := runtime.FuncForPC(callers[i])
				file, line := fn.FileLine(callers[i])
				fmt.Printf(">> %d >> %s:%d\n", i, file, line)
			}

			t.Fatal(a.Function + " - incorrect position")
		}
	}
}

func Test_AlertAtStart(t *testing.T) {
	alerts_called = 0
	func() {
		defer New(alert_function("tm2s", t)).AlertAtStart().Get().Check()
	}()
	if alerts_called != 1 {
		t.Fatal()
	}
}

func Test_AlertEnums(t *testing.T) {
	var checkAlertType = func(at int, t *testing.T) func(a *AlertData) {
		return func(a *AlertData) {
			if a.AlertType != at {
				t.Fatal()
			}
		}
	}

	func() {
		defer New(checkAlertType(START, t)).AlertAtStart().Get().Check()
	}()

	func() {
		defer New(checkAlertType(FINISH, t)).AlertAtEnd().Get().Check()
	}()

	func() {
		defer New(checkAlertType(LESS_LIMIT, t)).AlertIfLess(time.Second).Get().Check()
	}()
	func() {
		defer New(checkAlertType(MORE_LIMIT, t)).AlertIfMore(time.Microsecond).Get().Check()
		time.Sleep(time.Millisecond)
	}()

}

func Test_AlertAtEnd(t *testing.T) {
	alerts_called = 0
	func() {
		defer New(alert_function("tm3e", t)).AlertAtEnd().Get().Check()
	}()
	if alerts_called != 1 {
		t.Fatal()
	}

	alerts_called = 0
	func() {
		defer New(alert_function("tm3e2", t)).Get().AlertAtEnd().Check()
	}()
	if alerts_called != 1 {
		t.Fatal()
	}

	New(nil).Get().Check()
}

func Test_AlertIfMore(t *testing.T) {
	var tm1a = New(alert_function("tm1a3", t)).AlertIfMore(200 * time.Millisecond)

	alerts_called = 0
	func() {
		defer tm1a.Get().Check()
		time.Sleep(time.Millisecond * 250)
	}()
	if alerts_called != 1 {
		t.Fatal()
	}

	alerts_called = 0
	func() {
		defer tm1a.Get().Check()
		defer tm1a.Get().Check()
		time.Sleep(time.Millisecond * 250)
	}()
	if alerts_called != 2 {
		t.Fatal()
	}
}

func Test_AlertIfMore2(t *testing.T) {
	var tm1a = New(alert_function("tm1a2", t))

	alerts_called = 0
	func() {
		defer tm1a.Get().AlertIfMore(200 * time.Millisecond).Check()
		time.Sleep(time.Millisecond * 250)
	}()
	if alerts_called != 1 {
		t.Fatal()
	}
}

func Test_AlertIfMore3(t *testing.T) {
	var tm1a = New(alert_function("tm1a2b", t)).AlertIfMore(time.Millisecond)

	alerts_called = 0
	func() {
		defer tm1a.Get().AlertIfMore(200 * time.Millisecond).Check()
		time.Sleep(time.Millisecond * 50)
	}()
	if alerts_called != 0 {
		t.Fatal()
	}
}

func Test_AlertIfLess(t *testing.T) {
	var tm1a = New(alert_function("tm1a", t)).AlertIfLess(200 * time.Millisecond)

	alerts_called = 0
	func() {
		defer tm1a.Get().Check()
	}()
	if alerts_called != 1 {
		t.Fatal()
	}

	alerts_called = 0
	func() {
		defer tm1a.Get().Check()
		defer tm1a.Get().Check()
		time.Sleep(10 * time.Millisecond)
	}()
	if alerts_called != 2 {
		t.Fatal()
	}
}

func Test_AlertIfLess2(t *testing.T) {
	var tm1a = New(alert_function("tm1a2", t))

	alerts_called = 0
	func() {
		defer tm1a.Get().AlertIfLess(200 * time.Millisecond).Check()
	}()
	if alerts_called != 1 {
		t.Fatal()
	}
}

func a50(t *testing.T) {
	var tm100_300 = New(alert_function("tm3d", t)).AlertAtStart().AlertAtEnd().AlertIfLess(100 * time.Millisecond).AlertIfMore(300 * time.Millisecond)
	defer tm100_300.Get().Check() //2 times: start , less 100
	time.Sleep(50 * time.Millisecond)
}

func a350(t *testing.T) {
	var tm100_300 = New(alert_function("tm3d", t)).AlertAtStart().AlertAtEnd().AlertIfLess(100 * time.Millisecond).AlertIfMore(300 * time.Millisecond)
	defer tm100_300.Get().Check() //2 times: start , more 100
	time.Sleep(350 * time.Millisecond)
}

func Test_a50(t *testing.T) {
	alerts_called = 0
	a50(t)
	if alerts_called != 2 {
		t.Fatal(alerts_called)
	}
	alerts_called = 0
	a350(t)
	if alerts_called != 2 {
		t.Fatal(alerts_called)
	}
}

func a250(t *testing.T) {
	var tm200_300 = New(alert_function("tm4a", t))
	defer tm200_300.Get().AlertIfLess(200 * time.Millisecond).AlertIfMore(300 * time.Millisecond).Check() //never
	time.Sleep(250 * time.Millisecond)
}

func Test_a250(t *testing.T) {
	alerts_called = 0
	a250(t)
	if alerts_called != 0 {
		t.Fatal(alerts_called)
	}
}

type obj struct {
}

var ob obj

func (o *obj) c(t *testing.T) {
	var tm100_300 = New(alert_function("tm4b", t)).AlertAtStart().AlertAtEnd().AlertIfLess(100 * time.Millisecond).AlertIfMore(300 * time.Millisecond)
	defer tm100_300.Get().Check() //2 times: start, more
	time.Sleep(350 * time.Millisecond)
}

func ct(t *testing.T) {
	ob.c(t)
}

func ct2(t *testing.T) {
	ct(t)
}

func Test_level1(t *testing.T) {
	alerts_called = 0
	func() {
		ct(t)
	}()
	if alerts_called != 2 {
		t.Fatal(alerts_called)
	}
}

func Test_defaultAlert(t *testing.T) {
	defer New(nil).AlertAtStart().Get().AlertAtEnd().Check()
	//triggers [START] C:/gopath/src/github.com/MasterDimmy/timemark/timemark_test.go:266 function [Test_defaultAlert2.func1] worked 0s !
}

func Test_defaultAlert2(t *testing.T) {
	defer New(nil).AlertAtStart().AlertAtEnd().Get().Check()
	//triggers [FINISH] C:/gopath/src/github.com/MasterDimmy/timemark/timemark_test.go:267 function [Test_defaultAlert2.func1] worked 0s !
}

func Test_changedTimeAlert(t *testing.T) {
	errs := 0

	alert := func(a *AlertData) {
		errs++
		alert_function("changedTime", t)(a)
	}

	tm := New(alert).AlertIfMore(time.Second)

	tm2 := tm.AlertIfMore(time.Second * 2)

	if tm.moreLimit != time.Second || tm2.moreLimit != time.Second*2 {
		t.Fatal("not equal")
	}

	func() {
		defer tm.Get().Check()
		defer tm2.Get().Check()
		time.Sleep(time.Second + time.Second/2)
	}()

	if errs != 1 {
		t.Fatal("must be 1 alert!")
	}
}
