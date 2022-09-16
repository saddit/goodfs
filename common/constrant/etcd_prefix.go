package constrant

import "fmt"

type etcdPrefix struct {
	PeersInfo string
	HashSlot  string
	Registry  string
}

var EtcdPrefix = etcdPrefix{
	PeersInfo: "peers_info",
	HashSlot:  "hashslot",
	Registry:  "registry",
}

func (e *etcdPrefix) FmtPeersInfo(groupId, id string) string {
	return fmt.Sprintf("%s/%s/%s", e.PeersInfo, groupId, id)
}

func (e *etcdPrefix) FmtHashSlot(groupName, serviceName, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s", e.HashSlot, groupName, serviceName, id)
}

func (e *etcdPrefix) FmtRegistry(groupName, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s", e.Registry, groupName, serviceName)
}
