<template>
  <a class="btn-pri">
    {{ t('link') }}
  </a>
  <select v-model="$i18n.locale" @select="setLocale">
    <option v-for="lang in availableLocales" :key="`locale-${lang}`" :value="lang">{{ lang }}</option>
  </select>
</template>

<script setup lang="ts">
import {useStore} from "@/store";
import {useI18n} from "vue-i18n";

const store = useStore()
useHead({
  title: "Portal"
})

const {t, availableLocales} = useI18n({
  inheritLocale: true
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
zh:
  link: '链接'
</i18n>