package cst

import "fmt"

type etcdPrefix struct {
	Sep           []byte
	PeersInfo     string
	HashSlot      string
	Registry      string
	ObjectCap     string
	ApiCredential string
	SystemInfo    string
}

var EtcdPrefix = etcdPrefix{
	Sep:           []byte("/"),
	PeersInfo:     "peers_info",
	HashSlot:      "hash_slot",
	Registry:      "registry",
	ObjectCap:     "object_cap",
	ApiCredential: "api_credential",
	SystemInfo:    "sys_info",
}

func (e *etcdPrefix) FmtPeersInfo(groupId, id string) string {
	return fmt.Sprintf("%s/%s/%s", e.PeersInfo, groupId, id)
}

func (e *etcdPrefix) FmtRegistry(groupName, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s", e.Registry, groupName, serviceName)
}

func (e *etcdPrefix) FmtHashSlot(groupName, serviceName, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s", e.HashSlot, groupName, serviceName, id)
}

func (e *etcdPrefix) FmtObjectCap(groupName, serviceName, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s", e.ObjectCap, groupName, serviceName, name)
}

func (e *etcdPrefix) FmtSystemInfo(groupName, serviceName, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s", e.SystemInfo, groupName, serviceName, id)
}
