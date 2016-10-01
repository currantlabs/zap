package zap

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/mgutz/ansi"
)

var (
	ansiPool = sync.Pool{
		New: func() interface{} {
			return &ansiEncoder{
				textEncoder: textEncoder{
					bytes: make([]byte, 0, _initialBufSize),
				},
			}
		},
	}

	resetColor = ansi.ColorCode("reset")
	// Default colors for levels.
	defaultDebugColor = ansi.ColorCode("green")
	defaultInfoColor  = ansi.ColorCode("green")
	defaultWarnColor  = ansi.ColorCode("yellow")
	defaultErrorColor = ansi.ColorCode("red")
	defaultPanicColor = ansi.ColorCode("red")
	defaultFatalColor = ansi.ColorCode("red")
)

type ansiEncoder struct {
	textEncoder

	debugColor string
	infoColor  string
	warnColor  string
	errorColor string
	panicColor string
	fatalColor string
}

// A ANSIOption is used to set options for a ANSI encoder.
type ANSIOption interface {
	apply(*ansiEncoder)
}

type ansiOptionFunc func(*ansiEncoder)

func (opt ansiOptionFunc) apply(enc *ansiEncoder) {
	opt(enc)
}

// NewANSIEncoder creates a line-oriented text encoder whose output is optimized
// for human, rather than machine, consumption. Log levels are color-coded using
// ANSI escape codes. By default, the encoder uses RFC3339-formatted timestamps.
func NewANSIEncoder(options ...ANSIOption) Encoder {

	enc := ansiPool.Get().(*ansiEncoder)
	enc.truncate()
	enc.timeFmt = time.RFC3339
	enc.debugColor = defaultDebugColor
	enc.infoColor = defaultInfoColor
	enc.warnColor = defaultWarnColor
	enc.errorColor = defaultErrorColor
	enc.panicColor = defaultPanicColor
	enc.fatalColor = defaultFatalColor
	for _, opt := range options {
		opt.apply(enc)
	}
	return enc
}

func (enc *ansiEncoder) Clone() Encoder {
	clone := ansiPool.Get().(*ansiEncoder)
	clone.truncate()
	clone.bytes = append(clone.bytes, enc.bytes...)
	clone.timeFmt = enc.timeFmt
	clone.debugColor = enc.debugColor
	clone.infoColor = enc.infoColor
	clone.warnColor = enc.warnColor
	clone.errorColor = enc.errorColor
	clone.panicColor = enc.panicColor
	clone.fatalColor = enc.fatalColor
	return clone
}

func (enc *ansiEncoder) Free() {
	ansiPool.Put(enc)
}

func (enc *ansiEncoder) WriteEntry(sink io.Writer, name string, msg string, lvl Level, t time.Time) error {
	if sink == nil {
		return errNilSink
	}

	final := textPool.Get().(*textEncoder)
	final.truncate()

	enc.addLevelColor(final, lvl)
	enc.textEncoder.addLevel(final, lvl)
	enc.textEncoder.addTime(final, t)
	enc.textEncoder.addName(final, name)
	enc.textEncoder.addMessage(final, msg)

	if len(enc.textEncoder.bytes) > 0 {
		final.bytes = append(final.bytes, ' ')
		final.bytes = append(final.bytes, enc.textEncoder.bytes...)
	}
	enc.clearLevelColor(final, lvl)
	final.bytes = append(final.bytes, '\n')

	expectedBytes := len(final.bytes)
	n, err := sink.Write(final.bytes)
	final.Free()
	if err != nil {
		return err
	}
	if n != expectedBytes {
		return fmt.Errorf("incomplete write: only wrote %v of %v bytes", n, expectedBytes)
	}
	return nil
}

func (enc *ansiEncoder) addLevelColor(final *textEncoder, lvl Level) {
	switch lvl {
	case DebugLevel:
		final.bytes = append(final.bytes, enc.debugColor...)
	case InfoLevel:
		final.bytes = append(final.bytes, enc.infoColor...)
	case WarnLevel:
		final.bytes = append(final.bytes, enc.warnColor...)
	case ErrorLevel:
		final.bytes = append(final.bytes, enc.errorColor...)
	case PanicLevel:
		final.bytes = append(final.bytes, enc.panicColor...)
	case FatalLevel:
		final.bytes = append(final.bytes, enc.fatalColor...)
	default:
	}
}

func (enc *ansiEncoder) clearLevelColor(final *textEncoder, lvl Level) {
	switch lvl {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel, PanicLevel, FatalLevel:
		final.bytes = append(final.bytes, resetColor...)
	default:
	}
}

// AnsiTextOption allows passing TextOptions to an ANSI Encoder
func AnsiTextOption(to TextOption) ANSIOption {
	return ansiOptionFunc(func(enc *ansiEncoder) {
		to.apply(&enc.textEncoder)
	})
}
