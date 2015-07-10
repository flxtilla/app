// Package xrr contains common error handling functionality for flotilla.
package xrr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
)

const (
	ErrorTypeFlotilla = 1 << iota
	ErrorTypeInternal = 1 << iota
	ErrorTypeExternal = 1 << iota
	ErrorTypePanic    = 1 << iota
	ErrorTypeAll      = 0xffffffff
)

var (
	unknown   = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

type Erroror interface {
	Error(err error, typ uint32, meta interface{})
	Frror(format string, typ uint32, meta interface{}, parameters ...interface{})
	Errors() ErrorMsgs
}

type erroror struct {
	errors ErrorMsgs
}

func DefaultErroror() Erroror {
	return &erroror{}
}

func (e *erroror) Error(err error, typ uint32, meta interface{}) {
	e.errors = append(e.errors, ErrorMsg{
		Err:  err.Error(),
		Type: typ,
		Meta: meta,
	})
}

func (e *erroror) Frror(err string, typ uint32, meta interface{}, parameters ...interface{}) {
	e.errors = append(e.errors, ErrorMsg{
		Err:        err,
		Type:       typ,
		Meta:       meta,
		parameters: parameters,
	})
}

func (e *erroror) Errors() ErrorMsgs {
	return e.errors
}

type ErrorMsg struct {
	Err        string      `json:"error"`
	Type       uint32      `json:"-"`
	Meta       interface{} `json:"meta"`
	parameters []interface{}
}

func (e ErrorMsg) Error() string {
	return fmt.Sprintf(e.Err, e.parameters...)
}

func NewError(err string, params ...interface{}) ErrorMsg {
	return ErrorMsg{Err: err, parameters: params, Type: ErrorTypeFlotilla}
}

type ErrorMsgs []ErrorMsg

func (a ErrorMsgs) ByType(typ uint32) ErrorMsgs {
	if len(a) == 0 {
		return a
	}
	result := make(ErrorMsgs, 0, len(a))
	for _, msg := range a {
		if msg.Type&typ > 0 {
			result = append(result, msg)
		}
	}
	return result
}

func (a ErrorMsgs) String() string {
	if len(a) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for i, msg := range a {
		text := fmt.Sprintf("Error #%02d: %s\nMeta: %v\n", (i + 1), msg.Error(), msg.Meta)
		buffer.WriteString(text)
	}
	return buffer.String()
}

// stack returns a nicely formated stack frame, skipping skip frames
func Stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", Function(pc), Source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func Source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return unknown
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func Function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return unknown
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
