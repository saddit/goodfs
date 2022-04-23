package tool

import (
	"fmt"
	"goodfs/apiserver/global"
	"net/http"
	"strconv"
)

func RepairExistFilter(ip string) {
	resp, err := global.Http.Get(fmt.Sprintf("http://%v/help/exist_filter", ip))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic("Error Response Status:" + resp.Status)
	}
	cnt, err := strconv.Atoi(resp.Header.Get("Count"))
	if err != nil {
		panic(err)
	}
	if err = global.ExistFilter.DecodeBuckets(resp.Body, uint(cnt)); err != nil {
		panic(err)
	}
}
