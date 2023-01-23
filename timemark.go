package timemark

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
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
	"MORE_LIMIT",
	"LESS_LIMIT",
	"FINISH",
}

type AlertData struct {
	AlertType    int
	AlertTypeStr string

	File     string //invoker's file name
	Function string //invoker's function name
	Line     int    //invoker's file pos

	When  time.Time //when happened
	Spent time.Duration
}

//called to inform your function spend not such time as exptected
type AlertFunc func(a *AlertData)

type timeMarker struct {
	af    AlertFunc
	mutex sync.Mutex

	tmLimits
}

func defaultAlertFunction(a *AlertData) {
	fmt.Printf("[%s] %s:%d function [%s] worked %s !\n", a.AlertTypeStr, a.File, a.Line, a.Function, a.Spent.Truncate(time.Millisecond).String())
}

func New(af AlertFunc) *timeMarker {
	if af == nil {
		af = defaultAlertFunction
	}

	return &timeMarker{
		tmLimits: tmLimits{
			moreLimit: time.Duration(100 * 365 * 24 * time.Hour), //not set
			lessLimit: time.Duration(0),                          //not set
		},

		af: af, //alert function
	}
}

type tmLimits struct {
	moreLimit time.Duration
	lessLimit time.Duration

	alertAtStart bool //whether we log starting to watch function (when it launched)
	alertAtEnd   bool
}

func (tm *timeMarker) AlertAtStart() *timeMarker {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.alertAtStart = true
	return tm
}

//singleChecker whould have singleChecker as it never runs Get()

func (tm *timeMarker) AlertAtEnd() *timeMarker {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.alertAtEnd = true
	return tm
}

func (tm *singleChecker) AlertAtEnd() *singleChecker {
	tm.tm.mutex.Lock()
	defer tm.tm.mutex.Unlock()
	tm.alertAtEnd = true
	return tm
}

func (tm *timeMarker) AlertIfMore(t time.Duration) *timeMarker {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.moreLimit = t
	return tm
}

func (tm *timeMarker) AlertIfLess(t time.Duration) *timeMarker {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.lessLimit = t
	return tm
}

func (tm *singleChecker) AlertIfMore(t time.Duration) *singleChecker {
	tm.tm.mutex.Lock()
	defer tm.tm.mutex.Unlock()
	tm.moreLimit = t
	return tm
}

func (tm *singleChecker) AlertIfLess(t time.Duration) *singleChecker {
	tm.tm.mutex.Lock()
	defer tm.tm.mutex.Unlock()
	tm.lessLimit = t
	return tm
}

//do not trigger more then N alerts per D duration for same line of code
func (tm *singleChecker) MaxSameAlertsRate(n int, d time.Duration) *singleChecker {
	tm.tm.mutex.Lock()
	defer tm.tm.mutex.Unlock()
	tm.lessLimit = t
	return tm
}

type singleChecker struct {
	start time.Time
	tm    *timeMarker

	tmLimits
}

var replacer = strings.NewReplacer("command-line-arguments.", "")

func (tm *timeMarker) Get() *singleChecker {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	ret := &singleChecker{
		start: time.Now(),
		tm:    tm,

		tmLimits: tm.tmLimits,
	}

	if tm.alertAtStart {
		file, line, function, _ := getPosition()

		tm.af(&AlertData{
			AlertType:    START,
			AlertTypeStr: alertTypeStr[START],

			Function: function,
			File:     file,
			Line:     line,

			When: ret.start,
		})
	}

	return ret
}

//invokes alert if checked function work more or less time expected
//only one alert will be triggered at finish: LESS_LIMIT , MORE_LIMIT or FINISH
func (sc *singleChecker) Check() {
	sc.tm.mutex.Lock()
	defer sc.tm.mutex.Unlock()

	now := time.Now()
	file, line, function, _ := getPosition()

	if sc.moreLimit > time.Duration(0) && now.After(sc.start.Add(sc.moreLimit)) {
		sc.tm.af(&AlertData{
			AlertType:    MORE_LIMIT,
			AlertTypeStr: alertTypeStr[MORE_LIMIT],

			Function: function,
			File:     file,
			Line:     line,

			When:  now,
			Spent: time.Since(sc.start),
		})
		return
	}

	if sc.lessLimit > time.Duration(0) && now.Before(sc.start.Add(sc.lessLimit)) {
		sc.tm.af(&AlertData{
			AlertType:    LESS_LIMIT,
			AlertTypeStr: alertTypeStr[LESS_LIMIT],

			Function: function,
			File:     file,
			Line:     line,

			When:  now,
			Spent: time.Since(sc.start),
		})
		return
	}

	if sc.alertAtEnd {
		sc.tm.af(&AlertData{
			AlertType:    FINISH,
			AlertTypeStr: alertTypeStr[FINISH],

			Function: function,
			File:     file,
			Line:     line,

			When:  now,
			Spent: time.Since(sc.start),
		})
		return
	}
}

var iAmHere = func() string {
	callers2 := make([]uintptr, 30)
	wr := runtime.Callers(2, callers2)
	callers2 = callers2[:wr]

	fn := runtime.FuncForPC(callers2[0])
	file, _ := fn.FileLine(callers2[0])
	//	fmt.Printf("HERE %s:%d -> %s\n", file, line, function)
	return file
}()

//file, line, function, []callers
func getPosition() (string, int, string, []uintptr) {
	callers2 := make([]uintptr, 30)
	wr := runtime.Callers(2, callers2)
	callers2 = callers2[:wr]

	deep := 0
	found := false
	for i := 0; i < wr; i++ {
		fn := runtime.FuncForPC(callers2[i])
		file, _ := fn.FileLine(callers2[i])
		if file == iAmHere {
			found = true
		}
		if file != iAmHere && found {
			deep = i
			break
		}
	}

	fn := runtime.FuncForPC(callers2[deep])
	file, line := fn.FileLine(callers2[deep])
	function := replacer.Replace(fn.Name())

	return file, line, function, callers2
}
