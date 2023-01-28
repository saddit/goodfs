package webapi

import "adminserver/internal/usecase/pool"

func GetSchema() string {
	if pool.Config.EnabledApiTLS {
		return "https"
	}
	return "http"
}
