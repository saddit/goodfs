package logs

type Config struct {
	EmailConfig
	Level    Level  `yaml:"level" env:"LEVEL" env-default:"info"`
	StoreDir string `yaml:"store-dir" env:"STORE_DIR"`
}

type EmailConfig struct {
	TargetEmails []string `yaml:"target-emails" env:"TARGET_EMAILS" env-separator:","`
	SendEmail    string   `yaml:"send-email" env:"SEND_EMAIL"`
	Password     string   `yaml:"password" env:"PASSWORD"`
	SmtpHost     string   `yaml:"smtp-host" env:"SMTP_HOST"`
	SmtpPort     string   `yaml:"smtp-port" env:"SMTP_PORT" env-default:"583"` // SmtpPort default is '583'
}
