<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div v-if="infos.length > 0"
       class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard class="cursor-pointer" v-for="info in infos" :info="info"
                @click="openDialog(info.serverId)"></ServerCard>
  </div>
  <div v-else class="w-full my-5 text-center text-gray-600 text-lg font-medium">
    {{ $t('no-servers') }}
  </div>
  <div class="mb-4 mt-8">
    <!-- capacity card -->
    <CapCard v-if="capInfo.total > 0" class="w-[32%]" :cap-info="capInfo"/>
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <UsageLine class="mb-4" :type="$cst.statTypeCpu" :server-no="$cst.dataServerNo"/>
  <UsageLine :type="$cst.statTypeMem" :server-no="$cst.dataServerNo"/>
  <!-- Migration dialog -->
  <ModalTemplate title="Join or Leave" v-model="openMigrateDialog">
    <template #panel>
      <div class="py-6 px-8 grid-cols-1 grid gap-y-2">
        <button class="btn-pri" @click="clusterCmd('join')">Join cluster</button>
        <button class="btn-pri" @click="clusterCmd('leave')">Leave cluster</button>
        <button class="btn-revert" @click="openMigrateDialog = false">Close</button>
      </div>
    </template>
  </ModalTemplate>
</template>

<script setup lang="ts">
const infos = ref<ServerInfo[]>([])
const capInfo = ref<DiskInfo>({used: 0, total: 0, free: 0})
const store = useStore()
const openMigrateDialog = ref(false)
let migrateServId = ""
const {t} = useI18n({inheritLocale: true})

function updateInfo(state: any) {
  if (infos.value.length > 0) {
    return
  }
  let stats = state.serverStat.dataServer
  for (let k in stats) {
    let v = stats[k]
    infos.value.push(v)
    capInfo.value.used += v.sysInfo.diskInfo.used
    capInfo.value.total += v.sysInfo.diskInfo.total
    capInfo.value.free += v.sysInfo.diskInfo.free
  }
}

function openDialog(servId: string) {
  migrateServId = servId
  openMigrateDialog.value = true
}

async function clusterCmd(cmd: string) {
  try {
    if (cmd == 'join') {
      await api.objects.join(migrateServId)
    } else if (cmd == 'leave') {
      await api.objects.leave(migrateServId)
    }
    useToast().success(t('req-success'))
  } catch (err: any) {
    useToast().error(err.message)
  }
}

store.$subscribe((mutation, state) => {
  updateInfo(state)
})

onBeforeMount(() => {
  updateInfo(store)
})
</script>

<style scoped>

</style>

<route lang="json">
{
  "meta": {
    "title": "data-server"
  }
}
</route>