<template>
  <div class="w-1/4 flex inline-flex items-center px-8">
    <div class="mr-3 whitespace-nowrap text-gray-900">{{ t('server-id') }}</div>
    <SelectBox class="flex-grow" v-model="selectedId" :options="serverIds"></SelectBox>
  </div>
  <textarea disabled rows="30" class="text-sm m-8 p-3 resize-none border-0 bg-gray-800 text-gray-300 min-h-32 rounded"
            :value="configContent"/>
</template>

<script setup lang="ts">
const serverIds = ref<string[]>([])
const store = useStore()
const selectedId = ref("")
const configContent = ref("")

const {t} = useI18n({inheritLocale: true})

watch(selectedId, () => {
    loadConfigContent(selectedId.value)
})

onBeforeMount(() => {
    api.serverStat.stat().then((res) => {
        let ids = []
        for (let k in res.apiServer) {
            ids.push(k)
        }
        for (let k in res.metaServer) {
            ids.push(k)
        }
        for (let k in res.dataServer) {
            ids.push(k)
        }
        serverIds.value = ids
        selectedId.value = ids[0]
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
})

function loadConfigContent(id: string) {
    api.serverStat.config(id).then(r => {
        configContent.value = r
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
}

</script>

<style scoped>

</style>

<i18n lang="yaml">
en:
  server-id: 'Server ID'
zh:
  server-id: '服务ID'
</i18n>

<route lang="json">
{
  "meta": {
    "title": "config",
    "icon": "folder-open"
  }
}
</route>