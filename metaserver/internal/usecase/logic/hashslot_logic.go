package logic

import (
	"common/hashslot"
	"common/logs"
	"metaserver/internal/usecase/pool"
)

type HashSlot struct{}

func NewHashSlot() HashSlot { return HashSlot{} }

func (h HashSlot) IsKeyOnThisServer(key string) (bool, string) {
	id, err := pool.HashSlot.GetKeyIdentify(key)
	if err != nil {
		logs.Std().Error(err)
		return false, ""
	}
	return id == pool.HttpHostPort, id
}

func (HashSlot) GetSlotsProvider() (hashslot.IEdgeProvider, error) {
	return pool.HashSlot.GetEdgeProvider(false)
}

func (h HashSlot) SaveToEtcd(id string, info *hashslot.SlotInfo) error {
	return pool.HashSlot.Save(id, info)
}

func (HashSlot) RemoveFromEtcd(id string) error {
	return pool.HashSlot.Remove(id)
}

// AutoMigrate
// TODO 如何确保过时的数据全部更新完毕: 由迁移双方自行控制，迁移成功后更新双方slot信息
// TODO 何时迁移: 指令触发时，指定 A to B with 10-100
// TODO 如何迁移：将K不在该服务器上的key通过RPC流服务迁移出去，也就是说我需要编写HashSlotRpcServer
// TODO 何时失败：10-100不完全属于A、A或B繁忙、迁移过程中发生异常中断
// TODO 合适成功：key-value全部迁移过去。etcd中的slots信息为迁移完成后的slot信息
func (h HashSlot) AutoMigrate(cur []string) {

}

// SaveByConfig
// TODO 只有第一次才从配置文件保存slots信息
// TODO 但每次都需要更新Peers信息
// TODO Peers的配置文件需要相同的 hash-slot.id
func (h HashSlot) SaveByConfig(peers []string) error {
	var info *hashslot.SlotInfo
	if info, exist, err := pool.HashSlot.Get(pool.Config.HashSlot.ID); !exist {
		if err != nil {
			return err
		}
		info.Location = pool.HttpHostPort
		info.Slots = pool.Config.HashSlot.Slots
	}
	info.Peers = peers
	return h.SaveToEtcd(pool.Config.HashSlot.ID, info)
}