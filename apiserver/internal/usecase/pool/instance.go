package pool

import (
	"apiserver/config"
	"apiserver/internal/usecase/selector"
	"net/http"
	"time"

	"github.com/838239178/goodmq"
)

var (
	Config       *config.Config
	Http         *http.Client
	Balancer     selector.Selector
	AmqpTemplate *goodmq.AmqpProvider
	Amqp         *goodmq.AmqpConnection
)

func InitPool(cfg *config.Config) {
	var e error

	Config = cfg

	Http = &http.Client{Timeout: 5 * time.Second}

	goodmq.RecoverDelay = 3 * time.Second
	Amqp = goodmq.NewAmqpConnection(cfg.AmqpAddress)
	if AmqpTemplate, e = Amqp.NewProvider(); e != nil {
		panic(e)
	}

	Balancer = selector.NewSelector(cfg.SelectStrategy)
}

func Close() {
	AmqpTemplate.Close()
	Amqp.Close()
	Http.CloseIdleConnections()
}
