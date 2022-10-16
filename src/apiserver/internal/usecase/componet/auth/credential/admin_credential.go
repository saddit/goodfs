package credential

//go:generate msgp -tests=false

type AdminCredentail struct {
	Username string `msg:"username"`
	Password string `msg:"password"`
}
