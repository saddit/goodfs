<template>
  <div class="text-2xl text-gray-900 font-bold mb-4">{{ $t('overview') }}</div>
  <div class="grid gap-y-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-8 justify-items-center placeholder:py-2">
    <ServerCard v-for="info in infos" :info="info"></ServerCard>
  </div>
</template>

<script setup lang="ts">
const infos = ref<ServerInfo[]>([])

onBeforeMount(() => {
})
</script>

<style scoped>

</style>

<route lang="json">
{
  "meta": {
    "title": "meta-server"
  }
}
</route>