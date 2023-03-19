package webapi

import (
	"common/performance"
	"common/util"
	"context"
	"crypto/tls"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 10 * time.Minute,
	}
	httpClient = &http.Client{Transport: &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
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
	req.Header.Set("Keep-Alive", "timeout=300, max=6000")
}

func Close() {
	httpClient.CloseIdleConnections()
}
