package webapi

import (
	"bytes"
	"common/response"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 5 * time.Minute,
	}
	cli = &http.Client{Transport: &http2.Transport{
		// So http2.Transport doesn't complain the URL scheme isn't 'https'
		AllowHTTP: true,
		// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
	}}
)

func UpdateVersionLocates(ip, versionHash string, shardIndex int, locate string) error {
	uri := fmt.Sprintf("http://%s/version/locate", ip)
	bt, err := json.Marshal(map[string]interface{}{
		"hash":        versionHash,
		"locateIndex": shardIndex,
		"locate":      locate,
	})
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(bt))
	req.Header.Set("Content-Type", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return response.NewError(resp.StatusCode, response.MessageFromJSONBody(resp.Body))
	}
	return nil
}
