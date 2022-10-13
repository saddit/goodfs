package credential

type PasswordToken struct {
	emptyExtra
	Username string
	Password string
}

func NewPasswordToken(username, pwd string) *PasswordToken {
	return &PasswordToken{Username: username, Password: pwd}
}

func (pt *PasswordToken) GetUsername() string {
	return pt.Username
}

func (pt *PasswordToken) GetPassword() string {
	return pt.Password
}
