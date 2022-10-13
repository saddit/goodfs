package constrant

import "fmt"

type etcdPrefix struct {
	PeersInfo string
	HashSlot  string
	Registry  string
	ObjectCap string
}

var EtcdPrefix = etcdPrefix{
	PeersInfo: "peers_info",
	HashSlot:  "hashslot",
	Registry:  "registry",
	ObjectCap: "object_cap",
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
