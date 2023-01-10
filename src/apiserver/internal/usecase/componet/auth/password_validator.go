package auth

import (
	"apiserver/internal/usecase/componet/auth/credential"
	"common/cst"
	"common/logs"
	"common/response"
	"common/util"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type PasswordValidator struct {
	cli clientv3.KV
	cfg *PasswordConfig
}

func NewPasswordValidator(cli clientv3.KV, cfg *PasswordConfig) *PasswordValidator {
	pv := &PasswordValidator{cli, cfg}
	pv.init(cfg)
	return pv
}

func (pv *PasswordValidator) init(cfg *PasswordConfig) {
	if !cfg.Enable {
		return
	}
	resp, err := pv.cli.Get(context.Background(), cst.EtcdPrefix.ApiCredential)
	if err != nil {
		panic(err)
	}
	if len(resp.Kvs) != 0 || cfg.Username == "" || cfg.Password == "" {
		logs.Std().Info("exist credential, skip init admin credential")
		return
	}
	bt, err := util.EncodeMsgp(&credential.AdminCredential{
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		panic(err)
	}
	_, err = pv.cli.Put(context.Background(), cst.EtcdPrefix.ApiCredential, string(bt))
	if err != nil {
		panic(err)
	}
	logs.Std().Infof("password authenticator init success: username=%s", cfg.Username)
}

func (pv *PasswordValidator) Verify(token Credential) error {
	resp, err := pv.cli.Get(context.Background(), cst.EtcdPrefix.ApiCredential)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("no api-credential provided from etcd")
	}
	var admin credential.AdminCredential
	if err := util.DecodeMsgp(&admin, resp.Kvs[0].Value); err != nil {
		return err
	}
	if token.GetUsername() != admin.Username || token.GetPassword() != admin.Password {
		logs.Std().Tracef("%s:%s not match %s:%s", token.GetUsername(), token.GetPassword(), admin.Username, admin.Password)
		return response.NewError(http.StatusUnauthorized, "username or password wrong")
	}
	return nil
}

func (pv *PasswordValidator) Middleware(c *gin.Context) (bool, error) {
	if usr, pwd, ok := c.Request.BasicAuth(); ok && pv.cfg.Enable {
		err := pv.Verify(credential.NewPasswordToken(usr, pwd))
		return err == nil, err
	}
	return false, nil
}
