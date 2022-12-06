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
    <CapCard class="w-[32%]" :cap-info="capInfo"/>
    <div class="bg-white shadow-md rounded-md p-3 w-1/4 gap-y-4 gap-x-2 grid grid-rows-3 grid-cols-3">
      <span class="font-bold text-xl justify-self-start text-gray-500">{{ t('tool-box') }}</span>
      <button class="btn-pri" @click="openMigrateDialog = true">{{t('start-migrate')}}</button>
    </div>
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <UsageLine class="mb-4" :type="$cst.statTypeCpu" :server-no="$cst.metaServerNo"/>
  <UsageLine :type="$cst.statTypeMem" :server-no="$cst.metaServerNo"/>
  <!-- Data migration dialog -->
  <ModalTemplate v-model="openMigrateDialog" :title="t('migration')">
    <template #panel>
      <div class="mt-6 grid grid-cols-3 items-center gap-y-4 gap-x-1 rounded-md">
        <!-- row 1 source -->
        <span class="text-sm text-gray-700">{{ t('src-server') }}</span>
        <SelectBox class="col-span-2" v-model="migrateReq.srcServerId"
                   :value="(v: ServerInfo) => v.serverId"
                   :format="(v: ServerInfo) => `${v.serverId}(${getSlotsString(v.serverId)})`"
                   :options="migrateOptions"></SelectBox>
        <!-- row 2 target -->
        <span class="text-sm text-gray-700">{{ t('dest-server') }}</span>
        <SelectBox class="col-span-2" v-model="migrateReq.destServerId"
                   :value="(v: ServerInfo) => v.serverId"
                   :format="(v: ServerInfo) => `${v.serverId}(${getSlotsString(v.serverId)})`"
                   :options="migrateOptions"></SelectBox>
        <!-- row 3 slots -->
        <span class="text-sm text-gray-700">{{ t('which-slots') }}</span>
        <input id="slots" name="slots" required type="text" class="text-input col-span-2"/>
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
</template>

<script setup lang="ts">
let slots: { [key: string]: SlotsInfo } = {}
const infos = ref<ServerInfo[]>([])
const capInfo = ref<DiskInfo>({used: 0, total: 0, free: 0})
const migrateReq = ref<MetaMigrateReq>({srcServerId: "", destServerId: "", slots: []})
const openMigrateDialog = ref(false)
const store = useStore()
const formErrMsg = ref("")
const {t} = useI18n({inheritLocale: true})

const migrateOptions = computed(() => {
  let masters: ServerInfo[] = []
  for (let i in infos.value) {
    if(infos.value[i].isMaster) {
      masters.push(infos.value[i])
    }
  }
  return masters
})

function closeMigrateDialog() {
  openMigrateDialog.value = false
  formErrMsg.value = ""
  migrateReq.value = {srcServerId: "", destServerId: "", slots: []}
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
  if (infos.value.length > 0) {
    return
  }
  let stats = state.serverStat.metaServer
  for (let k in stats) {
    let v = stats[k]
    infos.value.push(v)
    capInfo.value.used += v.sysInfo.diskInfo.used
    capInfo.value.total += v.sysInfo.diskInfo.total
    capInfo.value.free += v.sysInfo.diskInfo.free
  }
}

function getSlotsDetail() {
  api.metadata.slotsDetail().then(res => {
    slots = res
  }).catch((err: Error) => {
    useToast().error(err.message)
  })
}

function startMigrate() {
  api.metadata.startMigrate(migrateReq.value)
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
zh:
  migration: '数据迁移'
  src-server: '源服务器'
  dest-server: '目标服务器'
  which-slots: '迁移序列'
  start-migrate: '数据迁移'
  tool-box: '工具箱'
</i18n>