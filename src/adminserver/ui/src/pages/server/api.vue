<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div
      class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
  <div class="mt-8 text-2xl text-gray-900 font-bold mb-4">{{ $t('monitor') }}</div>
  <div class="bg-white rounded-2xl border border-gray-200 outline outline-3 outline-offset-4 outline-indigo-600 w-full h-32">
  </div>
  <div class="inline-flex justify-center text-sm text-indigo-600 font-medium pt-2.5 w-full">{{ $t('cpu-usage') }}</div>
  <div class="mt-4 bg-white rounded-2xl border border-gray-200 outline outline-3 outline-offset-4 outline-indigo-600 w-full h-32">
  </div>
  <div class="inline-flex justify-center text-sm text-indigo-600 font-medium pt-2.5 w-full">{{ $t('mem-usage') }}</div>
</template>

<script setup lang="ts">
const infos = ref<ServerInfo[]>([])

onBeforeMount(() => {
  for (let i = 0; i < 3; i++) {
    infos.value.push({
      serverId: 'Node-A',
      httpAddr: '192.168.0.1:8080',
      rpcAddr: '192.168.0.1:4040',
      sysInfo: {
        cpuStatus: {
          usedPercent: 0.9,
          logicalCount: 16,
          physicalCount: 8
        },
        memStatus: {
          used: 0,
          free: 0,
          self: 0,
          all: 0
        },
        diskInfo: {
          used: 0,
          total: 0,
          free: 0
        }
      }
    })
  }
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