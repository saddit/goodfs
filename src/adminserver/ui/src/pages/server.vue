<template>
  <TabGroup class="bg-white w-full flex items-end border-b border-gray-200">
    <TabList>
      <Tab v-for="rt in tabs" v-slot="{ selected }" as="template" class="outline-none">
        <div v-if="rt.meta"
             class="w-32 cursor-pointer transition-all pb-2 px-2 text-center"
             :class="{ 'border-indigo-600 border-b-2 text-indigo-600': selected }"
             @click="$router.push(rt.path)">
          {{ $t(rt.meta.title) }}
        </div>
      </Tab>
    </TabList>
  </TabGroup>
  <div class="w-full h-full px-8 py-6 bg-gray-100 overflow-y-auto">
    <RouterView></RouterView>
  </div>
</template>

<script setup lang="ts">
import {routes} from "vue-router/auto/routes";
import type {RouteRecordRaw} from "vue-router/auto";

api.serverStat.stat().then((res)=>{
  useStore().setServerInfo(res)
}).catch((err: Error)=>{
  useToast().error(err.message)
})

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