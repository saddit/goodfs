package webapi

import "adminserver/internal/usecase/pool"

func GetSchema() string {
	if pool.Config.TLS {
		return "https"
	}
	return "http"
}
