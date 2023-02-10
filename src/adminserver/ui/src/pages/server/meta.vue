<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div v-if="infos.length > 0"
       class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
  <div v-else class="w-full my-5 text-center text-gray-600 text-lg font-medium">
    {{ $t('no-servers') }}
  </div>
  <div class="mb-4 mt-8 flex flex-wrap space-x-4">
    <!-- capacity card -->
    <CapCard v-if="capInfo.total > 0" class="w-[32%]" :cap-info="capInfo"/>
    <div class="w-1/4" v-if="infos.length > 0">
      <!-- tool box -->
      <div class="bg-white h-24 mb-2 shadow-md rounded-md p-3 gap-y-1 gap-x-2 grid grid-rows-2 grid-cols-2">
        <span class="font-bold text-xl justify-self-start text-gray-500">{{ t('tool-box') }}</span>
        <button class="btn-pri text-sm" @click="openMigrateDialog = true">{{ t('start-migrate') }}</button>
        <button class="btn-pri text-sm" @click="openRaftCmdDialog = true">{{ t('raft-cmd') }}</button>
      </div>
      <SlotsCard class="h-28" :value="slotRanges"></SlotsCard>
    </div>
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <UsageLine class="mb-4" :type="$cst.statTypeCpu" :server-no="$cst.metaServerNo"/>
  <UsageLine :type="$cst.statTypeMem" :server-no="$cst.metaServerNo"/>
  <!-- Data migration dialog -->
  <ModalTemplate v-model="openMigrateDialog" :title="t('migration')" @close="closeMigrateDialog">
    <template #panel>
      <div class="mt-6 grid grid-cols-3 items-center gap-y-4 gap-x-1 rounded-md">
        <!-- row 1 source -->
        <span class="text-sm text-gray-700">{{ t('src-server') }}</span>
        <SelectBox class="col-span-2" v-model="migrateReq.srcServerId"
                   :value="(v: ServerInfo) => v.serverId"
                   :format="(v: ServerInfo) => `${v.serverId}(${getSlotsString(v.serverId)})`"
                   :options="masters"></SelectBox>
        <!-- row 2 target -->
        <span class="text-sm text-gray-700">{{ t('dest-server') }}</span>
        <SelectBox class="col-span-2" v-model="migrateReq.destServerId"
                   :value="(v: ServerInfo) => v.serverId"
                   :format="(v: ServerInfo) => `${v.serverId}(${getSlotsString(v.serverId)})`"
                   :options="masters"></SelectBox>
        <!-- row 3 slots -->
        <span class="text-sm text-gray-700">{{ t('which-slots') }}</span>
        <input v-model="migrateReq.slotsStr" id="slots" name="slots" required type="text" placeholder="0-1000,2000-2500"
               class="text-input col-span-2"/>
        <!-- row 4 button -->
        <span></span>
        <button class="btn-revert max-h-10 mx-1" @click="closeMigrateDialog">{{ $t('btn-cancel') }}</button>
        <button class="btn-pri max-h-10 mx-1" @click="startMigrate">{{ $t('btn-ok') }}</button>
        <!-- row 5 error message-->
        <template v-if="formErrMsg">
          <span></span>
          <span class="justify-self-start text-sm text-red-500 col-span-2">{{ formErrMsg }}</span>
        </template>
      </div>
    </template>
  </ModalTemplate>
  <!-- Raft cmd dialog -->
  <ModalTemplate v-model="openRaftCmdDialog" :title="t('raft-cmd')" @close="closeRaftCmdDialog">
    <template #panel>
      <div class="mt-6 w-[28rem] grid grid-cols-3 items-center gap-y-4 gap-x-1">
        <span>{{ t('select-master') }}</span>
        <SelectBox class="col-span-2" v-model="selectedRaftMaster"
                   :value="(v: ServerInfo) => v.serverId"
                   :format="(v: ServerInfo) => v.serverId"
                   :options="masters"></SelectBox>
        <template v-if="selectedRaftMaster">
          <div class="col-span-3 px-2 text-sm inline-flex items-center justify-self-end" v-for="slave in slaves">
            <span class="text-center mr-3">{{ slave.serverId + `(${slave.httpAddr})` }}</span>
            <button class="btn-revert py-1 px-2 text-sm" @click="leaveCluster(slave.serverId)">{{
                t('remove')
              }}
            </button>
          </div>
          <div class="col-span-3 px-2 inline-flex items-center justify-self-end">
            <input type="text" class="text-input-sm mr-2" :placeholder="t('input-serv-id')" v-model="invitedServId"/>
            <button class="btn-pri py-1 px-2 text-sm" @click="joinCluster(selectedRaftMaster, invitedServId)">
              {{ t('invite') }}
            </button>
          </div>
        </template>
        <!-- row button -->
        <span class="col-span-2"></span>
        <button class="btn-revert max-h-10 mx-1" @click="closeRaftCmdDialog">{{ $t('btn-close') }}</button>
      </div>
    </template>
  </ModalTemplate>
</template>

<script setup lang="ts">

let slots: { [key: string]: SlotsInfo } = {}
const slotRanges = ref<SlotRange[]>([])
const infos = ref<ServerInfo[]>([])
const capInfo = ref<DiskInfo>({used: 0, total: 0, free: 0})
const migrateReq = ref<MetaMigrateReq>({srcServerId: "", destServerId: "", slots: [], slotsStr: ""})
const openMigrateDialog = ref(false)
const selectedRaftMaster = ref("")
const invitedServId = ref("")
const slaves = ref<ServerInfo[]>([])
const openRaftCmdDialog = ref(false)
const store = useStore()
const formErrMsg = ref("")
const {t} = useI18n({inheritLocale: true})

watch(selectedRaftMaster, v => {
    v ? getSlaves(v) : undefined
})

function getSlotRanges() {
    let res: SlotRange[] = []
    for (let idx in slots) {
        let arr = slots[idx].slots
        if (arr.length == 0) {
            res.push({
                identify: slots[idx].id,
                start: 0,
                end: 0
            })
            continue
        }
        for (let slot of arr) {
            let sp = slot.split("-")
            res.push({
                identify: slots[idx].id,
                start: parseInt(sp[0]),
                end: parseInt(sp[1])
            })
        }
    }
    return res.sort((a, b) => {
        return a.start - b.start
    })
}

const masters = computed(() => {
    let masters: ServerInfo[] = []
    for (let i in infos.value) {
        if (infos.value[i].isMaster) {
            masters.push(infos.value[i])
        }
    }
    return masters
})

function getSlaves(masterId: string) {
    api.metadata.getPeers(masterId).then(v => {
        slaves.value = v
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
}

function leaveCluster(servId: string) {
    api.metadata.leaveCluster(servId).then(() => {
        useToast().success(t('req-success'))
    }).catch((e: Error) => {
        useToast().error(e.message)
    })
}

function joinCluster(masterId: string, servId: string) {
    api.metadata.joinLeader(masterId, servId).then(() => {
        useToast().success(t('req-success'))
    }).catch((e: Error) => {
        useToast().error(e.message)
    })
}

function closeMigrateDialog() {
    openMigrateDialog.value = false
    formErrMsg.value = ""
    migrateReq.value = {srcServerId: "", destServerId: "", slots: [], slotsStr: ""}
}


function closeRaftCmdDialog() {
    openRaftCmdDialog.value = false
    selectedRaftMaster.value = ""
    slaves.value = []
}

function getSlotsString(id: string): string {
    for (let k in slots) {
        if (slots[k].serverId === id) {
            return slots[k].slots.join(",")
        }
    }
    return "unknown slots"
}

function updateInfo(state: any) {
    let infoList: ServerInfo[] = []
    let cap = {used: 0, total: 0, free: 0}
    let stats = state.serverStat.metaServer
    for (let k in stats) {
        let v = stats[k]
        infoList.push(v)
        cap.used += v.sysInfo.diskInfo.used
        cap.total += v.sysInfo.diskInfo.total
        cap.free += v.sysInfo.diskInfo.free
    }
    infos.value = infoList
    capInfo.value = cap
}

function getSlotsDetail() {
    api.metadata.slotsDetail().then(res => {
        slots = res
        slotRanges.value = getSlotRanges()
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
}

function startMigrate() {
    if (migrateReq.value.srcServerId == migrateReq.value.destServerId) {
        formErrMsg.value = t('do-not-choose-same-id')
        return
    }
    if (!/^(\d+-\d+,?)+$/.test(migrateReq.value.slotsStr)) {
        formErrMsg.value = t('err-format-of-slots')
        return;
    }
    migrateReq.value.slots = migrateReq.value.slotsStr.split(",")
    api.metadata.startMigrate(migrateReq.value)
        .then(() => {
            useToast().success(t('req-success'))
            closeMigrateDialog()
        })
        .catch((err: Error) => {
            formErrMsg.value = err.message
        })
}

store.$subscribe((mutation, state) => {
    updateInfo(state)
})

onBeforeMount(() => {
    updateInfo(store)
    getSlotsDetail()
})
</script>

<style scoped>
.text-input {
    @apply block appearance-none rounded-md border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm
}

.text-input-sm {
    @apply block appearance-none rounded-md border border-gray-300 px-2 py-1 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 text-sm
}
</style>

<route lang="json">
{
  "meta": {
    "title": "meta-server"
  }
}
</route>

<i18n lang="yaml">
en:
  migration: 'Data Migration'
  src-server: 'Migrate from'
  dest-server: 'Migrate to'
  which-slots: 'Migrate slots'
  start-migrate: "Migration"
  tool-box: 'Tool Box'
  raft-cmd: 'Cluster Cmd'
  select-master: 'Select Leader'
  remove: 'Leave'
  invite: 'Invite'
  input-serv-id: 'Please input server id'
  do-not-choose-same-id: 'Do not choose same server id'
  err-format-of-slots: 'Err format of slots'
zh:
  migration: '数据迁移'
  src-server: '源服务器'
  dest-server: '目标服务器'
  which-slots: '迁移序列'
  start-migrate: '数据迁移'
  tool-box: '工具箱'
  raft-cmd: '集群管理'
  select-master: '选择主节点'
  remove: '脱离'
  invite: '加入'
  input-serv-id: '请输入 server id'
  do-not-choose-same-id: '请勿选择相同的 server id'
  err-format-of-slots: 'slots 格式有误'
</i18n>