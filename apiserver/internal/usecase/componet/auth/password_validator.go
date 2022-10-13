package auth

import (
	"apiserver/internal/usecase/componet/auth/credential"
	"common/constrant"
	"common/logs"
	"common/response"
	"common/util"
	"context"
	"errors"
	"net/http"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type PasswordValidator struct {
	cli clientv3.KV
}

func NewPasswordValidator(cli clientv3.KV, cfg *Config) *PasswordValidator {
	pv :=  &PasswordValidator{cli}
	pv.init(cfg)
	return pv
}

func (pv *PasswordValidator) init(cfg *Config) {
	resp, err := pv.cli.Get(context.Background(), constrant.EtcdPrefix.ApiCredentail)
	if err != nil {
		panic(err)
	}
	if len(resp.Kvs) != 0 || cfg.Username == "" || cfg.Password == "" {
		logs.Std().Info("exist credentail, skip init admin credentail")
		return
	}
	bt, err := util.EncodeMsgp(&credential.AdminCredentail{
		Username: cfg.Username, 
		Password: cfg.Password,
	})
	if err != nil {
		panic(err)
	}
	_, err = pv.cli.Put(context.Background(), constrant.EtcdPrefix.ApiCredentail, string(bt))
	if err != nil {
		panic(err)
	}
}

func (pv *PasswordValidator) Verify(token Credential) error {
	resp, err := pv.cli.Get(context.Background(), constrant.EtcdPrefix.ApiCredentail)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("no api-credentail provided from etcd")
	}
	var admin credential.AdminCredentail
	if err := util.DecodeMsgp(&admin, resp.Kvs[0].Value); err != nil {
		return err	
	}
	if token.GetUsername() != admin.Username || token.GetPassword() != admin.Password {
		return response.NewError(http.StatusForbidden, "username or password wrong")
	}
	return nil
}