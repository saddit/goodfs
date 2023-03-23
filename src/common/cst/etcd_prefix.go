package cst

import "fmt"

type etcdPrefix struct {
	Sep            []byte
	HashSlot       string
	Registry       string
	ObjectCap      string
	ApiCredential  string
	SystemInfo     string
	Configure      string
	LocationSubKey string
}

var EtcdPrefix = etcdPrefix{
	Sep:            []byte("/"),
	HashSlot:       "hash_slot",
	Registry:       "registry",
	ObjectCap:      "object_cap",
	ApiCredential:  "api_credential",
	SystemInfo:     "sys_info",
	Configure:      "configure",
	LocationSubKey: "good.fs.location",
}

func (e *etcdPrefix) FmtRegistry(groupName, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s", groupName, e.Registry, serviceName)
}

func (e *etcdPrefix) FmtHashSlot(groupName, id string) string {
	return fmt.Sprintf("%s/%s/%s", groupName, e.HashSlot, id)
}

func (e *etcdPrefix) FmtSystemInfo(groupName, serviceName, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s", groupName, e.SystemInfo, serviceName, id)
}

func (e *etcdPrefix) FmtConfigure(groupName, id string) string {
	return fmt.Sprintf("%s/%s/%s", groupName, e.Configure, id)
}
