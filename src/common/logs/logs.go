package logs

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
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
		FieldsOrder:     []string{"component"},
	})
}

func SetLevel(l Level) {
	s := strings.ToUpper(string(l))
	switch Level(s) {
	case Trace:
		logrus.SetLevel(logrus.TraceLevel)
	case Debug:
		logrus.SetLevel(logrus.DebugLevel)
	default:
		fallthrough
	case Info:
		logrus.SetLevel(logrus.InfoLevel)
	case Warn:
		logrus.SetLevel(logrus.WarnLevel)
	case Error:
		logrus.SetLevel(logrus.ErrorLevel)
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
