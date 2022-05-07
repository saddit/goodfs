package global

import (
	"github.com/syndtr/goleveldb/leveldb"
	"goodfs/apiserver/config"
	"goodfs/apiserver/service/selector"
	"net/http"

	"github.com/838239178/goodmq"
)

var (
	Config         config.Config
	AmqpConnection *goodmq.AmqpConnection
	Http           *http.Client
	LocalDB        *leveldb.DB
	Balancer       selector.Selector
)
