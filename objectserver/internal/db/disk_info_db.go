package db

import clientv3 "go.etcd.io/etcd/client/v3"

//TODO(feat): get or set disk info to etcd

type DiskInfo struct {
	cli *clientv3.KV
}

func NewDiskInfo(c *clientv3.KV) *DiskInfo {
	return &DiskInfo{c}
}
