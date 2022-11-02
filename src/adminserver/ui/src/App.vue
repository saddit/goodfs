<script setup lang="ts">
import { ChevronLeftIcon, ClipboardIcon } from "@heroicons/vue/20/solid";
import { ArrowLeftCircleIcon } from '@heroicons/vue/24/outline'
import { useI18n } from "vue-i18n";
import { routes } from "vue-router/auto/routes";

const tabClosed = ref(false)

const { t } = useI18n({
  inheritLocale: true
})

function closeTabs() {
  tabClosed.value = !tabClosed.value
}

function title(metaTitle: string): string {
  if (!metaTitle) {
    return ''
  }
  let res = t(metaTitle)
  useHead({
    title: res
  })
  return res
}
</script>

<template>
  <main class="flex h-full">
    <!-- left tab-bar -->
    <div class="h-full pri-bg-gradient text-gray-300 transition-[width]" :class="[tabClosed ? 'w-20' : 'w-72']">
      <!-- head -->
      <div
        class="pl-4 pr-8 py-3 min-h-18 sm:pl-6 sm:pr-9 lg:pr-12 lg:pl-8 text-white flex flex-col relative border-indigo-500 border-b">
        <span class="text-sm h-4 font-bold px-0.5" v-show="!tabClosed">GooDFS</span>
        <span class="text-2xl font-light" v-show="!tabClosed">CONSOLE</span>
        <ChevronLeftIcon :class="[tabClosed ? 'rotate-180 mx-auto w-10 h-10' : 'absolute bottom-4 right-2 w-6 h-6']"
          class="cursor-pointer transition-transform transform" @click="closeTabs" />
      </div>
      <!-- routes -->
      <div class="flex flex-col py-3 pl-2">
        <template v-for="rt in routes" :key="rt.name">
          <div @click="$router.push(rt.path)" v-if="rt.meta && !rt.meta.hideTab"
            :class="[$route.name == rt.name ? 'text-white': 'text-gray-300']"
            class="flex items-center ml-4 my-2 pl-2 rounded-md py-4 transition-colors cursor-pointer select-none hover:bg-indigo-800 hover:text-white">
            <font-awesome-icon :icon="rt.meta.icon" class="ml-2 mr-6 text-xl"/>
            <span class="whitespace-nowrap overflow-hidden transition-opacity duration-300"
              :class="[tabClosed ? 'opacity-0' : 'opacity-100']">
              {{ t(rt.meta.title) }}
            </span>
          </div>
        </template>
      </div>
    </div>
    <!-- right content -->
    <div class="flex flex-col w-full">
      <div class="inline-flex items-center">
        <ArrowLeftCircleIcon @click="$router.back()"
          class="w-7 h-7 mr-3 ml-4 text-indigo-500 transition-transform transform active:scale-75 cursor-pointer" />
        <div class="text-xl py-7 font-bold text-gray-900">
          {{ title($route.meta.title) }}
        </div>
      </div>
      <RouterView />
    </div>
  </main>
</template>

<style scoped>
.pri-bg-gradient {
  @apply bg-gradient-to-r from-indigo-600 to-indigo-800
}
</style>
