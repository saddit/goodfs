package logs

type Config struct {
	Email    EmailConfig `yaml:"email" env-prefix:"EMAIL"`
	Level    Level       `yaml:"level" env:"LEVEL" env-default:"info"`
	Caller   bool        `yaml:"caller" env:"CALLER"`
	StoreDir string      `yaml:"store-dir" env:"STORE_DIR"`
}

type EmailConfig struct {
	Target   string `yaml:"target" env:"TARGET"`
	Sender   string `yaml:"sender" env:"SENDER"`
	Password string `yaml:"password" env:"PASSWORD"`
	SmtpHost string `yaml:"smtp-host" env:"SMTP_HOST"`
	SmtpPort string `yaml:"smtp-port" env:"SMTP_PORT" env-default:"587"` // SmtpPort default is '587'
}

func (ec *EmailConfig) IsValid() bool {
	return ec.Sender != "" && ec.Target != "" && ec.SmtpHost != "" && ec.Password != ""
}
