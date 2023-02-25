package webapi

import (
	"common/performance"
	"common/util"
	"net"
	"net/http"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 5 * time.Minute,
	}
	httpClient = http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          2500,
		IdleConnTimeout:       10 * time.Minute,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}}
	performCollector performance.Collector
)

func SetPerformanceCollector(c performance.Collector) {
	performCollector = c
}

func perform(written bool) func() {
	if performCollector == nil {
		return func() {}
	}
	t := time.Now()
	return func() {
		performCollector.PutAsync(
			util.IfElse(written, performance.ActionWrite, performance.ActionRead),
			performance.KindOfHTTP,
			time.Since(t),
		)
	}
}

func keepalive(req *http.Request) {
	req.Header.Set("Keep-Alive", "timeout=300, max=2500")
}

func Close() {
	httpClient.CloseIdleConnections()
}
