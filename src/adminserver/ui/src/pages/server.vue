<template>
  <TabGroup class="bg-gray-100 w-full flex items-end border-b border-gray-200">
    <TabList>
      <Tab v-for="rt in tabs" v-slot="{ selected }" as="template" class="outline-none">
        <div
          class="w-32 cursor-pointer transition-colors text-base border border-b-0 border-gray-200 rounded-t-2xl bg-white mt-4 py-2 px-2 text-center"
          :class="{
            'bg-indigo-600 text-white': selected
          }" @click="$router.push(rt.path)">
          {{ $t(rt.meta!.title) }}
        </div>
      </Tab>
    </TabList>
  </TabGroup>
  <div class="w-full h-full p-1 bg-gray-100 overflow-y-auto">
    <RouterView></RouterView>
  </div>
</template>

<script setup lang="ts">
import { routes } from "vue-router/auto/routes";
import type { RouteRecordRaw, RouteRecordName } from "vue-router/auto";

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