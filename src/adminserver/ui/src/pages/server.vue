<template>
  <TabGroup class="bg-white w-full flex items-end border-b border-gray-200"
            :selected-index="selectedTab"
            @change="onSelectTab">
    <TabList>
      <Tab v-for="rt in tabs" v-slot="{ selected }" as="template" class="outline-none">
        <div v-if="rt.meta"
             class="w-32 cursor-pointer transition-all pb-2 px-2 text-center border-b-2"
             :class="{ 'border-indigo-600 text-indigo-600': selected }">
          {{ $t(rt.meta.title) }}
        </div>
      </Tab>
    </TabList>
  </TabGroup>
  <div class="group p-0 btn-pri w-10 h-10 fixed right-8 top-36 rounded-full" @click="refreshBtn">
    <ArrowPathIcon class="text-white w-5 h-5 transform duration-300 transition-transform group-hover:rotate-180"/>
  </div>
  <div class="w-full h-full px-8 py-6 bg-gray-100 overflow-y-auto">
    <RouterView></RouterView>
  </div>
</template>

<script setup lang="ts">
import {routes} from "vue-router/auto/routes";
import type {RouteRecordRaw} from "vue-router/auto";
import {ArrowPathIcon} from "@heroicons/vue/24/outline"
import {useI18n} from "vue-i18n";
import {useRouter} from "vue-router";

onBeforeMount(() => {
    selectedTab.value = useStore().selectedServerTab
    api.serverStat.stat().then((res) => {
        useStore().setServerInfo(res)
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
    onSelectTab(useStore().selectedServerTab)
})

const {t} = useI18n()
const router = useRouter()

function onSelectTab(index: number) {
    selectedTab.value = index
    useStore().setSelectedServerTab(index)
    router.push(tabs[index].path)
}

function refreshBtn() {
    api.serverStat.stat().then((res) => {
        useStore().setServerInfo(res)
        useToast().success(t('refresh-success'))
    }).catch((err: Error) => {
        useToast().error(err.message)
    })
}

const selectedTab = ref(0)
const tabs: RouteRecordRaw[] = []

for (let i in routes) {
    if (routes[i].name == '/server') {
        routes[i].children!.forEach(v => {
            v.name != "/server/" ? tabs.push(v) : ''
        })
    }
}
</script>

<style scoped>

</style>

<route lang="json">
{
  "meta": {
    "title": "server",
    "icon": "server"
  }
}
</route>

<i18n lang="yaml">
en:
  refresh-success: 'Refresh Success!'
zh:
  refresh-success: '刷新成功'
</i18n>