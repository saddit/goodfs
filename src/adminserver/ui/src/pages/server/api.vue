<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ t('overview') }}</div>
  <div class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 justify-items-start placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
const infos = ref<ServerInfo[]>([])

const { t } = useI18n({
  inheritLocale: true
})

onBeforeMount(() => {
  for (let i = 0; i < 3; i++) {
    infos.value.push({
      serverId: 'Node-A',
      httpAddr: '192.168.0.1:8080',
      rpcAddr: '192.168.0.1:4040',
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

<i18n lang="yaml">
en:
  overview: 'Overview'
zh:
  overview: '总览'
</i18n>