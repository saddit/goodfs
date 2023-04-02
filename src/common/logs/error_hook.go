package logs

import (
	"github.com/sirupsen/logrus"
	"net"
	"net/smtp"
)

type ErrorNotifyHook struct {
	EmailConfig
}

func (eh *ErrorNotifyHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel}
}

func (eh *ErrorNotifyHook) Fire(entry *logrus.Entry) error {
	bt, err := entry.Bytes()
	if err != nil {
		return err
	}
	if eh.SmtpPort == "" {
		eh.SmtpPort = "583"
	}
	addr := net.JoinHostPort(eh.SmtpHost, eh.SmtpPort)
	auth := smtp.PlainAuth("", eh.SendEmail, eh.Password, eh.SmtpHost)
	return smtp.SendMail(addr, auth, eh.SendEmail, eh.TargetEmails, bt)
}
