package global

import (
	"goodfs/apiserver/config"
	"goodfs/apiserver/service/selector"
	"net/http"

	"github.com/838239178/goodmq"
)

var (
	Config         config.Config
	AmqpConnection *goodmq.AmqpConnection
	Http           *http.Client
	Balancer       selector.Selector
	//LocalDB        *leveldb.DB
)
