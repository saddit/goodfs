<template>
  <div class="flex items-baseline space-x-3">
    <div class="btn-pri">
      {{ t('link') }}
    </div>
    <div class="flex items-baseline">
      <label class="text-gray-900 text-sm mx-2">{{t('locale-label')}}</label>
      <PriListBox v-model="locale" :options="availableLocales" class="w-72">
      </PriListBox>
    </div>
  </div>
</template>

<script setup lang="ts">
import {useStore} from "@/store";
import {useI18n} from "vue-i18n";

const store = useStore()
useHead({
  title: "Portal"
})

const {t, locale, availableLocales} = useI18n({
  inheritLocale: true
})

watch(locale, (lang) => {
  setLocale(lang.toString())
})

function setLocale(lang: string) {
  store.setLocale(lang)
}
</script>

<route lang="json">
{
"meta": {
"title": "Portal"
}
}
</route>

<i18n lang="yaml">
en:
  link: 'linked'
  locale-label: 'locale'
zh:
  link: '链接'
  locale-label: '语言'
</i18n>