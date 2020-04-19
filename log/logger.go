// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel string
type LogLevelValue int

const (
	traceLevel LogLevelValue = iota + 1
	debugLevel
	infoLevel
	warningLevel
	errorLevel
	fatalLevel
	TRACE LogLevel = LogLevel("TRACE")
	DEBUG LogLevel = LogLevel("DEBUG")
	INFO  LogLevel = LogLevel("INFO")
	WARN  LogLevel = LogLevel("WARN")
	ERROR LogLevel = LogLevel("ERROR")
	FATAL LogLevel = LogLevel("FATAL")
)

func strRightPad(val string, padding int) string {
	if len(val) < padding {
		val = val + strings.Repeat(" ", padding-len(val))
	} else if len(val) > padding {
		val = val[:(padding-3)] + "..."
	}
	return val
}

func (l LogLevel) String() string {
	return strRightPad(string(l), 5)
}
func LogLevelFromString(level string) LogLevel {
	switch strings.ToUpper(level) {
	case strings.TrimSpace(TRACE.String()):
		return TRACE
	case strings.TrimSpace(DEBUG.String()):
		return DEBUG
	case strings.TrimSpace(WARN.String()):
		return WARN
	case strings.TrimSpace(ERROR.String()):
		return ERROR
	case strings.TrimSpace(FATAL.String()):
		return FATAL
	default:
		return INFO
	}
}

type Logger interface {
	Tracef(format string, in ...interface{})
	Trace(in ...interface{})
	Debugf(format string, in ...interface{})
	Debug(in ...interface{})
	Infof(format string, in ...interface{})
	Info(in ...interface{})
	Warnf(format string, in ...interface{})
	Warn(in ...interface{})
	Errorf(format string, in ...interface{})
	Error(in ...interface{})
	Fatalf(format string, in ...interface{})
	Fatal(in ...interface{})
	Printf(format string, in ...interface{})
	Println(in ...interface{})
	SetVerbosity(verbosity LogLevel)
	GetVerbosity() LogLevel
	Successf(format string, in ...interface{})
	Success(in ...interface{})
	Failuref(format string, in ...interface{})
	Failure(in ...interface{})
	IsAffiliated() bool
	AffiliateTo(l Logger)
	AffiliateLog(affiliateAppName string, level LogLevelValue, in ...interface{})
	AffiliateLogf(affiliateAppName string, level LogLevelValue, format string, in ...interface{})
	AffiliateWrite(affiliateAppName string, buff []byte)
	AddEchoWriter(key string, writer io.Writer)
	RemoveEchoWriter(key string)
}

type logger struct {
	sync.Mutex  // ensures atomic writes; protects the following fields
	verbosity   LogLevelValue
	onScreen    bool
	prefix      string               // prefix on each line to identify the logger (but see Lmsgprefix)
	flag        int                  // properties
	out         io.Writer            // destination for output
	buf         []byte               // for accumulating text to write
	mainLogger  *logger              // Main logger, for affiliated Sub-Loggers
	logRotator  LogRotator           //Log Rotator in case of file writer
	echoWriters map[string]io.Writer //MAp of Writers to echo the logs to

}

func (l *logger) IsAffiliated() bool {
	return l.mainLogger != nil
}

func (lg *logger) AffiliateTo(l Logger) {
	lg.mainLogger = l.(*logger)
}

func (l *logger) AffiliateLog(affiliateAppName string, level LogLevelValue, in ...interface{}) {
	l.logEvent(affiliateAppName, level, in...)
}
func (l *logger) AffiliateLogf(affiliateAppName string, level LogLevelValue, format string, in ...interface{}) {
	l.logEvent(affiliateAppName, level, fmt.Sprintf(format, in...))
}

func (l *logger) AffiliateWrite(affiliateAppName string, buff []byte) {
	var buffer *bytes.Buffer = bytes.NewBuffer([]byte("[" + affiliateAppName + "] "))
	buffer.Write(buff)
	l.write(buffer.Bytes())

}

func (l *logger) Tracef(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, traceLevel, format, in...)
	} else {
		l.log(traceLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Trace(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, traceLevel, in...)
	} else {
		l.log(traceLevel, in...)
	}
}

func (l *logger) Debugf(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, debugLevel, format, in...)
	} else {
		l.log(debugLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Debug(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, debugLevel, in...)
	} else {
		l.log(debugLevel, in...)
	}
}

func (l *logger) Infof(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, infoLevel, format, in...)
	} else {
		l.log(infoLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Info(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, infoLevel, in...)
	} else {
		l.log(infoLevel, in...)
	}
}

func (l *logger) Warnf(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, warningLevel, format, in...)
	} else {
		l.log(warningLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Warn(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, warningLevel, in...)
	} else {
		l.log(warningLevel, in...)
	}
}

func (l *logger) Errorf(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, errorLevel, format, in...)
	} else {
		l.log(errorLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Error(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, errorLevel, in...)
	} else {
		l.log(errorLevel, in...)
	}
}

func (l *logger) Fatalf(format string, in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLogf(l.prefix, fatalLevel, format, in...)
	} else {
		l.log(fatalLevel, fmt.Sprintf(format, in...))
	}
}

func (l *logger) Fatal(in ...interface{}) {
	if l.IsAffiliated() {
		l.mainLogger.AffiliateLog(l.prefix, fatalLevel, in...)
	} else {
		l.log(fatalLevel, in...)
	}
}

func (l *logger) Printf(format string, in ...interface{}) {
	var buf []byte = []byte(fmt.Sprintf(format, in...))
	if l.onScreen {
		color.LightWhite.Printf(string(buf))
	} else {
		if l.IsAffiliated() {
			l.mainLogger.AffiliateWrite(l.prefix, buf)
		} else {
			l.write(buf)
		}
	}
}

func (l *logger) Println(in ...interface{}) {
	var buf []byte = []byte(fmt.Sprint(in...) + "\n")
	if l.onScreen {
		color.LightWhite.Printf(string(buf))
	} else {
		if l.IsAffiliated() {
			l.mainLogger.AffiliateWrite(l.prefix, buf)
		} else {
			l.write(buf)
		}
	}
}

func (l *logger) SetVerbosity(verbosity LogLevel) {
	l.verbosity = toVerbosityLevelValue(verbosity)
}
func (l *logger) GetVerbosity() LogLevel {
	return toVerbosityLevel(l.verbosity)
}

func (l *logger) Successf(format string, in ...interface{}) {
	var itfs string = " SUCCESS " + fmt.Sprintf(format, in...) + "\n"
	if l.IsAffiliated() {
		l.mainLogger.outputLogger(l.prefix, color.Green, 2, itfs)
	} else {
		l.output(color.Green, 2, itfs)
	}
}

func (l *logger) Success(in ...interface{}) {
	var itfs string = " SUCCESS " + fmt.Sprint(in...) + "\n"
	if l.IsAffiliated() {
		l.mainLogger.outputLogger(l.prefix, color.Green, 2, itfs)
	} else {
		l.output(color.Green, 2, itfs)
	}
}

func (l *logger) Failuref(format string, in ...interface{}) {
	var itfs string = " FAILURE " + fmt.Sprintf(format, in...) + "\n"
	if l.IsAffiliated() {
		l.mainLogger.outputLogger(l.prefix, color.Red, 2, itfs)
	} else {
		l.output(color.Red, 2, itfs)
	}
}

func (l *logger) Failure(in ...interface{}) {
	var itfs string = " FAILURE " + fmt.Sprint(in...) + "\n"
	if l.IsAffiliated() {
		l.mainLogger.outputLogger(l.prefix, color.Red, 2, itfs)
	} else {
		l.output(color.Red, 2, itfs)
	}
}

func (l *logger) write(buff []byte) {
	l.out.Write(buff)
}

func (l *logger) log(level LogLevelValue, in ...interface{}) {
	l.logEvent(l.prefix, level, in...)
}
func (l *logger) logEvent(appName string, level LogLevelValue, in ...interface{}) {
	if level >= l.verbosity {
		var itfs string = " " + string(toVerbosityLevel(level)) + " " + fmt.Sprint(in...) + "\n"
		switch string(toVerbosityLevel(level)) {
		case "DEBUG":
			l.outputLogger(appName, color.Yellow, 2, itfs)
			break
		case "TRACE":
			l.outputLogger(appName, color.Yellow, 2, itfs)
			break
		case "WARN":
			l.outputLogger(appName, color.LightYellow, 2, itfs)
			break
		case "INFO":
			l.outputLogger(appName, color.LightWhite, 2, itfs)
			break
		case "ERROR":
			l.outputLogger(appName, color.LightRed, 2, itfs)
			break
		case "FATAL":
			l.outputLogger(appName, color.Red, 2, itfs)
			break
		default:
			l.outputLogger(appName, color.White, 2, itfs)
		}
	}
}

const (
	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
	Ltime                         // the time in the local time zone: 01:23:23
	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                     // full file name and line number: /a/b/c/d.go:23
	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
	LstdFlags     = Ldate | Ltime // initial values for the standard logger
)

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (l *logger) formatHeader(prefix string, buf *[]byte, t time.Time, file string, line int) {
	*buf = append(*buf, prefix...)
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (l *logger) AddEchoWriter(key string, writer io.Writer) {
	defer func() {
		_ = recover()
		l.Unlock()
	}()
	l.Lock()
	if writer != nil {
		l.echoWriters[key] = writer
	}
}
func (l *logger) RemoveEchoWriter(echoKey string) {
	defer func() {
		_ = recover()
		l.Unlock()
	}()
	l.Lock()
	echoWriters := make(map[string]io.Writer)
	if _, ok := l.echoWriters[echoKey]; ok {
		for key, value := range l.echoWriters {
			if key != echoKey {
				echoWriters[key] = value
			}
		}
		l.echoWriters = echoWriters
	}
}

func (l *logger) reloadWriter() {
	var out io.Writer
	var state bool
	if l.logRotator != nil {
		out, state = l.logRotator.GetDefaultWriter()
	}
	if state && out != nil {
		l.out = out
	}
}

func (l *logger) echo() {
	for _, w := range l.echoWriters {
		if w != nil {
			_, _ = w.Write(l.buf)
		}
	}
}

func (l *logger) output(color color.Color, calldepth int, s string) error {
	return l.outputLogger(l.prefix, color, calldepth, s)
}
func (l *logger) outputLogger(prefix string, color color.Color, calldepth int, s string) error {
	defer func() {
		if r := recover(); r == nil {
			l.echo()
		}
		if l.logRotator != nil && l.logRotator.IsEnabled() {
			l.logRotator.Hook(int64(len(l.buf)))
		}
		l.Unlock()
	}()
	l.Lock()
	now := time.Now() // get this early.
	var file string
	var line int
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
	}
	l.buf = l.buf[:0]
	l.formatHeader(prefix, &l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	if l.onScreen {
		color.Printf(string(l.buf))
		return nil
	} else {
		_, err := l.out.Write(l.buf)
		return err
	}
}

func NewLogger(appName string, verbosity LogLevel) Logger {
	return &logger{
		verbosity:   toVerbosityLevelValue(verbosity),
		onScreen:    true,
		out:         os.Stdout,
		prefix:      "[" + appName + "] ",
		flag:        LstdFlags | LUTC,
		mainLogger:  nil,
		buf:         []byte{},
		logRotator:  nil, // Just to highlight the rotator state
		echoWriters: make(map[string]io.Writer),
	}
}

func NewFileLogger(appName string, logRotator LogRotator, verbosity LogLevel) (Logger, error) {
	var out io.Writer
	if logRotator != nil {
		out, _ = logRotator.GetDefaultWriter()
	}
	if out == nil {
		return nil, errors.New("log.Logger: Unable to receive rotator write stream")
	}
	logger := &logger{
		verbosity:   toVerbosityLevelValue(verbosity),
		onScreen:    false,
		out:         out,
		prefix:      "[" + appName + "] ",
		flag:        LstdFlags | LUTC,
		mainLogger:  nil,
		buf:         []byte{},
		logRotator:  logRotator,
		echoWriters: make(map[string]io.Writer),
	}
	logger.logRotator.UpdateCallBack(logger.reloadWriter)
	return logger, nil
}

func VerbosityLevelFromString(verbosity string) LogLevel {
	switch strings.ToUpper(verbosity) {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	}
	return INFO
}

func toVerbosityLevelValue(verbosity LogLevel) LogLevelValue {
	switch strings.ToUpper(string(verbosity)) {
	case "TRACE":
		return traceLevel
	case "DEBUG":
		return debugLevel
	case "INFO":
		return infoLevel
	case "WARN":
		return warningLevel
	case "ERROR":
		return errorLevel
	case "FATAL":
		return fatalLevel
	}
	return infoLevel
}

func toVerbosityLevel(verbosity LogLevelValue) LogLevel {
	switch verbosity {
	case traceLevel:
		return TRACE
	case debugLevel:
		return DEBUG
	case infoLevel:
		return INFO
	case warningLevel:
		return WARN
	case errorLevel:
		return ERROR
	case fatalLevel:
		return FATAL
	}
	return INFO
}

func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
