package pool

import (
	"apiserver/config"
	"apiserver/internal/usecase/selector"
	"net/http"
	"time"
)

var (
	Config   *config.Config
	Http     *http.Client
	Balancer selector.Selector
)

func InitPool(cfg *config.Config) {
	Config = cfg

	Http = &http.Client{Timeout: 5 * time.Second}

	Balancer = selector.NewSelector(cfg.SelectStrategy)
}

func Close() {
	Http.CloseIdleConnections()
}
