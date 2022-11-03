<template>
  <div class="bg-gray-100 w-full flex items-end">
    <div v-for="rt in tabs"
         :class="activeTab === rt.name ? 'bg-indigo-600 text-white' : ''"
         @click="activeTab = rt.name"
         class="w-32 cursor-pointer transition-colors text-base border border-b-0 border-gray-200 rounded-t-2xl bg-white mt-4 py-2 px-2 text-center">
      {{ $t(rt.meta.title) }}
    </div>
  </div>
  <div class="w-full">
    <RouterView></RouterView>
  </div>
</template>

<script setup lang="ts">
import {routes} from "vue-router/auto/routes";

const activeTab = ref("/server/api")

const tabs = []

for (let i in routes) {
  if (routes[i].name == '/server') {
    routes[i].children!.forEach(v => {
      v.name != "/server/" ? tabs.push(v): ''
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