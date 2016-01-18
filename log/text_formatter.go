package log

import (
	"bytes"
	"sort"
	"time"
)

type TextFormatter struct {
	ForceColors     bool
	DisableColors   bool
	TimestampFormat string
	DisableSorting  bool
}

func DefaultTextFormatter() Formatter {
	return &TextFormatter{
		true,
		false,
		time.StampNano,
		false,
	}
}

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

//func LogFmt(s State) string {
//st := s.RStatus
//md := s.RMethod
//return fmt.Sprintf("%v |%s %3d %s| %12v | %s |%s %s %-7s %s",
//s.RStop.Format("2006/01/02 - 15:04:05"),
//StatusColor(st), st, reset,
//s.RLatency,
//s.RRequester,
//MethodColor(md), reset, md,
//s.RPath,
//)
//}

func (lv Level) Color() string {
	switch lv {
	case LPanic:
		return red
	case LFatal:
		return magenta
	case LError:
		return cyan
	case LWarn:
		return yellow
	case LInfo:
		return green
	case LDebug:
		return blue
	}
	return white
}

//func StatusColor(code int) (color string) {
//	switch {
//	case code >= 200 && code <= 299:
//		color = green
//	case code >= 300 && code <= 399:
//		color = white
//	case code >= 400 && code <= 499:
//		color = yellow
//	default:
//		color = red
//	}
//	return color
//}

//func MethodColor(method string) (color string) {
//	switch {
//	case method == "GET":
//		color = blue
//	case method == "POST":
//		color = cyan
//	case method == "PUT":
//		color = yellow
//	case method == "DELETE":
//		color = red
//	case method == "PATCH":
//		color = green
//	case method == "HEAD":
//		color = magenta
//	case method == "OPTIONS":
//		color = white
//	}
//	return color
//}

func (t *TextFormatter) Format(e Entry) ([]byte, error) {
	fs := e.Fields()
	var keys []string = make([]string, 0, len(fs))
	for _, k := range fs {
		keys = append(keys, k.Key)
	}

	if !t.DisableSorting {
		sort.Strings(keys)
	}

	b := &bytes.Buffer{}
	return b.Bytes(), nil
}
