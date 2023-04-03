package logs

import (
	"fmt"
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
	content := fmt.Sprintf("%s [%s] %s\r\n", entry.Time, entry.Level, entry.Message)
	if entry.HasCaller() {
		content += fmt.Sprint(" at ", entry.Caller.File, ":", entry.Caller.Line)
	}
	if eh.SmtpPort == "" {
		eh.SmtpPort = "587"
	}
	addr := net.JoinHostPort(eh.SmtpHost, eh.SmtpPort)
	auth := smtp.PlainAuth("", eh.Sender, eh.Password, eh.SmtpHost)
	msg := []byte("To: " + eh.Target + "\r\n" + "Subject: [GoodFS] ERROR HAPPENED!\r\n\r\n")
	msg = append(msg, []byte(content)...)
	return smtp.SendMail(addr, auth, eh.Sender, []string{eh.Target}, msg)
}
