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
	go func() {
		defer func() { _ = recover() }()
		errMsg := fmt.Sprintf("%s [%s] %s", entry.Time.Format("2006-01-02 15:04:05"), entry.Level, entry.Message)
		if entry.HasCaller() {
			errMsg += fmt.Sprint(" at ", entry.Caller.File, ":", entry.Caller.Line)
		}
		content := fmt.Sprintf(
			`
			<html><body style="padding: 2rem">
			  <h1>An Error Occurs</h1>
			  <div style="color: red; background-color: rgba(255, 0, 0, 0.10); border-radius: 0.5rem; padding: 1rem;">%s</div>
			</body></html>
			`,
			errMsg,
		)
		if eh.SmtpPort == "" {
			eh.SmtpPort = "587"
		}
		addr := net.JoinHostPort(eh.SmtpHost, eh.SmtpPort)
		auth := smtp.PlainAuth("", eh.Sender, eh.Password, eh.SmtpHost)
		msg := []byte("From: " + eh.Sender + "\r\n" +
			"To: " + eh.Target + "\r\n" +
			"Subject: [GoodFS] ERROR OCCURS!\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
			"\r\n")
		msg = append(msg, []byte(content)...)
		_ = smtp.SendMail(addr, auth, eh.Sender, []string{eh.Target}, msg)
	}()
	return nil
}
