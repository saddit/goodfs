<script setup lang="ts">
import {ChevronLeftIcon, UserCircleIcon} from "@heroicons/vue/20/solid";
import {ArrowLeftCircleIcon, ArrowLeftOnRectangleIcon} from '@heroicons/vue/24/outline'
import {useI18n} from "vue-i18n";
import {routes} from "vue-router/auto/routes";
import {useStore} from "@/store";

const store = useStore()

const needLogin = ref(store.basicAuth == "")

const {t} = useI18n({inheritLocale: true})

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
    <LoginDialog v-model="needLogin"/>
    <!-- left tab-bar -->
    <div class="h-full pri-bg-gradient text-gray-300 transition-[width]" :class="[store.tabClosed ? 'w-20' : 'w-64']">
      <!-- head -->
      <div
          class="pl-4 pr-8 py-3 min-h-18 sm:pl-6 sm:pr-9 lg:pr-12 lg:pl-8 text-white flex flex-col relative border-indigo-500 border-b">
        <span class="text-sm h-4 font-bold px-0.5" v-show="!store.tabClosed">GooDFS</span>
        <span class="text-2xl font-light" v-show="!store.tabClosed">CONSOLE</span>
        <ChevronLeftIcon
            :class="[store.tabClosed ? 'rotate-180 mx-auto w-10 h-10' : 'absolute bottom-4 right-2 w-6 h-6']"
            class="cursor-pointer transition-transform transform" @click="store.closeTab"/>
      </div>
      <!-- routes -->
      <div class="flex flex-col py-3 pl-2 overflow-y-auto no-scrollbar">
        <template v-for="rt in routes" :key="rt.name">
          <div @click="$router.push(rt.path)" v-if="rt.meta && !rt.meta.hideTab"
               :class="[$route.name === rt.name ? 'text-white' : 'text-gray-300']"
               class="flex items-center ml-4 my-2 pl-2 rounded-md py-4 transition-colors cursor-pointer select-none hover:bg-indigo-800 hover:text-white">
            <font-awesome-icon :icon="rt.meta.icon" class="ml-2 mr-6 text-xl"/>
            <span class="whitespace-nowrap overflow-hidden transition-opacity duration-300"
                  :class="[store.tabClosed ? 'opacity-0' : 'opacity-100']">
              {{ t(rt.meta.title) }}
            </span>
          </div>
        </template>
      </div>
    </div>
    <!-- right content -->
    <div class="flex flex-col w-full">
      <div class="inline-flex items-center py-6 px-4">
        <ArrowLeftCircleIcon @click="$router.back()"
                             class="w-7 h-7 mr-3 text-indigo-500 transition-transform transform active:scale-75 cursor-pointer"/>
        <div class="text-xl flex-grow font-bold text-gray-900">
          {{ title($route.meta.title) }}
        </div>
        <div class="inline-flex items-center text-gray-500">
          <UserCircleIcon class="w-9 h-9"/>
          <!-- popover menu  -->
          <Popover class="relative mx-2">
            <PopoverButton>
              <span class="pop-btn">{{ store.user }}</span>
            </PopoverButton>
            <PopTransition>
              <PopoverPanel class="pop-panel">
                <div class="grid gap-2 bg-white p-3">
                  <div class="pop-panel-item">
                    <ArrowLeftOnRectangleIcon class="w-5 h-5 text-indigo-600 mr-2"/>
                    <span>{{ t('login-out') }}</span>
                  </div>
                </div>
              </PopoverPanel>
            </PopTransition>
          </Popover>
        </div>
      </div>
      <RouterView/>
    </div>
  </main>
</template>

<style scoped>
.pri-bg-gradient {
  @apply bg-gradient-to-r from-indigo-600 to-indigo-800
}

.pop-btn {
  @apply underline mx-1 outline-none hover:text-indigo-500 select-none transition-colors focus:outline-none
}

.pop-panel {
  @apply absolute overflow-hidden rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 left-1/2 z-10 mt-3 w-screen max-w-[10rem] -translate-x-full transform px-1 sm:px-0
}

.pop-panel-item {
  @apply inline-flex select-none justify-center items-center rounded-lg p-2 transition duration-150 ease-in-out hover:text-indigo-600 hover:bg-indigo-50
}
</style>

<i18n lang="yaml">
en:
  login-out: 'Sign out'
zh:
  login-out: '退出登陆'
</i18n>
