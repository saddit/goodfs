package logs

import (

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func init() {
	// init log
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		HideKeys:    true,
		FieldsOrder: []string{"component"},
	})
}

func Std() *logrus.Logger {
	return logrus.StandardLogger()
}

func New(name string) *logrus.Entry {
	return Std().WithField("component", name)
}
