<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div v-if="infos.length > 0"
    class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
  <div v-else class="w-full my-5 text-center text-gray-600 text-lg font-medium">
    {{$t('no-servers')}}
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <UsageLine class="mb-4" :type="$cst.statTypeCpu" :server-no="$cst.metaServerNo" />
  <UsageLine :type="$cst.statTypeMem" :server-no="$cst.metaServerNo" />
</template>

<script setup lang="ts">
const infos = ref<ServerInfo[]>([])
const store = useStore()

function updateInfo(state: any) {
  if (infos.value.length > 0) {
    return
  }
  let stats = state.serverStat.apiServer
  for (let k in stats) {
    infos.value.push(stats[k])
  }
}

store.$subscribe((mutation, state)=>{
  updateInfo(state)
})

onBeforeMount(()=>{
  updateInfo(store)
})
</script>

<style scoped>

</style>

<route lang="json">
{
  "meta": {
    "title": "api-server"
  }
}
</route>