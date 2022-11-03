<template>
  <div class="flex items-baseline space-x-3">
    <div class="btn-pri">
      {{ t('link') }}
    </div>
    <div class="flex items-baseline">
      <label class="text-gray-900 text-sm mx-2">{{  t('locale-label')  }}</label>
      <PriListBox @change="setLocale" v-model="$i18n.locale" :options="$i18n.availableLocales" :format="t" class="w-72">
      </PriListBox>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useStore } from "@/store";
import { useI18n } from "vue-i18n";

const {t} = useI18n({
  inheritLocale: true
})

const store = useStore()

function setLocale(lang: string) {
  store.setLocale(lang)
}
</script>

<route lang="json">
{
  "meta": {
    "title": "setting",
    "icon": "tools"
  }
}
</route>

<i18n lang="yaml">
en:
  link: 'linked'
  locale-label: 'locale'
  en: 'English(US)'
  zh: 'Chinese-Simple'
zh:
  link: '链接'
  locale-label: '语言'
  en: '英语(美国)'
  zh: '中文简体'
</i18n>