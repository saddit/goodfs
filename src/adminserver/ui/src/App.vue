<script setup lang="ts">
import {ChevronLeftIcon, ClipboardIcon} from "@heroicons/vue/20/solid";

const tabClosed = ref(false)

function closeTabs() {
  tabClosed.value = !tabClosed.value
}

</script>

<template>
  <main class="flex h-full">
    <!-- left tab-bar -->
    <div class="h-full pri-bg-gradient text-gray-300 transition-[width]" :class="[tabClosed ? 'w-20' : 'w-72']">
      <!-- head -->
      <div class="pl-4 pr-8 py-3 min-h-18 sm:pl-6 sm:pr-9 lg:pr-12 lg:pl-8 text-white flex flex-col relative border-indigo-500 border-b">
        <div class="text-sm h-4 font-bold px-0.5" v-show="!tabClosed">GooDFS</div>
        <div class="text-2xl font-light" v-show="!tabClosed">CONSOLE</div>
        <ChevronLeftIcon :class="[tabClosed ? 'rotate-180 mx-auto w-10 h-10' : 'absolute bottom-4 right-1 w-6 h-6']"
                         class="cursor-pointer transition-transform transform"
                         @click="closeTabs"/>
      </div>
      <!-- routes -->
      <div class="flex flex-col py-3 pl-2">
        <div v-for="idx in [1,2,3,4,5]" :key="idx"
             class="inline-flex items-center ml-4 my-2 pl-2 rounded-md py-4 hover:bg-indigo-800">
          <ClipboardIcon class="w-6 h-6"/>
          <span class="pl-4 pr-5" v-show="!tabClosed">Account Setting {{ idx }}</span>
        </div>
      </div>
    </div>
    <!-- right content -->
    <div class="flex flex-col w-full">
      <div class="text-xl px-8 py-7 font-bold text-gray-900">
        {{ $t($route.meta.title) }}
      </div>
      <router-view/>
    </div>
  </main>
</template>

<style scoped>
.pri-bg-gradient {
  @apply bg-gradient-to-r from-indigo-600 to-indigo-800
}
</style>
