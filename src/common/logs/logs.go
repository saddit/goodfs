package logs

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"strings"
)

type Level string

const (
	Trace Level = "TRACE"
	Debug Level = "DEBUG"
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
)

func init() {
	// init log
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:        true,
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			return fmt.Sprint(" at ", frame.File, ":", frame.Line, "\r\n")
		},
		FieldsOrder: []string{"component"},
	})
}

func WithConfig(cfg *Config) {
	SetLevel(cfg.Level)
	EnableNotify(&ErrorNotifyHook{EmailConfig: cfg.Email})
	WithCaller(cfg.Caller)
}

func WithCaller(t bool) {
	logrus.SetReportCaller(t)
}

func ToLogLevel(l Level) logrus.Level {
	s := strings.ToUpper(string(l))
	switch Level(s) {
	case Trace:
		return logrus.TraceLevel
	case Debug:
		return logrus.DebugLevel
	default:
		fallthrough
	case Info:
		return logrus.InfoLevel
	case Warn:
		return logrus.WarnLevel
	case Error:
		return logrus.ErrorLevel
	}
}

func SetLevel(l Level) {
	logrus.SetLevel(ToLogLevel(l))
}

func EnableNotify(hook *ErrorNotifyHook) {
	if !hook.EmailConfig.IsValid() {
		println("enable error notify fail, required value is empty")
		return
	}
	logrus.AddHook(hook)
}

func SetOutput(w io.Writer) {
	logrus.SetOutput(w)
	ft, ok := logrus.StandardLogger().Formatter.(*nested.Formatter)
	if !ok {
		return
	}
	switch w {
	case os.Stderr, os.Stdout:
		ft.NoColors = false
	default:
		ft.NoColors = true
	}
}

func Std() *logrus.Logger {
	return logrus.StandardLogger()
}

func New(name string) *logrus.Entry {
	return Std().WithField("component", name)
}

func IsDebug() bool {
	return Std().Level == logrus.DebugLevel
}

func IsTrace() bool {
	return Std().Level == logrus.TraceLevel
}
