package timemark

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

//only one alert will present if set AlertAtStart().AlertIfMore().AlertIfLess()
const (
	START      = iota //we begin to inspect function
	MORE_LIMIT        //function exited with exceeding time limit set
	LESS_LIMIT
	FINISH //function exited without time limit
)

var alertTypeStr = []string{
	"START",
	"FINISH",
	"MORE_LIMIT",
	"LESS_LIMIT",
}

type AlertData struct {
	AlertType    int
	AlertTypeStr string

	File     string //invoker's file name
	Function string //invoker's function name
	Line     int    //invoker's file pos

	Callers []uintptr //stack frame  for runtime.CallersFrame for future investigation if needed

	When  time.Time //when happened
	Spent time.Duration
}

//formats AlertData.Callers into:
// func1:23 -> :35 -> func2:12 -> func3:23
//not more then count are exported
func (a *AlertData) CallersTree(count int) string {
	if count > len(a.Callers) {
		count = len(a.Callers)
	}
	ret := ""
	prev_fl := ""
	for i, v := range a.Callers {
		if i < len(a.Callers)-count {
			continue
		}
		fn := runtime.FuncForPC(v)

		var fl string
		var ln int
		if fn != nil {
			fl, ln = fn.FileLine(v)
		}
		if ln > 0 && !(fl == "asm_amd64.s" && ln == 1582) {
			if len(ret) != 0 {
				ret += "->"
			}
			if prev_fl == fl {
				fl = ""
			} else {
				prev_fl = fl
			}
			if fl != "" {
				is := strings.LastIndex(fl, "/")
				if is > 0 {
					fl = fl[is+1:]
				}
			}

			ret += fmt.Sprintf("%s:%d", fl, ln)
		}
	}
	return ret
}

//called to inform your function spend not such time as exptected
type AlertFunc func(a *AlertData)

type timeMarker struct {
	af AlertFunc

	tmLimits
}

func defaultAlertFunction(a *AlertData) {
	fmt.Printf("%s:%d function [%s] worked %s !\n", a.File, a.Line, a.Function, a.Spent.Truncate(time.Millisecond).String())
}

func New(af AlertFunc) *timeMarker {
	if af == nil {
		af = defaultAlertFunction
	}

	return &timeMarker{
		tmLimits: tmLimits{
			MoreLimit: time.Duration(100 * 365 * 24 * time.Hour), //not set
			LessLimit: time.Duration(0),                          //not set
		},

		af: af, //alert function
	}
}

type tmLimits struct {
	MoreLimit time.Duration
	LessLimit time.Duration

	alertAtStart bool //whether we log starting to watch function (when it launched)
	alertAtEnd   bool
}

func (tm *timeMarker) AlertAtStart() *timeMarker {
	tm.alertAtStart = true
	return tm
}

func (tm *singleChecker) AlertAtStart() *singleChecker {
	tm.alertAtStart = true
	return tm
}

func (tm *timeMarker) AlertAtEnd() *timeMarker {
	tm.alertAtEnd = true
	return tm
}

func (tm *singleChecker) AlertAtEnd() *singleChecker {
	tm.alertAtEnd = true
	return tm
}

func (tm *timeMarker) AlertIfMore(t time.Duration) *timeMarker {
	atomic.StoreInt64((*int64)(&tm.MoreLimit), int64(t))
	return tm
}

func (tm *timeMarker) AlertIfLess(t time.Duration) *timeMarker {
	atomic.StoreInt64((*int64)(&tm.LessLimit), int64(t))
	return tm
}

func (tm *singleChecker) AlertIfMore(t time.Duration) *singleChecker {
	atomic.StoreInt64((*int64)(&tm.MoreLimit), int64(t))
	return tm
}

func (tm *singleChecker) AlertIfLess(t time.Duration) *singleChecker {
	atomic.StoreInt64((*int64)(&tm.LessLimit), int64(t))
	return tm
}

type singleChecker struct {
	Line     int
	File     string
	Function string    //caller function name
	Callers  []uintptr //stack callers functions ierarchy

	start time.Time
	tm    *timeMarker

	tmLimits
}

var replacer = strings.NewReplacer("command-line-arguments.", "")

func (tm *timeMarker) Get() *singleChecker {
	ret := &singleChecker{
		start: time.Now(),
		tm:    tm,

		tmLimits: tm.tmLimits,

		Callers: make([]uintptr, 30),
	}

	wr := runtime.Callers(1, ret.Callers)
	if wr < 30 {
		ret.Callers = ret.Callers[:wr]
	}

	if wr > 0 {
		fni := 0
		if wr > 2 {
			fni = 2
		}

		fn := runtime.FuncForPC(ret.Callers[fni])
		ret.File, ret.Line = fn.FileLine(ret.Callers[fni])
		ret.Function = replacer.Replace(fn.Name())

		// взять +1 строку после данной
		// command-line-arguments.(*singleChecker).AlertAtStart
		for i := 1; i < wr; i++ {
			fn = runtime.FuncForPC(ret.Callers[i])
			if fn != nil {
				fname := fn.Name()
				if strings.Contains(fname, "command-line-arguments") &&
					strings.Contains(fname, "*singleChecker") {
					continue
				}
				ret.Function = replacer.Replace(fn.Name())
				ret.File, ret.Line = fn.FileLine(ret.Callers[i])
				break
			}
		}
	}

	callers := make([]uintptr, len(ret.Callers))
	j := 0
	for i := wr - 1; i >= 0; i-- {
		callers[j] = ret.Callers[i]
		j++
	}
	ret.Callers = callers

	if tm.alertAtStart {
		tm.af(&AlertData{
			AlertType:    START,
			AlertTypeStr: alertTypeStr[START],

			Function: ret.Function,
			File:     ret.File,
			Line:     ret.Line,

			Callers: ret.Callers,

			When: ret.start,
		})
	}

	return ret
}

//invokes alert if checked function work more or less time expected
func (sc *singleChecker) Check() {
	now := time.Now()

	if now.After(sc.start.Add(sc.MoreLimit)) {
		sc.tm.af(&AlertData{
			AlertType:    MORE_LIMIT,
			AlertTypeStr: alertTypeStr[MORE_LIMIT],

			Function: sc.Function,
			File:     sc.File,
			Line:     sc.Line,

			Callers: sc.Callers,

			When:  now,
			Spent: time.Since(sc.start),
		})
		return
	}

	if now.Before(sc.start.Add(sc.LessLimit)) {
		sc.tm.af(&AlertData{
			AlertType:    LESS_LIMIT,
			AlertTypeStr: alertTypeStr[LESS_LIMIT],

			Function: sc.Function,
			File:     sc.File,
			Line:     sc.Line,

			Callers: sc.Callers,

			When:  now,
			Spent: time.Since(sc.start),
		})
		return
	}

	if sc.alertAtEnd {
		sc.tm.af(&AlertData{
			AlertType:    FINISH,
			AlertTypeStr: alertTypeStr[FINISH],

			Function: sc.Function,
			File:     sc.File,
			Line:     sc.Line,

			Callers: sc.Callers,

			When:  now,
			Spent: time.Since(sc.start),
		})
	}
}
