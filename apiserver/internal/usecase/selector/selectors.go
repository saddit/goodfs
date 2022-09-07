package selector

import (
	"log"
	"strings"
)

type SelectStrategy string

type Select interface {
	Select([]string) string
}

type Selector interface {
	Select
	Pop([]string) ([]string, string)
	Strategy() SelectStrategy
}

func NewSelector(str string) Selector {
	var sec Selector

	switch strings.ToLower(str) {
	case string(Random):
		sec = &RandomSelector{}
	case string(MaxFreeDisk):
		sec = &MaxFreeDiskSelector{}
	default:
		log.Panicf("Not allowed selector strategy: %v", str)
	}
	return sec
}

type IPSelector struct {
	Selector
	IPs  []string
	used []string
}

func NewIPSelector(selector Selector, ips []string) *IPSelector{
	return &IPSelector{Selector: selector, IPs: ips, used: make([]string, 0, len(ips))}
}

// Select 负载均衡选增IP，能够自动处理IP不足的情况
func (i *IPSelector) Select() string {
	if i.used == nil {
		i.used = make([]string, 0, len(i.IPs))
	}
	var ip string
	if len(i.IPs) > 0 {
		i.IPs, ip = i.Selector.Pop(i.IPs)
		i.used = append(i.used, ip)
	} else {
		i.used, ip = i.Selector.Pop(i.used)
		i.IPs = append(i.IPs, ip)
	}
	return ip
}