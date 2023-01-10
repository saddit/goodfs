package credential

//go:generate msgp -tests=false

type AdminCredential struct {
	Username string `msg:"username"`
	Password string `msg:"password"`
}
