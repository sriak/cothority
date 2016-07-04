// Package log is an output-library that can print nicely formatted
// messages to the screen.
//
// There are log-level messages that will be printed according to the
// current debug-level set. Furthermore a set of common messages exist
// that are printed according to a chosen format.
//
// The log-level messages are:
//	log.Lvl1("Important information")
//	log.Lvl2("Less important information")
//	log.Lvl3("Eventually flooding information")
//	log.Lvl4("Definitively flooding information")
//	log.Lvl5("I hope you never need this")
// in your program, then according to the debug-level one or more levels of
// output will be shown. To set the debug-level, use
//	log.SetDebugVisible(3)
// which will show all `Lvl1`, `Lvl2`, and `Lvl3`. If you want to turn
// on just one output, you can use
//	log.LLvl2("Less important information")
// By adding a single 'L' to the method, it *always* gets printed.
//
// You can also add a 'f' to the name and use it like fmt.Printf:
//	log.Lvlf1("Level: %d/%d", now, max)
//
// The common messages are:
//	log.Print("Simple output")
//	log.Info("For your information")
//	log.Warn("Only a warning")
//	log.Error("This is an error, but continues")
//	log.Panic("Something really went bad - calls panic")
//	log.Fatal("No way to continue - calls os.Exit")
//
// These messages are printed according to the value of 'Format':
// - Format == FormatLvl - same as log.Lvl
// - Format == FormatPython - with some nice python-style formatting
// - Format == FormatNone - just as plain text
//
// The log-package also takes into account the following environment-variables:
//	DEBUG_LVL // will act like SetDebugVisible
//	DEBUG_TIME // if 'true' it will print the date and time
//	DEBUG_COLOR // if 'false' it will not use colors
// But for this the function ParseEnv() or AddFlags() has to be called.
package log

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/daviddengcn/go-colortext"
)

// For debugging purposes we can change the output-writer
var stdOut io.Writer
var stdErr io.Writer

func init() {
	stdOut = os.Stdout
	stdErr = os.Stderr
}

const (
	lvlWarning = iota - 20
	lvlError
	lvlFatal
	lvlPanic
	lvlInfo
	lvlPrint
)

// These formats can be used in place of the debugVisible
const (
	// FormatPython uses [x] and others to indicate what is shown
	FormatPython = -1
	// FormatNone is just pure print
	FormatNone = 0
)

// NamePadding - the padding of functions to make a nice debug-output - this is automatically updated
// whenever there are longer functions and kept at that new maximum. If you prefer
// to have a fixed output and don't remember oversized names, put a negative value
// in here
var NamePadding = 40

// LinePadding of line-numbers for a nice debug-output - used in the same way as
// NamePadding
var LinePadding = 3

// StaticMsg - if this variable is set, it will be outputted between the
// position and the message
var StaticMsg = ""

// These are information-debugging levels that can be turned on or off.
// Every logging greater than 'DebugVisible' will be discarded. So you can
// Log at different levels and easily turn on or off the amount of logging
// generated by adjusting the 'DebugVisible' variable.
var debugVisible = 1

// If showTime is true, it will print the time for each line of debug-output
var showTime = false

// If useColors is true, debug-output will be colored
var useColors = true

var debugMut sync.RWMutex

// outputLines can be false to suppress outputting of lines in tests
var outputLines = true

var regexpPaths, _ = regexp.Compile(".*/")

func lvl(lvl, skip int, args ...interface{}) {
	debugMut.Lock()
	defer debugMut.Unlock()

	if lvl > debugVisible {
		return
	}
	pc, _, line, _ := runtime.Caller(skip)
	name := regexpPaths.ReplaceAllString(runtime.FuncForPC(pc).Name(), "")
	lineStr := fmt.Sprintf("%d", line)

	// For the testing-framework, we check the resulting string. So as not to
	// have the tests fail every time somebody moves the functions, we put
	// the line-# to 0
	if !outputLines {
		line = 0
	}

	if len(name) > NamePadding && NamePadding > 0 {
		NamePadding = len(name)
	}
	if len(lineStr) > LinePadding && LinePadding > 0 {
		LinePadding = len(name)
	}
	fmtstr := fmt.Sprintf("%%%ds: %%%dd", NamePadding, LinePadding)
	caller := fmt.Sprintf(fmtstr, name, line)
	if StaticMsg != "" {
		caller += "@" + StaticMsg
	}
	message := fmt.Sprintln(args...)
	bright := lvl < 0
	lvlAbs := lvl
	if bright {
		lvlAbs *= -1
	}
	lvlStr := strconv.Itoa(lvlAbs)
	if lvl < 0 {
		lvlStr += "!"
	}
	switch lvl {
	case lvlPrint:
		fg(ct.White, true)
		lvlStr = "I"
	case lvlWarning:
		fg(ct.Green, true)
		lvlStr = "W"
	case lvlError:
		fg(ct.Red, false)
		lvlStr = "E"
	case lvlFatal:
		fg(ct.Red, true)
		lvlStr = "F"
	case lvlPanic:
		fg(ct.Red, true)
		lvlStr = "P"
	default:
		if lvl != 0 {
			if lvlAbs <= 5 {
				colors := []ct.Color{ct.Yellow, ct.Cyan, ct.Green, ct.Blue, ct.Cyan}
				fg(colors[lvlAbs-1], bright)
			}
		}
	}
	str := fmt.Sprintf(": (%s) - %s", caller, message)
	if showTime {
		ti := time.Now()
		str = fmt.Sprintf("%s.%09d%s", ti.Format("06/02/01 15:04:05"), ti.Nanosecond(), str)
	}
	str = fmt.Sprintf("%-2s%s", lvlStr, str)
	if lvl < lvlPrint {
		fmt.Fprint(stdErr, str)
	} else {
		fmt.Fprint(stdOut, str)
	}
	if useColors {
		ct.ResetColor()
	}
}

func fg(c ct.Color, bright bool) {
	if useColors {
		ct.Foreground(c, bright)
	}
}

// Needs two functions to keep the caller-depth the same and find who calls us
// Lvlf1 -> Lvlf -> lvl
// or
// Lvl1 -> lvld -> lvl
func lvlf(l int, f string, args ...interface{}) {
	lvl(l, 3, fmt.Sprintf(f, args...))
}
func lvld(l int, args ...interface{}) {
	lvl(l, 3, args...)
}

// Lvl1 debug output is informational and always displayed
func Lvl1(args ...interface{}) {
	lvld(1, args...)
}

// Lvl2 is more verbose but doesn't spam the stdout in case
// there is a big simulation
func Lvl2(args ...interface{}) {
	lvld(2, args...)
}

// Lvl3 gives debug-output that can make it difficult to read
// for big simulations with more than 100 hosts
func Lvl3(args ...interface{}) {
	lvld(3, args...)
}

// Lvl4 is only good for test-runs with very limited output
func Lvl4(args ...interface{}) {
	lvld(4, args...)
}

// Lvl5 is for big data
func Lvl5(args ...interface{}) {
	lvld(5, args...)
}

// Lvlf1 is like Lvl1 but with a format-string
func Lvlf1(f string, args ...interface{}) {
	lvlf(1, f, args...)
}

// Lvlf2 is like Lvl2 but with a format-string
func Lvlf2(f string, args ...interface{}) {
	lvlf(2, f, args...)
}

// Lvlf3 is like Lvl3 but with a format-string
func Lvlf3(f string, args ...interface{}) {
	lvlf(3, f, args...)
}

// Lvlf4 is like Lvl4 but with a format-string
func Lvlf4(f string, args ...interface{}) {
	lvlf(4, f, args...)
}

// Lvlf5 is like Lvl5 but with a format-string
func Lvlf5(f string, args ...interface{}) {
	lvlf(5, f, args...)
}

// LLvl1 *always* prints
func LLvl1(args ...interface{}) { lvld(-1, args...) }

// LLvl2 *always* prints
func LLvl2(args ...interface{}) { lvld(-2, args...) }

// LLvl3 *always* prints
func LLvl3(args ...interface{}) { lvld(-3, args...) }

// LLvl4 *always* prints
func LLvl4(args ...interface{}) { lvld(-4, args...) }

// LLvl5 *always* prints
func LLvl5(args ...interface{}) { lvld(-5, args...) }

// LLvlf1 *always* prints
func LLvlf1(f string, args ...interface{}) { lvlf(-1, f, args...) }

// LLvlf2 *always* prints
func LLvlf2(f string, args ...interface{}) { lvlf(-2, f, args...) }

// LLvlf3 *always* prints
func LLvlf3(f string, args ...interface{}) { lvlf(-3, f, args...) }

// LLvlf4 *always* prints
func LLvlf4(f string, args ...interface{}) { lvlf(-4, f, args...) }

// LLvlf5 *always* prints
func LLvlf5(f string, args ...interface{}) { lvlf(-5, f, args...) }

// TestOutput sets the DebugVisible to 0 if 'show'
// is false, else it will set DebugVisible to 'level'
//
// Usage: TestOutput( test.Verbose(), 2 )
func TestOutput(show bool, level int) {
	debugMut.Lock()
	defer debugMut.Unlock()

	if show {
		debugVisible = level
	} else {
		debugVisible = 0
	}
}

// SetDebugVisible set the global debug output level in a go-rountine-safe way
func SetDebugVisible(lvl int) {
	debugMut.Lock()
	defer debugMut.Unlock()
	debugVisible = lvl
}

// DebugVisible returns the actual visible debug-level
func DebugVisible() int {
	debugMut.RLock()
	defer debugMut.RUnlock()
	return debugVisible
}

// SetShowTime allows for turning on the flag that adds the current
// time to the debug-output
func SetShowTime(show bool) {
	debugMut.Lock()
	defer debugMut.Unlock()
	showTime = show
}

// ShowTime returns the current setting for showing the time in the debug
// output
func ShowTime() bool {
	debugMut.Lock()
	defer debugMut.Unlock()
	return showTime
}

// SetUseColors can turn off or turn on the use of colors in the debug-output
func SetUseColors(show bool) {
	debugMut.Lock()
	defer debugMut.Unlock()
	useColors = show
}

// UseColors returns the actual setting of the color-usage in log
func UseColors() bool {
	debugMut.Lock()
	defer debugMut.Unlock()
	return useColors
}

// MainTest can be called from TestMain. It will parse the flags and
// set the DebugVisible to 3, then run the tests and check for
// remaining go-routines.
func MainTest(m *testing.M) {
	flag.Parse()
	TestOutput(testing.Verbose(), 3)
	code := m.Run()
	AfterTest(nil)
	os.Exit(code)
}

// ParseEnv looks at the following environment-variables:
//   DEBUG_LVL - for the actual debug-lvl - default is 1
//   DEBUG_TIME - whether to show the timestamp - default is false
//   DEBUG_COLOR - whether to color the output - default is true
func ParseEnv() {
	var err error
	dv := os.Getenv("DEBUG_LVL")
	if dv != "" {
		debugVisible, err = strconv.Atoi(dv)
		Lvl3("Setting level to", dv, debugVisible, err)
		if err != nil {
			Error("Couldn't convert", dv, "to debug-level")
		}
	}
	dt := os.Getenv("DEBUG_TIME")
	if dt != "" {
		showTime, err = strconv.ParseBool(dt)
		Lvl3("Setting showTime to", dt, showTime, err)
		if err != nil {
			Error("Couldn't convert", dt, "to boolean")
		}
	}
	dc := os.Getenv("DEBUG_COLOR")
	if dc != "" {
		useColors, err = strconv.ParseBool(dc)
		Lvl3("Setting useColor to", dc, showTime, err)
		if err != nil {
			Error("Couldn't convert", dc, "to boolean")
		}
	}
}

// RegisterFlags adds the flags and the variables for the debug-control using
// the standard flag-package.
func RegisterFlags() {
	ParseEnv()
	flag.IntVar(&debugVisible, "debug", DebugVisible(), "Change debug level (0-5)")
	flag.BoolVar(&showTime, "debug-time", ShowTime(), "Shows the time of each message")
	flag.BoolVar(&useColors, "debug-color", UseColors(), "Colors each message")
}